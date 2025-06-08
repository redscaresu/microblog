package main

import (
	"errors"
	"fmt"
	"microblog/pkg/handlers"
	"microblog/pkg/models"
	"microblog/pkg/repository"
	"net"
	"net/http"
	"os"
	"sync"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {

	if os.Getenv("AUTH_USERNAME") == "" {
		return errors.New("please set AUTH_USERNAME")
	}

	if os.Getenv("AUTH_PASSWORD") == "" {
		return errors.New("please set AUTH_PASSWORD")
	}

	host := os.Getenv("DB_HOST")
	if host == "" {
		return errors.New("please set DB_HOST")
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		return errors.New("please set DB_PORT")
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		return errors.New("please set DB_PASSWORD")
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		return errors.New("please set DB_USER")
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		return errors.New("please set DB_NAME")
	}

	psqlInfo := repository.GeneratePSQL(host, port, password, user, dbName)

	psStore, err := repository.New(psqlInfo, "./sql/create_tables.sql")
	if err != nil {
		return fmt.Errorf("unable to connect to database due to error: %v", err)
	}

	cacheMu := sync.RWMutex{}
	app := handlers.NewApplication(os.Getenv("AUTH_USERNAME"),
		os.Getenv("AUTH_PASSWORD"),
		psStore,
		[]*models.BlogPost{},
		&cacheMu)

	netListener, err := net.Listen("tcp", ":8080")
	addr := netListener.Addr().String()
	if err != nil {
		return fmt.Errorf("unable to listen due to error: %v", err)
	}
	netListener.Close()

	serveMux := http.NewServeMux()

	err = handlers.RegisterRoutes(serveMux, addr, app)
	if err != nil {
		return fmt.Errorf("unable to register handlers due to error: %v", err)
	}

	return nil
}
