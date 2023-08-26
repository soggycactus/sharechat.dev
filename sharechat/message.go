package sharechat

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type MessageRepository interface {
	InsertMessage(ctx context.Context, message Message) (*Message, error)
	GetMessagesByRoom(ctx context.Context, roomID string) (*[]Message, error)
}

type Message struct {
	ID       string `json:"id" db:"message_id"`
	RoomID   string `json:"roomId" db:"room_id"`
	MemberID string `json:"memberId" db:"member_id"`
	// We do not save MemberName in the database, but we return
	// it from the DB for ease of use on the client-side.
	MemberName string      `json:"memberName" db:"member_name"`
	Type       MessageType `json:"type" db:"type"`
	Message    string      `json:"message" db:"message"`
	// Sent is generated by the database
	Sent time.Time `json:"sent" db:"sent"`
}

type MessageType string

const (
	Chat         MessageType = "chat"
	MemberJoined MessageType = "joined"
	MemberLeft   MessageType = "left"
	SendFailed   MessageType = "failed"
)

// MarshalBinary implements encoding.BinaryMarshaler
// This is needed to publish Messages to the Queue
func (m Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

func NewChatMessage(member Member, message []byte) Message {
	return Message{
		ID:         uuid.New().String(),
		RoomID:     member.RoomID,
		MemberID:   member.ID,
		MemberName: member.Name,
		Type:       Chat,
		Message:    string(message),
	}
}

func NewMemberJoinedMessage(member Member) Message {
	return Message{
		ID:         uuid.New().String(),
		RoomID:     member.RoomID,
		MemberID:   member.ID,
		MemberName: member.Name,
		Type:       MemberJoined,
		Message:    fmt.Sprintf("%s joined the room.", member.Name),
	}
}

func NewMemberLeftMessage(member Member) Message {
	return Message{
		ID:         uuid.New().String(),
		RoomID:     member.RoomID,
		MemberID:   member.ID,
		MemberName: member.Name,
		Type:       MemberLeft,
		Message:    fmt.Sprintf("%s left the room.", member.Name),
	}
}

func NewSendFailedMessage(member Member) Message {
	return Message{
		ID:         uuid.New().String(),
		RoomID:     member.RoomID,
		MemberID:   member.ID,
		MemberName: member.Name,
		Type:       SendFailed,
		Message:    "failed to send message",
	}
}
