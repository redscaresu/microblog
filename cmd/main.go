package main

import (
	"errors"
	"fmt"
	microblog "microblog/internal"
	"microblog/internal/repository"
	"net"
	"net/http"
	"os"
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

	psStore, err := repository.New()
	if err != nil {
		return fmt.Errorf("unable to connect to database due to error: %v", err)
	}

	app := microblog.NewApplication(os.Getenv("AUTH_USERNAME"),
		os.Getenv("AUTH_PASSWORD"),
		psStore)

	netListener, err := net.Listen("tcp", ":8080")
	addr := netListener.Addr().String()
	if err != nil {
		return fmt.Errorf("unable to listen due to error: %v", err)
	}
	netListener.Close()

	serveMux := http.NewServeMux()

	err = microblog.RegisterRoutes(serveMux, addr, app)
	if err != nil {
		return fmt.Errorf("unable to register handlers due to error: %v", err)
	}

	return nil
}
