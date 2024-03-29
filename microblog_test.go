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
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestListenAndServe_UsesGivenStore(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	blogPost := &microblog.BlogPost{ID: id, Title: "title", Content: "content"}
	store := &microblog.MemoryPostStore{BlogPosts: []microblog.BlogPost{*blogPost}}

	addr := newTestServer(t, store)

	resp, err := http.Get("http://" + addr.String())
	if err != nil {
		t.Fatal(err)
	}

	read, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("test fail")
	}
	got := string(read)
	want := fmt.Sprintf("[{%s title content}]", id)

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestSubmitHandler(t *testing.T) {

	// store := newTestDBConnection(t)
	store := &microblog.MemoryPostStore{}
	addr := newTestServer(t, store)
	body := strings.NewReader("{\"title\":\"boo\",\"content\":\"foo\"}")

	endpoint := fmt.Sprintf("http://%v/submit", addr)
	req, err := http.NewRequest("POST", endpoint, body)
	require.NoError(t, err)

	req.Header.Add("Authorization", "Basic "+basicAuth())
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	want := &microblog.BlogPost{}
	err = json.NewDecoder(resp.Body).Decode(want)
	require.NoError(t, err)

	got, err := store.GetByID(want.ID)
	require.NoError(t, err)

	if cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, &got))
	}
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
