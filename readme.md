# Pastry [![GoDoc](https://godoc.org/github.com/uhthomas/pastry/pkg/pastry?status.svg)](https://godoc.org/github.com/uhthomas/pastry/pkg/pastry)

The [Pastry DHT](https://www.freepastry.org/) written in Go. Written specifically for [Pastry Search](https://github.com/uhthomas/pastrysearch).

# Status
Under development.

## Example
```go
package main

import (
	"context"
	"crypto/rand"
	"io"
	"log"
	
	"github.com/uhthomas/pastry/pkg/pastry"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/sync/errgroup"
)

func main() {
	// Generate key for node
	var seed [ed25519.SeedSize]byte
	if _, err := io.ReadFull(rand.Reader, seed[:]); err != nil {
		log.Fatal(err)
	}
	
	n, err := pastry.New(
		// Pass logger to node
		pastry.DebugLogger,
		// Pass ed25519 seed to node
		pastry.Seed(seed[:]),
		// Use a forwarding func to log forwarded requests or modify next
		pastry.Forward(pastry.ForwarderFunc(func(key, b, next []byte) {
			// message <key> with <b> is being forwarded to <next>
		})),
		// Handle received messages
		pastry.Deliver(pastry.DelivererFunc(func(key, b []byte) {
		
		})),
	)
	if err != nil {
		log.Fatal(err)
	}
	
	// Connect to another node -- bootstrap 
	go n.DialAndAccept("tcp", "localhost:1234")
	
	// Listen for other nodes
	if err := n.ListenAndServe(context.Background(), "tcp", "localhost"); err != nil {
		log.Fatal(err)
	}
}
```