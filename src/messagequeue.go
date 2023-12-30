package src

import (
	"log"
	"time"
)

// MessageQueue ...
type MessageQueue interface {
	Name() string
	Publish(*Message) error
	Consume() (*Message, error)
	Delete(id string) error
}

// NewMessageQueue ...
func NewMessageQueue(name string) MessageQueue {
	return &messageQueue{
		name:              name,
		returnToQueueTime: 1 * time.Minute,
		q:                 NewQueue[Message](),
		kv:                NewKVStore[string, Message](),
	}
}

// messageQueue ...
type messageQueue struct {
	name              string
	returnToQueueTime time.Duration
	q                 Queue[Message]
	kv                KVStore[string, Message]
}

// Name ...
func (mq *messageQueue) Name() string {
	return mq.name
}

// Publish ...
func (mq *messageQueue) Publish(m *Message) error {
	return mq.q.Enqueue(*m)
}

// Consume ...
func (mq *messageQueue) Consume() (*Message, error) {
	m, err := mq.q.Dequeue()
	if err != nil {
		return nil, err
	}

	if err := mq.kv.Store(m.ID, m); err != nil {
		return nil, err
	}

	time.AfterFunc(mq.returnToQueueTime, func() {
		if err := mq.makeAvailable(m.ID); err != nil {
			if err == ErrNotFound {
				return
			}
			panic(err)
		}
	})

	return &m, nil
}

// makeAvailable ...
func (mq *messageQueue) makeAvailable(id string) error {
	m, err := mq.kv.GetAndDelete(id)
	if err != nil {
		return err
	}
	defer func() {
		log.Printf("makeAvailable: %+v\n", id)
	}()

	return mq.q.Enqueue(m)
}

// Delete ...
func (mq *messageQueue) Delete(id string) error {
	_, err := mq.kv.GetAndDelete(id)
	return err
}
