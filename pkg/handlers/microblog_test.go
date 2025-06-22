package handlers_test

import (
	"encoding/json"
	"fmt"
	"io"
	"microblog/pkg/cache"
	"microblog/pkg/handlers"
	"microblog/pkg/models"
	"microblog/pkg/repository"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkdown(t *testing.T) {
	input := `> This is a test blockquote`
	output := handlers.RenderMarkdown(input)
	fmt.Println("Test Input:", input)
	fmt.Println("Test Output:", output)
	// Should output: <blockquote><p>This is a test blockquote</p></blockquote>
}

func TestMarkdownAgain(t *testing.T) {
	input := `> "What has been will be again,
> what has been done will be done again;
> there is nothing new under the sun."

Can the past provide us with a lens with which to understand the present in terms of what the impact of AI tooling will have on software development?

I read this article, which I think articulates perfectly the skepticism software engineers have towards AI coding assistants.`

	output := handlers.RenderMarkdown(input)
	fmt.Println("=== ACTUAL CONTENT TEST ===")
	fmt.Println("Input:", input)
	fmt.Println("Output:", output)
	fmt.Println("===========================")
}

func TestListenAndServe_NoCache(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	blogPost := &models.BlogPost{
		ID:            id,
		Name:          "foo",
		Title:         "foo",
		Content:       "boo",
		FormattedDate: "1 June, 2025"}
	store := &repository.MemoryPostStore{BlogPosts: []*models.BlogPost{blogPost}}
	cache := cache.New([]*models.BlogPost{}, &sync.Mutex{})

	server := newTestServer(t, store, cache)
	defer server.Close()

	resp, err := http.Get(server.URL)
	require.NoError(t, err)

	read, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	got := string(read)
	assert.Contains(t, got, "<h2><a href=\"/post/foo")
	assert.Contains(t, got, "<p>foo</p>")
	assert.Contains(t, got, "<p>boo</p>")
	assert.Contains(t, got, "<h3 style=\"color: grey; font-size: 0.9em;\">1 June, 2025</h3>")
}

func TestListenAndServe_CacheHit(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	blogPost := &models.BlogPost{
		ID:            id,
		Name:          "foo",
		Title:         "foo",
		Content:       "boo",
		FormattedDate: "1 June, 2025"}
	store := &repository.MemoryPostStore{BlogPosts: []*models.BlogPost{blogPost}}
	cache := cache.New([]*models.BlogPost{}, &sync.Mutex{})

	server := newTestServer(t, store, cache)
	defer server.Close()

	// cache is empty
	require.Len(t, cache.BlogPosts, 0)

	resp, err := http.Get(server.URL)
	require.NoError(t, err)

	// Database accessed
	require.Equal(t, 1, store.AccessCounter)

	// cache is hydrated on the first Get to the homepage
	require.Len(t, cache.BlogPosts, 1)
	require.Equal(t, []*models.BlogPost{
		{
			Title:         "<p>foo</p>\n",
			Name:          "foo",
			Content:       "<p>boo</p>\n",
			ID:            id,
			FormattedDate: blogPost.FormattedDate,
		},
	},
		cache.BlogPosts)

	read, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	got := string(read)
	assert.Contains(t, got, "<h2><a href=\"/post/foo")
	assert.Contains(t, got, "<p>foo</p>")
	assert.Contains(t, got, "<p>boo</p>")
	assert.Contains(t, got, "<h3 style=\"color: grey; font-size: 0.9em;\">1 June, 2025</h3>")

	// test cache hit
	resp, err = http.Get(server.URL)
	require.NoError(t, err)

	// Database has still only been accessed once which means the cache has been used
	require.Equal(t, 1, store.AccessCounter)

	read, err = io.ReadAll(resp.Body)
	require.NoError(t, err)

	got = string(read)
	assert.Contains(t, got, "<h2><a href=\"/post/foo")
	assert.Contains(t, got, "<p>foo</p>")
	assert.Contains(t, got, "<p>boo</p>")
	assert.Contains(t, got, "<h3 style=\"color: grey; font-size: 0.9em;\">1 June, 2025</h3>")
}

func TestSubmitHandler(t *testing.T) {
	t.Parallel()

	store := &repository.MemoryPostStore{}
	cache := cache.New([]*models.BlogPost{}, &sync.Mutex{})

	server := newTestServer(t, store, cache)
	defer server.Close()

	form := url.Values{}
	form.Add("title", "Original Title")
	form.Add("content", "Original Content")

	req, err := http.NewRequest(http.MethodPost, server.URL+"/submit", strings.NewReader(form.Encode()))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("foo", "foo")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var createdPost models.BlogPost
	err = json.NewDecoder(resp.Body).Decode(&createdPost)
	require.NoError(t, err)
	assert.Equal(t, "Original Title", createdPost.Title)
	assert.Equal(t, "Original Content", createdPost.Content)
	assert.Equal(t, createdPost.ID, createdPost.ID)
}

func TestUpdatePostHandler(t *testing.T) {
	t.Parallel()

	store := &repository.MemoryPostStore{}
	cache := cache.New([]*models.BlogPost{}, &sync.Mutex{})

	server := newTestServer(t, store, cache)
	defer server.Close()

	form := url.Values{}
	form.Add("title", "Original Title")
	form.Add("content", "Original Content")

	req, err := http.NewRequest(http.MethodPost, server.URL+"/submit", strings.NewReader(form.Encode()))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("foo", "foo")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var createdPost models.BlogPost
	err = json.NewDecoder(resp.Body).Decode(&createdPost)
	require.NoError(t, err)

	editForm := url.Values{}
	editForm.Add("id", createdPost.ID.String())
	editForm.Add("title", "Updated Title")
	editForm.Add("content", "Updated Content")

	reqEdit, err := http.NewRequest(http.MethodPost, server.URL+"/updatepost", strings.NewReader(editForm.Encode()))
	require.NoError(t, err)
	reqEdit.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqEdit.SetBasicAuth("foo", "foo")

	editResp, err := http.DefaultClient.Do(reqEdit)
	require.NoError(t, err)
	defer editResp.Body.Close()

	require.Equal(t, http.StatusOK, editResp.StatusCode)

	updatedPost, err := store.GetByID(createdPost.ID)
	require.NoError(t, err)
	require.NotNil(t, updatedPost)

	assert.Equal(t, "Updated Title", updatedPost.Title)
	assert.Equal(t, "Updated Content", updatedPost.Content)
	assert.Equal(t, createdPost.ID, updatedPost.ID)

	// Assert that the cache has been hydrated when a blogpost is updated
	assert.Equal(t, "<p>Updated Title</p>\n", cache.BlogPosts[0].Title)
	assert.Equal(t, "<p>Updated Content</p>\n", cache.BlogPosts[0].Content)
}

func TestUpdateHandlerBasicAuthError(t *testing.T) {
	t.Parallel()

	store := &repository.MemoryPostStore{}
	cache := cache.New([]*models.BlogPost{}, &sync.Mutex{})

	server := newTestServer(t, store, cache)
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL+"/submit", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestEditPostHandlerBasicAuthError(t *testing.T) {
	t.Parallel()

	store := &repository.MemoryPostStore{}
	cache := cache.New([]*models.BlogPost{}, &sync.Mutex{})

	server := newTestServer(t, store, cache)
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL+"/editpost?name=doesnotexist", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestSubmitHandlerBasicAuthError(t *testing.T) {
	t.Parallel()

	store := &repository.MemoryPostStore{}
	cache := cache.New([]*models.BlogPost{}, &sync.Mutex{})

	server := newTestServer(t, store, cache)
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL+"/newpost", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestIsAuthenticatedWhenCorrectPasswordProvidedReturnsTrue(t *testing.T) {

	t.Setenv(handlers.MicroblogToken, "password123")

	got := handlers.IsAuthenticated("password123")
	want := true

	assert.Equal(t, want, got)
}

func TestIsAuthenticatedReturnsFalseWhenIncorrectPasswordProvided(t *testing.T) {

	t.Setenv(handlers.MicroblogToken, "password123")

	got := handlers.IsAuthenticated("hotdog")
	want := false

	assert.Equal(t, want, got)
}

func TestGetBlogPostByName_ServesFromCacheOnSubsequentRequests(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	blogPost := &models.BlogPost{ID: id, Name: "testtitle", Title: "Test Title", Content: "Test Content", FormattedDate: "1 June, 2025"}
	store := &repository.MemoryPostStore{BlogPosts: []*models.BlogPost{blogPost}}
	cache := cache.New([]*models.BlogPost{}, &sync.Mutex{})

	server := newTestServer(t, store, cache)
	defer server.Close()

	// First request
	req1, err := http.NewRequest(http.MethodGet, server.URL+"/blogpost?name=testtitle", nil)
	require.NoError(t, err)
	resp1, err := http.DefaultClient.Do(req1)
	require.NoError(t, err)
	assert.Equal(t, 1, store.AccessCounter)

	body1, err := io.ReadAll(resp1.Body)
	require.NoError(t, err)
	content1 := string(body1)
	assert.Contains(t, content1, "Test Title")
	assert.Contains(t, content1, "Test Content")

	// Second request should use cache
	req2, err := http.NewRequest(http.MethodGet, server.URL+"/blogpost?name=testtitle", nil)
	require.NoError(t, err)

	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()
	require.Equal(t, http.StatusOK, resp2.StatusCode)

	//Confirm that we only accessed the database once and served this from cache.
	assert.Equal(t, 1, store.AccessCounter)

	body2, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)
	content2 := string(body2)

	// Contents should be the same
	assert.Equal(t, content1, content2)
}

func TestGetBlogPostByName_NoCache(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	blogPost := &models.BlogPost{ID: id, Name: "testtitle", Title: "Test Title", Content: "Test Content", FormattedDate: "1 June, 2025"}
	store := &repository.MemoryPostStore{BlogPosts: []*models.BlogPost{blogPost}}
	cache := cache.New([]*models.BlogPost{}, &sync.Mutex{})

	server := newTestServer(t, store, cache)
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL+"/blogpost?name=testtitle", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	// database should be access once here
	assert.Equal(t, 1, store.AccessCounter)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	content1 := string(body)
	assert.Contains(t, content1, "Test Title")
	assert.Contains(t, content1, "Test Content")
}

func newTestServer(t *testing.T, store repository.PostStore, cache *cache.Cache) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux, handlers.NewApplication("foo", "foo", store, cache))
	server := httptest.NewServer(mux)
	return server
}
