package microblog

import (
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/json"
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

type Application struct {
	Auth struct {
		Username string
		Password string
	}
	Poststore PostStore
}

func (app *Application) submitFormHandler(w http.ResponseWriter, r *http.Request) {
	err := RenderHTMLTemplate(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w)
}

func (app *Application) readAllHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := app.Poststore.GetAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprint(w, posts)
}

func (app *Application) submitHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// title := r.FormValue("title")
	newBlogPost := &BlogPost{}
	err = json.NewDecoder(r.Body).Decode(newBlogPost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = app.Poststore.Create(*newBlogPost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Post submitted successfully!")
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

func ListenAndServe(addr string, app Application) error {

	if app.Auth.Username == "" {
		log.Fatal("basic auth username must be provided")
	}

	if app.Auth.Password == "" {
		log.Fatal("basic auth password must be provided")
	}

	customMux := http.NewServeMux()

	customMux.HandleFunc("/", app.readAllHandler)
	customMux.HandleFunc("/submitform", app.basicAuth(app.submitFormHandler))
	customMux.HandleFunc("/submit", app.basicAuth(app.submitHandler))

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
