package sharechat

import (
	"context"
	"log"
	"sync"
)

type NewControllerInput struct {
	RoomRepo    RoomRepository
	MemberRepo  MemberRepository
	MessageRepo MessageRepository
	Queue       Queue
}

func NewController(input NewControllerInput) *Controller {
	return &Controller{
		roomRepo:    input.RoomRepo,
		memberRepo:  input.MemberRepo,
		messageRepo: input.MessageRepo,
		queue:       input.Queue,
		mu:          new(sync.Mutex),
		roomCache:   make(map[string]*Room),
	}
}

type Controller struct {
	roomRepo    RoomRepository
	memberRepo  MemberRepository
	messageRepo MessageRepository
	queue       Queue

	mu        *sync.Mutex
	roomCache map[string]*Room
}

func (c *Controller) CreateRoom(ctx context.Context, name string, callbackFn ...func(*Message)) (*Room, error) {
	room := NewRoom(name)
	if len(callbackFn) != 0 {
		// we don't support passing in multiple callback functions
		room = room.WithCallbackInbound(callbackFn[0])
	}
	c.addRoomToCache(room)
	if err := c.startRoom(ctx, room); err != nil {
		c.deleteRoomFromCache(room.ID)
		return nil, err
	}

	err := c.Subscribe(ctx, room)
	if err != nil {
		shutdownErr := c.shutdownRoom(ctx, room.ID)
		if shutdownErr == ErrRoomNotShutdown {
			// ErrRoomNotShutdown means the room hasn't processed our shutdown request yet.
			// We close the inbound channel to force the goroutine to stop.
			room.CloseInbound()
		}
		c.deleteRoomFromCache(room.ID)
		return nil, err
	}

	if err := c.roomRepo.InsertRoom(ctx, room); err != nil {
		shutdownErr := c.shutdownRoom(ctx, room.ID)
		if shutdownErr == ErrRoomNotShutdown {
			// ErrRoomNotShutdown means the room hasn't processed our shutdown request yet.
			// We close the inbound channel to force the goroutine to stop.
			room.CloseInbound()
		}
		c.deleteRoomFromCache(room.ID)
		return nil, err
	}

	return room, nil
}

func (c *Controller) ServeRoom(ctx context.Context, roomID string, connection Connection) error {
	var room *Room
	cacheResult, ok := c.getRoomFromCache(roomID)
	if !ok {
		repoResult, err := c.roomRepo.GetRoom(ctx, roomID)
		if err != nil {
			return err
		}
		room = repoResult
		c.addRoomToCache(room)
		if err := c.startRoom(ctx, room); err != nil {
			c.deleteRoomFromCache(room.ID)
			return err
		}
		if err := c.Subscribe(ctx, room); err != nil {
			shutdownErr := c.shutdownRoom(ctx, room.ID)
			if shutdownErr == ErrRoomNotShutdown {
				// ErrRoomNotShutdown means the room hasn't processed our shutdown request yet.
				// We close the inbound channel to force the goroutine to stop.
				room.CloseInbound()
			}
			c.deleteRoomFromCache(room.ID)
			return err
		}
	} else {
		room = cacheResult
	}

	member := NewMember("test", room.ID, connection)

	go member.Listen()
	err := member.ListenReady(ctx)
	if err != nil {
		// Do not allow the member to Listen if the goroutine
		// does not start within our context deadline.
		member.CloseInbound()
		return err
	}

	go c.Publish(ctx, member)

	go member.Broadcast()
	err = member.BroadcastReady(ctx)
	if err != nil {
		defer member.StopListen()
		// Do not allow the member to Broadcast if the
		// goroutine does not start within our context deadline.
		// This also halts the Publish goroutine started previously.
		member.CloseOutbound()
		return err
	}

	message, err := c.memberRepo.InsertMember(ctx, *member)
	if err != nil {
		defer member.StopListen()
		defer member.StopBroadcast()
		return err
	}

	// don't allow the Member to broadcast until we return
	defer func(member *Member) {
		member.startBroadcast <- struct{}{}
	}(member)

	room.AddMember(member)

	if err := c.queue.Publish(ctx, *message); err != nil {
		// At this point the member & joined message have been saved in the database,
		// so we don't stop the goroutines we've started.
		return &ErrFailedToPublish{err: err}
	}

	return nil
}

