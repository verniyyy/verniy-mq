package src

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

// MQManager ...
type MQManager interface {
	CreateQueue(userID, name string) error
	GetQueue(userID, name string) (MessageQueue, error)
	ListQueues(userID string) ([]MessageQueue, error)
	DeleteQueue(userID, name string) error
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
func (m *mqManager) CreateQueue(userID, name string) error {
	id, err := newQueueID(userID, name)
	if err != nil {
		return err
	}

	return m.mqList.Store(id, NewMessageQueue(name))
}

// GetQueue ...
func (m *mqManager) GetQueue(userID, name string) (MessageQueue, error) {
	id, err := newQueueID(userID, name)
	if err != nil {
		return nil, err
	}

	return m.mqList.Get(id)
}

// ListQueues ...
func (m *mqManager) ListQueues(userID string) ([]MessageQueue, error) {
	keys, values, err := m.mqList.GetAll()
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return make([]MessageQueue, 0), nil
	}

	arg := userID
	result := make([]MessageQueue, 0)
	for i, k := range keys {
		userID, name, err := decodeQueueID(k)
		if err != nil {
			return nil, err
		}
		if userID != arg {
			continue
		}
		if name != values[i].Name() {
			return nil, fmt.Errorf("invalid queue")
		}
		result = append(result, values[i])
	}

	return result, nil
}

// DeleteQueue ...
func (m *mqManager) DeleteQueue(userID, name string) error {
	id, err := newQueueID(userID, name)
	if err != nil {
		return err
	}

	return m.mqList.Delete(id)
}

// queueID ...
type queueID []byte

// newQueueID ...
func newQueueID(userID, name string) (queueID, error) {
	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(map[string]string{
		"userID": userID,
		"name":   name,
	}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// decodeQueueID ...
func decodeQueueID(buf []byte) (userID, name string, err error) {
	v := make(map[string]string)
	if err := gob.NewDecoder(bytes.NewBuffer(buf)).Decode(&v); err != nil {
		return "", "", err
	}

	return v["userID"], v["name"], nil
}
