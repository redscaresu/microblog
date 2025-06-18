package handlers

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"microblog/pkg/cache"
	"microblog/pkg/models"
	"microblog/pkg/repository"
	"net/http"
	"regexp"
	"strings"
	"time"

	htmltemplate "html/template"
	texttemplate "text/template"

	"github.com/google/uuid"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var (
	//go:embed templates/*
	templates embed.FS
	md        goldmark.Markdown
)

func init() {
	md = goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)
}

var funcMap = texttemplate.FuncMap{
	"safeHTML": func(s string) htmltemplate.HTML {
		return htmltemplate.HTML(s)
	},
}

const re = `[^a-zA-Z0-9\s]+`

type Application struct {
	Auth      *Auth
	PostStore repository.PostStore
	Cache     *cache.Cache
}

type Auth struct {
	UserName string
	Password string
}

func NewApplication(userName, passWord string, postStore repository.PostStore, cache *cache.Cache) *Application {

	return &Application{
		Auth: &Auth{
			UserName: userName,
			Password: passWord,
		},
		PostStore: postStore,
		Cache:     cache,
	}
}

func RegisterRoutes(mux *http.ServeMux, app *Application) {
	mux.HandleFunc("/", app.Home)
	mux.HandleFunc("/blogpost", app.GetBlogPostByName)
	mux.HandleFunc("/submit", app.basicAuth(app.Submit))
	mux.HandleFunc("/editpost", app.basicAuth(app.EditPostHandler))
	mux.HandleFunc("/newpost", app.basicAuth(app.NewPostHandler))
	mux.HandleFunc("/updatepost", app.basicAuth(app.UpdatePostHandler))
	mux.HandleFunc("/deletepost", app.basicAuth(app.DeletePostHandler))
	mux.HandleFunc("/rebuildcache", app.basicAuth(app.RebuildCacheHandler))
}

