package cache

import (
	"microblog/pkg/models"
	"sync"
)

type Cache struct {
	BlogPosts []*models.BlogPost
	Mutex     *sync.Mutex
}

func New(blogPosts []*models.BlogPost, mutex *sync.Mutex) *Cache {

	return &Cache{
		BlogPosts: blogPosts,
		Mutex:     mutex,
	}
}

func (c *Cache) Lock() {
	c.Mutex.Lock()
}

func (c *Cache) Unlock() {
	c.Mutex.Unlock()
}

func (c *Cache) LoadCache(blogPosts []*models.BlogPost) {
	c.Mutex.Lock()
	c.BlogPosts = blogPosts
	c.Mutex.Unlock()
}

func (c *Cache) InvalidateCache() {
	c.Mutex.Lock()
	c.BlogPosts = nil
	c.Mutex.Unlock()
}
