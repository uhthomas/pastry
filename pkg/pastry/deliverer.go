package pastry

// Deliverer will be called when the node is the closest the message can be routed to.
type Deliverer interface {
	Deliver(key, b []byte)
}

// DelivererFunc is an adapter to allow the use of ordinary functions as Deliverers.
type DelivererFunc func(key, b []byte)

// Deliver calls f(key, b)
func (f DelivererFunc) Deliver(key, b []byte) { f(key, b) }
