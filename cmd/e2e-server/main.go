package main

import (
	"log"
	"microblog/pkg/cache"
	"microblog/pkg/handlers"
	"microblog/pkg/models"
	"microblog/pkg/repository"
	"net/http"
	"sync"
)

func main() {
	mux := http.NewServeMux()
	postStore := &repository.MemoryPostStore{BlogPosts: []*models.BlogPost{}}
	postCache := cache.New([]*models.BlogPost{}, &sync.Mutex{})
	app := handlers.NewApplication("foo", "foo", postStore, postCache)
	handlers.RegisterRoutes(mux, app)

	log.Fatal(http.ListenAndServe(":18080", mux))
}
