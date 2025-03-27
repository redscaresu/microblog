package microblog_test

import (
	"context"
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
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gotest.tools/assert"
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

func TestCreateWithContainer(t *testing.T) {
	store, cleanup := setupTestContainer(t)
	defer cleanup()

	now := time.Now()
	want := microblog.NewBlogPost()
	want.ID = uuid.New()
	want.Title = "Test Title"
	want.Content = "Test Content"
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
	store, cleanup := setupTestContainer(t)
	defer cleanup()

	now := time.Now()
	want1 := microblog.NewBlogPost()
	want1.ID = uuid.New()
	want1.Title = "Test Title 1"
	want1.Content = "Test Content 1"
	want1.CreatedAt = now
	want1.UpdatedAt = now

	want2 := microblog.NewBlogPost()
	want2.ID = uuid.New()
	want2.Title = "Test Title 1"
	want2.Content = "Test Content 1"
	want2.CreatedAt = now
	want2.UpdatedAt = now

	err := store.Create(*want1)
	require.NoError(t, err)

	err = store.Create(*want2)
	require.NoError(t, err)

	got, err := store.GetAll()
	require.NoError(t, err)

	assert.Equal(t, 2, len(got))
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

func setupTestContainer(t *testing.T) (*microblog.PostgresStore, func()) {
	t.Helper()

	// Create a PostgreSQL container
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15", // Use the desired PostgreSQL version
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	// Get the container's host and port
	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Failed to get container port: %v", err)
	}

	// Build the PostgreSQL connection string
	dsn := fmt.Sprintf("host=%s port=%s user=postgres password=postgres dbname=testdb sslmode=disable", host, port.Port())

	// Connect to the database
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Ping the database to ensure it's ready
	err = db.Ping()
	if err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("PostgreSQL container is ready!")

	// Run the SQL script to create tables
	runSQLScript(t, db, "sql/create_tables.sql")

	// Return the PostgresStore and a cleanup function
	return &microblog.PostgresStore{DB: db}, func() {
		db.Close()
		container.Terminate(ctx)
	}
}

func runSQLScript(t *testing.T, db *sql.DB, scriptPath string) {
	t.Helper()

	// Read the SQL script file
	script, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("Failed to read SQL script: %v", err)
	}

	// Execute the SQL script
	_, err = db.Exec(string(script))
	if err != nil {
		t.Fatalf("Failed to execute SQL script: %v", err)
	}

	log.Println("SQL script executed successfully!")
}
