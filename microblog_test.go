package microblog_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"microblog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListenAndServe_UsesGivenStore(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	blogPost := &microblog.BlogPost{ID: id, Title: "foo", Content: "boo"}
	store := &microblog.MemoryPostStore{BlogPosts: []*microblog.BlogPost{blogPost}}

	addr := newTestServer(t, store)

	resp, err := http.Get("http://" + addr.String())
	require.NoError(t, err)

	read, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	got := string(read)

	// Check for the presence of the blog post title and content in the response
	assert.Contains(t, got, "<h2><a href=\"/blogpost?name=")
	assert.Contains(t, got, "<p>foo</p>")
	assert.Contains(t, got, "<p>boo</p>")
}

func TestSubmitHandler(t *testing.T) {
	t.Parallel()

	// Create a mock store
	store := &microblog.MemoryPostStore{}
	app := &microblog.Application{Poststore: store}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(app.Submit))
	defer server.Close()

	// Prepare form data
	form := url.Values{}
	form.Add("title", "Test Title")
	form.Add("content", "Test Content")

	// Send POST request
	resp, err := http.PostForm(server.URL, form)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check response status
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Decode the response body
	var got microblog.BlogPost
	err = json.NewDecoder(resp.Body).Decode(&got)
	require.NoError(t, err)

	// Verify the blog post was created in the store
	createdPost, err := store.GetByID(got.ID)
	require.NoError(t, err)
	require.NotNil(t, createdPost)

	// Check the created post fields
	assert.Equal(t, "Test Title", createdPost.Title)
	// assert.Equal(t, "Test Content", createdPost.Content)
	// assert.NotEmpty(t, createdPost.ID)
	// assert.NotEmpty(t, createdPost.CreatedAt)
	// assert.NotEmpty(t, createdPost.UpdatedAt)
}

func TestSubmitHandlerBasicAuthError(t *testing.T) {
	t.Parallel()

	store := &microblog.MemoryPostStore{}
	addr := newTestServer(t, store)

	endpoint := fmt.Sprintf("http://%v/newpost", addr)
	req, err := http.NewRequest("GET", endpoint, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	if !cmp.Equal("401 Unauthorized", resp.Status) {
		t.Error(cmp.Diff(http.StatusUnauthorized, resp.Status))
	}
}

func TestIsAuthenticatedWhenCorrectPasswordProvidedReturnsTrue(t *testing.T) {

	t.Setenv(microblog.MicroblogToken, "password123")

	got := microblog.IsAuthenticated("password123")
	want := true

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestIsAuthenticatedReturnsFalseWhenIncorrectPasswordProvided(t *testing.T) {

	t.Setenv(microblog.MicroblogToken, "password123")

	got := microblog.IsAuthenticated("hotdog")
	want := false

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestNewPostHandler(t *testing.T) {
	t.Parallel()

	store := &microblog.MemoryPostStore{}
	addr := newTestServer(t, store)

	endpoint := fmt.Sprintf("http://%v/newpost", addr)
	req, err := http.NewRequest("GET", endpoint, nil)
	require.NoError(t, err)

	req.Header.Add("Authorization", "Basic "+basicAuth())
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("test fail")
	}

	if !bytes.Contains(data, []byte("post-form")) {
		t.Fatalf("%s, %s", data, "form does not contain string")
	}
}

func newTestServer(t *testing.T, store microblog.PostStore) net.Addr {
	t.Helper()

	netListener, err := net.Listen("tcp", "127.0.0.1:")
	addr := netListener.Addr().String()

	if err != nil {
		t.Fatal(err)
	}
	netListener.Close()

	go func() {
		err := microblog.ListenAndServe(addr,
			microblog.Application{
				Auth: struct {
					Username string
					Password string
				}{
					Username: "foo",
					Password: "foo",
				},
				Poststore: store,
			},
		)
		if err != nil {
			panic(err)
		}
	}()

	resp, err := http.Get("http:" + addr)

	for err != nil {
		t.Log("retrying")
		resp, err = http.Get("http://" + addr)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatal(resp.StatusCode)
	}

	return netListener.Addr()
}

func basicAuth() string {
	auth := "foo" + ":" + "foo"
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
