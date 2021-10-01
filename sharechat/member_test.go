package sharechat_test

import (
	"errors"
	"testing"

	"github.com/soggycactus/sharechat.dev/sharechat"
	"github.com/soggycactus/sharechat.dev/sharechat/memory"
	"github.com/soggycactus/sharechat.dev/sharechat/mock"
	"github.com/stretchr/testify/assert"
)

func TestListen(t *testing.T) {
	// channels will notify us when the subscribe has been received
	memberDone := make(chan struct{})

	messageRepo := memory.NewMessageRepo()
	memberRepo := memory.NewMemberRepo(&messageRepo)

	// Create a test room
	room := sharechat.NewRoom("test", &memberRepo, &messageRepo).
		WithNoErrorLogs()

	connection := mock.NewConnection().WithWriteMessageResult(nil)
	member := sharechat.NewMember("test", room, connection).
		WithTestNoop(func(i interface{}) {
			memberDone <- struct{}{}
		})

	go room.Start()
	go member.Listen()

	// wait until the member is subscribed to send a message
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

func TestBroadcast(t *testing.T) {
	// memberDone will notify us when the subscribe has been received
	memberDone := make(chan interface{})

	messageRepo := memory.NewMessageRepo()
	memberRepo := memory.NewMemberRepo(&messageRepo)

	testCases := []struct {
		name    string
		message []byte
		err     error
	}{
		{name: "Successful Send", message: []byte("hello!"), err: nil},
		{name: "Expected Close Error", message: nil, err: sharechat.ExpectedCloseError},
		{name: "Generic Error", message: nil, err: errors.New("generic error")},
	}

	for _, test := range testCases {
		// Create a test room
		room := sharechat.NewRoom("test", &memberRepo, &messageRepo).
			WithNoErrorLogs()

		connection := mock.NewConnection().WithReadBytesResult(test.message, test.err)
		member := sharechat.NewMember("test", room, connection).
			WithTestNoop(func(i interface{}) {
				memberDone <- i
			})

		go room.Start()
		go member.Broadcast()
		r := <-memberDone

		if test.message != nil {
			result, ok := r.(sharechat.Message)
			assert.True(t, ok, "result should be casted successfully")
			assert.Equal(t, string(test.message), result.Message, "message content should be same")
		}
	}
}
