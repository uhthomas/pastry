package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	mplex "github.com/libp2p/go-libp2p-mplex"
	tptu "github.com/libp2p/go-libp2p-transport-upgrader"
	tcp "github.com/libp2p/go-tcp-transport"
	"github.com/multiformats/go-multiaddr"
	"github.com/uhthomas/pastry"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/sync/errgroup"
)

func main() {
	addr := flag.String("addr", "/ip4/0.0.0.0/tcp/2376", "The multiaddress address to listen on")
	maddr, err := multiaddr.NewMultiaddr(*addr)
	if err != nil {
		log.Fatal(err)
	}
	dial := flag.String("dial", "", "a comma separated list of multiaddrs to connect to")
	flag.Parse()

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
		pastry.Transport(tcp.NewTCPTransport(&tptu.Upgrader{
			Muxer: new(mplex.Transport),
		})),
		// Use a forwarding func to log forwarded requests or modify next
		pastry.Forward(pastry.ForwarderFunc(func(ctx context.Context, next, key []byte, r io.Reader) error {
			// forwarding to <next> with <key> and body <r>
			return nil
		})),
		// Handle received messages
		pastry.Deliver(pastry.DelivererFunc(func(ctx context.Context, key []byte, r io.Reader) error {
			// message <key> with body <r> delivered
			var s strings.Builder
			if _, err := io.Copy(&s, r); err != nil {
				return err
			}
			log.Printf(
				"Message %s delivered with body %s",
				base64.RawURLEncoding.EncodeToString(key),
				s.String(),
			)
			return nil
		})),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error { return n.ListenAndServe(ctx, maddr) })

	g.Go(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		select {
		case <-c:
			cancel()
			return n.Close()
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	go func() {
		defer cancel()
		for r := bufio.NewScanner(os.Stdin); r.Scan(); {
			b := r.Bytes()
			h := blake2b.Sum256(b)
			if err := n.Route(ctx, h[:], bytes.NewReader(b)); err != nil {
				log.Fatal(err)
			}
		}
	}()

	log.Printf("Listening on %s\n", *addr)

	if s := strings.Fields(strings.TrimSpace(*dial)); len(s) > 0 {
		log.Printf("Connecting to %d nodes\n", len(s))
		for _, addr := range s {
			log.Printf("Connecting to %s\n", addr)
			mmaddr, err := multiaddr.NewMultiaddr(addr)
			if err != nil {
				log.Fatal(err)
			}
			// this isn't going to work, it's really just a temporary thing
			if err := n.DialAndAccept(ctx, mmaddr); err != nil {
				log.Fatal(err)
			}
		}
	}

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
