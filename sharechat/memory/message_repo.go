package memory

import "github.com/soggycactus/sharechat.dev/sharechat"

type MessageRepo struct {
	Messages map[string]sharechat.Message
}

func NewMessageRepo() MessageRepo {
	return MessageRepo{
		Messages: make(map[string]sharechat.Message),
	}
}

func (m *MessageRepo) InsertMessage(message sharechat.Message) error {
	m.Messages[message.ID] = message
	return nil
}

func (m *MessageRepo) GetMessagesByRoom(roomID string) (*[]sharechat.Message, error) {
	messages := make([]sharechat.Message, 0, len(m.Messages))
	for _, message := range m.Messages {
		if message.RoomID == roomID {
			messages = append(messages, message)
		}
	}

	return &messages, nil
}
