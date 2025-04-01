package models

import (
	"time"

	"github.com/google/uuid"
)

type BlogPost struct {
	ID            uuid.UUID
	Name          string
	Title         string
	Content       string
	FormattedDate string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewBlogPost() *BlogPost {
	blogpost := &BlogPost{}
	return blogpost
}
