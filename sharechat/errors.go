package sharechat

import (
	"errors"
	"fmt"
)

// ErrSendTimedOut is returned when a send has blocked for too long
var ErrSendTimedOut = errors.New("send timed out")

// ErrRoomNotReady is returned when a room is not yet ready
var ErrRoomNotReady = errors.New("room is not ready")

// ErrRoomNotShutdown is returned when a room has not yet processed a shutdown request
var ErrRoomNotShutdown = errors.New("room is not shutdown")

// ErrRoomNotReceiving is returned when a room does not successfully receive messages
var ErrRoomNotReceiving = errors.New("room not receiving")

// ErrExpectedClose is a wrapper error that signifies the connection closed expectedly.
// We do not need to log these errors.
var ErrExpectedClose = errors.New("expected close error")

// ErrNotListening is returned when a Member is not successfully listening
var ErrNotListening = errors.New("listen is not ready")

// ErrNotSubscribed is returned when a Queue is not successfully subscribed to a topic
var ErrNotSubscribed = errors.New("subscribe is not ready")

// ErrNotBroadcasting is returned when a Member is not broadcasting
var ErrNotBroadcasting = errors.New("broadcast is not ready")

// ErrFailedToPublish is returned when the Controller cannot publish to the Queue
type ErrFailedToPublish struct {
	err error
}

func (e *ErrFailedToPublish) Error() string {
	return fmt.Sprintf("failed to publish to queue: %s", e.err)
}

func (e *ErrFailedToPublish) Unwrap() error {
	return e.err
}
