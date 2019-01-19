package pastry

import (
	"bytes"
	"sort"
)

type Leafset struct {
	parent      *Node
	size        int
	left, right []*Peer
}

func (l *Leafset) Closest(k []byte) *Peer {
	if bytes.Compare(k, l.parent.PublicKey) < 0 {
		return l.closest(k, l.left)
	} else if bytes.Compare(k, l.parent.PublicKey) > 0 {
		return l.closest(k, l.right)
	}
	return nil
}

// needs modification to determine closest(a, b, c)
// oh well lol
func (l *Leafset) closest(k []byte, s []*Peer) *Peer {
	return s[sort.Search(len(s), func(i int) bool {
		return bytes.Compare(s[i].PublicKey, k) >= 0
	})]
}

func (l *Leafset) Insert(p *Peer) bool {
	if bytes.Compare(p.PublicKey, l.parent.PublicKey) < 0 {
		return l.insert(p, l.left)
	} else if bytes.Compare(p.PublicKey, l.parent.PublicKey) > 0 {
		return l.insert(p, l.right)
	}
	return false
}

func (l *Leafset) insert(p *Peer, s []*Peer) bool {
	i := sort.Search(len(s), func(i int) bool {
		return bytes.Compare(s[i].PublicKey, p.PublicKey) >= 0
	})
	if i >= l.size {
		return false
	}
	if i < len(s) && bytes.Equal(s[i].PublicKey, p.PublicKey) {
		return true
	}
	s = append(s, nil)
	copy(s[i+1:], s[i:])
	s[i] = p
	if len(s) > l.size {
		// we don't want to block the insert and we don't care about
		// errors
		go s[len(s)-1].Close()
		s = s[:len(s)-1]
	}
	return true

	// for i := 0; i < len(s); i++ {
	// 	if bytes.Equal(s[i].PublicKey, n.ID) {
	// 		return true
	// 	}
	// 	if bytes.Compare(s[i].PublicKey, n.ID) < 0 {
	// 		s = append(s, nil)
	// 		copy(s[i+1:], s[i:])
	// 		s[i] = n
	// 		if len(s) > l.size {
	// 			// we don't want to black the insert and we don't care about
	// 			// errors
	// 			go s[len(s)-1].Close()
	// 			s = s[:len(s)-1]
	// 		}
	// 		return true
	// 	}
	// }
	// return false
}

func (l *Leafset) Remove(p *Peer) bool {
	if bytes.Compare(p.PublicKey, l.parent.PublicKey) < 0 {
		return l.remove(p, l.left)
	} else if bytes.Compare(p.PublicKey, l.parent.PublicKey) > 0 {
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
