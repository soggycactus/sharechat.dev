//go:build unit || all

package sharechat_test

import (
	"testing"

	"github.com/soggycactus/sharechat.dev/sharechat"
	"github.com/soggycactus/sharechat.dev/sharechat/memory"
	"github.com/stretchr/testify/assert"
)

func TestRoomOutbound(t *testing.T) {
	// doneCh will notify us when the subscribe has been received
	doneCh := make(chan struct{})

	messageRepo := memory.NewMessageRepo()
	memberRepo := memory.NewMemberRepo(&messageRepo)

	// Create a test room with a hook that notifies doneCh when a subscribe is received
	room := sharechat.NewRoom("test", &memberRepo, &messageRepo).
		WithTestNoop(func() {
			doneCh <- struct{}{}
		}).
		WithNoErrorLogs()
	member := sharechat.NewMember("test", room, nil)
	room = room.WithMembers(*member)

	go room.Start()
	message := sharechat.NewChatMessage(*member, []byte("hello world!"))
	room.Outbound(message)
	// receiving from doneCh blocks until the subscribe is complete
	<-doneCh
	// it is now safe to make our assertions
	assert.Equal(t, message, messageRepo.Messages[message.ID], "message should be recorded")
}

func TestRoomSubscribe(t *testing.T) {
	// doneCh will notify us when the subscribe has been received
	doneCh := make(chan struct{})

	messageRepo := memory.NewMessageRepo()
	memberRepo := memory.NewMemberRepo(&messageRepo)

	// Create a test room with a hook that notifies doneCh when a subscribe is received
	room := sharechat.NewRoom("test", &memberRepo, &messageRepo).
		WithTestNoop(func() {
			doneCh <- struct{}{}
		}).
		WithNoErrorLogs()

	go room.Start()
	member := sharechat.NewMember("test", room, nil)
	room.Subscribe(*member)
	// receiving from doneCh blocks until the subscribe is complete
	<-doneCh
	// it is now safe to make our assertions
	result := room.Members()[member.ID]
	assert.Equal(t, member, result, "member should be in room")
}

func TestRoomLeave(t *testing.T) {
	// doneCh will notify us when the subscribe has been received
	doneCh := make(chan struct{})

	messageRepo := memory.NewMessageRepo()
	memberRepo := memory.NewMemberRepo(&messageRepo)

	// Create a test room with a hook that notifies doneCh when a subscribe is received
	room := sharechat.NewRoom("test", &memberRepo, &messageRepo).
		WithTestNoop(func() {
			doneCh <- struct{}{}
		}).
		WithNoErrorLogs()
	member := sharechat.NewMember("test", room, nil)
	room = room.WithMembers(*member)

	go room.Start()
	room.Leave(*member)
	// receiving from doneCh blocks until the subscribe is complete
	<-doneCh
	// it is now safe to make our assertions
	assert.Empty(t, room.Members(), "room should have deleted member")
}
