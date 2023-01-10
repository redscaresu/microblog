package microblog

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5438
	user     = "postgres"
	password = "postgres"
	dbname   = "postgres"
)

type PostgresStore struct {
	DB *sql.DB
}

type BlogPost struct {
	Blog_Id   int64  `json:"id"`
	Blog_Post string `json:"blog_post,omitempty"`
}

func New() *PostgresStore {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
	return &PostgresStore{DB: db}

}

func (p *PostgresStore) GetAll() ([]BlogPost, error) {

	blogPosts := []BlogPost{}

	rows, err := p.DB.Query("SELECT * FROM blog;")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		bp := BlogPost{}
		err := rows.Scan(&bp.Blog_Id, &bp.Blog_Post)
		if err != nil {
			panic(err)
		}
		blogPosts = append(blogPosts, bp)
	}

	fmt.Println(blogPosts)
	return blogPosts, nil

}

func (p *PostgresStore) Create(post string) error {
	return nil
}
