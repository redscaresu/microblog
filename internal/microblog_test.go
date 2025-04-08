package microblog_test

import (
	"encoding/json"
	"fmt"
	"io"
	microblog "microblog/internal"
	"microblog/internal/models"
	"microblog/internal/repository"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListenAndServe_UsesGivenStore(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	blogPost := &models.BlogPost{ID: id, Title: "foo", Content: "boo", FormattedDate: "1 June, 2025"}
	store := &repository.MemoryPostStore{BlogPosts: []*models.BlogPost{blogPost}}

	addr := newTestServer(t, store)

	resp, err := http.Get("http://" + addr.String())
	require.NoError(t, err)

	read, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	got := string(read)
	assert.Contains(t, got, "<h2><a href=\"/blogpost?name=")
	assert.Contains(t, got, "<p>foo</p>")
	assert.Contains(t, got, "<p>boo</p>")
	assert.Contains(t, got, "<h3 style=\"color: grey; font-size: 0.9em;\">1 June, 2025</h3>")
}

func TestSubmitHandler(t *testing.T) {
	t.Parallel()

	store := &repository.MemoryPostStore{}
	app := &microblog.Application{PostStore: store}

	server := httptest.NewServer(http.HandlerFunc(app.Submit))
	defer server.Close()

	form := url.Values{}
	form.Add("title", "Test Title")
	form.Add("content", "Test Content")

	resp, err := http.PostForm(server.URL, form)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var got models.BlogPost
	err = json.NewDecoder(resp.Body).Decode(&got)
	require.NoError(t, err)

	createdPost, err := store.GetByID(got.ID)
	require.NoError(t, err)
	require.NotNil(t, createdPost)

	assert.Equal(t, "Test Title", createdPost.Title)
	assert.Equal(t, "Test Content", createdPost.Content)
	assert.NotEmpty(t, createdPost.ID)
	assert.NotEmpty(t, createdPost.CreatedAt)
	assert.NotEmpty(t, createdPost.UpdatedAt)
	assert.NotEmpty(t, createdPost.FormattedDate)
}

func TestUpdatePostHandler(t *testing.T) {
	t.Parallel()

	store := &repository.MemoryPostStore{}
	app := &microblog.Application{PostStore: store}

	submitServer := httptest.NewServer(http.HandlerFunc(app.Submit))
	defer submitServer.Close()

	form := url.Values{}
	form.Add("title", "Original Title")
	form.Add("content", "Original Content")

	resp, err := http.PostForm(submitServer.URL, form)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var createdPost models.BlogPost
	err = json.NewDecoder(resp.Body).Decode(&createdPost)
	require.NoError(t, err)

	updateServer := httptest.NewServer(http.HandlerFunc(app.UpdatePostHandler))
	defer updateServer.Close()

	editForm := url.Values{}
	editForm.Add("id", createdPost.ID.String())
	editForm.Add("title", "Updated Title")
	editForm.Add("content", "Updated Content")

	editResp, err := http.PostForm(updateServer.URL, editForm)
	require.NoError(t, err)
	defer editResp.Body.Close()

	require.Equal(t, http.StatusOK, editResp.StatusCode)

	updatedPost, err := store.GetByID(createdPost.ID)
	require.NoError(t, err)
	require.NotNil(t, updatedPost)

	assert.Equal(t, "Updated Title", updatedPost.Title)
	assert.Equal(t, "Updated Content", updatedPost.Content)
	assert.Equal(t, createdPost.ID, updatedPost.ID)
}

func TestUpdateHandlerBasicAuthError(t *testing.T) {
	t.Parallel()

	store := &repository.MemoryPostStore{}
	addr := newTestServer(t, store)

	endpoint := fmt.Sprintf("http://%v/updatepost", addr)
	req, err := http.NewRequest("GET", endpoint, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestEditPostHandlerBasicAuthError(t *testing.T) {
	t.Parallel()

	store := &repository.MemoryPostStore{}
	addr := newTestServer(t, store)

	endpoint := fmt.Sprintf("http://%v/editpost?name=%s", addr, "doesnotexit")
	req, err := http.NewRequest("GET", endpoint, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestSubmitHandlerBasicAuthError(t *testing.T) {
	t.Parallel()

	store := &repository.MemoryPostStore{}
	addr := newTestServer(t, store)

	endpoint := fmt.Sprintf("http://%v/newpost", addr)
	req, err := http.NewRequest("GET", endpoint, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestIsAuthenticatedWhenCorrectPasswordProvidedReturnsTrue(t *testing.T) {

	t.Setenv(microblog.MicroblogToken, "password123")

	got := microblog.IsAuthenticated("password123")
	want := true

	assert.Equal(t, want, got)
}

func TestIsAuthenticatedReturnsFalseWhenIncorrectPasswordProvided(t *testing.T) {

	t.Setenv(microblog.MicroblogToken, "password123")

	got := microblog.IsAuthenticated("hotdog")
	want := false

	assert.Equal(t, want, got)

}

func newTestServer(t *testing.T, store repository.PostStore) net.Addr {
	t.Helper()

	netListener, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	addr := netListener.Addr().String()
	netListener.Close()

	mux := http.NewServeMux()
	go func() {
		err := microblog.RegisterRoutes(mux,
			addr,
			microblog.NewApplication("foo", "foo", store))
		require.NoError(t, err)
	}()

	resp, err := http.Get("http:" + addr)
	for err != nil {
		t.Log("retrying")
		resp, err = http.Get("http://" + addr)
		require.NoError(t, err)
	}

	if resp.StatusCode != http.StatusOK {
		require.NoError(t, err)
	}

	return netListener.Addr()
}
