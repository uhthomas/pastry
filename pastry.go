package pastry

import (
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
	go n.listen(l)
	return n, nil
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
	n.send(p, Message{key, b})
}

// Send data directly to a node, bypassing routing.
func (n *Node) Send(to net.Addr, m Message) error {

	return nil
}

func (n *Node) send(w io.Writer, m Message) error {

	return nil
}

func (n *Node) accept(conn net.Conn) (*Peer, error) {
	p, err := NewPeerConn(conn, n)
	if err != nil {
		return nil, err
	}
	return p, err
}

func (n *Node) listen(l net.Listener) error {
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go n.accept(conn)
	}
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
