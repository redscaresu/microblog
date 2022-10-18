package microblog

import (
	"github.com/google/uuid"
)

type MapPostStore struct {
	Post map[string]string
}

func (m MapPostStore) ListAllPost() []string {
	posts := []string{}
	for _, v := range m.Post {
		posts = append(posts, v)
	}
	return posts
}

func (m *MapPostStore) CreatePost(post string) error {
	m.Post[uuid.NewString()] = post
	return nil
}
