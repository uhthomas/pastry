package pastry

import (
	"github.com/lucas-clemente/quic-go"
	"golang.org/x/crypto/ed25519"
)

type Peer struct {
	PublicKey ed25519.PublicKey
	Node      *Node
	quic.Session
}
