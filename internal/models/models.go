package models

import (
	"time"

	"github.com/google/uuid"
)

type BlogPost struct {
	ID            uuid.UUID
	Title         string
	Content       string
	Name          string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	FormattedDate string
}

func NewBlogPost() *BlogPost {
	blogpost := &BlogPost{}
	return blogpost
}
