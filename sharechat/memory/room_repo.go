package memory

import (
	"fmt"

	"github.com/soggycactus/sharechat.dev/sharechat"
)

type RoomRepo struct {
	Rooms map[string]sharechat.Room
}

func NewRoomRepo() RoomRepo {
	rooms := make(map[string]sharechat.Room)
	return RoomRepo{Rooms: rooms}
}

func (m *RoomRepo) InsertRoom(room *sharechat.Room) error {
	m.Rooms[room.ID] = *room
	return nil
}

func (m *RoomRepo) GetRoom(roomID string) (*sharechat.Room, error) {
	if room, ok := m.Rooms[roomID]; !ok {
		return nil, fmt.Errorf("room %s does not exist", roomID)
	} else {
		return &room, nil
	}
}

func (m *RoomRepo) GetRoomMembers(roomID string) (*[]sharechat.Member, error) {
	room, ok := m.Rooms[roomID]
	if !ok {
		return nil, fmt.Errorf("room %s not found", roomID)
	}

	members := []sharechat.Member{}

	for _, member := range room.Members() {
		members = append(members, *member)
	}

	return &members, nil
}

func (m *RoomRepo) DeleteRoom(roomID string) error {
	delete(m.Rooms, roomID)
	return nil
}
