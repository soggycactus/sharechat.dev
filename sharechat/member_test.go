//go:build unit || all

package sharechat_test

import (
	"testing"

	"github.com/soggycactus/sharechat.dev/sharechat"
	"github.com/soggycactus/sharechat.dev/sharechat/memory"
	"github.com/soggycactus/sharechat.dev/sharechat/mock"
	"github.com/stretchr/testify/assert"
)

func TestListen(t *testing.T) {
	// channels will notify us when the subscribe has been received
	memberDone := make(chan struct{})
	roomDone := make(chan struct{})

	messageRepo := memory.NewMessageRepo()
	memberRepo := memory.NewMemberRepo(&messageRepo)

	// Create a test room
	room := sharechat.NewRoom("test", &memberRepo, &messageRepo).
		WithTestNoop(func() {
			roomDone <- struct{}{}
		})
		//WithNoErrorLogs()

	connection := mock.NewConnection().WithWriteMessageResult(nil)
	member := sharechat.NewMember("test", room, connection).
		WithListenNoop(func() {
			memberDone <- struct{}{}
		})

	go room.Start()
	// wait for the room to be ready before listening
	<-roomDone
	go member.Listen()

	// wait until the member is listening before subscribing
	<-memberDone
	room.Subscribe(*member)
	// wait until the member is subscribed before sending a message
	<-roomDone
	<-memberDone
	<-memberDone
	message := sharechat.NewChatMessage(*member, []byte("hello world!"))
	member.Inbound(message)

	// wait until the message is received to verify the result
	<-memberDone
	result, ok := connection.InboundMessages()[message.ID]
	assert.True(t, ok, "message should be recorded")

	// We don't compare the structs outright here because it results in a race condition
	// since both messages have the same memory reference
	assert.Equal(t, message.Message, result.Message, "message content should be equal")
	assert.Equal(t, message.Type, result.Type, "message type should be equal")
	assert.Equal(t, message.Sent, result.Sent, "message timestamp should be equal")
	assert.Equal(t, message.RoomID, result.RoomID, "message room ids should be equal")
	assert.Equal(t, message.Member.ID, result.Member.ID, "member IDs should be equal")
	assert.Equal(t, message.Member.Name, result.Member.Name, "member names should be equal")
}
