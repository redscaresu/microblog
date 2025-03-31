package main

import (
	"fmt"
	"log"
	microblog "microblog/internal"
	"net"
	"os"
)

func main() {

	app := microblog.Application{}
	// connect to DB
	psStore, err := microblog.New()
	if err != nil {
		log.Fatal(err)
	}

	app.Poststore = psStore

	netListener, err := net.Listen("tcp", ":8080")
	addr := netListener.Addr().String()

	if err != nil {
		log.Fatal(err)
	}
	netListener.Close()

	app.Auth.Username = os.Getenv("AUTH_USERNAME")
	app.Auth.Password = os.Getenv("AUTH_PASSWORD")

	err = microblog.ListenAndServe(addr, app)
	if err != nil {
		fmt.Println(err)
	}
}
