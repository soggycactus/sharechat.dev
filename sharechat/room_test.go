//go:build unit || all

package sharechat_test

import (
	"context"
	"testing"
	"time"

	"github.com/soggycactus/sharechat.dev/sharechat"
	"github.com/stretchr/testify/assert"
)

func TestRoomOutbound(t *testing.T) {
	// doneCh will notify us when the subscribe has been received
	doneCh := make(chan sharechat.Message)

	// Create a test room with a hook that notifies doneCh when a subscribe is received
	room := sharechat.NewRoom("test").
		WithCallbackInbound(func(message *sharechat.Message) {
			doneCh <- *message
		}).
		WithNoErrorLogs()
	member := sharechat.NewMember("test", room.ID, nil)
	room = room.WithMembers(*member)

	ctx, fn := context.WithDeadline(context.Background(), time.Now().Add(1*time.Millisecond))
	defer fn()

	go room.Start(ctx)
	// wait for the room to be ready
	err := room.Ready(ctx)
	if err != nil {
		t.Fatalf("room is not ready: %v", err)
	}

	// send a message
	message := sharechat.NewChatMessage(*member, []byte("hello world!"))
	err = room.Inbound(context.Background(), message)
	if err != nil {
		t.Fatalf("could not send message before timeout: %v", err)
	}

	// receiving from doneCh blocks until the message is received
	result := <-doneCh
	// We don't compare the structs outright here because it results in a race condition
	// since both messages have the same memory reference
	assert.Equal(t, message.Message, result.Message, "message content should be equal")
	assert.Equal(t, message.Type, result.Type, "message type should be equal")
	assert.Equal(t, message.Sent, result.Sent, "message timestamp should be equal")
	assert.Equal(t, message.RoomID, result.RoomID, "message room ids should be equal")
	assert.Equal(t, message.MemberID, result.MemberID, "member IDs should be equal")
	assert.Equal(t, message.MemberName, result.MemberName, "member names should be equal")
}

func TestRoomLeave(t *testing.T) {
	// doneCh will notify us when the subscribe has been received
	doneCh := make(chan interface{})

	// Create a test room with a hook that notifies doneCh when a subscribe is received
	room := sharechat.NewRoom("test").
		WithCallbackInbound(func(message *sharechat.Message) {
			doneCh <- *message
		}).
		WithNoErrorLogs()

	member := sharechat.NewMember("test", room.ID, nil)
	room = room.WithMembers(*member)

	ctx, fn := context.WithDeadline(context.Background(), time.Now().Add(1*time.Millisecond))
	defer fn()

	go room.Start(ctx)
	// wait for the room to be ready
	err := room.Ready(ctx)
	if err != nil {
		t.Fatalf("room is not ready: %v", err)
	}

	// remove a member
	leaveMessage := sharechat.NewMemberLeftMessage(*member)
	err = room.Inbound(context.Background(), leaveMessage)
	if err != nil {
		t.Fatalf("could not send message before timeout: %v", err)
	}

	// receiving from doneCh blocks until the removal is complete
	<-doneCh
	// it is now safe to make our assertions
	assert.Empty(t, room.Members(), "room should have deleted member")
}
