package main

import (
	"fmt"
	"log"
	"net"

	pastry2 "github.com/uhthomas/pastry/pkg/pastry"
)

func main() {
	l, err := net.Listen("tcp", ":2000")
	if err != nil {
		log.Fatal(err)
	}
	n, err := pastry2.NewNode(nil)
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
