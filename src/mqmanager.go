package src

import (
	"bytes"
	"encoding/gob"
)

// MQManager ...
type MQManager interface {
	CreateQueue(id, name string) error
	GetQueue(id, name string) (MessageQueue, error)
	DeleteQueue(id, name string) error
}

// NewMQManager ...
func NewMQManager() MQManager {
	return &mqManager{}
}

// mqManager ...
type mqManager struct {
	mqList KVStore[[]byte, MessageQueue]
}

// CreateQueue ...
func (m *mqManager) CreateQueue(id, name string) error {
	hash, err := newQueueHash(id, name)
	if err != nil {
		return err
	}

	return m.mqList.Store(hash, NewMessageQueue())
}

// GetQueue ...
func (m *mqManager) GetQueue(id, name string) (MessageQueue, error) {
	hash, err := newQueueHash(id, name)
	if err != nil {
		return nil, err
	}

	return m.mqList.Get(hash)
}

// DeleteQueue ...
func (m *mqManager) DeleteQueue(id, name string) error {
	hash, err := newQueueHash(id, name)
	if err != nil {
		return err
	}

	return m.mqList.Delete(hash)
}

// queueHash ...
type queueHash []byte

// newQueueHash ...
func newQueueHash(id, name string) (queueHash, error) {
	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(map[string]string{
		"id":   id,
		"name": name,
	}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
