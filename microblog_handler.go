package microblog

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
)

func ListenAndServe(addr string, ps PostStore) error {
	//pass netListener as string not netListener object
	customMux := http.NewServeMux()

	customMux.HandleFunc("/write", func(w http.ResponseWriter, r *http.Request) {
		userPassword := r.FormValue("password")
		if IsAuthenticated(userPassword) {
			bg := &BlogPost{Blog_Id: int(uuid.New().ID()), Blog_Post: r.FormValue("blog")}
			log.Println(bg)
			err := ps.Create(*bg)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, err)
				return
			}
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, "awesome blog post")
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "access denied")
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
