package microblog

type PostStore interface {
	Create(BlogPost) error
	GetAll() ([]BlogPost, error)
}

type BlogPost struct {
	ID      int64
	Title   string
	Content string
}

func NewBlogPost() *BlogPost {
	blogpost := &BlogPost{}
	return blogpost
}
