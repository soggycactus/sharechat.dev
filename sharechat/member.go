package sharechat

import (
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MemberRepository interface {
	InsertMember(member Member) (*Message, error)
	GetMembersByRoom(roomID string) (*[]Member, error)
	DeleteMember(member Member) (*Message, error)
}

type Member struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	room *Room
	// inbound receives Messages from Room
	inbound chan Message
	// disconnect relays a message from Broadcast to Listen
	// that the user has gone offline & closes the goroutine
	disconnect chan disconnect
	conn       *websocket.Conn
}

type disconnect struct{}

func NewMember(name string, room Room, conn *websocket.Conn) *Member {
	return &Member{
		ID:         uuid.New().String(),
		Name:       name,
		room:       &room,
		inbound:    make(chan Message),
		disconnect: make(chan disconnect),
		conn:       conn,
	}
}

func (s *Member) RoomID() string {
	return s.room.ID
}

// Listen receives messages from the Room and forwards them to the websocket connection
func (s *Member) Listen() {
	log.Printf("listening for messages for %s", s.Name)
	s.room.subscribe <- *s
	for {
		select {
		case message := <-s.inbound:
			if err := s.conn.WriteJSON(message); err != nil {
				log.Printf("failed to write message to Member %s: %v", s.Name, err)
			}
		case <-s.disconnect:
			return
		}

	}
}

// Broadcast receives messages from the websocket connection and forwards them to the Room
func (s *Member) Broadcast() {
	for {
		_, bytes, err := s.conn.ReadMessage()
		if err != nil {
			var closeErr *websocket.CloseError
			if errors.As(err, &closeErr) {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("close error for member %s: %v", s.Name, closeErr)
				}
				s.disconnect <- disconnect{}
				s.room.leave <- *s
				return
			}

			log.Printf("failed to read websocket for member %s: %v", s.Name, err)
			continue
		}

		message := NewChatMessage(*s, bytes)
		s.room.outbound <- message
	}
}
