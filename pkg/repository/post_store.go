package repository

import (
	"microblog/pkg/models"

	"github.com/google/uuid"
)

type PostStore interface {
	Create(*models.BlogPost) error
	GetAll() ([]*models.BlogPost, error)
	GetByID(id uuid.UUID) (*models.BlogPost, error)
	GetByName(name string) (*models.BlogPost, error)
	FetchLast10BlogPosts() ([]*models.BlogPost, error)
	Delete(id uuid.UUID) error
	Update(*models.BlogPost) error
}
