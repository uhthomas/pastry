package pastry

type Handler interface {
	Handle(key, b []byte)
}

type HandlerFunc func(key, b []byte)

func (f HandlerFunc) Handle(key, b []byte) { f(key, b) }
