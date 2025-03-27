package microblog_test

import (
	"database/sql"
	"fmt"
	"log"
	"microblog"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {

	os.Setenv("LOCAL", "local")
	os.Setenv("HOST", "127.0.0.1")
	os.Setenv("PORT", "5438")
	os.Setenv("PASSWORD", "postgres")
	os.Setenv("DB_NAME", "postgres")
	os.Setenv("USER", "postgres")

	_, err := microblog.New()
	require.NoError(t, err)
}

func TestCreate(t *testing.T) {

	store := newTestDBConnection(t)
	now := time.Now()
	want := microblog.NewBlogPost()
	want.ID = uuid.New()
	want.Title = uuid.NewString()
	want.Content = uuid.NewString()
	want.CreatedAt = now
	want.UpdatedAt = now
	err := store.Create(*want)
	require.NoError(t, err)

	got, err := store.GetByID(want.ID)
	require.NoError(t, err)

	if !cmp.Equal(want, &got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestGetAll(t *testing.T) {
	store := newTestDBConnection(t)
	got, err := store.GetAll()
	require.NoError(t, err)

	if (len(got)) < 1 {
		t.Error(1, got)
	}
}

func TestGetError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer db.Close()

	store := &microblog.PostgresStore{DB: db}

	invalidID := uuid.New()

	mock.ExpectQuery("SELECT \\* FROM blog WHERE blog_id = (.+)").
		WithArgs(invalidID).
		WillReturnError(sql.ErrNoRows) // Simulating a "no rows found" error

	result, err := store.GetByID(invalidID)

	if err != sql.ErrNoRows {
		t.Errorf("Expected error: %v, got: %v", sql.ErrNoRows, err)
	}

	if result != (microblog.BlogPost{}) {
		t.Errorf("Expected empty result, got: %v", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestCreateError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer db.Close()

	store := &microblog.PostgresStore{DB: db}

	blogpost := microblog.BlogPost{
		ID:      uuid.New(),
		Title:   "Test Post",
		Content: "Test Content",
	}

	mock.ExpectQuery("insert into blog values (.+)").WillReturnError(sql.ErrTxDone)

	err = store.Create(blogpost)

	if err != sql.ErrTxDone {
		t.Errorf("Expected error: %v, got: %v", sql.ErrTxDone, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetAllError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer db.Close()

	store := &microblog.PostgresStore{DB: db}

	mock.ExpectQuery("SELECT \\* FROM blog;").
		WillReturnError(sql.ErrTxDone)

	result, err := store.GetAll()

	if err != sql.ErrTxDone {
		t.Errorf("Expected error: %v, got: %v", sql.ErrTxDone, err)
	}

	if result != nil {
		t.Errorf("Expected nil result, got: %v", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func newTestDBConnection(t *testing.T) *microblog.PostgresStore {
	t.Helper()
	port := "5438"
	host := "127.0.0.1"
	user := "postgres"
	password := "postgres"
	dbName := "postgres"

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		t.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		t.Fatal(err)
	}

	log.Print("Successfully connected!")
	return &microblog.PostgresStore{DB: db}
}
