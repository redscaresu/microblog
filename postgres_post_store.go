package microblog

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type PostgresStore struct {
	DB *sql.DB
}

func New() *PostgresStore {

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	unixSocketPath := os.Getenv("INSTANCE_UNIX_SOCKET")

	local := os.Getenv("LOCAL")

	var psqlInfo string

	psqlInfo = fmt.Sprintf("user=%s password=%s database=%s host=%s",
		user, password, dbName, unixSocketPath)

	if local == "local" {
		host := os.Getenv("HOST")
		port := os.Getenv("PORT")
		password := os.Getenv("PASSWORD")
		user := os.Getenv("USER")
		psqlInfo = fmt.Sprintf("host=%s port=%s user=%s "+
			"password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbName)
	}

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
		err := rows.Scan(&bp.ID, &bp.Title, &bp.Content)
		if err != nil {
			panic(err)
		}
		blogPosts = append(blogPosts, *bp)
	}

	return blogPosts, nil
}

func (p *PostgresStore) Create(blogpost BlogPost) error {

	_, err := p.DB.Query("insert into blog values ($1,$2,$3);", blogpost.ID, blogpost.Title, blogpost.Content)
	if err != nil {
		panic(err)
	}
	return nil
}

func (p *PostgresStore) Get(id uuid.UUID) (BlogPost, error) {

	bp := NewBlogPost()

	err := p.DB.QueryRow("SELECT * FROM blog WHERE blog_id = $1;", id).
		Scan(&bp.ID, &bp.Title, &bp.Content)
	if err != nil {
		panic(err)
	}

	return *bp, nil

}
