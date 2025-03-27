package microblog

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type PostgresStore struct {
	DB *sql.DB
}

func New() (*PostgresStore, error) {

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
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	log.Print("Successfully connected!")
	return &PostgresStore{DB: db}, nil
}

func (p *PostgresStore) GetAll() ([]BlogPost, error) {

	blogPosts := []BlogPost{}

	rows, err := p.DB.Query("SELECT * FROM blog;")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		bp := NewBlogPost()
		err := rows.Scan(&bp.ID, &bp.Title, &bp.Content, &bp.Name, &bp.CreatedAt, &bp.UpdatedAt)
		if err != nil {
			return nil, err
		}
		blogPosts = append(blogPosts, *bp)
	}

	return blogPosts, nil
}

func (p *PostgresStore) Create(blogpost BlogPost) error {

	rows, err := p.DB.Query("insert into blog values ($1,$2,$3,$4,$5,$6);", blogpost.ID, blogpost.Title, blogpost.Content, blogpost.Name, blogpost.CreatedAt, blogpost.UpdatedAt)
	if err != nil {
		return err
	}

	fmt.Println(rows, err)
	return nil
}

func (p *PostgresStore) Delete(id uuid.UUID) error {

	rows, err := p.DB.Query("DELETE FROM blog WHERE blog_id = $1;", id)
	if err != nil {
		return err
	}

	fmt.Println(rows, err)
	return nil
}

func (p *PostgresStore) Update(blogpost BlogPost) error {
	rows, err := p.DB.Query("UPDATE blog SET blog_title = $1, blog_post = $2, updated_at = $3 WHERE blog_id = $4;", blogpost.Title, blogpost.Content, blogpost.UpdatedAt, blogpost.ID)
	if err != nil {
		return err
	}
	fmt.Println(rows, err)
	return nil
}

func (p *PostgresStore) GetByID(id uuid.UUID) (BlogPost, error) {

	bp := NewBlogPost()

	err := p.DB.QueryRow("SELECT * FROM blog WHERE blog_id = $1;", id).
		Scan(&bp.ID, &bp.Title, &bp.Content, &bp.Name, &bp.CreatedAt, &bp.UpdatedAt)
	if err != nil {
		return BlogPost{}, err
	}

	return *bp, nil
}

func (p *PostgresStore) GetByName(name string) (BlogPost, error) {

	bp := NewBlogPost()

	if name == "" {
		return BlogPost{}, fmt.Errorf("name is empty")
	}

	err := p.DB.QueryRow("SELECT * FROM blog WHERE blog_name = $1;", name).
		Scan(&bp.ID, &bp.Title, &bp.Content, &bp.Name, &bp.CreatedAt, &bp.UpdatedAt)
	if err != nil {
		return BlogPost{}, err
	}

	return *bp, nil
}

func (p *PostgresStore) FetchLast10BlogPosts() ([]BlogPost, error) {

	blogPosts := []BlogPost{}

	rows, err := p.DB.Query("SELECT * FROM blog LIMIT 10;")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		bp := NewBlogPost()
		err := rows.Scan(&bp.ID, &bp.Title, &bp.Content, &bp.Name, &bp.CreatedAt, &bp.UpdatedAt)
		if err != nil {
			return nil, err
		}
		blogPosts = append(blogPosts, *bp)
	}

	return blogPosts, nil
}
