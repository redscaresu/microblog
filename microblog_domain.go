package microblog

import (
	"time"

	"github.com/google/uuid"
)

type PostStore interface {
	Create(BlogPost) error
	GetAll() ([]BlogPost, error)
	GetByID(id uuid.UUID) (BlogPost, error)
	GetByName(name string) (BlogPost, error)
	FetchLast10BlogPosts() ([]BlogPost, error)
	Delete(id uuid.UUID) error
	Update(BlogPost) error
}

type BlogPost struct {
	ID        uuid.UUID
	Name      string
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewBlogPost() *BlogPost {
	blogpost := &BlogPost{}
	return blogpost
}
