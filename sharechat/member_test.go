//go:build unit || all

package sharechat_test

import (
	"context"
	"errors"
	"testing"

	"github.com/soggycactus/sharechat.dev/sharechat"
	"github.com/soggycactus/sharechat.dev/sharechat/mock"
	"github.com/stretchr/testify/assert"
)

func TestListen(t *testing.T) {
	// channels will notify us when the subscribe has been received
	memberDone := make(chan struct{})
	roomDone := make(chan struct{})

	// Create a test room
	room := sharechat.NewRoom("test").
		WithCallbackInbound(func(message *sharechat.Message) {
			roomDone <- struct{}{}
		}).
		WithNoErrorLogs()

	connection := mock.NewConnection().WithWriteMessageResult(nil)
	member := sharechat.NewMember("test", room.ID, connection).
		WithCallbackListen(func() {
			memberDone <- struct{}{}
		})

	go room.Start(context.Background())
	// wait for the room to be ready before listening
	err := room.Ready(context.Background())
	if err != nil {
		t.Fatal("room should be ready")
	}

	// wait until the member is listening before subscribing
	go member.Listen()
	err = member.ListenReady(context.Background())
	if err != nil {
		t.Fatal("listen should be ready")
	}

	// subscribe member to the room
	subscribeMessage := sharechat.NewMemberJoinedMessage(*member)
	err = room.Inbound(context.Background(), subscribeMessage)
	if err != nil {
		t.Fatal("subscribe send should be succeed")
	}

	// wait until the member is subscribed before sending a message
	<-roomDone
	<-memberDone
	message := sharechat.NewChatMessage(*member, []byte("hello world!"))
	err = member.Inbound(context.Background(), message)
	if err != nil {
		t.Fatal("message send should be succeed")
	}

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
	testCases := []struct {
		name    string
		message []byte
		err     error
	}{
		{name: "Successful Send", message: []byte("hello!"), err: nil},
		{name: "Expected Close Error", message: nil, err: sharechat.ErrExpectedClose},
		{name: "Generic Error", message: nil, err: errors.New("generic error")},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			// create a new member
			connection := mock.NewConnection().WithReadBytesResult(test.message, test.err)
			member := sharechat.NewMember("test", "test", connection)

			// kick off broadcast goroutine
			go member.Broadcast()
			err := member.BroadcastReady(context.Background())
			if err != nil {
				t.Fatalf("broadcast is not ready: %v", err)
			}

			// kick off a goroutine to listen receive the outbound
			messageChan := make(chan sharechat.Message)
			go func(done chan sharechat.Message) {
				message := member.Outbound()
				done <- message
			}(messageChan)

			// start the broadcast
			member.StartBroadcast()

			// receive the broadcast result
			result := <-messageChan

			if test.message != nil {
				assert.Equal(t, string(test.message), result.Message, "message content should be same")
				assert.Equal(t, sharechat.Chat, result.Type, "message should be chat type")
			} else {
				assert.Equal(t, sharechat.MemberLeft, result.Type, "message should be member left type")
			}
		})
	}
}
