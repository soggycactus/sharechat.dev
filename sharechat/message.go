package sharechat

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type MessageRepository interface {
	InsertMessage(ctx context.Context, message Message) (*Message, error)
	GetMessages(ctx context.Context, options GetMessageOptions) (*[]Message, error)
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

// MarshalBinary implements encoding.BinaryMarshaler
// This is needed to publish Messages to the Queue
func (m Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

type MessageType string

const (
	Chat         MessageType = "chat"
	MemberJoined MessageType = "joined"
	MemberLeft   MessageType = "left"
	SendFailed   MessageType = "failed"
)

type MessageCursor struct {
	ID   string
	Sent time.Time
}

func (m *MessageCursor) IsEmpty() bool {
	return m.ID == "" && m.Sent.IsZero()
}

func (m *MessageCursor) Encode() string {
	if m.IsEmpty() {
		return ""
	}

	concat := fmt.Sprintf("%s,%s", m.ID, m.Sent.Format(time.RFC3339Nano))
	return base64.StdEncoding.EncodeToString([]byte(concat))
}

func (m *MessageCursor) DecodeFromString(s string) error {
	if s == "" {
		return nil
	}

	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}

	parts := strings.Split(string(decoded), ",")
	if len(parts) != 2 {
		return fmt.Errorf("invalid cursor format")
	}

	m.ID = parts[0]
	m.Sent, err = time.Parse(time.RFC3339Nano, parts[1])
	if err != nil {
		return err
	}

	return nil
}

type GetMessageOptions struct {
	Limit  int
	RoomID string
	After  MessageCursor
	Before MessageCursor
}

var ErrInvalidOptions = errors.New("invalid options:")

func (g *GetMessageOptions) Validate() error {
	if g.Limit < 0 {
		return errors.Join(ErrInvalidOptions, errors.New("limit must be greater than 0"))
	}

	if !g.After.IsEmpty() && !g.Before.IsEmpty() {
		return errors.Join(ErrInvalidOptions, errors.New("cannot specify both after and before"))
	}

	return nil
}

type GetMessagesResult struct {
	Messages []Message
	Next     MessageCursor
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
