package redis

import (
	"context"
	"encoding/json"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/soggycactus/sharechat.dev/sharechat"
)

type Queue struct {
	redisClient *redis.Client
}

func (q *Queue) Ping(ctx context.Context) error {
	return q.redisClient.Ping(ctx).Err()
}

func (q *Queue) Publish(ctx context.Context, message sharechat.Message) error {
	return q.redisClient.Publish(ctx, message.RoomID, message).Err()
}

func (q *Queue) Subscribe(ctx context.Context, roomID string, controller chan sharechat.Message, done chan struct{}) {
	topic := q.redisClient.Subscribe(ctx, roomID)
	channel := topic.Channel()

	for {
		select {
		case msg := <-channel:
			var message *sharechat.Message
			if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
				log.Printf("could not unmarshal message: %v", err)
				break
			}

			controller <- *message
		case <-done:
			return
		}
	}

}
