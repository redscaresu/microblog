package microblog

import (
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/google/uuid"
)

var (
	//go:embed templates/*
	templates embed.FS
)

type Application struct {
	Auth struct {
		Username string
		Password string
	}
	Poststore PostStore
}

func ListenAndServe(addr string, app Application) error {

	if app.Auth.Username == "" {
		log.Fatal("basic auth username must be provided")
	}

	if app.Auth.Password == "" {
		log.Fatal("basic auth password must be provided")
	}

	customMux := http.NewServeMux()

	customMux.HandleFunc("/", app.Home)
	customMux.HandleFunc("/getlast5blogposts", app.basicAuth(app.NewPostHandler))
	customMux.HandleFunc("/submit", app.basicAuth(app.Submit))
	customMux.HandleFunc("/newpost", app.basicAuth(app.NewPostHandler))

	err := http.ListenAndServe(addr, customMux)
	return err
}

func (app *Application) basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte(app.Auth.Username))
			expectedPasswordHash := sha256.Sum256([]byte(app.Auth.Password))

			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func (app *Application) NewPostHandler(w http.ResponseWriter, r *http.Request) {
	blog := template.Must(template.New("main").ParseFS(templates, "templates/newpost.gohtml"))
	err := blog.Execute(w, "foo")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w)
}

func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	// Assuming you have a "last5blogposts.gohtml" template file in your templates directory
	tpl, err := template.ParseFS(templates, "templates/home.gohtml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute the template with the blog post data
	err = tpl.Execute(w, "foo")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (app *Application) GetLast5BlogPosts(w http.ResponseWriter, r *http.Request) {

	last5Posts, err := app.Poststore.FetchLast5BlogPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprint(w, last5Posts)

}

func (app *Application) ReadAllHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := app.Poststore.GetAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprint(w, posts)
}

func (app *Application) Submit(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	newBlogPost := &BlogPost{}
	err = json.NewDecoder(r.Body).Decode(newBlogPost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newBlogPost.ID = uuid.New()

	err = app.Poststore.Create(*newBlogPost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(newBlogPost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Post submitted successfully!")
}
