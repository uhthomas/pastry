package pastry

// Forwarder can modify the contents of redirect the message elsewhere or set next to nil to deliver to itself.
type Forwarder interface {
	Forward(key, b, next []byte)
}

// ForwarderFunc is an adapter to allow the use of ordinary functions as Forwarders.
type ForwarderFunc func(key, b, next []byte)

// Forward calls f(key, b, next)
func (f ForwarderFunc) Forward(key, b, next []byte) { f(key, b, next) }
