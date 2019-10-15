package pastry

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/libp2p/go-libp2p-core/transport"
)

type Option func(*Node) error

func Transport(tpt transport.Transport) Option {
	return func(n *Node) error {
		n.transport = tpt
		return nil
	}
}

func Seed(seed []byte) Option {
	return func(n *Node) error {
		n.key = ed25519.NewKeyFromSeed(seed)
		return nil
	}
}

func RandomSeed(n *Node) error {
	var seed [ed25519.SeedSize]byte
	if _, err := io.ReadFull(rand.Reader, seed[:]); err != nil {
		return err
	}
	return Seed(seed[:])(n)
}

func Forward(f Forwarder) Option {
	return func(n *Node) error {
		n.forwarder = f
		return nil
	}
}

func Deliver(d Deliverer) Option {
	return func(n *Node) error {
		n.deliverer = d
		return nil
	}
}

func Logger(l *log.Logger) Option {
	return func(n *Node) error {
		n.logger = l
		return nil
	}
}

func DiscardLogger(n *Node) error {
	return Logger(log.New(ioutil.Discard, "", 0))(n)
}

func DebugLogger(n *Node) error {
	return Logger(
		log.New(os.Stdout, base64.RawURLEncoding.EncodeToString(n.PublicKey())+" ", log.Ldate|log.Ltime),
	)(n)
}
