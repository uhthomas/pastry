package pastry

import (
	"crypto/rand"
	"encoding/gob"
	"errors"
	"io"
	"net"

	"golang.org/x/crypto/ed25519"
)

type Node struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
	Leafset    Leafset
	Forward    func(key, b, next []byte)
	c          chan Message
}

func NewNode(k ed25519.PrivateKey) (n *Node, err error) {
	if k == nil {
		_, k, err = ed25519.GenerateKey(nil)
		if err != nil {
			return nil, err
		}
	}
	l, err := net.Listen("tcp", ":9001")
	if err != nil {
		return nil, err
	}
	n = &Node{PrivateKey: k, PublicKey: k.Public().(ed25519.PublicKey)}
	go n.serve(l)
	return n, nil
}

func (n *Node) Accept(l net.Listener) error {
	for {
		conn, err := l.Accpet()
		if err != nil {
			return err
		}
		go n.Serve(conn)
	}
}

func (n *Node) Serve(conn net.Conn) error {
	defer conn.Close()
	// send our public key
	// read their public key
	if _, err := conn.Write(n.PublicKey); err != nil {
		return err
	}
	var k [ed25519.PublicKeySize]byte
	if _, err := io.ReadFull(conn, k[:]); err != nil {
		return err
	}

	// send them a challenge
	// read their challenge
	// a - our challenge
	// b - their challenge - then our signature - then their signature
	var a, b [ed25519.SignatureSize]byte
	if _, err := io.ReadFull(rand.Reader, a[:]); err != nil {
		return err
	}
	if _, err := conn.Write(a[:]); err != nil {
		return err
	}
	if _, err := io.ReadFull(conn, b[:]); err != nil {
		return err
	}

	// send our signature
	// read their signature
	if _, err := conn.Write(ed25519.Sign(n.PrivateKey, b[:])); err != nil {
		return err
	}
	if _, err := io.ReadFull(conn, b[:]); err != nil {
		return err
	}

	// verify
	if !ed25519.Verify(k[:], a[:], b[:]) {
		return errors.New("invalid signature")
	}

	k = ed25519.PublicKey(k[:])
	d, e := gob.NewDecoder(conn), gob.NewEncoder(conn)
	var m struct{ Key, Value []byte }
	for {
		if err := d.Decode(&m); err != nil {
			return err
		}
		switch {
		case m.Key == nil:
			go func() { n.c <- m }()
		case m.Value == nil:
			e.Encode(m)
		default:
			n.Route(m)
		}
	}
}

// Send data to the node closest to the key.
func (n *Node) Route(key, b []byte) {
	p := n.Leafset.Closest(key)
	if p == nil {
		go func() { n.c <- Message{key, b} }()
		return
	}
	if n.Forward != nil {
		n.Forward(key, b, p.PublicKey)
	}
	n.send(p.Encoder, Message{key, b})
}

// Send data directly to a node, bypassing routing.
func (n *Node) Send(to net.Addr, m Message) error {

	return nil
}

func (n *Node) send(e *gob.Encoder, m Message) error {

	return nil
}

// "Callbacks"
// not thread safe
// func (n *Node) Deliver(key, msg []byte) {

// }

// func (n *Node) Forward(key, next, msg []byte) {
// 	c := n.Leafset.Closest(next)
// }

// func (n *Node) NewLeafSet(leafset []byte) {}

func (n *Node) Next() (Message, error) {
	return <-n.c, nil
}
