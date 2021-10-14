package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/soggycactus/sharechat.dev/sharechat"
)

const (
	InsertMessageQuery     = "INSERT INTO messages (message_id, type, message, room_id, member_id) VALUES ($1,$2,$3,$4,$5)"
	GetMessagesByRoomQuery = `
	SELECT message_id, m.room_id, m.member_id, me.name member_name, type, message, sent
	FROM messages m JOIN members me ON m.member_id = me.member_id 
	WHERE m.room_id=$1
	`
)

func NewMessageRepository(db *sql.DB, driver string) *MessageRepository {
	return &MessageRepository{db: db, driver: driver}
}

type MessageRepository struct {
	db     *sql.DB
	driver string
}

func (m *MessageRepository) InsertMessage(ctx context.Context, message sharechat.Message) error {
	db := sqlx.NewDb(m.db, m.driver)

	return executeTransaction(
		ctx,
		*db,
		InsertMessageQuery,
		message.ID,
		message.Type,
		message.Message,
		message.RoomID,
		message.MemberID,
	)
}

func (m *MessageRepository) GetMessagesByRoom(ctx context.Context, roomID string) (*[]sharechat.Message, error) {
	db := sqlx.NewDb(m.db, m.driver)

	var messages []sharechat.Message
	if err := db.SelectContext(ctx, &messages, GetMessagesByRoomQuery, roomID); err != nil {
		return nil, err
	}

	return &messages, nil
}
