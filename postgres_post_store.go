package microblog

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
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
		bp := NewBlogPost()
		err := rows.Scan(&bp.Blog_Id, &bp.Blog_Post)
		if err != nil {
			panic(err)
		}
		blogPosts = append(blogPosts, *bp)
	}

	return blogPosts, nil
}

func (p *PostgresStore) Create(blogpost BlogPost) (BlogPost, error) {
	blogpost.Blog_Id = uuid.NewString()
	return blogpost, nil
}
