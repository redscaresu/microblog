package cache_test

import (
	"microblog/pkg/cache"
	"microblog/pkg/models"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestLoadCache(t *testing.T) {

	id1 := uuid.New()
	id2 := uuid.New()
	blogPost1 := &models.BlogPost{
		ID:            id1,
		Title:         "first title",
		Content:       "first content",
		Name:          "first name",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		FormattedDate: "1 June, 2025",
	}
	blogPost2 := &models.BlogPost{
		ID:            id2,
		Title:         "second title",
		Content:       "second content",
		Name:          "second name",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		FormattedDate: "2 June, 2025",
	}

	cache := cache.New([]*models.BlogPost{}, &sync.Mutex{})
	cache.Load([]*models.BlogPost{blogPost1, blogPost2})

	assert.Equal(t, []*models.BlogPost{blogPost1, blogPost2}, cache.BlogPosts)
}

func TestInvalidateCache(t *testing.T) {

	c := cache.New(
		[]*models.BlogPost{
			{
				ID:    uuid.New(),
				Title: "title",
			},
		},
		&sync.Mutex{})

	c.Invalidate()
	assert.Nil(t, c.BlogPosts)
}

func TestConcurrentLoadAndInvalidate(t *testing.T) {
	m := &sync.Mutex{}
	c := cache.New(nil, m)
	posts := []*models.BlogPost{
		{
			ID:    uuid.New(),
			Title: "Concurrent Title"},
	}

	done := make(chan struct{})
	go func() {
		c.Load(posts)
		done <- struct{}{}
	}()
	go func() {
		c.Invalidate()
		done <- struct{}{}
	}()
	<-done
	<-done
}
