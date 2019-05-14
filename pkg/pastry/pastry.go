package pastry

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net"
	"time"

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
	logger     *log.Logger
}

func New(opts ...Option) (*Node, error) {
	n := &Node{conns: make(map[net.Conn]struct{})}
	n.Apply(opts...)
	if n.privateKey == nil && n.publicKey == nil {
		_, k, err := ed25519.GenerateKey(nil)
		if err != nil {
			return nil, err
		}
		Key(k)(n)
	}
	if n.logger == nil {
		n.logger = log.New(ioutil.Discard, "", 0)
	}
	n.Leafset = NewLeafset(n)
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
		n.logger.Print("Accepting conn")
		go n.Accept(conn)
	}
}

func (n *Node) DialAndAccept(network, address string) error {
	conn, err := (&net.Dialer{KeepAlive: 10 * time.Second}).Dial(network, address)
	if err != nil {
		return err
	}
	return n.Accept(conn)
}

func (n *Node) Accept(conn net.Conn) (err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	// send our public key
	// read their public key
	n.logger.Println("writing public key")
	if _, err := conn.Write(n.publicKey); err != nil {
		return err
	}
	n.logger.Println("receiving public key")
	var k [ed25519.PublicKeySize]byte
	if _, err := io.ReadFull(conn, k[:]); err != nil {
		return err
	}

	// send them a challenge
	// read their challenge
	// a - our challenge
	// b - their challenge - then after we've signed it, their signature
	n.logger.Println("generating challenge")
	var a, b [ed25519.SignatureSize]byte
	if _, err := io.ReadFull(rand.Reader, a[:]); err != nil {
		return err
	}
	n.logger.Println("writing challenge")
	if _, err := conn.Write(a[:]); err != nil {
		return err
	}
	n.logger.Println("receiving challenge")
	if _, err := io.ReadFull(conn, b[:]); err != nil {
		return err
	}

	// send our signature
	// read their signature
	n.logger.Println("sending our signature")
	if _, err := conn.Write(ed25519.Sign(n.privateKey, b[:])); err != nil {
		return err
	}
	n.logger.Println("reading their signature")
	if _, err := io.ReadFull(conn, b[:]); err != nil {
		return err
	}

	// verify
	n.logger.Println("verifying signature")
	if !ed25519.Verify(k[:], a[:], b[:]) {
		return errors.New("invalid signature")
	}
	n.logger.Printf(
		"signature %s verified! their public key is: %s\n",
		base64.RawURLEncoding.EncodeToString(b[:]),
		base64.RawURLEncoding.EncodeToString(k[:]),
	)

	p := n.newPeer(k[:], conn)

	if ok := n.Leafset.Insert(p); !ok {
		return errors.New("peer either already exists or does not fit in leafset")
	}

	go func() {
		defer p.Close()
		defer n.Leafset.Remove(p)

		d := gob.NewDecoder(conn)
		var m Message
		for {
			if err := d.Decode(&m); err != nil {
				n.logger.Printf("error: %s\n", err)
				return
			}
			switch {
			case m.Key == nil:
				if n.deliverer != nil {
					go n.deliverer.Deliver(m.Key, m.Data)
				}
			case m.Data == nil:
				if err := p.Encode(m); err != nil {
					return
				}
			default:
				go n.Route(m.Key, m.Data)
			}
		}
	}()

	return nil
}

// Send data to the node closest to the key.
func (n *Node) Route(key, b []byte) {
	n.logger.Printf("Routing %s\n", base64.RawURLEncoding.EncodeToString(key))
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
	p.Encode(Message{key, b})
}

func (n *Node) newPeer(k ed25519.PublicKey, conn net.Conn) *Peer {
	return &Peer{
		PublicKey: ed25519.PublicKey(k[:]),
		Node:      n,
		Encoder:   gob.NewEncoder(conn),
		Closer:    conn,
	}
}

func (n *Node) Close() error { return n.Leafset.Close() }
