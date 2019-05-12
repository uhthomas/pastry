package pastry

type Forwarder interface {
	Forward(key, b, next []byte)
}

type ForwarderFunc func(key, b, next []byte)

func (f ForwarderFunc) Forward(key, b, next []byte) { f(key, b, next) }
