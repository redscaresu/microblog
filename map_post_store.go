package microblog

import (
	"github.com/google/uuid"
)

type MapPostStore struct {
	Post map[string]string
}

func (m MapPostStore) GetAll() ([]string, error) {
	posts := []string{}
	for _, v := range m.Post {
		posts = append(posts, v)
	}
	return posts, nil
}

func (m MapPostStore) Create(post string) error {
	m.Post[uuid.NewString()] = post
	return nil
}
