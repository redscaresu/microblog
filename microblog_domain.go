package microblog

type PostStore interface {
	Create(BlogPost) error
	GetAll() ([]BlogPost, error)
}

type BlogPost struct {
	Blog_Id   string `json:"id"`
	Blog_Post string `json:"blog_post,omitempty"`
}

func NewBlogPost() *BlogPost {
	blogpost := &BlogPost{}
	return blogpost
}
