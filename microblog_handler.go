package microblog

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	//go:embed templates/*
	templates embed.FS
)

func ListenAndServe(addr string, ps PostStore) error {
	//pass netListener as string not netListener object
	customMux := http.NewServeMux()

	// customMux.HandleFunc("/write", func(w http.ResponseWriter, r *http.Request) {
	// 	userPassword := r.FormValue("password")
	// 	if IsAuthenticated(userPassword) {
	// 		bg := &BlogPost{Blog_Id: int(uuid.New().ID()), Blog_Post: r.FormValue("blog")}
	// 		log.Println(bg)
	// 		err := ps.Create(*bg)
	// 		if err != nil {
	// 			w.WriteHeader(http.StatusInternalServerError)
	// 			fmt.Fprint(w, err)
	// 			return
	// 		}
	// 		w.WriteHeader(http.StatusCreated)
	// 		fmt.Fprint(w, "awesome blog post")
	// 		return
	// 	}
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	fmt.Fprint(w, "access denied")
	// })

	customMux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		posts, err := ps.GetAll()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
			return
		}
		fmt.Fprint(w, posts)
	})

	customMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// posts, err := ps.GetAll()
		// if err != nil {
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	fmt.Fprint(w, err)
		// 	return
		// }

		err := RenderHTMLTemplate(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w)
	})

	err := http.ListenAndServe(addr, customMux)
	fmt.Println(err)
	return err
}

func RenderHTMLTemplate(w io.Writer) error {
	blog := template.Must(template.New("main").ParseFS(templates, "templates/home.gohtml"))
	err := blog.Execute(w, "foo")
	if err != nil {
		log.Panic(err)
	}
	return nil
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
