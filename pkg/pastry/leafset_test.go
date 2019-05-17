package pastry_test

import (
	"bytes"
	"testing"

	"github.com/uhthomas/pastry/pkg/pastry"
	"golang.org/x/crypto/ed25519"
)

func TestLeafSet_Closest(t *testing.T) {
	n, err := pastry.New(pastry.Seed(make([]byte, ed25519.SeedSize)))
	if err != nil {
		t.Fatal(err)
	}

	left, right := n.PublicKey(), n.PublicKey()
	left[0]--
	right[0]++

	l := pastry.NewLeafSet(n)
	if ok := l.Insert(&pastry.Peer{PublicKey: left}); !ok {
		t.Fatal("could not insert left")
	}
	if ok := l.Insert(&pastry.Peer{PublicKey: right}); !ok {
		t.Fatal("could not insert right")
	}

	t.Run("should return closest peer on the left", func(t *testing.T) {
		if p := l.Closest(make([]byte, ed25519.PublicKeySize)); p == nil || !bytes.Equal(p.PublicKey, left) {
			t.Fatal("p is either nil or incorrect")
		}
	})

	t.Run("should return closest peer on the right", func(t *testing.T) {
		b := make([]byte, ed25519.PublicKeySize)
		b[0] = 4
		if p := l.Closest(b); p == nil || !bytes.Equal(p.PublicKey, right) {
			t.Fatal("p is either nil or incorrect")
		}
	})
}
