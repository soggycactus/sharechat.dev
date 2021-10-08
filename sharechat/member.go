package sharechat

import (
	"context"
	"log"
	"sync"

	"github.com/google/uuid"
)

type MemberRepository interface {
	InsertMember(member Member) (*Message, error)
	GetMembersByRoom(roomID string) (*[]Member, error)
	DeleteMember(member Member) (*Message, error)
}

type Member struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	RoomID string `json:"room_id"`
	conn   Connection
	// inbound receives Messages from Room
	inbound chan Message
	// outbound forwards Messages to the Controller
	outbound chan Message
	// channels to indicate goroutines are ready
	readyListen    chan struct{}
	readyBroadcast chan struct{}
	// startBroadcast allows the Member to begin sending messages
	startBroadcast chan struct{}
	// channels to shut down the goroutines
	stopListen    chan struct{}
	stopBroadcast chan struct{}
	// callbacks for synchronizing goroutines during tests
	callbackListen func()
	// sync.Once to safely close inbound
	closeInbound *sync.Once
	// sync.Once to safely close outbound
	closeOutbound *sync.Once
}

func NewMember(name string, roomID string, conn Connection) *Member {
	return &Member{
		ID:             uuid.New().String(),
		Name:           name,
		RoomID:         roomID,
		inbound:        make(chan Message),
		outbound:       make(chan Message),
		readyListen:    make(chan struct{}),
		readyBroadcast: make(chan struct{}),
		startBroadcast: make(chan struct{}),
		stopListen:     make(chan struct{}),
		stopBroadcast:  make(chan struct{}),
		conn:           conn,
		callbackListen: func() {},
		closeInbound:   &sync.Once{},
		closeOutbound:  &sync.Once{},
	}
}

// Listen receives messages from the Room and forwards them to the websocket connection
func (m *Member) Listen() {
	m.readyListen <- struct{}{}
	for {
		select {
		case message, ok := <-m.inbound:
			if !ok {
				return
			}
			if err := m.conn.WriteMessage(message); err != nil {
				log.Printf("failed to write message to Member %s: %v", m.Name, err)
			}
		case <-m.stopListen:
			m.callbackListen()
			return
		}
		m.callbackListen()
	}
}

// Broadcast receives messages from the websocket connection and forwards them to the Room
func (m *Member) Broadcast() {
	m.readyBroadcast <- struct{}{}
	<-m.startBroadcast
	for {
		select {
		case <-m.stopBroadcast:
			return
		default:
			bytes, err := m.conn.ReadBytes()
			if err != nil {
				if err != ErrExpectedClose {
					log.Printf("failed to read websocket for member %s: %v", m.Name, err)
				}
				m.outbound <- NewMemberLeftMessage(*m)
				return
			}

			message := NewChatMessage(*m, bytes)
			m.outbound <- message
		}
	}
}

func (m *Member) Inbound(ctx context.Context, message Message) error {
	select {
	case m.inbound <- message:
		return nil
	case <-ctx.Done():
		return ErrSendTimedOut
	}
}

func (m *Member) CloseInbound() {
	m.closeInbound.Do(func() {
		close(m.inbound)
	})
}

func (m *Member) Outbound() Message {
	return <-m.outbound
}

func (m *Member) CloseOutbound() {
	m.closeOutbound.Do(func() {
		close(m.outbound)
	})
}

func (m *Member) ListenReady(ctx context.Context) error {
	select {
	case <-m.readyListen:
		return nil
	case <-ctx.Done():
		return ErrNotListening
	}
}

func (m *Member) BroadcastReady(ctx context.Context) error {
	select {
	case <-m.readyBroadcast:
		return nil
	case <-ctx.Done():
		return ErrNotBroadcasting
	}
}

func (m *Member) StartBroadcast() {
	m.startBroadcast <- struct{}{}
}

func (m *Member) StopListen() {
	m.stopListen <- struct{}{}
}

func (m *Member) StopBroadcast() {
	m.stopBroadcast <- struct{}{}
}

func (m *Member) WithCallbackListen(f func()) *Member {
	m.callbackListen = f
	return m
}
