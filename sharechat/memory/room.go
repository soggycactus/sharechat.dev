package memory

import (
	"context"
	"fmt"

	"github.com/soggycactus/sharechat.dev/sharechat"
)

type RoomRepo struct {
	Rooms map[string]*sharechat.Room
}

func NewRoomRepo() *RoomRepo {
	rooms := make(map[string]*sharechat.Room)
	return &RoomRepo{Rooms: rooms}
}

func (m *RoomRepo) InsertRoom(ctx context.Context, room *sharechat.Room) error {
	m.Rooms[room.ID] = room
	return nil
}

func (m *RoomRepo) GetRoom(ctx context.Context, roomID string) (*sharechat.Room, error) {
	if room, ok := m.Rooms[roomID]; !ok {
		return nil, fmt.Errorf("room %s does not exist", roomID)
	} else {
		return room, nil
	}
}

func (m *RoomRepo) DeleteRoom(ctx context.Context, roomID string) error {
	delete(m.Rooms, roomID)
	return nil
}
