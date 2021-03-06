package memory

import (
	"context"
	"sync"
	"time"

	"github.com/soggycactus/sharechat.dev/sharechat"
)

type MessageRepo struct {
	mu       sync.Mutex
	Messages map[string]sharechat.Message
}

func NewMessageRepo() *MessageRepo {
	return &MessageRepo{
		Messages: make(map[string]sharechat.Message),
	}
}

func (m *MessageRepo) InsertMessage(ctx context.Context, message sharechat.Message) (*sharechat.Message, error) {
	m.mu.Lock()
	message.Sent = time.Now()
	m.Messages[message.ID] = message
	m.mu.Unlock()
	return &message, nil
}

func (m *MessageRepo) GetMessagesByRoom(ctx context.Context, roomID string) (*[]sharechat.Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	messages := make([]sharechat.Message, 0, len(m.Messages))
	for _, message := range m.Messages {
		if message.RoomID == roomID {
			messages = append(messages, message)
		}
	}

	return &messages, nil
}
