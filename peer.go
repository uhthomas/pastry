package pastry

import (
	"crypto/ed25519"

	"github.com/libp2p/go-libp2p-core/mux"
)

type Peer struct {
	PublicKey ed25519.PublicKey
	Node      *Node
	mux.MuxedConn
	//quic.Session
}
