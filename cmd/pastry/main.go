package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/uhthomas/pastry/pkg/pastry"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/sync/errgroup"
)

func main() {
	addr := flag.String("addr", ":2376", "The address to listen on")
	dial := flag.String("dial", "", "a comma separated list of address to connect to")
	flag.Parse()

	// Generate key for node
	_, k, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatal(err)
	}

	l := log.New(os.Stdout, base64.RawURLEncoding.EncodeToString(k.Public().(ed25519.PublicKey))+" ", log.Ldate|log.Ltime)

	n, err := pastry.New(
		pastry.Logger(l),
		// Pass private key to node
		pastry.Key(k),
		// Use a forwarding func to log forwarded requests or modify next
		pastry.Forward(pastry.ForwarderFunc(func(key, b, next []byte) {
			// message <key> with <b> is being forwarded to <next>
			l.Printf("%s forwarding\n", base64.RawURLEncoding.EncodeToString(key))
		})),
		// Handle received messages
		pastry.Deliver(pastry.DelivererFunc(func(key, b []byte) {
			l.Printf("%s delivered with content %s\n", base64.RawURLEncoding.EncodeToString(key), string(b))
		})),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error { return n.ListenAndServe(ctx, "tcp", *addr) })

	g.Go(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		select {
		case <-c:
			cancel()
			var g errgroup.Group
			g.Go(n.Close)
			g.Go(os.Stdin.Close)
			return g.Wait()
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	go func() {
		defer cancel()
		r := bufio.NewScanner(os.Stdin)
		for r.Scan() {
			n.Route([]byte("some-key"), r.Bytes())
		}
	}()

	l.Printf("Listening on %s\n", *addr)

	if s := strings.Fields(strings.TrimSpace(*dial)); len(s) > 0 {
		l.Printf("Connecting to %d nodes\n", len(s))
		for _, addr := range s {
			l.Printf("Connecting to %s\n", addr)
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				log.Fatal(err)
			}
			go n.Accept(conn)
		}
	}

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
