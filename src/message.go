package src

import (
	"io"

	"github.com/oklog/ulid/v2"
)

// Message ...
type Message struct {
	ID   string
	Data []byte
}

// NewMessage ...
func NewMessage(rs RandomStringer, r io.Reader) (*Message, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:   rs.Generate(),
		Data: data,
	}, nil
}

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

const MessageIDSize = 26

// RandomStringer ...
type RandomStringer interface {
	Generate() string
}

var ULIDGenerator = ulidGenerator{}
var _ RandomStringer = ulidGenerator{}

// ULIDGenerator ...
type ulidGenerator struct{}

// Generate() ...
func (ulidGenerator) Generate() string {
	return ulid.Make().String()
}
