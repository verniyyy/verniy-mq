package src

import "time"

// MessageQueue ...
type MessageQueue interface {
	Publish(*Message) error
	Consume() (*Message, error)
	Delete(id string) error
}

// NewMessageQueue ...
func NewMessageQueue() MessageQueue {
	return &messageQueue{
		returnToQueueTime: 1 * time.Minute,
		q:                 NewQueue[Message](),
		kv:                NewKVStore[string, Message](),
	}
}

// messageQueue ...
type messageQueue struct {
	returnToQueueTime time.Duration
	q                 Queue[Message]
	kv                KVStore[string, Message]
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
	if !m.IsEmpty() {
		return nil, nil
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

	return mq.q.Enqueue(m)
}

// Delete ...
func (mq *messageQueue) Delete(id string) error {
	return mq.kv.Delete(id)
}
