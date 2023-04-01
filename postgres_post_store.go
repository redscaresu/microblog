package microblog

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var (
	host     = os.Getenv("INSTANCE_HOST")
	port     = os.Getenv("DB_PORT")
	user     = os.Getenv("DB_USER")
	password = os.Getenv("DB_USER")
	dbname   = os.Getenv("DB_NAME")
)

type PostgresStore struct {
	DB *sql.DB
}

func New() *PostgresStore {

	psqlInfo := fmt.Sprintf("host=%s port=%v user=%s "+
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

func (p *PostgresStore) Create(blogpost BlogPost) error {

	_, err := p.DB.Query("insert into blog values ($1,$2);", blogpost.Blog_Id, blogpost.Blog_Post)
	if err != nil {
		panic(err)
	}
	return nil
}
