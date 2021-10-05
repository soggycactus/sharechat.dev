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
	// channels to shut down the goroutines
	stopListen    chan struct{}
	stopBroadcast chan struct{}
	conn          Connection
	// hooks for synchronizing goroutines during tests
	listenNoop   func()
	brodcastNoop func()
}

func NewMember(name string, room *Room, conn Connection) *Member {
	return &Member{
		ID:            uuid.New().String(),
		Name:          name,
		room:          room,
		inbound:       make(chan Message),
		stopListen:    make(chan struct{}),
		stopBroadcast: make(chan struct{}),
		conn:          conn,
		listenNoop:    func() {},
		brodcastNoop:  func() {},
	}
}

func (s *Member) RoomID() string {
	return s.room.ID
}

// Listen receives messages from the Room and forwards them to the websocket connection
func (s *Member) Listen() {
	log.Printf("listening for messages for %s", s.Name)
	for {
		s.listenNoop()
		select {
		case message := <-s.inbound:
			if err := s.conn.WriteMessage(message); err != nil {
				log.Printf("failed to write message to Member %s: %v", s.Name, err)
			}
		case <-s.stopListen:
			s.listenNoop()
			return
		}
		s.listenNoop()
	}
}

// Broadcast receives messages from the websocket connection and forwards them to the Room
func (s *Member) Broadcast() {
	for {
		select {
		case <-s.stopBroadcast:
			s.brodcastNoop()
			return
		default:
			bytes, err := s.conn.ReadBytes()
			if err != nil {
				log.Print("error")
				if err != ExpectedCloseError {
					log.Printf("failed to read websocket for member %s: %v", s.Name, err)
				}
				// s.room.Leave(*s)
				s.brodcastNoop()
				return
			}

			message := NewChatMessage(*s, bytes)
			log.Print("sending")
			s.room.Outbound(message)
			log.Print("sent")
			s.brodcastNoop()
		}
	}
}

func (m *Member) Inbound(message Message) {
	m.inbound <- message
}

func (m *Member) StopListen() {
	m.stopListen <- struct{}{}
}

func (m *Member) StopBroadcast() {
	m.stopBroadcast <- struct{}{}
}

func (m *Member) WithListenNoop(f func()) *Member {
	m.listenNoop = f
	return m
}

func (m *Member) WithBroadcastNoop(f func()) *Member {
	m.brodcastNoop = f
	return m
}
