package microblog

import "github.com/google/uuid"

type PostStore interface {
	Create(BlogPost) error
	GetAll() ([]BlogPost, error)
	Get(id uuid.UUID) (BlogPost, error)
}

type BlogPost struct {
	ID      uuid.UUID
	Title   string
	Content string
}

func NewBlogPost() *BlogPost {
	blogpost := &BlogPost{}
	return blogpost
}
