package pastry

import (
	"context"
	"io"
)

// Deliverer will be called when the node is the closest the message can be routed to.
type Deliverer interface {
	Deliver(ctx context.Context, key []byte, r io.Reader) error
}

// DelivererFunc is an adapter to allow the use of ordinary functions as Deliverers.
type DelivererFunc func(ctx context.Context, key []byte, r io.Reader) error

// Deliver calls f(key, b)
func (f DelivererFunc) Deliver(ctx context.Context, key []byte, r io.Reader) error {
	return f(ctx, key, r)
}