func (app *Application) basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte(app.Auth.UserName))
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
	tpl, err := texttemplate.ParseFS(templates, "templates/newpost.gohtml")
	if err != nil {
		log.Printf("Failed to load template: %v", err)
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}

	err = tpl.Execute(w, nil)
	if err != nil {
		log.Printf("Failed to render template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

func (app *Application) EditPostHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	name := queryParams.Get("name")
	if name == "" {
		http.Error(w, "name is empty", http.StatusBadRequest)
		return
	}

	blog, err := app.PostStore.GetByName(name)
	if err != nil {
		log.Printf("Error getting post by name %s: %v", name, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tpl, err := texttemplate.ParseFS(templates, "templates/editpost.gohtml")
	if err != nil {
		log.Printf("Error parsing editpost.gohtml template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tpl.Execute(w, blog)
	if err != nil {
		log.Printf("Error executing editpost.gohtml template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	tpl, err := texttemplate.ParseFS(templates, "templates/home.gohtml")
	if err != nil {
		log.Printf("Error parsing home.gohtml template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(app.Cache.BlogPosts) < 1 {
		// cache miss, lets fetch from the database
		unNormalizedblogPosts, err := app.PostStore.FetchLast10BlogPosts()
		if err != nil {
			log.Printf("Error fetching last 10 blog posts: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// inflate the cache with what has come from the DB
		app.Cache.LoadCache(unNormalizedblogPosts)
	}

	// if we miss the miss the cache then app.Cache is initialized from line 183
	// if we hit the cache then we just immediately use the current app.Cache
	normalizedBlogPost := normalizeBlogPost(app.Cache.BlogPosts)

	err = tpl.Execute(w, normalizedBlogPost)
	if err != nil {
		log.Printf("Error executing home.gohtml template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (app *Application) GetBlogPostByID(w http.ResponseWriter, r *http.Request) {

	queryParams := r.URL.Query()
	id := queryParams.Get("id")

	blog, err := app.PostStore.GetByID(uuid.MustParse(id))
	if err != nil {
		log.Printf("Error getting blog post by ID %s: %v", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(blog)
	if err != nil {
		log.Printf("Error marshalling blog post ID %s: %v", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(resp))
}

func (app *Application) GetBlogPostByName(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	name := queryParams.Get("name")

	app.Cache.Lock()
	if len(app.Cache.BlogPosts) > 0 {
		for _, cachedPost := range app.Cache.BlogPosts {
			//blog exists in cache
			if cachedPost.Name == name {
				blog := cachedPost
				blog.Content = RenderMarkdown(blog.Content)
				blog.Title = RenderMarkdown(blog.Title)
				log.Printf("Processed Content for %s: %s", name, blog.Content)

				tpl, err := texttemplate.New("blogpost.gohtml").Funcs(funcMap).ParseFS(templates, "templates/blogpost.gohtml")
				if err != nil {
					log.Printf("Error parsing blogpost.gohtml template: %v", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				err = tpl.Execute(w, blog)
				if err != nil {
					log.Printf("Error executing blogpost.gohtml template: %v", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
		app.Cache.Unlock()
		return
	}
	// unlock the cache from the above.
	app.Cache.Unlock()

	// cache miss, lets fetch from the database
	unNormalizedblogPosts, err := app.PostStore.FetchLast10BlogPosts()
	if err != nil {
		log.Printf("Error fetching last 10 blog posts: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// inflate the cache with what has come from the DB
	app.Cache.LoadCache(unNormalizedblogPosts)

	var blog *models.BlogPost
	for _, cachedBlogPost := range app.Cache.BlogPosts {
		if cachedBlogPost.Name == name {
			blog = cachedBlogPost
		}
	}

	log.Printf("Original Content for %s: %s", name, blog.Content)
	blog.Content = RenderMarkdown(blog.Content)
	blog.Title = RenderMarkdown(blog.Title)

	log.Printf("Processed Content for %s: %s", name, blog.Content)

	tpl, err := texttemplate.New("blogpost.gohtml").Funcs(funcMap).ParseFS(templates, "templates/blogpost.gohtml")
	if err != nil {
		log.Printf("Error parsing blogpost.gohtml template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tpl.Execute(w, blog)
	if err != nil {
		log.Printf("Error executing blogpost.gohtml template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

	ID := uuid.New()

	titleCopy := title
	name := strings.ReplaceAll(titleCopy, " ", "-")
	name = strings.ToLower(name)

	rexp := regexp.MustCompile(re)
	name = rexp.ReplaceAllString(name, "")

	now := time.Now().UTC()
	newBlogPost := &models.BlogPost{
		ID:            ID,
		Name:          name,
		Title:         title,
		Content:       content,
		CreatedAt:     now,
		UpdatedAt:     now,
		FormattedDate: formattedDate(now),
	}

	err = app.PostStore.Create(newBlogPost)
	if err != nil {
		log.Printf("Error creating post: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	blogPosts, err := app.PostStore.FetchLast10BlogPosts()
	if err != nil {
		http.Error(w, "unable to fetch last 10 blog posts", http.StatusInternalServerError)
		return
	}

	// rehydrate the cache with what has come out of the DB
	app.Cache.LoadCache(blogPosts)

	err = json.NewEncoder(w).Encode(newBlogPost)
	if err != nil {
		log.Printf("Error encoding new blog post: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, "cache reloaded")
	fmt.Fprintf(w, "Post submitted successfully!")
}

func (app *Application) UpdatePostHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Printf("Unable to parse form: %v", err)
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	id := r.FormValue("id")
	title := r.FormValue("title")
	content := r.FormValue("content")

	log.Printf("Updating Post - ID: %s, Title: %s, Content: %s\n", id, title, content)

	idUUID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("Invalid ID for update: %s, error: %v", id, err)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	newBlogPost := &models.BlogPost{
		ID:        idUUID,
		Title:     title,
		Content:   content,
		UpdatedAt: now,
	}

	err = app.PostStore.Update(newBlogPost)
	if err != nil {
		log.Printf("Error updating post ID %s: %v", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	blogPosts, err := app.PostStore.FetchLast10BlogPosts()
	if err != nil {
		http.Error(w, "unable to fetch last 10 blog posts", http.StatusInternalServerError)
		return
	}

	// reinflate the cache with what has come out of the DB
	app.Cache.LoadCache(blogPosts)
	fmt.Fprintf(w, "cache reloaded")
	fmt.Fprintf(w, "Post updated successfully!")
}

func (app *Application) DeletePostHandler(w http.ResponseWriter, r *http.Request) {

	queryParams := r.URL.Query()
	id := queryParams.Get("id")

	idUUID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("Invalid ID for delete: %s, error: %v", id, err)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = app.PostStore.Delete(idUUID)
	if err != nil {
		log.Printf("Error deleting post ID %s: %v", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Cache.InvalidateCache()
	fmt.Fprint(w, "Cache deleted")
	fmt.Fprintf(w, "Post deleted successfully!")
}

func (app *Application) RebuildCacheHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	app.Cache.InvalidateCache()
	log.Println("Cache invalidated.")

	allPosts, err := app.PostStore.FetchLast10BlogPosts()
	if err != nil {
		log.Printf("Error fetching posts from store to rebuild cache: %v", err)
		http.Error(w, "Failed to rebuild cache: could not fetch posts", http.StatusInternalServerError)
		return
	}

	app.Cache.LoadCache(allPosts)
	log.Printf("Cache rebuilt successfully with %d posts.", len(allPosts))

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Cache invalidated and rebuilt successfully with %d posts.\n", len(allPosts))
}

func normalizeBlogPost(unNormalizedBlogPosts []*models.BlogPost) []*models.BlogPost {
	preview := 300

	normalizedBlogPosts := make([]*models.BlogPost, len(unNormalizedBlogPosts))

	for i := range unNormalizedBlogPosts {
		normalizedBlogPosts[i] = &models.BlogPost{
			ID:            unNormalizedBlogPosts[i].ID,
			Title:         unNormalizedBlogPosts[i].Title,
			Content:       unNormalizedBlogPosts[i].Content,
			Name:          unNormalizedBlogPosts[i].Name,
			CreatedAt:     unNormalizedBlogPosts[i].CreatedAt,
			UpdatedAt:     unNormalizedBlogPosts[i].UpdatedAt,
			FormattedDate: unNormalizedBlogPosts[i].FormattedDate,
		}
		var contentBuf bytes.Buffer
		if err := md.Convert([]byte(normalizedBlogPosts[i].Content), &contentBuf); err != nil {
			log.Printf("Error converting blog post content to HTML: %v\n", err)
		} else {
			normalizedBlogPosts[i].Content = contentBuf.String()
		}
		var titleBuf bytes.Buffer
		if err := md.Convert([]byte(normalizedBlogPosts[i].Title), &titleBuf); err != nil {
			log.Printf("Error converting blog post title to HTML: %v\n", err)
		} else {
			normalizedBlogPosts[i].Title = titleBuf.String()
		}
	}

	for i := range normalizedBlogPosts {
		if len(normalizedBlogPosts[i].Content) > preview {
			normalizedBlogPosts[i].Content = unNormalizedBlogPosts[i].Content[:preview] + "..."
		}
	}
	return normalizedBlogPosts
}

func formattedDate(now time.Time) string {
	return fmt.Sprintf("%s %d, %d", now.Month().String(), now.Day(), now.Year())
}

func RenderMarkdown(content string) string {
	var buf bytes.Buffer
	if err := md.Convert([]byte(content), &buf); err != nil {
		log.Printf("Error converting markdown to HTML: %v", err)
		return content
	}
	parsedContent := buf.String()

	log.Println("=== MARKDOWN DEBUG ===")
	log.Println("Input:", content)
	log.Println("Output:", parsedContent)
	log.Println("======================")

	return parsedContent
}
