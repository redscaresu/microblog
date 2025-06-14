package repository

import (
	"microblog/pkg/models"
	"time"

	"github.com/google/uuid"
)

type MemoryPostStore struct {
	BlogPosts []*models.BlogPost
}

func (s *MemoryPostStore) GetAll() ([]*models.BlogPost, error) {
	return s.BlogPosts, nil
}

func (s *MemoryPostStore) Create(blogpost *models.BlogPost) error {
	s.BlogPosts = append(s.BlogPosts, blogpost)
	return nil
}

func (s *MemoryPostStore) GetByID(id uuid.UUID) (*models.BlogPost, error) {
	for _, v := range s.BlogPosts {
		if v.ID == id {
			return v, nil
		}
	}
	return &models.BlogPost{}, nil
}

func (s *MemoryPostStore) GetByName(name string) (*models.BlogPost, error) {
	for _, v := range s.BlogPosts {
		if v.Name == name {
			return v, nil
		}
	}
	return &models.BlogPost{}, nil
}

func (s *MemoryPostStore) FetchLast10BlogPosts() ([]*models.BlogPost, error) {
	return s.BlogPosts, nil
}

func (s *MemoryPostStore) Delete(id uuid.UUID) error {
	return nil
}

func (s *MemoryPostStore) Update(updatedBlogpost *models.BlogPost) error {
	for _, v := range s.BlogPosts {
		if v.ID == updatedBlogpost.ID {
			v.Content = updatedBlogpost.Content
			v.Title = updatedBlogpost.Title
			v.UpdatedAt = time.Now().UTC()
		}
	}
	return nil
}
