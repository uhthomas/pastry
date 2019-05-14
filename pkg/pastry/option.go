package pastry

import (
	"log"

	"golang.org/x/crypto/ed25519"
)

type Option func(*Node)

func Key(k ed25519.PrivateKey) Option {
	return func(n *Node) {
		n.privateKey = k
		n.publicKey = k.Public().(ed25519.PublicKey)
	}
}

func Forward(f Forwarder) Option  { return func(n *Node) { n.forwarder = f } }
func Deliver(d Deliverer) Option  { return func(n *Node) { n.deliverer = d } }
func Logger(l *log.Logger) Option { return func(n *Node) { n.logger = l } }
