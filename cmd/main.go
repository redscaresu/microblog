package main

import (
	"fmt"
	"log"
	"microblog"
	"net"
)

func main() {

	// connect to DB
	psStore := microblog.New()
	psStore.GetAll()

	netListener, err := net.Listen("tcp", ":8080")
	addr := netListener.Addr().String()

	if err != nil {
		log.Fatal(err)
	}
	netListener.Close()

	err = microblog.ListenAndServe(addr, psStore)
	if err != nil {
		fmt.Println(err)
	}
}
