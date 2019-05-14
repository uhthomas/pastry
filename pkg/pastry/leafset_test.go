package pastry_test

import (
	"bytes"
	"testing"

	"github.com/uhthomas/pastry/pkg/pastry"
	"golang.org/x/crypto/ed25519"
)

func TestLeafset_Closest(t *testing.T) {
	var (
		left   = make([]byte, ed25519.PublicKeySize)
		middle = make([]byte, ed25519.PublicKeySize)
		right  = make([]byte, ed25519.PublicKeySize)
	)
	left[0] = 1
	middle[0] = 2
	right[0] = 3

	n, err := pastry.New(pastry.Key(append(make([]byte, ed25519.SeedSize), middle...)))
	if err != nil {
		t.Fatal(err)
	}

	l := pastry.NewLeafset(n)
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
