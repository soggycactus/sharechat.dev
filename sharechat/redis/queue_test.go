//go:build int || all

package redis_test

import (
	"context"
	"testing"
	"time"

	redisv8 "github.com/go-redis/redis/v8"
	"github.com/soggycactus/sharechat.dev/sharechat"
	"github.com/soggycactus/sharechat.dev/sharechat/redis"
	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	client := redisv8.NewClient(&redisv8.Options{
		Addr:     "0.0.0.0:6379",
		Username: "",
		Password: "",
	})
	queue := redis.NewQueue(*client)
	ctx, fn := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer fn()

	err := queue.Ping(ctx)
	if err != nil {
		t.Fatal(err)
	}

	controller := make(chan sharechat.Message)
	done := make(chan struct{})
	ready := make(chan struct{})
	receive, err := queue.Subscribe(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}

	go receive(controller, done, ready)
	<-ready

	member := sharechat.NewMember("test", "test", nil)
	err = queue.Publish(ctx, sharechat.NewChatMessage(*member, []byte("hello world!")))
	if err != nil {
		t.Fatal(err)
	}

	message := <-controller
	assert.Equal(t, member.ID, message.MemberID, "message should have correct Member")
	assert.Equal(t, "hello world!", message.Message, "message content should be same")
}
