# Pastry
The [Pastry DHT](https://www.freepastry.org/) written in Go. Written specifically for [Pastry Search](https://github.com/uhthomas/pastrysearch).

# Status
Under development.

## Usage
```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/uhthomas/pastry/pkg/pastry"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/sync/errgroup"
)

func main() {
	// Generate key for node
	_, k, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatal(err)
	}
	n, err := pastry.New(
		// Pass private key to node
		pastry.Key(k),
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

	ctx, cancel := context.WithCancel(context.Background())

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error { return n.ListenAndServe(ctx, "tcp", "localhost") })

	g.Go(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		select {
		case <-c:
			cancel()
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
```