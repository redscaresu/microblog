package microblog

import "github.com/google/uuid"

type MemoryPostStore struct {
	BlogPosts []*BlogPost
}

func (s *MemoryPostStore) GetAll() ([]*BlogPost, error) {
	return s.BlogPosts, nil
}

func (s *MemoryPostStore) Create(blogpost *BlogPost) error {
	s.BlogPosts = append(s.BlogPosts, blogpost)
	return nil
}

func (s *MemoryPostStore) GetByID(id uuid.UUID) (*BlogPost, error) {
	for _, v := range s.BlogPosts {
		if v.ID == id {
			return v, nil
		}
	}
	return &BlogPost{}, nil
}

func (s *MemoryPostStore) GetByName(name string) (*BlogPost, error) {
	return NewBlogPost(), nil
}

func (s *MemoryPostStore) FetchLast10BlogPosts() ([]*BlogPost, error) {
	return s.BlogPosts, nil
}

func (s *MemoryPostStore) Delete(id uuid.UUID) error {
	return nil
}

func (s *MemoryPostStore) Update(blogpost *BlogPost) error {
	return nil
}
