package pastry

import (
	ci "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/mux"
)

type Peer struct {
	PublicKey ci.PubKey
	Node      *Node
	mux.MuxedConn
	//quic.Session
}
