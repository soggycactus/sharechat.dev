package sharechat

import "context"

type Queue interface {
	Publish(ctx context.Context, message Message) error
	Subscribe(ctx context.Context, room *Room, controller chan Message, done chan struct{})
}
