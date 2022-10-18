package microblog

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func ListenAndServe(m *MapPostStore) error {

	TmpDir := "/Users/countdoo/work/microblog/"

	h1 := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, TmpDir+"blog.txt")
	}

	h2 := func(w http.ResponseWriter, r *http.Request) {
		CreateBlogEntry(w, r)
	}

	http.HandleFunc("/write", h2)
	http.HandleFunc("/", h1)

	return http.ListenAndServe(":8080", nil)
}

func CreateBlogEntry(w http.ResponseWriter, r *http.Request) {

	file, err := os.Create("blog.txt")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	defer file.Close()

	_, err = io.Copy(file, strings.NewReader(r.FormValue("text")))
	if err != nil {
		fmt.Println(err)
	}
}
