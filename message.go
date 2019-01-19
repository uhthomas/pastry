package pastry

type Message struct {
	Key, Data []byte
}

func NewMessage(k, b []byte) *Message {
	return &Message{k, b}
}
