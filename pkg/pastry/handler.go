package pastry

type Deliverer interface {
	Deliver(key, b []byte)
}

type DelivererFunc func(key, b []byte)

func (f DelivererFunc) Deliver(key, b []byte) { f(key, b) }
