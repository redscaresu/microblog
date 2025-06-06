package models

import (
	"time"

	"github.com/google/uuid"
)

type BlogPost struct {
	ID            uuid.UUID `json:"id" db:"blog_id"`
	Title         string    `json:"title" db:"blog_title"`
	Content       string    `json:"content" db:"blog_post"`
	Name          string    `json:"name" db:"blog_name"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	FormattedDate string    `json:"formatted_date" db:"formatted_date"`
}

func NewBlogPost() *BlogPost {
	blogpost := &BlogPost{}
	return blogpost
}
