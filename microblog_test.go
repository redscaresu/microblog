package microblog_test

import (
	"io"
	"microblog"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// func TestServerReturnsHelloWorld(t *testing.T) {

// 	defer func() {
// 		if r := recover(); r != nil {
// 			fmt.Println("Recovered. Error:\n", r)
// 			t.Fatal(r)
// 		}
// 	}()

// 	go microblog.ListenAndServe()

// 	resp, err := http.Get("http://127.0.0.1:8080/")

// 	for err != nil {
// 		t.Log("retrying")
// 		resp, err = http.Get("http://127.0.0.1:8080/")
// 	}

// 	if resp.StatusCode != http.StatusOK {
// 		t.Fatal(resp.StatusCode)
// 	}

// 	read, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		t.Fatal("test fail")
// 	}
// 	got := string(read)
// 	want := "<h1>hello world</h1>"

// 	if !cmp.Equal(want, got) {
// 		t.Error(cmp.Diff(want, got))
// 	}
// }

func TestMapStorePost(t *testing.T) {
	m := &microblog.MapPostStore{Post: map[string]string{"1": "foo"}}
	go microblog.ListenAndServe(m)

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
	want := "foo"

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}

}
