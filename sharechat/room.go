package sharechat

import (
	"context"
	"log"
	"sync"

	"github.com/google/uuid"
)

type RoomRepository interface {
	InsertRoom(ctx context.Context, room *Room) error
	GetRoom(ctx context.Context, roomID string) (*Room, error)
	DeleteRoom(ctx context.Context, roomID string) error
}

type Room struct {
	ID   string `json:"room_id" db:"room_id"`
	Name string `json:"room_name" db:"room_name"`
	// inbound forwards Messages to members
	inbound chan Message
	// shutdown stops the Room
	shutdown chan struct{}
	// ready is used to communicate the Start goroutine is running
	ready chan struct{}
	// sync.Once to close inbound
	closeInbound *sync.Once
	// callbackInbound is a hook for synchronizing goroutines in tests.
	callbackInbound func(*Message)
	// flag used in tests to turn off error logging
	logErrors bool
	// sync.Mutex to make members thread-safe
	mu *sync.Mutex
	// members holds the local Members of a room
	members map[string]*Member
}

func NewRoom(name string) *Room {
	return &Room{
		ID:              uuid.New().String(),
		Name:            name,
		inbound:         make(chan Message),
		shutdown:        make(chan struct{}),
		ready:           make(chan struct{}),
		closeInbound:    new(sync.Once),
		callbackInbound: func(*Message) {},
		logErrors:       true,
		mu:              new(sync.Mutex),
		members:         make(map[string]*Member),
	}
}

// SetupRoom creates the necessary in-memory data structures needed to operate the room.
// This is called inside of Repositories that fetch the room from a database.
func SetupRoom(room *Room) {
	room.inbound = make(chan Message)
	room.shutdown = make(chan struct{})
	room.ready = make(chan struct{})
	room.closeInbound = new(sync.Once)
	room.callbackInbound = func(*Message) {}
	room.logErrors = true
	room.mu = new(sync.Mutex)
	room.members = make(map[string]*Member)
}

func (r *Room) AddMember(member *Member) {
	r.mu.Lock()
	r.members[member.ID] = member
	r.mu.Unlock()
}

func (r *Room) Members() map[string]*Member {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.members
}

// Room constantly listens for messages and forwards them to existing members
func (r *Room) Start(ctx context.Context) {
	r.ready <- struct{}{}
	for {
		message, ok := <-r.inbound
		if !ok {
			for _, member := range r.members {
				member.CloseInbound()
				member.StopBroadcast()
			}
			return
		}
		switch message.Type {
		case MemberLeft:
			delete(r.members, message.MemberID)
			for _, member := range r.members {
				r.notifyMember(ctx, message, *member)
			}
		default:
			for _, member := range r.members {
				r.notifyMember(ctx, message, *member)
			}
		}
		r.callbackInbound(&message)
	}
}

func (r *Room) Inbound(ctx context.Context, message Message) error {
	select {
	case r.inbound <- message:
		return nil
	case <-ctx.Done():
		return ErrSendTimedOut
	}
}

func (r *Room) CloseInbound() {
	r.closeInbound.Do(func() {
		close(r.inbound)
	})
}

func (r *Room) Ready(ctx context.Context) error {
	select {
	case <-r.ready:
		return nil
	case <-ctx.Done():
		return ErrRoomNotReady
	}
}

func (r *Room) notifyMember(ctx context.Context, message Message, member Member) {
	select {
	case member.inbound <- message:
	// if the send blocks, log it and move on
	case <-ctx.Done():
		if r.logErrors {
			log.Printf("failed to notify member %s of message %s: context deadline exceeded", member.ID, message.ID)
		}
	}
}

func (r *Room) WithMembers(member ...Member) *Room {
	for _, mem := range member {
		r.members[mem.ID] = &mem
	}
	return r
}

func (r *Room) WithCallbackInbound(f func(*Message)) *Room {
	r.callbackInbound = f
	return r
}

func (r *Room) WithNoErrorLogs() *Room {
	r.logErrors = false
	return r
}
