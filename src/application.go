package src

import (
	"context"
	"encoding/json"

	"github.com/verniyyy/verniy-mq/src/util"
)

// MessageQueueApplication ...
type MessageQueueApplication struct {
	mqManager MQManager
}

// NewMessageQueueApplication ...
func NewMessageQueueApplication(mqm MQManager) MessageQueueApplication {
	return MessageQueueApplication{
		mqManager: mqm,
	}
}

// CreateQueue ...
func (a MessageQueueApplication) CreateQueue(ctx context.Context, userID, name string) error {
	return a.mqManager.CreateQueue(userID, name)
}

// ListQueues ...
func (a MessageQueueApplication) ListQueues(ctx context.Context, userID string) (ListQueuesOutput, error) {
	mqList, err := a.mqManager.ListQueues(userID)
	if err != nil {
		return ListQueuesOutput{}, err
	}

	out := ListQueuesOutput{
		Queues: make([]string, len(mqList)),
	}
	for i, q := range mqList {
		out.Queues[i] = q.Name()
	}

	return out, nil
}

type ListQueuesOutput struct {
	Queues []string `json:"queues"`
}

func (o ListQueuesOutput) EncodeJSON() ([]byte, error) {
	return json.Marshal(o)
}

// DeleteQueue ...
func (a MessageQueueApplication) DeleteQueue(ctx context.Context, userID, name string) error {
	return a.mqManager.DeleteQueue(userID, name)
}

// Publish ...
func (a MessageQueueApplication) Publish(ctx context.Context, userID, name string, data []byte) error {
	m, err := NewMessage(util.GenULID, data)
	if err != nil {
		return err
	}

	mq, err := a.mqManager.GetQueue(userID, name)
	if err != nil {
		return err
	}

	return mq.Publish(m)
}

// Consume ...
func (a MessageQueueApplication) Consume(ctx context.Context, userID, name string) (*Message, error) {
	mq, err := a.mqManager.GetQueue(userID, name)
	if err != nil {
		return nil, err
	}

	return mq.Consume()
}

// Delete ...
func (a MessageQueueApplication) Delete(ctx context.Context, userID, name, messageID string) error {
	mq, err := a.mqManager.GetQueue(userID, name)
	if err != nil {
		return err
	}

	return mq.Delete(messageID)
}
