package repository

import (
	"database/sql"
	"fmt"
	"log"
	"microblog/internal/models"
	"os"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type PostgresStore struct {
	DB *sql.DB
}

func New() (*PostgresStore, error) {

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	password := os.Getenv("DB_PASSWORD")
	user := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")

	var psqlInfo string
	if os.Getenv("LOCAL") == "local" {
		psqlInfo = fmt.Sprintf("host=%s port=%s user=%s "+
			"password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbName)
	} else {
		psqlInfo = fmt.Sprintf("host=%s port=%s user=%s "+
			"password=%s dbname=%s sslmode=require options=databaseid%%3D%s",
			host, port, user, password, dbName, os.Getenv("DB_ID"))
	}

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	query, err := os.ReadFile("path/to/database.sql")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(string(query))
	if err != nil {
		return nil, err
	}

	log.Print("Successfully connected!")
	return &PostgresStore{DB: db}, nil
}

func (p *PostgresStore) GetAll() ([]*models.BlogPost, error) {

	blogPosts := []*models.BlogPost{}

	rows, err := p.DB.Query("SELECT * FROM blog;")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		bp := models.NewBlogPost()
		err := rows.Scan(&bp.ID, &bp.Title, &bp.Content, &bp.Name, &bp.FormattedDate, &bp.CreatedAt, &bp.UpdatedAt)
		if err != nil {
			return nil, err
		}
		blogPosts = append(blogPosts, bp)
	}

	return blogPosts, nil
}

func (p *PostgresStore) Create(blogpost *models.BlogPost) error {

	rows, err := p.DB.Query("insert into blog values ($1,$2,$3,$4,$5,$6,$7);", blogpost.ID, blogpost.Title, blogpost.Content, blogpost.Name, blogpost.FormattedDate, blogpost.CreatedAt, blogpost.UpdatedAt)
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

func (p *PostgresStore) Update(blogpost *models.BlogPost) error {
	rows, err := p.DB.Query("UPDATE blog SET blog_title = $1, blog_post = $2, updated_at = $3 WHERE blog_id = $4;", blogpost.Title, blogpost.Content, blogpost.UpdatedAt, blogpost.ID)
	if err != nil {
		return err
	}
	fmt.Println(rows, err)
	return nil
}

func (p *PostgresStore) GetByID(id uuid.UUID) (*models.BlogPost, error) {

	bp := models.NewBlogPost()

	err := p.DB.QueryRow("SELECT * FROM blog WHERE blog_id = $1;", id).
		Scan(&bp.ID, &bp.Title, &bp.Content, &bp.Name, &bp.FormattedDate, &bp.CreatedAt, &bp.UpdatedAt)
	if err != nil {
		return &models.BlogPost{}, err
	}

	return bp, nil
}

func (p *PostgresStore) GetByName(name string) (*models.BlogPost, error) {

	bp := models.NewBlogPost()

	if name == "" {
		return &models.BlogPost{}, fmt.Errorf("name is empty")
	}

	err := p.DB.QueryRow("SELECT * FROM blog WHERE blog_name = $1;", name).
		Scan(&bp.ID, &bp.Title, &bp.Content, &bp.Name, &bp.FormattedDate, &bp.CreatedAt, &bp.UpdatedAt)
	if err != nil {
		return &models.BlogPost{}, err
	}

	return bp, nil
}

func (p *PostgresStore) FetchLast10BlogPosts() ([]*models.BlogPost, error) {

	blogPosts := []*models.BlogPost{}

	rows, err := p.DB.Query("SELECT * FROM blog LIMIT 10;")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		bp := models.NewBlogPost()
		err := rows.Scan(&bp.ID, &bp.Title, &bp.Content, &bp.Name, &bp.FormattedDate, &bp.CreatedAt, &bp.UpdatedAt)
		if err != nil {
			return nil, err
		}
		blogPosts = append(blogPosts, bp)
	}

	return blogPosts, nil
}
