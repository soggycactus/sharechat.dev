package redis

import (
	"context"
	"encoding/json"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/soggycactus/sharechat.dev/sharechat"
)

func NewQueue(host, user, password string) *Queue {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     host,
		Username: user,
		Password: password,
	})
	return &Queue{redisClient: redisClient}
}

type Queue struct {
	redisClient *redis.Client
}

func (q *Queue) Ping(ctx context.Context) error {
	return q.redisClient.Ping(ctx).Err()
}

func (q *Queue) Publish(ctx context.Context, message sharechat.Message) error {
	return q.redisClient.Publish(ctx, message.RoomID, message).Err()
}

func (q *Queue) Subscribe(ctx context.Context, roomID string) (func(chan sharechat.Message, chan struct{}, chan struct{}), error) {
	topic := q.redisClient.Subscribe(ctx, roomID)
	// Receive forces us to wait on a response from Redis
	_, err := topic.Receive(ctx)
	if err != nil {
		return nil, err
	}
	channel := topic.Channel()

	return func(controller chan sharechat.Message, done, ready chan struct{}) {
		ready <- struct{}{}
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
	}, nil
}
