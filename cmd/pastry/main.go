package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/uhthomas/pastry/pkg/pastry"
)

func main() {
	l, err := net.Listen("tcp", ":2000")
	if err != nil {
		log.Fatal(err)
	}
	n, err := pastry.NewNode(
		pastry.Forward(pastry.ForwarderFunc(func(key, b, next []byte) {

		})),
	)
	if err != nil {
		log.Fatal(err)
	}
	go n.Accept(l)
	for {
		ctx := context.Background()
		m, err := n.Next(ctx)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(m)
	}
}
