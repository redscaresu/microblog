package microblog

import (
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"text/template"

	"github.com/google/uuid"
	"github.com/russross/blackfriday/v2"
)

var (
	//go:embed templates/*
	templates embed.FS
)

const re = `[^a-zA-Z0-9\s]+`

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
	customMux.HandleFunc("/blogpost", app.GetBlogPostByName)
	customMux.HandleFunc("/getlast5blogposts", app.GetLast10BlogPosts)
	customMux.HandleFunc("/submit", app.basicAuth(app.Submit))
	customMux.HandleFunc("/editpost", app.basicAuth(app.EditPostHandler))
	customMux.HandleFunc("/newpost", app.basicAuth(app.NewPostHandler))
	customMux.HandleFunc("/updatepost", app.basicAuth(app.UpdatePostHandler))
	customMux.HandleFunc("/deletepost", app.basicAuth(app.DeletePostHandler))

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
	http.ServeFile(w, r, "templates/newpost.gohtml")
}

func (app *Application) EditPostHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	name := queryParams.Get("name")
	if name == "" {
		http.Error(w, "name is empty", http.StatusBadRequest)
		return
	}

	blog, err := app.Poststore.GetByName(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	tpl, err := template.ParseFS(templates, "templates/editpost.gohtml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tpl.Execute(w, blog)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFS(templates, "templates/home.gohtml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	blogPosts, err := app.Poststore.FetchLast10BlogPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	normalizedBlogPost := normalizeBlogPost(blogPosts)

	err = tpl.Execute(w, normalizedBlogPost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (app *Application) GetLast10BlogPosts(w http.ResponseWriter, r *http.Request) {

	last5Posts, err := app.Poststore.FetchLast10BlogPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}
	resp, err := json.Marshal(last5Posts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	fmt.Fprint(w, string(resp))
}

func (app *Application) GetBlogPostByID(w http.ResponseWriter, r *http.Request) {

	queryParams := r.URL.Query()
	id := queryParams.Get("id")

	blog, err := app.Poststore.GetByID(uuid.MustParse(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}
	resp, err := json.Marshal(blog)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	fmt.Fprint(w, string(resp))
}

func (app *Application) GetBlogPostByName(w http.ResponseWriter, r *http.Request) {

	queryParams := r.URL.Query()
	name := queryParams.Get("name")

	blog, err := app.Poststore.GetByName(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	blog.Content = string(blackfriday.Run([]byte(blog.Content)))
	blog.Title = string(blackfriday.Run([]byte(blog.Title)))

	tpl, err := template.ParseFS(templates, "templates/blogpost.gohtml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tpl.Execute(w, blog)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "Title is empty", http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Content is empty", http.StatusBadRequest)
		return
	}

	content = content[:len(content)-1]

	ID := uuid.New()

	titleCopy := title
	name := strings.ReplaceAll(titleCopy, " ", "-")
	name = strings.ToLower(name)

	rexp := regexp.MustCompile(re)
	name = rexp.ReplaceAllString(name, "")

	now := time.Now().UTC()
	newBlogPost := &BlogPost{
		ID:        ID,
		Name:      name,
		Title:     title,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}

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

func (app *Application) UpdatePostHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	id := r.FormValue("id")
	title := r.FormValue("title")
	content := r.FormValue("content")

	fmt.Printf("ID: %s, Title: %s, Content: %s\n", id, title, content)

	idUUID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	newBlogPost := &BlogPost{
		ID:        idUUID,
		Title:     title,
		Content:   content,
		UpdatedAt: now,
	}

	err = app.Poststore.Update(*newBlogPost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Post updated successfully!")
}

func (app *Application) DeletePostHandler(w http.ResponseWriter, r *http.Request) {

	queryParams := r.URL.Query()
	id := queryParams.Get("id")

	idUUID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = app.Poststore.Delete(idUUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Post deleted successfully!")
}

func normalizeBlogPost(blogPost []BlogPost) []BlogPost {
	preview := 20
	for i := range blogPost {
		blogPost[i].Content = string(blackfriday.Run([]byte(blogPost[i].Content)))
		blogPost[i].Title = string(blackfriday.Run([]byte(blogPost[i].Title)))
	}

	for i := range blogPost {
		if len(blogPost[i].Content) > preview {
			blogPost[i].Content = blogPost[i].Content[:preview] + "..."
		}
	}
	return blogPost
}
