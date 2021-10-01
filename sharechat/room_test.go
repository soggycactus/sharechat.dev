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
	// We don't compare the structs outright here because it results in a race condition
	// since both messages have the same memory reference
	result := messageRepo.Messages[message.ID]
	assert.Equal(t, message.Message, result.Message, "message content should be equal")
	assert.Equal(t, message.Type, result.Type, "message type should be equal")
	assert.Equal(t, message.Sent, result.Sent, "message timestamp should be equal")
	assert.Equal(t, message.RoomID, result.RoomID, "message room ids should be equal")
	assert.Equal(t, message.Member.ID, result.Member.ID, "member IDs should be equal")
	assert.Equal(t, message.Member.Name, result.Member.Name, "member names should be equal")
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
	assert.Equal(t, member.Name, result.Name, "member names should be equal")
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
