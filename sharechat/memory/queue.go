package memory

import (
	"context"
	"errors"

	"github.com/soggycactus/sharechat.dev/sharechat"
)

func NewQueue() *Queue {
	return &Queue{messages: make(chan sharechat.Message)}
}

type Queue struct {
	messages chan sharechat.Message
}

func (q *Queue) Ping(ctx context.Context) error {
	return nil
}

func (q *Queue) Publish(ctx context.Context, message sharechat.Message) error {
	select {
	case q.messages <- message:
		return nil
	case <-ctx.Done():
		return errors.New("context deadline exceeded")
	}
}

func (q *Queue) Subscribe(ctx context.Context, roomID string, controller chan sharechat.Message, done chan struct{}) {
	for {
		select {
		case <-done:
			return
		case message := <-q.messages:
			controller <- message
		}
	}
}
