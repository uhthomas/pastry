package pastry

import (
	"context"
	"io"
)

// Forwarder can modify the contents of redirect the message elsewhere or set next to nil to deliver to itself.
type Forwarder interface {
	Forward(ctx context.Context, next, key []byte, r io.Reader) error
}

// ForwarderFunc is an adapter to allow the use of ordinary functions as Forwarders.
type ForwarderFunc func(ctx context.Context, next, key []byte, r io.Reader) error

// Forward calls f(key, b, next)
func (f ForwarderFunc) Forward(ctx context.Context, next, key []byte, r io.Reader) error {
	return f(ctx, next, key, r)
}
