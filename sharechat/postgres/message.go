package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/soggycactus/sharechat.dev/sharechat"
)

const (
	insertMessageQuery = `
	INSERT INTO messages (message_id, type, message, room_id, member_id) VALUES ($1,$2,$3,$4,$5) 
	RETURNING message_id, type, message, room_id, member_id, sent
	`
	getMessagesQuery = "SELECT message_id, type, message, sent, room_id, member_id FROM messages WHERE 1=1"
)

func NewMessageRepository(db *sql.DB, driver string) *MessageRepository {
	return &MessageRepository{db: db, driver: driver}
}

type MessageRepository struct {
	db     *sql.DB
	driver string
}

func (m *MessageRepository) InsertMessage(ctx context.Context, message sharechat.Message) (*sharechat.Message, error) {
	db := sqlx.NewDb(m.db, m.driver)
	var result sharechat.Message

	if err := db.QueryRowxContext(
		ctx,
		insertMessageQuery,
		message.ID,
		message.Type,
		message.Message,
		message.RoomID,
		message.MemberID,
	).StructScan(&result); err != nil {
		return nil, err
	}

	// make sure to preserve the member name
	result.MemberName = message.MemberName
	return &result, nil
}

func (m *MessageRepository) GetMessages(ctx context.Context, options sharechat.GetMessageOptions) ([]sharechat.Message, error) {
	db := sqlx.NewDb(m.db, m.driver)
	var result []sharechat.Message

	query, err := m.buildGetMessagesQuery(options)
	if err != nil {
		return nil, err
	}

	if err := db.SelectContext(ctx, &result, *query); err != nil {
		return nil, err
	}

	return result, nil
}

func (m *MessageRepository) buildGetMessagesQuery(options sharechat.GetMessageOptions) (*string, error) {
	builder := strings.Builder{}
	if _, err := builder.WriteString(getMessagesQuery); err != nil {
		return nil, err
	}

	// default to 100 records per query
	if options.Limit == 0 {
		options.Limit = 100
	}

	if options.RoomID != "" {
		if _, err := builder.WriteString(fmt.Sprintf(" AND room_id='%s'", options.RoomID)); err != nil {
			return nil, err
		}
	}

	switch {
	case !options.After.IsEmpty():
		if _, err := builder.WriteString(fmt.Sprintf(" AND (sent, message_id) > ('%s', '%s')", options.After.Sent.Format(time.RFC3339Nano), options.After.ID)); err != nil {
			return nil, err
		}

		if _, err := builder.WriteString(" ORDER BY sent ASC, message_id"); err != nil {
			return nil, err
		}
	case !options.Before.IsEmpty():
		if _, err := builder.WriteString(fmt.Sprintf(" AND (sent, message_id) < ('%s', '%s')", options.Before.Sent.Format(time.RFC3339Nano), options.Before.ID)); err != nil {
			return nil, err
		}

		if _, err := builder.WriteString(" ORDER BY sent DESC, message_id"); err != nil {
			return nil, err
		}
	default:
		if _, err := builder.WriteString(" ORDER BY sent DESC, message_id"); err != nil {
			return nil, err
		}
	}

	if options.Limit > 0 {
		if _, err := builder.WriteString(fmt.Sprintf(" LIMIT %d", options.Limit)); err != nil {
			return nil, err
		}
	}

	query := builder.String()
	return &query, nil
}
