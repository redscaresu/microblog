// go:build integration

package microblog_test

import (
	"database/sql"
	"fmt"
	"microblog"
	"os"
	"testing"

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
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
	return &microblog.PostgresStore{DB: db}
}
