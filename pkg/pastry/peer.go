package pastry

import (
	"crypto/rand"
	"encoding/gob"
	"errors"
	"io"
	"net"
	"time"

	"golang.org/x/crypto/ed25519"
)

type Peer struct {
	PublicKey ed25519.PublicKey
	Node      *Node
	*gob.Encoder
	io.Closer
}

func NewPeer(addr string, n *Node) (*Peer, error) {
	conn, err := (&net.Dialer{KeepAlive: 30 * time.Second}).Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return NewPeerConn(conn, n)
}

func NewPeerConn(conn net.Conn, n *Node) (p *Peer, err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	// send our public key
	// read their public key
	if _, err := conn.Write(n.publicKey); err != nil {
		return nil, err
	}
	var k [ed25519.PublicKeySize]byte
	if _, err := io.ReadFull(conn, k[:]); err != nil {
		return nil, err
	}

	// send them a challenge
	// read their challenge
	// a - our challenge
	// b - their challenge and then our signature and then their signature
	var a, b [ed25519.SignatureSize]byte
	if _, err := io.ReadFull(rand.Reader, a[:]); err != nil {
		return nil, err
	}
	if _, err := conn.Write(a[:]); err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(conn, b[:]); err != nil {
		return nil, err
	}

	// send our signature
	// read their signature
	if _, err := conn.Write(ed25519.Sign(n.privateKey, b[:])); err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(conn, b[:]); err != nil {
		return nil, err
	}

	// verify
	if !ed25519.Verify(k[:], a[:], b[:]) {
		return nil, errors.New("invalid signature")
	}
	p = &Peer{ed25519.PublicKey(k[:]), n, gob.NewEncoder(conn), conn}
	go p.listen(conn)
	return p, nil
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
			go func() { p.Node.c <- m }()
		case m.Data == nil:
			// bootstrap
		default:
			go p.Node.Route(m.Key, m.Data)
		}
	}
}
