package microblog_test

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"microblog"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestListenAndServe_UsesGivenStore(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	blogPost := &microblog.BlogPost{ID: id, Title: "title", Content: "content"}
	store := &microblog.SlicePostStore{BlogPosts: []microblog.BlogPost{*blogPost}}

	addr := newTestServer(t, store)

	resp, err := http.Get("http://" + addr.String())
	if err != nil {
		t.Fatal(err)
	}

	read, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("test fail")
	}
	got := string(read)
	want := fmt.Sprintf("[{%s title content}]", id)

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}

}

func TestSubmitFormHandler(t *testing.T) {

	store := newTestDBConnection(t)

	addr := newTestServer(t, store)

	body := strings.NewReader("{\"title\":\"boo\",\"content\":\"foo\"}")

	endpoint := fmt.Sprintf("http://%v/submit", addr)
	req, err := http.NewRequest("POST", endpoint, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Authorization", "Basic "+basicAuth())

}

func basicAuth() string {
	auth := os.Getenv("AUTH_USERNAME") + ":" + os.Getenv(os.Getenv("AUTH_PASSWORD"))
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func TestIsAuthenticatedWhenCorrectPasswordProvidedReturnsTrue(t *testing.T) {

	t.Setenv(microblog.MicroblogToken, "password123")

	got := microblog.IsAuthenticated("password123")
	want := true

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestIsAuthenticatedReturnsFalseWhenIncorrectPasswordProvided(t *testing.T) {

	t.Setenv(microblog.MicroblogToken, "password123")

	got := microblog.IsAuthenticated("hotdog")
	want := false

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func newTestServer(t *testing.T, store microblog.PostStore) net.Addr {
	t.Helper()

	netListener, err := net.Listen("tcp", "127.0.0.1:")
	addr := netListener.Addr().String()

	if err != nil {
		t.Fatal(err)
	}
	netListener.Close()

	go func() {
		err := microblog.ListenAndServe(addr,
			microblog.Application{
				Auth: struct {
					Username string
					Password string
				}{
					Username: "foo",
					Password: "foo",
				},
				Poststore: store,
			},
		)
		if err != nil {
			panic(err)
		}
	}()

	resp, err := http.Get("http:" + addr)

	for err != nil {
		t.Log("retrying")
		resp, err = http.Get("http://" + addr)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatal(resp.StatusCode)
	}

	return netListener.Addr()
}

func newTestDBConnection(t *testing.T) *microblog.PostgresStore {

	port := "5438"
	host := "127.0.0.1"
	user := "postgres"
	password := "postgres"
	dbName := "postgres"

	var psqlInfo string

	psqlInfo = fmt.Sprintf("host=%s port=%s user=%s "+
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
