package memory

import (
	"context"
	"sort"
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

func (m *MessageRepo) GetMessages(ctx context.Context, options sharechat.GetMessageOptions) ([]sharechat.Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sortedMessages := []sharechat.Message{}
	for _, message := range m.Messages {
		sortedMessages = append(sortedMessages, message)
	}

	switch {
	case !options.Before.IsEmpty():
		sort.Slice(sortedMessages, func(i, j int) bool {
			return sortedMessages[j].Sent.Before(sortedMessages[i].Sent)
		})
	case !options.After.IsEmpty():
		sort.Slice(sortedMessages, func(i, j int) bool {
			return sortedMessages[i].Sent.Before(sortedMessages[j].Sent)
		})
	default:
		sort.Slice(sortedMessages, func(i, j int) bool {
			return sortedMessages[j].Sent.Before(sortedMessages[i].Sent)
		})
	}

	messages := []sharechat.Message{}
	for _, message := range sortedMessages {
		if options.Limit > 0 && len(messages) >= options.Limit {
			break
		}

		if message.ID == options.Before.ID || message.ID == options.After.ID {
			continue
		}

		if options.RoomID != "" && message.RoomID != options.RoomID {
			continue
		}

		if options.Before.ID != "" && message.Sent.After(options.Before.Sent) {
			continue
		}

		if options.After.ID != "" && message.Sent.Before(options.After.Sent) {
			continue
		}

		messages = append(messages, message)
	}

	return messages, nil
}
