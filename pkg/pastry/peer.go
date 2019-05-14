package pastry

import (
	"encoding/gob"
	"io"
	"net"

	"golang.org/x/crypto/ed25519"
)

type Peer struct {
	PublicKey ed25519.PublicKey
	Node      *Node
	*gob.Encoder
	io.Closer
}

func (p *Peer) listen(conn net.Conn) {
	defer conn.Close()
	d, e := gob.NewDecoder(conn), gob.NewEncoder(conn)
	_ = e
	var m Message
	for {
		if err := d.Decode(&m); err != nil {
			return
		}
		switch {
		case m.Key == nil:
			// this is for us.
			if d := p.Node.deliverer; d != nil {
				go d.Deliver(m.Key, m.Data)
			}
		case m.Data == nil:
			// bootstrap
		default:
			go p.Node.Route(m.Key, m.Data)
		}
	}
}
