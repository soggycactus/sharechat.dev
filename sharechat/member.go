package sharechat

import (
	"log"

	"github.com/google/uuid"
)

type MemberRepository interface {
	InsertMember(member Member) (*Message, error)
	GetMembersByRoom(roomID string) (*[]Member, error)
	DeleteMember(member Member) (*Message, error)
}

type Member struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Members are bound to exactly one Room, so embed it here
	room *Room
	// inbound receives Messages from Room
	inbound chan Message
	// disconnect relays a message from Broadcast to Listen
	// that the user has gone offline & closes the goroutine
	disconnect chan struct{}
	conn       Connection
	// testNoop is a hook for synchronizing goroutines in tests.
	// It is called during Listen
	testNoop func(interface{})
}

func NewMember(name string, room *Room, conn Connection) *Member {
	return &Member{
		ID:         uuid.New().String(),
		Name:       name,
		room:       room,
		inbound:    make(chan Message),
		disconnect: make(chan struct{}),
		conn:       conn,
		testNoop:   func(i interface{}) {},
	}
}

func (s *Member) RoomID() string {
	return s.room.ID
}

// Listen receives messages from the Room and forwards them to the websocket connection
func (s *Member) Listen() {
	log.Printf("listening for messages for %s", s.Name)
	s.room.Subscribe(*s)
	s.testNoop(nil)
	for {
		select {
		case message := <-s.inbound:
			if err := s.conn.WriteMessage(message); err != nil {
				log.Printf("failed to write message to Member %s: %v", s.Name, err)
			}
		case <-s.disconnect:
			return
		}
		s.testNoop(nil)
	}
}

// Broadcast receives messages from the websocket connection and forwards them to the Room
func (s *Member) Broadcast() {
	for {
		bytes, err := s.conn.ReadBytes()
		if err != nil {
			if err != ExpectedCloseError {
				log.Printf("failed to read websocket for member %s: %v", s.Name, err)
			}
			s.disconnect <- struct{}{}
			s.room.Leave(*s)
			s.testNoop(err)
			return
		}

		message := NewChatMessage(*s, bytes)
		s.room.Outbound(message)
		s.testNoop(message)
	}
}

func (m *Member) Inbound(message Message) {
	m.inbound <- message
}

func (m *Member) WithTestNoop(f func(interface{})) *Member {
	m.testNoop = f
	return m
}
