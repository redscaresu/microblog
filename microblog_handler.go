package microblog

import (
	"crypto/sha256"
	"crypto/subtle"
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

type application struct {
	auth struct {
		username string
		password string
	}
	poststore PostStore
}

func (app *application) protectedHandler(w http.ResponseWriter, r *http.Request) {
	err := RenderHTMLTemplate(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w)
}

func (app *application) unprotectedHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := app.poststore.GetAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprint(w, posts)
}

func (app *application) basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte(app.auth.username))
			expectedPasswordHash := sha256.Sum256([]byte(app.auth.password))

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

func ListenAndServe(addr string, ps PostStore) error {

	app := new(application)
	app.poststore = ps

	app.auth.username = os.Getenv("AUTH_USERNAME")
	app.auth.password = os.Getenv("AUTH_PASSWORD")

	if app.auth.username == "" {
		log.Fatal("basic auth username must be provided")
	}

	if app.auth.password == "" {
		log.Fatal("basic auth password must be provided")
	}

	customMux := http.NewServeMux()

	customMux.HandleFunc("/", app.unprotectedHandler)
	customMux.HandleFunc("/write", app.basicAuth(app.protectedHandler))

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
