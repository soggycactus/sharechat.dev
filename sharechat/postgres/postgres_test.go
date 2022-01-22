//go:build int || all

package postgres_test

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/soggycactus/sharechat.dev/sharechat"
	"github.com/soggycactus/sharechat.dev/sharechat/postgres"
	"github.com/stretchr/testify/assert"
)

func GetDB() *sql.DB {
	dbstring := "user=user dbname=public password=password host=0.0.0.0 sslmode=disable"
	driver := "postgres"

	db, err := sql.Open(driver, dbstring)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func TestMain(t *testing.M) {
	err := goose.Up(GetDB(), "../../migrations")
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(t.Run())
}

func TestAll(t *testing.T) {
	db := GetDB()
	ctx := context.Background()

	roomRepo := postgres.NewRoomRepository(db, "postgres")
	room := sharechat.NewRoom("test")
	err := roomRepo.InsertRoom(context.Background(), room)
	if err != nil {
		t.Fatal(err)
	}

	insertedRoom, err := roomRepo.GetRoom(ctx, room.ID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, room.ID, insertedRoom.ID, "room should have same ID")
	assert.Equal(t, room.Name, insertedRoom.Name, "room should have same name")

	memberRepo := postgres.NewMemberRepository(db, "postgres")
	member := sharechat.NewMember("test", room.ID, nil)

	insertTime := time.Now()
	joinedMessage, err := memberRepo.InsertMember(ctx, *member)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, member.ID, joinedMessage.MemberID, "inserted message should have member ID")
	assert.Equal(t, sharechat.MemberJoined, joinedMessage.Type, "inserted message should be member joined type")
	assert.Equal(t, member.Name, joinedMessage.MemberName, "joined message should equal member name")
	assert.True(t, insertTime.Before(joinedMessage.Sent), "joined message should have more recent timestamp")

	members, err := memberRepo.GetMembersByRoom(ctx, room.ID)
	if err != nil {
		t.Fatal(err)
	}

	foundMember := false
	for _, m := range *members {
		if m.ID == member.ID {
			foundMember = true
		}
	}

	assert.True(t, foundMember, "member should be in database")

	messageRepo := postgres.NewMessageRepository(db, "postgres")

	insertTime = time.Now()
	chatMessage := sharechat.NewChatMessage(*member, []byte("hello world!"))
	result, err := messageRepo.InsertMessage(ctx, chatMessage)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, chatMessage.ID, result.ID)
	assert.Equal(t, chatMessage.MemberID, result.MemberID)
	assert.Equal(t, chatMessage.Message, result.Message)
	assert.Equal(t, chatMessage.MemberName, result.MemberName)
	assert.Equal(t, chatMessage.RoomID, result.RoomID)
	assert.Equal(t, chatMessage.Type, result.Type)
	assert.True(t, insertTime.Before(result.Sent), "message timestamp should be more recent")

	deleteTime := time.Now()
	deletedMessage, err := memberRepo.DeleteMember(ctx, *member)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, member.ID, deletedMessage.MemberID, "inserted message should have member ID")
	assert.Equal(t, sharechat.MemberLeft, deletedMessage.Type, "inserted message should be member joined type")
	assert.Equal(t, member.Name, deletedMessage.MemberName, "deleted message should have equal member name")
	assert.True(t, deleteTime.Before(deletedMessage.Sent), "deleted message should have more recent timestamp")

	messages, err := messageRepo.GetMessagesByRoom(ctx, room.ID)

	foundJoinedMessage := false
	foundChatMessage := false
	foundDeletedMessage := false
	for _, m := range *messages {
		switch m.ID {
		case joinedMessage.ID:
			foundJoinedMessage = true
		case chatMessage.ID:
			foundChatMessage = true
		case deletedMessage.ID:
			foundDeletedMessage = true
		}
	}

	assert.True(t, foundJoinedMessage, "joined message should be in database")
	assert.True(t, foundChatMessage, "chat message should be in database")
	assert.True(t, foundDeletedMessage, "deleted message should be in database")

	members, err = memberRepo.GetMembersByRoom(ctx, room.ID)
	if err != nil {
		t.Fatal(err)
	}

	foundMember = false
	for _, m := range *members {
		if m.ID == member.ID {
			foundMember = true
		}
	}

	assert.False(t, foundMember, "member should no longer be in database")
}
