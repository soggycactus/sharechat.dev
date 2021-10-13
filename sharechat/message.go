package sharechat

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type MessageRepository interface {
	InsertMessage(ctx context.Context, message Message) error
	GetMessagesByRoom(ctx context.Context, roomID string) (*[]Message, error)
}

type Message struct {
	ID      string      `json:"id"`
	RoomID  string      `json:"room_id"`
	Member  Member      `json:"member"`
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
	Sent    time.Time   `json:"sent"`
}

type MessageType string

const (
	Chat         MessageType = "chat"
	MemberJoined MessageType = "joined"
	MemberLeft   MessageType = "left"
	SendFailed   MessageType = "failed"
)

func NewChatMessage(member Member, message []byte) Message {
	return Message{
		ID:      uuid.New().String(),
		RoomID:  member.RoomID,
		Member:  member,
		Type:    Chat,
		Message: string(message),
		Sent:    time.Now(),
	}
}

func NewMemberJoinedMessage(member Member) Message {
	return Message{
		ID:      uuid.New().String(),
		RoomID:  member.RoomID,
		Member:  member,
		Type:    MemberJoined,
		Message: fmt.Sprintf("%s joined the room.", member.Name),
		Sent:    time.Now(),
	}
}

func NewMemberLeftMessage(member Member) Message {
	return Message{
		ID:      uuid.New().String(),
		RoomID:  member.RoomID,
		Member:  member,
		Type:    MemberLeft,
		Message: fmt.Sprintf("%s left the room.", member.Name),
		Sent:    time.Now(),
	}
}

func NewSendFailedMessage(member Member) Message {
	return Message{
		ID:      uuid.New().String(),
		RoomID:  member.RoomID,
		Member:  member,
		Type:    SendFailed,
		Message: "failed to send message",
		Sent:    time.Now(),
	}
}
