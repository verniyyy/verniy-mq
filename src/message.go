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

// Bytes ...
func (m Message) Bytes() []byte {
	const headerSize = MessageIDSize

	headerBuf := [headerSize]byte{}
	copy(headerBuf[0:], []byte(m.ID))
	if m.Data == nil {
		return headerBuf[:]
	}

	return append(headerBuf[:], m.Data...)
}

// MessageIDSize ...
const MessageIDSize = 26
