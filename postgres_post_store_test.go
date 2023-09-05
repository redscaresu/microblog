//go:build integration

package microblog_test

import (
	"database/sql"
	"fmt"
	"microblog"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDBCreate(t *testing.T) {

	want := microblog.NewBlogPost()
	want.ID = uuid.New()
	want.Title = "foo"
	want.Content = "foo"
	store := newTestDBConnection(t)

	err := store.Create(*want)
	require.NoError(t, err)

	got, err := store.Get(want.ID)
	require.NoError(t, err)

	if !cmp.Equal(want, &got) {
		t.Error(cmp.Diff(want, got))
	}
}

func newTestDBConnection(t *testing.T) *microblog.PostgresStore {

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
