package src

// Message ...
type Message struct {
	ID   string
	Data []byte
}

// NewMessage ...
func NewMessage(rs RandomStringer, data []byte) (*Message, error) {
	return &Message{
		ID:   rs(),
		Data: data,
	}, nil
}

// RandomStringer ...
type RandomStringer func() string

// IsEmpty ...
func (m Message) IsEmpty() bool {
	return m.Data == nil
}

// Bytes ...
func (m Message) Bytes() []byte {
	const headerSize = MessageIDSize

	headerBuf := [headerSize]byte{}
	copy(headerBuf[0:], []byte(m.ID))

	return append(headerBuf[:], m.Data...)
}

// MessageIDSize ...
const MessageIDSize = 26
