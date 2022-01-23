//go:build unit || all

package sharechat_test

import (
	"context"
	"testing"
	"time"

	"github.com/soggycactus/sharechat.dev/sharechat"
	"github.com/soggycactus/sharechat.dev/sharechat/memory"
	"github.com/soggycactus/sharechat.dev/sharechat/mock"
	"github.com/stretchr/testify/assert"
)

func TestController(t *testing.T) {
	roomRepo := memory.NewRoomRepo()
	messageRepo := memory.NewMessageRepo()
	memberRepo := memory.NewMemberRepo(messageRepo)

	controller := sharechat.NewController(
		sharechat.NewControllerInput{
			RoomRepo:    roomRepo,
			MemberRepo:  memberRepo,
			MessageRepo: messageRepo,
			Queue:       memory.NewQueue(),
		},
	)

	ctx, fn := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	defer fn()

	roomProcessedMessage := make(chan sharechat.Message)
	room, err := controller.CreateRoom(ctx, func(m *sharechat.Message) {
		roomProcessedMessage <- *m
	})
	if err != nil {
		t.Fatalf("failed to create room: %v", err)
	}

	connection := mock.NewConnection().
		WithWriteMessageResult(nil).
		WithReadBytesResult([]byte("hello world"), nil)

	beforeSent := time.Now()

	err = controller.ServeRoom(ctx, room.ID, connection)
	if err != nil {
		t.Fatalf("failed to serve room: %v", err)
	}

	subscribe := <-roomProcessedMessage
	assert.Equal(t, sharechat.MemberJoined, subscribe.Type, "first message should be member joined type")
	chat := <-roomProcessedMessage
	assert.Equal(t, sharechat.Chat, chat.Type, "second message should be chat type")
	assert.Equal(t, "hello world", chat.Message, "message content should be equal")
	assert.Equal(t, 1, len(room.Members()), "room should have one member")
	assert.True(t, beforeSent.Before(chat.Sent), "message timestamp should be more recent")
}
