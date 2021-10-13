package sharechat

import "context"

type Queue interface {
	Publish(ctx context.Context, message Message) error
	Subscribe(ctx context.Context, roomID string, controller chan Message, done chan struct{})
}
