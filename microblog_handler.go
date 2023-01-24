package microblog

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
)

func ListenAndServe(addr string, ps PostStore) error {
	//pass netListener as string not netListener object
	customMux := http.NewServeMux()

	customMux.HandleFunc("/write", func(w http.ResponseWriter, r *http.Request) {
		bg := &BlogPost{Blog_Id: uuid.NewString(), Blog_Post: r.FormValue("bonbon")}
		_, err := ps.Create(*bg)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, "awesome blog post")
	})

	customMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		posts, err := ps.GetAll()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
			return
		}
		fmt.Fprint(w, posts)
	})

	err := http.ListenAndServe(addr, customMux)
	fmt.Println(err)
	return err
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
