package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Postgres Driver
	"github.com/soggycactus/sharechat.dev/sharechat"
)

const (
	InsertRoomQuery = "INSERT INTO rooms (room_id, room_name) VALUES ($1,$2)"
	GetRoomQuery    = "SELECT room_id, room_name FROM rooms WHERE room_id=$1"
)

func NewRoomRepository(db *sql.DB, driver string) *RoomRepository {
	return &RoomRepository{db: db, driver: driver}
}

type RoomRepository struct {
	db     *sql.DB
	driver string
}

func (r *RoomRepository) InsertRoom(ctx context.Context, room *sharechat.Room) error {
	db := sqlx.NewDb(r.db, r.driver)

	return executeTransaction(ctx, *db, InsertRoomQuery, room.ID, room.Name)
}

func (r *RoomRepository) GetRoom(ctx context.Context, roomID string) (*sharechat.Room, error) {
	db := sqlx.NewDb(r.db, r.driver)

	var room sharechat.Room
	err := db.GetContext(ctx, &room, GetRoomQuery, roomID)
	if err != nil {
		return nil, err
	}

	sharechat.SetupRoom(&room)

	return &room, nil
}

func (r *RoomRepository) DeleteRoom(ctx context.Context, roomID string) error {
	return errors.New("TODO: not implemented")
}
