package pastry

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"os"

	ci "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/transport"
)

type Option func(*Node) error

func Transport(tpt transport.Transport) Option {
	return func(n *Node) error {
		n.transport = tpt
		return nil
	}
}

func Key(pk ci.PrivKey) Option {
	return func(n *Node) error {
		n.key = pk
		return nil
	}
}

func RandomKey() Option {
	return func(n *Node) error {
		pk, _, err := ci.GenerateKeyPair(ci.Ed25519, 256)
		if err != nil {
			return err
		}
		n.key = pk
		return nil
	}
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
	keyBytes, err := n.PublicKey().Bytes()
	if err != nil {
		return err
	}
	return Logger(
		log.New(os.Stdout, base64.RawURLEncoding.EncodeToString(keyBytes)+" ", log.Ldate|log.Ltime),
	)(n)
}
