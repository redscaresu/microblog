package microblog

import "github.com/google/uuid"

type MemoryPostStore struct {
	BlogPosts []BlogPost
}

func (s MemoryPostStore) GetAll() ([]BlogPost, error) {
	return s.BlogPosts, nil
}

func (s MemoryPostStore) Create(blogpost BlogPost) error {
	blogpost.ID = uuid.New()
	return nil
}

func (s MemoryPostStore) Get(id uuid.UUID) (BlogPost, error) {

	return *NewBlogPost(), nil
}
