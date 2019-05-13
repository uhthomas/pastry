package pastry

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/sync/errgroup"
)

type Node struct {
	Leafset    Leafset
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	forwarder  Forwarder
	deliverer  Deliverer
	conns      map[net.Conn]struct{}
}

func New(opts ...Option) (*Node, error) {
	n := &Node{conns: make(map[net.Conn]struct{})}
	n.Apply(opts...)
	if n.privateKey == nil {
		_, k, err := ed25519.GenerateKey(nil)
		if err != nil {
			return nil, err
		}
		Key(k)(n)
	}
	return n, nil
}

func (n *Node) Apply(opts ...Option) {
	for _, opt := range opts {
		opt(n)
	}
}

func (n *Node) ListenAndServe(ctx context.Context, network, address string) error {
	l, err := net.Listen(network, address)
	if err != nil {
		return err
	}
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return n.Serve(l) })
	g.Go(func() error {
		<-ctx.Done()
		if err := l.Close(); err != nil {
			return err
		}
		return ctx.Err()
	})
	return g.Wait()
}

func (n *Node) Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		fmt.Println("Got conn! Accepting")
		go n.Accept(conn)
	}
}

func (n *Node) DialAndAccept(network, address string) error {
	conn, err := net.Dial(network, address)
	if err != nil {
		return err
	}
	return n.Accept(conn)
}

func (n *Node) Accept(conn net.Conn) error {
	defer conn.Close()

	// send our public key
	// read their public key
	log.Println("writing public key")
	if _, err := conn.Write(n.publicKey); err != nil {
		return err
	}
	log.Println("receiving public key")
	var k [ed25519.PublicKeySize]byte
	if _, err := io.ReadFull(conn, k[:]); err != nil {
		return err
	}

	// send them a challenge
	// read their challenge
	// a - our challenge
	// b - their challenge - then our signature - then their signature
	log.Println("generating challenge")
	var a, b [ed25519.SignatureSize]byte
	if _, err := io.ReadFull(rand.Reader, a[:]); err != nil {
		return err
	}
	log.Println("writing challenge")
	if _, err := conn.Write(a[:]); err != nil {
		return err
	}
	log.Println("receiving challenge")
	if _, err := io.ReadFull(conn, b[:]); err != nil {
		return err
	}

	// send our signature
	// read their signature
	log.Println("sending our signature")
	if _, err := conn.Write(ed25519.Sign(n.privateKey, b[:])); err != nil {
		return err
	}
	log.Println("reading their signature")
	if _, err := io.ReadFull(conn, b[:]); err != nil {
		return err
	}

	// verify
	log.Println("verifying signature")
	if !ed25519.Verify(k[:], a[:], b[:]) {
		return errors.New("invalid signature")
	}
	log.Printf("signature verified! their public key is: %s\n", base64.RawURLEncoding.EncodeToString(k[:]))

	n.conns[conn] = struct{}{}

	// k = ed25519.PublicKey(k[:])
	d, e := gob.NewDecoder(conn), gob.NewEncoder(conn)
	var m struct{ Key, Value []byte }
	for {
		if err := d.Decode(&m); err != nil {
			return err
		}
		switch {
		case m.Key == nil:
			if n.deliverer != nil {
				go n.deliverer.Deliver(m.Key, m.Value)
			}
		case m.Value == nil:
			e.Encode(m)
		default:
			n.Route(m.Key, m.Value)
		}
	}
}

// Send data to the node closest to the key.
func (n *Node) Route(key, b []byte) {
	p := n.Leafset.Closest(key)
	if p == nil {
		if n.deliverer != nil {
			n.deliverer.Deliver(key, b)
		}
		return
	}
	if n.forwarder != nil {
		n.forwarder.Forward(key, b, p.PublicKey)
	}
	// n.send(p.Encoder, Message{key, b})
}

func (n *Node) Close() error {
	var g errgroup.Group
	for conn := range n.conns {
		g.Go(conn.Close)
	}
	return g.Wait()
}
