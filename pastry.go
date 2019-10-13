package pastry

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/lucas-clemente/quic-go"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/sync/errgroup"
)

type InvalidSignatureError struct {
	PublicKey *[32]byte
	Signature [ed25519.SignatureSize]byte
}

func (err InvalidSignatureError) Error() string {
	return fmt.Sprintf(
		"invalid signature %s for public key %s",
		base64.RawURLEncoding.EncodeToString(err.Signature[:]),
		base64.RawURLEncoding.EncodeToString(err.PublicKey[:]),
	)
}

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

func (n *Node) PublicKey() ed25519.PublicKey {
	return n.key.Public().(ed25519.PublicKey)
}

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

func (n *Node) DialAndAccept(ctx context.Context, address string) error {
	conn, err := quic.DialAddrContext(ctx, address, nil, nil)
	if err != nil {
		return err
	}
	stream, err := conn.OpenStream()
	if err != nil {
		return err
	}
	return n.Accept(conn, stream)
}

// Accept takes the session and a pre-opened stream since we need to do the
// initial handshake. The only way to do that agnostically is to have a
// pre-opened stream.
func (n *Node) Accept(conn quic.Session, stream quic.Stream) (err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()
	defer stream.Close()

	publicKey, sharedKey, err := n.Handshake(stream, stream, rand.Reader)
	if err != nil {
		return err
	}

	_ = sharedKey

	p := &Peer{
		PublicKey: publicKey[:],
		Node:      n,
		Session:   conn,
	}

	if ok := n.Leafset.Insert(p); !ok {
		return errors.New("peer either already exists or does not fit in leafset")
	}

	go func() {
		defer conn.Close()
		defer n.Leafset.Remove(p)

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

// Handshake will generate an ephemeral key-pair and will then send our
// handshake to w. We will then read the peer's handshake from r, verifying the
// signature and then generating the shared secret.
//
// The handshake looks as follows:
// <-> [
//      [32] public key,
//      [32] ephemeral key public key,
//      (64) signature(private key, ephemeral public key),
// ]
func (n *Node) Handshake(
	w io.Writer,
	r, rand io.Reader,
) (
	publicKey [ed25519.PublicKeySize]byte,
	sharedKey [32]byte,
	err error,
) {
	pub, prv, err := box.GenerateKey(rand)
	if err != nil {
		return publicKey, sharedKey, err
	}

	b := make([]byte, ed25519.PublicKeySize+32+ed25519.SignatureSize)

	// Our public key
	nc := copy(b, n.PublicKey())

	// Our ephemeral public key
	nc += copy(b[nc:], pub[:])

	// The signature of our ephemeral public key
	copy(b[nc:], ed25519.Sign(n.key, pub[:]))

	if _, err := w.Write(b); err != nil {
		return publicKey, sharedKey, err
	}

	if _, err := io.ReadFull(r, b); err != nil {
		return publicKey, sharedKey, err
	}

	// their public key
	nc = copy(publicKey[:], b)

	// their ephemeral public key
	nc += copy(pub[:], b[nc:])

	// the signature of their ephemeral public key
	var sig [ed25519.SignatureSize]byte
	copy(sig[:], b[nc:])

	if !ed25519.Verify(
		// their public key
		publicKey[:],
		// their ephemeral public key
		pub[:],
		// the signature of their ephemeral public key
		sig[:],
	) {
		return publicKey, sharedKey, InvalidSignatureError{
			PublicKey: pub,
			Signature: sig,
		}
	}

	box.Precompute(
		// new shared ephemeral key
		&sharedKey,
		// their ephemeral public key
		pub,
		// our ephemeral private key
		prv,
	)

	return publicKey, sharedKey, nil
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
