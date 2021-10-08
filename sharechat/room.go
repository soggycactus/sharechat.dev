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
	ID   string `json:"id"`
	Name string `json:"name"`
	// members holds the local Members of a room
	members map[string]*Member
	// inbound forwards Messages to members
	inbound chan Message
	// shutdown stops the Room
	shutdown chan struct{}
	// ready is used to communicate the Start goroutine is running
	ready chan struct{}
	// stopped is used to communicate the Start goroutine has finished
	stopped chan struct{}
	// sync.Once to close inbound
	closeInbound *sync.Once
	// callbackInbound is a hook for synchronizing goroutines in tests.
	callbackInbound func(*Message)
	// flag used in tests to turn off error logging
	logErrors bool
}

func NewRoom(name string) *Room {
	return &Room{
		ID:              uuid.New().String(),
		Name:            name,
		members:         make(map[string]*Member),
		inbound:         make(chan Message),
		shutdown:        make(chan struct{}),
		ready:           make(chan struct{}),
		stopped:         make(chan struct{}),
		closeInbound:    &sync.Once{},
		callbackInbound: func(*Message) {},
		logErrors:       true,
	}
}

func (r *Room) Members() map[string]*Member {
	return r.members
}

// Room constantly listens for messages and forwards them to existing members
func (r *Room) Start(ctx context.Context) {
	r.ready <- struct{}{}
	for {
		select {
		case message, ok := <-r.inbound:
			if !ok {
				return
			}
			switch message.Type {
			case Chat:
				for _, member := range r.members {
					r.notifyMember(ctx, message, *member)
				}
			case MemberJoined:
				r.members[message.Member.ID] = &message.Member
				for _, member := range r.members {
					r.notifyMember(ctx, message, *member)
				}
			case MemberLeft:
				delete(r.members, message.Member.ID)
				for _, member := range r.members {
					r.notifyMember(ctx, message, *member)
				}
			}
			r.callbackInbound(&message)
		case <-r.shutdown:
			for _, member := range r.members {
				member.stopListen <- struct{}{}
				member.stopBroadcast <- struct{}{}
			}
			defer func() { r.stopped <- struct{}{} }()
			return
		}
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
