package pastry

import (
	"bytes"
	"crypto/ed25519"
	"sort"

	"golang.org/x/sync/errgroup"
)

const leafs = ed25519.PublicKeySize * 8 * 2

type LeafSet struct {
	parent      *Node
	left, right []*Peer
}

func NewLeafSet(n *Node) LeafSet { return LeafSet{parent: n} }

// Closest will return the closest peer to the given key. If the key is equal to the parent then nil is returned.
func (l *LeafSet) Closest(k []byte) *Peer {
	if c := bytes.Compare(k, l.parent.PublicKey()); c < 0 {
		return l.closest(k, l.left)
	} else if c > 0 {
		return l.closest(k, l.right)
	}
	return nil
}

func (l *LeafSet) closest(k []byte, s []*Peer) *Peer {
	if s == nil {
		return nil
	}
	i := sort.Search(len(s), func(i int) bool { return bytes.Compare(s[i].PublicKey, k) >= 0 })
	if i >= len(s) {
		i = len(s) - 1
	}
	return s[i]
}

func (l *LeafSet) Insert(p *Peer) (ok bool) {
	if c := bytes.Compare(p.PublicKey, l.parent.PublicKey()); c < 0 {
		l.left, ok = l.insert(p, l.left)
	} else if c > 0 {
		l.right, ok = l.insert(p, l.right)
	}
	return ok
}

func (l *LeafSet) insert(p *Peer, s []*Peer) ([]*Peer, bool) {
	i := sort.Search(len(s), func(i int) bool { return bytes.Compare(s[i].PublicKey, p.PublicKey) >= 0 })
	if i >= leafs {
		return s, false
	}
	if i >= len(s) || !bytes.Equal(s[i].PublicKey, p.PublicKey) {
		s = append(s, nil)
		copy(s[i+1:], s[i:])
		s[i] = p
		if len(s) > leafs {
			// we don't want to block the insert and we don't care about
			// errors
			go s[len(s)-1].Close()
			s = s[:len(s)-1]
		}
	}
	return s, true
}

func (l *LeafSet) Remove(p *Peer) (ok bool) {
	if c := bytes.Compare(p.PublicKey, l.parent.PublicKey()); c < 0 {
		l.left, ok = l.remove(p, l.left)
	} else if c > 0 {
		l.right, ok = l.remove(p, l.right)
	}
	return ok
}

func (l *LeafSet) remove(p *Peer, s []*Peer) ([]*Peer, bool) {
	if i := sort.Search(len(s), func(i int) bool {
		return bytes.Compare(s[i].PublicKey, p.PublicKey) >= 0
	}); i < len(s) && bytes.Equal(s[i].PublicKey, p.PublicKey) {
		copy(s[i:], s[i+1:])
		s[len(s)-1] = nil
		s = s[:len(s)-1]
		return s, true
	}
	return s, false
}

func (l *LeafSet) Close() error {
	var g errgroup.Group
	for _, p := range l.left {
		l.Remove(p)
		g.Go(p.Close)
	}
	for _, p := range l.right {
		l.Remove(p)
		g.Go(p.Close)
	}
	return g.Wait()
}
