package main

import (
	"context"
	"fmt"
	"log"

	"github.com/uhthomas/pastry/pkg/pastry"
	"golang.org/x/crypto/ed25519"
)

func main() {
	// Generate key for node
	_, k, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatal(err)
	}
	n, err := pastry.NewNode(
		// Pass private key to node
		pastry.Key(k),
		// Use a forwarding func to log forwarded requests or modify next
		pastry.Forward(pastry.ForwarderFunc(func(key, b, next []byte) {
			// message <key> with <b> is being forwarded to <next>
		})),
	)
	if err != nil {
		log.Fatal(err)
	}
	for {
		ctx := context.Background()
		m, err := n.Next(ctx)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(m)
	}
}
