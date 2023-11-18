package src

import (
	"context"
	"io"
)

// MessageQueueApplication ...
type MessageQueueApplication struct {
	mqManager MQManager
}

// NewMessageQueueApplication ...
func NewMessageQueueApplication() MessageQueueApplication {
	return MessageQueueApplication{
		mqManager: NewMQManager(),
	}
}

// CreateQueue ...
func (a MessageQueueApplication) CreateQueue(ctx context.Context, name string) error {
	return a.mqManager.CreateQueue("", name)
}

// DeleteQueue ...
func (a MessageQueueApplication) DeleteQueue(ctx context.Context, name string) error {
	return a.mqManager.DeleteQueue("", name)
}

// Publish ...
func (a MessageQueueApplication) Publish(ctx context.Context, name string, r io.Reader) error {
	m, err := NewMessage(ULIDGenerator, r)
	if err != nil {
		return err
	}

	mq, err := a.mqManager.GetQueue("", name)
	if err != nil {
		return err
	}

	return mq.Publish(m)
}

// Consume ...
func (a MessageQueueApplication) Consume(ctx context.Context, name string, r io.Reader) (*Message, error) {
	mq, err := a.mqManager.GetQueue("", name)
	if err != nil {
		return nil, err
	}

	return mq.Consume()
}

// Delete ...
func (a MessageQueueApplication) Delete(ctx context.Context, name string) error {
	id := ""
	mq, err := a.mqManager.GetQueue(id, name)
	if err != nil {
		return err
	}

	return mq.Delete(id)
}
