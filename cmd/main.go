package main

import (
	"fmt"
	"log"
	"microblog"
	"net"
)

func main() {

	m := microblog.MapPostStore{}
	m.Post = map[string]string{}

	netListener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	netListener.Close()

	err = microblog.ListenAndServe(netListener, m)
	if err != nil {
		fmt.Println(err)
	}

}
