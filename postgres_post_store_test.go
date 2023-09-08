//go:build integration

package microblog_test

import (
	"database/sql"
	"fmt"
	"log"
	"microblog"
	"os"
	"testing"

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
	want := microblog.NewBlogPost()
	want.ID = uuid.New()
	want.Title = uuid.NewString()
	want.Content = uuid.NewString()

	err := store.Create(*want)
	require.NoError(t, err)

	got, err := store.Get(want.ID)
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

func TestCreateErrorCase(t *testing.T) {
	// Create a new instance of go-sqlmock and a database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer db.Close()

	// Create a PostgresStore instance with the mock database connection
	store := &microblog.PostgresStore{DB: db}

	// Define the input data for your test
	blogpost := microblog.BlogPost{
		ID:      uuid.New(),
		Title:   "Test Post",
		Content: "Test Content",
	}

	// Expect the Query method to be called with an error
	mock.ExpectQuery("insert into blog values (.+)").WillReturnError(sql.ErrTxDone)

	// Call the Create method
	err = store.Create(blogpost)

	// Check if the error is as expected
	if err != sql.ErrTxDone {
		t.Errorf("Expected error: %v, got: %v", sql.ErrTxDone, err)
	}

	// Ensure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
