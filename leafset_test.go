package pastry_test

import (
	"testing"
)

func TestLeafSet_Closest(t *testing.T) {
	t.Skip("needs to be refactored")
	/*	b64u := base64.RawURLEncoding.EncodeToString

		n, err := pastry.New(pastry.RandomKey())
		if err != nil {
			t.Fatal(err)
		}

		left, err := n.PublicKey().Bytes()
		if err != nil {
			t.Fatal(err)
		}
		right, err := n.PublicKey().Bytes()
		if err != nil {
			t.Fatal(err)
		}
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
			p := l.Closest(b)
			if p == nil {
				t.Fatal("p is nil")
			}
			if !bytes.Equal(p.PublicKey, right) {
				t.Fatalf("got %s, want %s", b64u(p.PublicKey), b64u(right))
			}
		})*/
}
