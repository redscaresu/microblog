package main

import (
	"fmt"
	microblog "microblog/internal"
	"microblog/internal/repository"
	"net"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	app := microblog.Application{}
	psStore, err := repository.New()
	if err != nil {
		return err
	}

	app.Poststore = psStore

	netListener, err := net.Listen("tcp", ":8080")
	addr := netListener.Addr().String()
	if err != nil {
		return err
	}

	netListener.Close()

	app.Auth.Username = os.Getenv("AUTH_USERNAME")
	app.Auth.Password = os.Getenv("AUTH_PASSWORD")

	err = microblog.ListenAndServe(addr, app)
	if err != nil {
		return err
	}

	return nil
}
