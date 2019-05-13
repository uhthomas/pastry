package pastry

import (
	"context"
	"crypto/rand"
	"encoding/gob"
	"errors"
	"io"
	"net"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/sync/errgroup"
)

type Node struct {
	Leafset    Leafset
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	forwarder  Forwarder
	handler    Handler
	c          chan Message
}

func New(opts ...Option) (*Node, error) {
	n := new(Node)
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
		go n.Accept(conn)
	}
}

func (n *Node) Accept(conn net.Conn) error {
	defer conn.Close()

	// send our public key
	// read their public key
	if _, err := conn.Write(n.publicKey); err != nil {
		return err
	}
	var k [ed25519.PublicKeySize]byte
	if _, err := io.ReadFull(conn, k[:]); err != nil {
		return err
	}

	// send them a challenge
	// read their challenge
	// a - our challenge
	// b - their challenge - then our signature - then their signature
	var a, b [ed25519.SignatureSize]byte
	if _, err := io.ReadFull(rand.Reader, a[:]); err != nil {
		return err
	}
	if _, err := conn.Write(a[:]); err != nil {
		return err
	}
	if _, err := io.ReadFull(conn, b[:]); err != nil {
		return err
	}

	// send our signature
	// read their signature
	if _, err := conn.Write(ed25519.Sign(n.privateKey, b[:])); err != nil {
		return err
	}
	if _, err := io.ReadFull(conn, b[:]); err != nil {
		return err
	}

	// verify
	if !ed25519.Verify(k[:], a[:], b[:]) {
		return errors.New("invalid signature")
	}

	k = ed25519.PublicKey(k[:])
	d, e := gob.NewDecoder(conn), gob.NewEncoder(conn)
	var m struct{ Key, Value []byte }
	for {
		if err := d.Decode(&m); err != nil {
			return err
		}
		switch {
		case m.Key == nil:
			go func() { n.c <- m }()
		case m.Value == nil:
			e.Encode(m)
		default:
			n.Route(m)
		}
	}
}

// Send data to the node closest to the key.
func (n *Node) Route(key, b []byte) {
	p := n.Leafset.Closest(key)
	if p == nil {
		go func() { n.c <- Message{key, b} }()
		return
	}
	if n.forwarder != nil {
		n.forwarder.Forward(key, b, p.PublicKey)
	}
	n.send(p.Encoder, Message{key, b})
}

// Send data directly to a node, bypassing routing.
func (n *Node) Send(to net.Addr, m Message) error {

	return nil
}

func (n *Node) send(e *gob.Encoder, m Message) error {

	return nil
}

// "Callbacks"
// not thread safe
// func (n *Node) Deliver(key, msg []byte) {

// }

// func (n *Node) forward(key, next, msg []byte) {
// 	c := n.Leafset.Closest(next)
// }

// func (n *Node) NewLeafSet(leafset []byte) {}

func (n *Node) Next(ctx context.Context) (Message, error) {
	select {
	case m := <-n.c:
		return m, nil
	case <-ctx.Done():
		return Message{}, ctx.Err()
	}
}
