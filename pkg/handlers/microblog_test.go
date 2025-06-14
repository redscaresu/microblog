package handlers_test

import (
	"encoding/json"
	"fmt"
	"io"
	"microblog/pkg/handlers"
	"microblog/pkg/models"
	"microblog/pkg/repository"
	"net"
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

// RoundTripper is the custom RoundTripper that logs API calls
type CountingRoundTripper struct {
	Count int
}

// RoundTrip intercepts the HTTP request and logs the details
func (r *CountingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	r.Count++
	// Perform the request
	return http.DefaultTransport.RoundTrip(req)
}

func TestListenAndServe_CacheHit(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	blogPost := &models.BlogPost{ID: id, Title: "foo", Content: "boo", FormattedDate: "1 June, 2025"}
	store := &repository.MemoryPostStore{BlogPosts: []*models.BlogPost{blogPost}}

	addr := newTestServer(t, store)

	countRoundTripper := &CountingRoundTripper{}
	client := http.Client{
		Transport: countRoundTripper,
	}

	resp, err := client.Get("http://" + addr.String())
	require.NoError(t, err)
	assert.Equal(t, 1, countRoundTripper.Count)

	read, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	got := string(read)
	assert.Contains(t, got, "<h2><a href=\"/blogpost?name=")
	assert.Contains(t, got, "<p>foo</p>")
	assert.Contains(t, got, "<p>boo</p>")
	assert.Contains(t, got, "<h3 style=\"color: grey; font-size: 0.9em;\">1 June, 2025</h3>")

	// test cache hit
	resp, err = client.Get("http://" + addr.String())
	require.NoError(t, err)

	assert.Equal(t, 2, countRoundTripper.Count)

	read, err = io.ReadAll(resp.Body)
	require.NoError(t, err)

	got = string(read)
	assert.Contains(t, got, "<h2><a href=\"/blogpost?name=")
	assert.Contains(t, got, "<p>foo</p>")
	assert.Contains(t, got, "<p>boo</p>")
	assert.Contains(t, got, "<h3 style=\"color: grey; font-size: 0.9em;\">1 June, 2025</h3>")
}

func TestSubmitHandler(t *testing.T) {
	t.Parallel()

	store := &repository.MemoryPostStore{}
	cache := &handlers.Cache{
		BlogPosts: []*models.BlogPost{},
		Mutex:     &sync.Mutex{},
	}
	app := handlers.NewApplication("", "", store, cache)
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
	cache := &handlers.Cache{
		BlogPosts: []*models.BlogPost{},
		Mutex:     &sync.Mutex{},
	}
	app := handlers.NewApplication("", "", store, cache)
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

	store := &repository.MemoryPostStore{}
	addr := newTestServer(t, store)

	submitURL := fmt.Sprintf("http://%s/submit", addr.String())

	form := url.Values{}
	form.Add("title", "Test Title")
	form.Add("content", "Test Content")

	// Set basic auth for the submit request
	req, err := http.NewRequest("POST", submitURL, strings.NewReader(form.Encode()))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("foo", "foo")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var createdPost models.BlogPost
	err = json.NewDecoder(resp.Body).Decode(&createdPost)
	require.NoError(t, err)

	// Now test the cache behavior for the GetBlogPostByName endpoint
	// First request should populate the cache
	blogPostURL := fmt.Sprintf("http://%s/blogpost?name=testtitle", addr.String())

	countingTransport := &CountingRoundTripper{}
	cacheClient := &http.Client{Transport: countingTransport}

	// First request
	resp1, err := cacheClient.Get(blogPostURL)
	require.NoError(t, err)
	defer resp1.Body.Close()
	require.Equal(t, http.StatusOK, resp1.StatusCode)
	assert.Equal(t, 1, countingTransport.Count)

	body1, err := io.ReadAll(resp1.Body)
	require.NoError(t, err)
	content1 := string(body1)
	assert.Contains(t, content1, "Test Title")
	assert.Contains(t, content1, "Test Content")

	// Second request should use cache
	resp2, err := cacheClient.Get(blogPostURL)
	require.NoError(t, err)
	defer resp2.Body.Close()
	require.Equal(t, http.StatusOK, resp2.StatusCode)
	assert.Equal(t, 2, countingTransport.Count)

	body2, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)
	content2 := string(body2)

	// Contents should be the same
	assert.Equal(t, content1, content2)
}

// // delete it from the cache
// app.Cache.InvalidateCache()

// addr := newTestServer(t, store)

// countRoundTripper := &CountingRoundTripper{}
// client := http.Client{
// 	Transport: countRoundTripper,
// }

// blogPostURL := fmt.Sprintf("http://%s/blogpost?name=%s", addr.String(), "testtitle")

// resp, err = client.Get(blogPostURL)
// require.NoError(t, err, "First request to GetBlogPostByName should not error. URL used: "+blogPostURL)
// defer resp.Body.Close()

// require.Equal(t, http.StatusOK, resp.StatusCode)

// var gotAgain models.BlogPost
// err = json.NewDecoder(resp.Body).Decode(&got)
// require.NoError(t, err)

// fmt.Println(gotAgain)
// assert.NotNil(t, gotAgain)

func newTestServer(t *testing.T, store repository.PostStore) net.Addr {
	t.Helper()

	netListener, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	addr := netListener.Addr().String()
	netListener.Close()

	cache := &handlers.Cache{
		BlogPosts: []*models.BlogPost{},
		Mutex:     &sync.Mutex{},
	}

	mux := http.NewServeMux()
	go func() {
		err := handlers.RegisterRoutes(mux,
			addr,
			handlers.NewApplication("foo",
				"foo",
				store,
				cache))
		require.NoError(t, err)
	}()

	resp, err := http.Get("http:" + addr)
	for err != nil {
		t.Log("retrying")
		resp, err = http.Get("http://" + addr)
		t.Logf("unable to get endpoint, err is %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatal("non statusok return", resp.StatusCode)
	}

	return netListener.Addr()
}
