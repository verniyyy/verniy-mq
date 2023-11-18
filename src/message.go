package src

import "github.com/oklog/ulid/v2"

// Message ...
type Message struct {
	ID   string
	Data []byte
}

// NewMessage ...
func NewMessage(rs RandomStringer, data []byte) *Message {
	return &Message{
		ID:   rs.Generate(),
		Data: data,
	}
}

// IsEmpty ...
func (m Message) IsEmpty() bool {
	return m.Data == nil
}

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
