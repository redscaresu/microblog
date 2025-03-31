package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"microblog/internal/models"
	"microblog/internal/repository"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestCreateWithContainer(t *testing.T) {
	store, cleanup := setupTestContainer(t)
	defer cleanup()

	now := time.Now().UTC().Format(time.RFC3339)
	nowTime, err := time.Parse(time.RFC3339, now)
	require.NoError(t, err)

	want := models.NewBlogPost()
	want.ID = uuid.New()
	want.Title = "Test Title"
	want.Content = "Test Content"
	want.CreatedAt = nowTime
	want.UpdatedAt = nowTime

	err = store.Create(want)
	require.NoError(t, err)

	got, err := store.GetByID(want.ID)
	require.NoError(t, err)

	got.UpdatedAt = got.UpdatedAt.UTC()
	got.CreatedAt = got.CreatedAt.UTC()

	assert.Equal(t, want, got)
}

func TestGetAll(t *testing.T) {
	store, cleanup := setupTestContainer(t)
	defer cleanup()

	now := time.Now().UTC().Format(time.RFC3339)
	nowTime, err := time.Parse(time.RFC3339, now)
	require.NoError(t, err)

	want1 := models.NewBlogPost()
	want1.ID = uuid.New()
	want1.Name = "Test Name 1"
	want1.Title = "Test Title 1"
	want1.Content = "Test Content 1"
	want1.CreatedAt = nowTime.UTC() // Ensure UTC
	want1.UpdatedAt = nowTime.UTC() // Ensure UTC

	want2 := models.NewBlogPost()
	want2.ID = uuid.New()
	want2.Name = "Test Name 2"
	want2.Title = "Test Title 2"
	want2.Content = "Test Content 2"
	want2.CreatedAt = nowTime.UTC() // Ensure UTC
	want2.UpdatedAt = nowTime.UTC() // Ensure UTC

	var wantSlice []*models.BlogPost
	wantSlice = append(wantSlice, want1, want2)

	err = store.Create(want1)
	require.NoError(t, err)

	err = store.Create(want2)
	require.NoError(t, err)

	got, err := store.GetAll()
	require.NoError(t, err)

	for i := range got {
		got[i].CreatedAt = got[i].CreatedAt.UTC()
		got[i].UpdatedAt = got[i].UpdatedAt.UTC()
	}

	assert.ElementsMatch(t, wantSlice, got)
	assert.Equal(t, 2, len(got))
}

func TestGetError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer db.Close()

	store := &repository.PostgresStore{DB: db}

	invalidID := uuid.New()

	mock.ExpectQuery("SELECT \\* FROM blog WHERE blog_id = (.+)").
		WithArgs(invalidID).
		WillReturnError(sql.ErrNoRows)

	result, err := store.GetByID(invalidID)
	assert.Empty(t, &result)
	assert.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
	assert.NoError(t, mock.ExpectationsWereMet(), "There were unfulfilled expectations: %s", err)
}

func TestCreateError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := &repository.PostgresStore{DB: db}

	blogpost := models.BlogPost{
		ID:      uuid.New(),
		Title:   "Test Post",
		Content: "Test Content",
	}

	mock.ExpectQuery("insert into blog values (.+)").WillReturnError(sql.ErrTxDone)
	err = store.Create(&blogpost)
	assert.ErrorIs(t, err, sql.ErrTxDone)
	assert.NoError(t, mock.ExpectationsWereMet(), "There were unfulfilled expectations: %s", err)

}

func TestGetAllError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()

	store := &repository.PostgresStore{DB: db}

	mock.ExpectQuery("SELECT \\* FROM blog;").
		WillReturnError(sql.ErrTxDone)

	result, err := store.GetAll()
	assert.Nil(t, result)
	assert.ErrorIs(t, err, sql.ErrTxDone)
	assert.NoError(t, mock.ExpectationsWereMet(), "There were unfulfilled expectations: %s", err)

}

func setupTestContainer(t *testing.T) (*repository.PostgresStore, func()) {
	t.Helper()

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
	require.NoError(t, err, "failed to start container")

	host, err := container.Host(ctx)
	require.NoError(t, err, "failed to get container host and port")

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err, "error mapping port")

	dsn := fmt.Sprintf("host=%s port=%s user=postgres password=postgres dbname=testdb sslmode=disable", host, port.Port())

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err, "failed to connect to database")

	err = db.Ping()
	require.NoError(t, err, "failed to ping database")

	log.Println("PostgreSQL container is ready!")

	runSQLScript(t, db, "../../sql/create_tables.sql")

	return &repository.PostgresStore{DB: db}, func() {
		db.Close()
		container.Terminate(ctx)
	}
}

func runSQLScript(t *testing.T, db *sql.DB, scriptPath string) {
	t.Helper()

	script, err := os.ReadFile(scriptPath)
	require.NoError(t, err)

	_, err = db.Exec(string(script))
	require.NoError(t, err)

	log.Println("SQL script executed successfully!")
}
