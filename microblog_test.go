package microblog_test

import (
	"fmt"
	"io"
	"microblog"
	"net"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestServerReturnsHelloWorld(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered. Error:\n", r)
			t.Fatal(r)
		}
	}()

	m := microblog.SlicePostStore{}
	m.BlogPosts = []microblog.BlogPost{}
	netListener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	netListener.Close()

	addr := netListener.Addr().String()

	go func() {
		err := microblog.ListenAndServe(addr, m)
		if err != nil {
			panic(err)
		}
	}()

	resp, err := http.Get("http://127.0.0.1:8080/")

	for err != nil {
		t.Log("retrying")
		resp, err = http.Get("http://127.0.0.1:8080/")
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatal(resp.StatusCode)
	}

	read, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("test fail")
	}
	got := string(read)
	want := "[bonbon]"

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestSliceStore(t *testing.T) {
	t.Parallel()

	blogPost := &microblog.BlogPost{Blog_Id: "1", Blog_Post: "bonbon"}
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
	want := "[{1 bonbon}]"

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}

}

func newTestServer(t *testing.T, store microblog.PostStore) net.Addr {

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
