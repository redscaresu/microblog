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

	m := microblog.MapPostStore{}
	m.Post = map[string]string{}
	netListener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	netListener.Close()

	go func() {
		err := microblog.ListenAndServe(netListener, m)
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

func TestMapStorePost(t *testing.T) {
	t.Parallel()
	m := &microblog.MapPostStore{Post: map[string]string{"1": "foo"}}

	netListener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	netListener.Close()

	go func() {
		err := microblog.ListenAndServe(netListener, m)
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
