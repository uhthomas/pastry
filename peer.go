package pastry

import (
	"crypto/ed25519"

	"github.com/lucas-clemente/quic-go"
)

type Peer struct {
	PublicKey ed25519.PublicKey
	Node      *Node
	quic.Session
}