func (c *Controller) addRoomToCache(room *Room) {
	c.mu.Lock()
	c.roomCache[room.ID] = room
	c.mu.Unlock()
}

func (c *Controller) deleteRoomFromCache(roomID string) {
	c.mu.Lock()
	delete(c.roomCache, roomID)
	c.mu.Unlock()
}

func (c *Controller) getRoomFromCache(roomID string) (*Room, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	room, ok := c.roomCache[roomID]
	return room, ok
}

func (c *Controller) startRoom(ctx context.Context, room *Room) error {
	go room.Start(ctx)

	select {
	case <-room.ready:
		return nil
	case <-ctx.Done():
		room.CloseInbound()
		return ErrRoomNotReady
	}
}

func (c *Controller) shutdownRoom(ctx context.Context, roomID string) error {
	if room, ok := c.roomCache[roomID]; ok {
		// shutdown the room, return if not received before the deadline
		select {
		case room.shutdown <- struct{}{}:
		case <-ctx.Done():
			// ensure the room cannot start again if the goroutine has already been closed
			room.CloseInbound()
			return ErrRoomNotReceiving
		}

		// wait for the room to be successfully shutdown
		select {
		case <-room.stopped:
		case <-ctx.Done():
			return ErrRoomNotShutdown
		}
	}
	return nil
}

func (c *Controller) Publish(ctx context.Context, member *Member) {
	for {
		select {
		case message, ok := <-member.outbound:
			if !ok {
				return
			}

			switch message.Type {
			case MemberLeft:
				message, err := c.memberRepo.DeleteMember(ctx, *member)
				if err != nil {
					log.Printf("failed to delete member %s: %v", member.ID, err)
					return
				}

				if err := c.queue.Publish(ctx, *message); err != nil {
					log.Printf("failed to publish member %s left: %v", member.ID, err)
				}
				return
			default:
				if err := c.messageRepo.InsertMessage(ctx, message); err != nil {
					log.Printf("failed to insert message: %v", err)
					member.inbound <- NewSendFailedMessage(*member)
					break
				}

				if err := c.queue.Publish(ctx, message); err != nil {
					log.Printf("failed to publish message: %v", err)
				}
			}
		case <-member.stopBroadcast:
			return
		}
	}
}

func (c *Controller) Subscribe(ctx context.Context, room *Room) error {
	fn, err := c.queue.Subscribe(ctx, room.ID)
	if err != nil {
		return err
	}

	controller := make(chan Message)
	done := make(chan struct{})
	ready := make(chan struct{})
	go fn(controller, done, ready)
	<-ready
	go func(controller chan Message, done, ready chan struct{}) {
		ready <- struct{}{}
		for {
			select {
			case <-room.shutdown:
				done <- struct{}{}
				return
			case message := <-controller:
				room.inbound <- message
			}
		}
	}(controller, done, ready)
	<-ready

	return nil
}

type GetRoomResponse struct {
	RoomID   string   `json:"room_id"`
	RoomName string   `json:"room_name"`
	Members  []Member `json:"members"`
}

func (c *Controller) GetRoom(ctx context.Context, roomID string) (*GetRoomResponse, error) {
	room, err := c.roomRepo.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}

	members, err := c.memberRepo.GetMembersByRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}

	return &GetRoomResponse{RoomID: room.ID, RoomName: room.Name, Members: *members}, nil
}

func (c *Controller) GetMessagesByRoom(ctx context.Context, roomID string) (*[]Message, error) {
	return c.messageRepo.GetMessagesByRoom(ctx, roomID)
}
