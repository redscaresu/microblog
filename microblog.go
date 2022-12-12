package microblog

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func ListenAndServe(ps PostStore) error {

	http.HandleFunc("/write", func(w http.ResponseWriter, r *http.Request) {
		err := ps.Create(r.FormValue("text"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, "awesome blog post")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		posts, err := ps.GetAll()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
			return
		}
		fmt.Fprint(w, posts)
	})

	return http.ListenAndServe(":8080", nil)
}

type PostStore interface {
	Create(string) error
	GetAll() ([]string, error)
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
