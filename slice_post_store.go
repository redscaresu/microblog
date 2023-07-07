package microblog

import "github.com/google/uuid"

type SlicePostStore struct {
	BlogPosts []BlogPost
}

func (s SlicePostStore) GetAll() ([]BlogPost, error) {
	blogPosts := []BlogPost{}
	for _, v := range s.BlogPosts {
		blogPosts = append(blogPosts, v)
	}

	return blogPosts, nil
}

func (s SlicePostStore) Create(blogpost BlogPost) error {

	blogpost.ID = int64(uuid.New().ID())
	return nil
}
