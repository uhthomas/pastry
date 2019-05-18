package pastry

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"log"

	"github.com/lucas-clemente/quic-go"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/sync/errgroup"
)

type Node struct {
	Leafset   LeafSet
	key       ed25519.PrivateKey
	forwarder Forwarder
	deliverer Deliverer
	logger    *log.Logger
}

func New(opts ...Option) (*Node, error) {
	n := new(Node)
	if err := n.Apply(append([]Option{
		DiscardLogger,
		RandomSeed,
	}, opts...)...); err != nil {
		return nil, err
	}
	n.Leafset = NewLeafSet(n)
	return n, nil
}

func (n *Node) Apply(opts ...Option) error {
	for _, opt := range opts {
		if err := opt(n); err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) PublicKey() ed25519.PublicKey { return n.key.Public().(ed25519.PublicKey) }

func (n *Node) ListenAndServe(ctx context.Context, address string) error {
	l, err := quic.ListenAddr(address, nil, nil)
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

func (n *Node) Serve(l quic.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		n.logger.Print("Accepting conn")
		go func() {
			stream, err := conn.AcceptStream()
			if err != nil {
				return
			}
			n.Accept(conn, stream)
		}()
	}
}

func (n *Node) DialAndAccept(address string) error {
	conn, err := quic.DialAddrContext(context.Background(), address, nil, nil)
	if err != nil {
		return err
	}
	stream, err := conn.OpenStream()
	if err != nil {
		return err
	}
	return n.Accept(conn, stream)
}

// Accept takes the session and a pre-opened stream since we need to do the initial handshake. The only way to do that
// agnostically is to have a pre-opened stream.
func (n *Node) Accept(conn quic.Session, stream quic.Stream) (err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()
	defer stream.Close()

	// send our public key
	// read their public key
	n.logger.Println("writing public key")
	if _, err := stream.Write(n.PublicKey()); err != nil {
		return err
	}
	n.logger.Println("receiving public key")
	var k [ed25519.PublicKeySize]byte
	if _, err := io.ReadFull(stream, k[:]); err != nil {
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
	if _, err := stream.Write(a[:]); err != nil {
		return err
	}
	n.logger.Println("receiving challenge")
	if _, err := io.ReadFull(stream, b[:]); err != nil {
		return err
	}

	// send our signature
	// read their signature
	n.logger.Println("sending our signature")
	if _, err := stream.Write(ed25519.Sign(n.key, b[:])); err != nil {
		return err
	}
	n.logger.Println("reading their signature")
	if _, err := io.ReadFull(stream, b[:]); err != nil {
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

	if ok := n.Leafset.Insert(k[:], conn); !ok {
		return errors.New("peer either already exists or does not fit in leafset")
	}

	go func() {
		defer conn.Close()
		defer n.Leafset.Remove(k[:])

		for {
			stream, err := conn.AcceptStream()
			if err != nil {
				return
			}
			go func() {
				defer stream.Close()
				var key [ed25519.PublicKeySize]byte
				if _, err := io.ReadFull(stream, key[:]); err != nil {
					return
				}
				n.Route(context.TODO(), key[:], stream)
			}()
		}
	}()

	return nil
}

// Send data to the node closest to the key.
func (n *Node) Route(ctx context.Context, key []byte, r io.Reader) error {
	n.logger.Printf("Routing %s\n", base64.RawURLEncoding.EncodeToString(key))

	p := n.Leafset.Closest(key)
	if p == nil {
		n.logger.Printf("Delivering %s\n", base64.RawURLEncoding.EncodeToString(key))
		if n.deliverer != nil {
			return n.deliverer.Deliver(ctx, key, r)
		}
		return nil
	}

	n.logger.Printf("Forwarding %s\n", base64.RawURLEncoding.EncodeToString(key))
	if n.forwarder != nil {
		if err := n.forwarder.Forward(ctx, p.PublicKey, key, r); err != nil {
			return err
		}
	}

	stream, err := p.OpenStream()
	if err != nil {
		return err
	}
	defer stream.Close()

	_, err = io.Copy(stream, io.MultiReader(bytes.NewReader(key), r))
	return err
}

func (n *Node) Close() error { return n.Leafset.Close() }
