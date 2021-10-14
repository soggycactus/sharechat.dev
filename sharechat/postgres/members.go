package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/soggycactus/sharechat.dev/sharechat"
)

const (
	InsertMemberQuery = `
	INSERT INTO members (member_id, name, room_id) VALUES (%v, %v, %v); 
	INSERT INTO messages (message_id, type, message, room_id, member_id) VALUES (%v, %v, %v, %v, %v);
	`
	GetMembersByRoomQuery = "SELECT member_id, name, room_id FROM members WHERE room_id=$1 AND is_deleted=false"
	DeleteMemberQuery     = `
	UPDATE members SET is_deleted=true WHERE member_id=%v;
	INSERT INTO messages (message_id, type, message, room_id, member_id) VALUES (%v, %v, %v, %v, %v);
	`
)

func NewMemberRepository(db *sql.DB, driver string) *MemberRepository {
	return &MemberRepository{db: db, driver: driver}
}

type MemberRepository struct {
	db     *sql.DB
	driver string
}

func (m *MemberRepository) InsertMember(ctx context.Context, member sharechat.Member) (*sharechat.Message, error) {
	db := sqlx.NewDb(m.db, m.driver)

	message := sharechat.NewMemberJoinedMessage(member)
	query := m.buildInsertQuery(member, message)
	if err := executeTransaction(ctx, *db, query); err != nil {
		return nil, err
	}
	return &message, nil
}

func (m *MemberRepository) buildInsertQuery(member sharechat.Member, message sharechat.Message) string {
	return fmt.Sprintf(
		InsertMemberQuery,
		fmt.Sprintf("'%s'", member.ID),
		fmt.Sprintf("'%s'", member.Name),
		fmt.Sprintf("'%s'", member.RoomID),
		fmt.Sprintf("'%s'", message.ID),
		fmt.Sprintf("'%s'", message.Type),
		fmt.Sprintf("'%s'", message.Message),
		fmt.Sprintf("'%s'", message.RoomID),
		fmt.Sprintf("'%s'", message.MemberID),
	)
}

func (m *MemberRepository) GetMembersByRoom(ctx context.Context, roomID string) (*[]sharechat.Member, error) {
	db := sqlx.NewDb(m.db, m.driver)

	var members []sharechat.Member
	if err := db.SelectContext(ctx, &members, GetMembersByRoomQuery, roomID); err != nil {
		return nil, err
	}

	return &members, nil
}

func (m *MemberRepository) DeleteMember(ctx context.Context, member sharechat.Member) (*sharechat.Message, error) {
	db := sqlx.NewDb(m.db, m.driver)

	message := sharechat.NewMemberLeftMessage(member)
	query := m.buildDeleteQuery(member, message)
	if err := executeTransaction(ctx, *db, query); err != nil {
		return nil, err
	}
	return &message, nil
}

func (m *MemberRepository) buildDeleteQuery(member sharechat.Member, message sharechat.Message) string {
	return fmt.Sprintf(
		DeleteMemberQuery,
		fmt.Sprintf("'%s'", member.ID),
		fmt.Sprintf("'%s'", message.ID),
		fmt.Sprintf("'%s'", message.Type),
		fmt.Sprintf("'%s'", message.Message),
		fmt.Sprintf("'%s'", message.RoomID),
		fmt.Sprintf("'%s'", message.MemberID),
	)
}
