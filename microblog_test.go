package microblog_test

import (
	"io"
	"microblog"
	"net"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestListenAndServe_UsesGivenStore(t *testing.T) {
	t.Parallel()

	blogPost := &microblog.BlogPost{ID: 1, Title: "blog"}
	store := &microblog.SlicePostStore{BlogPosts: []microblog.BlogPost{*blogPost}}

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
	want := "[{1 blog}]"

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
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

func newTestServer(t *testing.T, store microblog.PostStore) net.Addr {
	t.Helper()

	netListener, err := net.Listen("tcp", "127.0.0.1:")
	addr := netListener.Addr().String()

	if err != nil {
		t.Fatal(err)
	}
	netListener.Close()

	go func() {
		err := microblog.ListenAndServe(addr, store)
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
