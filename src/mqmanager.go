package src

import (
	"encoding/json"
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
	return &mqManager{
		mqList: NewKVStore[queueID, MessageQueue](),
	}
}

// mqManager ...
type mqManager struct {
	mqList KVStore[queueID, MessageQueue]
}

// CreateQueue ...
func (m *mqManager) CreateQueue(userID, name string) error {
	id := encodeQueueID(userID, name)

	_, err := m.mqList.Get(id)
	if err != nil && err != ErrNotFound {
		return err
	}
	if err == nil {
		return fmt.Errorf("queue name \"%s\" is already stored", name)
	}

	return m.mqList.Store(id, NewMessageQueue(name))
}

// GetQueue ...
func (m *mqManager) GetQueue(userID, name string) (MessageQueue, error) {
	id := encodeQueueID(userID, name)
	q, err := m.mqList.Get(id)
	if err != nil && err == ErrNotFound {
		return nil, fmt.Errorf("queue name \"%s\" is not found", name)
	}

	return q, nil
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
	id := encodeQueueID(userID, name)
	if _, err := m.mqList.Get(id); err != nil {
		return fmt.Errorf("queue name \"%s\" is not found", name)
	}

	return m.mqList.Delete(id)
}

// queueID ...
type queueID string

const queueIDEncodeFmt = "{\"userID\":\"%v\",\"name\":\"%v\"}"

// encodeQueueID ...
func encodeQueueID(userID, name string) queueID {
	return queueID(fmt.Sprintf(queueIDEncodeFmt, userID, name))
}

// decodeQueueID ...
func decodeQueueID(id queueID) (userID, name string, err error) {
	v := make(map[string]string)
	if err := json.Unmarshal([]byte(id), &v); err != nil {
		fmt.Printf("err: %v\n", err.Error())
		return "", "", err
	}

	return v["userID"], v["name"], nil
}
