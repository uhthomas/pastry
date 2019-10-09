package pastry

//
// import (
// 	"encoding/gob"
// 	"io"
// 	"net"
// 	"sync"
//
// 	"golang.org/x/crypto/ed25519"
// )
//
// type Peer interface {
// 	PublicKey() ed25519.PublicKey
// 	Encode(key, b []byte) error
// 	io.Closer
// }
//
// type peer struct {
// 	m         sync.Mutex
// 	conn      net.Conn
// 	Public ed25519.PublicKey
// 	enc       *gob.Encoder
// }
//
// func (p *peer) PublicKey() ed25519.PublicKey {
// 	return p.Public
// }
//
// func (p *peer) Encode(key, b []byte) error {
// 	p.m.Lock()
// 	defer p.m.Unlock()
// 	return p.enc.Encode(struct{ key, b []byte }{key, b})
// }
//
// func (p *peer) Close() error { return p.conn.Close() }
//
// func (p *peer) loop() {
// 	defer p.Close()
// 	d := gob.NewDecoder(p)
// 	for {
// 		var out struct{ key, b []byte }
// 		if err := d.Decode(&out); err != nil {
// 			return
// 		}
// 	}
// }
