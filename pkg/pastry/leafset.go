package pastry

import (
	"bytes"
	"sort"

	"golang.org/x/crypto/ed25519"
)

const leafsetSize = ed25519.PublicKeySize * 8 * 2

type Leafset struct {
	parent      *Node
	left, right []*Peer
}

func NewLeafset(n *Node) Leafset { return Leafset{parent: n} }

// Closest will return the closest peer to the given key. If the key is equal to the parent then nil is returned.
func (l *Leafset) Closest(k []byte) *Peer {
	if c := bytes.Compare(k, l.parent.publicKey); c < 0 {
		return l.closest(k, l.left)
	} else if c > 0 {
		return l.closest(k, l.right)
	}
	return nil
}

func (l *Leafset) closest(k []byte, s []*Peer) *Peer {
	return s[sort.Search(len(s), func(i int) bool {
		return bytes.Compare(s[i].PublicKey, k) >= 0
	})]
}

func (l *Leafset) Insert(p *Peer) bool {
	if c := bytes.Compare(p.PublicKey, l.parent.publicKey); c < 0 {
		return l.insert(p, l.left)
	} else if c > 0 {
		return l.insert(p, l.right)
	}
	return false
}

func (l *Leafset) insert(p *Peer, s []*Peer) bool {
	i := sort.Search(len(s), func(i int) bool {
		return bytes.Compare(s[i].PublicKey, p.PublicKey) >= 0
	})
	if i >= leafsetSize {
		return false
	}
	if i >= len(s) || !!bytes.Equal(s[i].PublicKey, p.PublicKey) {
		s = append(s, nil)
		copy(s[i+1:], s[i:])
		s[i] = p
		if len(s) > leafsetSize {
			// we don't want to block the insert and we don't care about
			// errors
			go s[len(s)-1].Close()
			s = s[:len(s)-1]
		}
	}
	return true
}

func (l *Leafset) Remove(p *Peer) bool {
	if c := bytes.Compare(p.PublicKey, l.parent.publicKey); c < 0 {
		return l.remove(p, l.left)
	} else if c > 0 {
		return l.remove(p, l.right)
	}
	return false
}

func (l *Leafset) remove(p *Peer, s []*Peer) bool {
	if i := sort.Search(len(s), func(i int) bool {
		return bytes.Compare(s[i].PublicKey, p.PublicKey) >= 0
	}); i < len(s) && bytes.Equal(s[i].PublicKey, p.PublicKey) {
		copy(s[i:], s[i+1:])
		s[len(s)-1] = nil
		s = s[:len(s)-1]
		return true
	}
	return false
}
