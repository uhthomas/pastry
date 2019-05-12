package main

import (
	"fmt"
	"log"
	"net"

	"github.com/uhthomas/pastry"
)

func main() {
	l, err := net.Listen("tcp", ":2000")
	if err != nil {
		log.Fatal(err)
	}
	n, err := pastry.NewNode(nil)
	if err != nil {
		log.Fatal(err)
	}
	// if you want to edit messages before they're forwarded
	n.Forward = func(key, b, next []byte) {}
	go n.Accept(l)
	for {
		m, err := n.Next()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(m)
	}
}
