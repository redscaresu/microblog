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

func (s SlicePostStore) Create(blogpost BlogPost) (BlogPost, error) {

	blogpost.Blog_Id = uuid.NewString()
	return blogpost, nil
}