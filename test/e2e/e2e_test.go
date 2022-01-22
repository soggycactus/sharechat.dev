//go:build e2e || all

package e2e

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/pressly/goose/v3"
	"github.com/soggycactus/sharechat.dev/sharechat"
	sharechathttp "github.com/soggycactus/sharechat.dev/sharechat/http"
	"github.com/soggycactus/sharechat.dev/sharechat/memory"
	"github.com/soggycactus/sharechat.dev/sharechat/postgres"
	"github.com/soggycactus/sharechat.dev/sharechat/redis"
)

func ServeNewRoom(e *httpexpect.Expect) {
	response := e.POST("/api/room").WithHeader("Content-Type", "application/json").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object()

	response.Keys().ContainsOnly("room_id", "room_name")
	roomID := response.Value("room_id").NotNull().String()
	response.Value("room_name").NotNull().String()

	ServeRoom(e, roomID.Raw())
}

func ServeRoom(e *httpexpect.Expect, roomID string) {
	ws1 := e.GET("/api/serve/{roomID}").WithPath("roomID", roomID).
		WithWebsocketUpgrade().
		Expect().
		Status(http.StatusSwitchingProtocols).
		Websocket()
	defer ws1.Disconnect()

	// first message is member joining
	ws1.Expect().
		JSON().
		Object().
		ValueEqual("type", sharechat.MemberJoined)

	// test writing a message
	ws1.WriteText("hello world!").
		Expect().
		JSON().
		Object().
		ValueEqual("message", "hello world!").
		ValueEqual("type", sharechat.Chat)

	// add another connection
	ws2 := e.GET("/api/serve/{roomID}").WithPath("roomID", roomID).
		WithWebsocketUpgrade().
		Expect().
		Status(http.StatusSwitchingProtocols).
		Websocket()
	defer ws2.Disconnect()

	// assert ws1 knows that ws2 has joined
	ws1.Expect().
		JSON().
		Object().
		ValueEqual("type", sharechat.MemberJoined)

	ws2ID := ws2.Expect().
		JSON().
		Object().
		ValueEqual("type", sharechat.MemberJoined).
		Value("member_id").String().Raw()

	ws2.WriteText("hello ws1!").
		Expect().
		JSON().
		Object().
		ValueEqual("message", "hello ws1!").
		ValueEqual("type", sharechat.Chat)

	// assert ws1 got the message
	ws1.Expect().
		JSON().
		Object().
		ValueEqual("message", "hello ws1!").
		ValueEqual("type", sharechat.Chat)

	// have ws1 leave the room
	ws1.Close().Expect().CloseMessage().NoContent()

	// assert ws2 knows that ws1 left the chat
	ws2.Expect().
		JSON().
		Object().
		ValueEqual("type", sharechat.MemberLeft)

	// get the Room details from the API
	roomResponse := e.GET("/api/room/{roomID}").WithPath("roomID", roomID).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object()

	roomResponse.Keys().ContainsOnly("room_id", "room_name", "members")
	// assert that only ws2 is still in the room
	members := roomResponse.Value("members").Array().NotEmpty()
	members.Length().Equal(1)
	members.Element(0).Object().ValueEqual("id", ws2ID)

	// a total of 5 messages were sent, assert they are all recorded
	e.GET("/api/room/{roomID}/messages").WithPath("roomID", roomID).
		Expect().
		Status(http.StatusOK).
		JSON().
		Array().NotEmpty().Length().Equal(5)
}

func TestMemory(t *testing.T) {
	roomRepo := memory.NewRoomRepo()
	messageRepo := memory.NewMessageRepo()
	memberRepo := memory.NewMemberRepo(messageRepo)
	controller := sharechat.NewController(sharechat.NewControllerInput{
		RoomRepo:    roomRepo,
		MessageRepo: messageRepo,
		MemberRepo:  memberRepo,
		Queue:       redis.NewQueue("0.0.0.0:6379", "", ""),
	})

	server := httptest.NewServer(sharechathttp.NewServer(controller).Handler)
	defer server.Close()
	e := httpexpect.New(t, server.URL)
	ServeNewRoom(e)
}

func TestServeNewRoom(t *testing.T) {
	dbstring := "user=user dbname=public password=password host=0.0.0.0 sslmode=disable"
	driver := "postgres"

	db, err := sql.Open(driver, dbstring)
	if err != nil {
		t.Fatal(err)
	}

	err = goose.Up(db, "../../migrations")
	if err != nil {
		t.Fatal(err)
	}

	roomRepo := postgres.NewRoomRepository(db, "postgres")
	messageRepo := postgres.NewMessageRepository(db, "postgres")
	memberRepo := postgres.NewMemberRepository(db, "postgres")
	controller := sharechat.NewController(sharechat.NewControllerInput{
		RoomRepo:    roomRepo,
		MessageRepo: messageRepo,
		MemberRepo:  memberRepo,
		Queue:       redis.NewQueue("0.0.0.0:6379", "", ""),
	})

	server := httptest.NewServer(sharechathttp.NewServer(controller).Handler)
	defer server.Close()
	e := httpexpect.New(t, server.URL)
	ServeNewRoom(e)
}

func TestServeExistingRoom(t *testing.T) {
	dbstring := "user=user dbname=public password=password host=0.0.0.0 sslmode=disable"
	driver := "postgres"

	db, err := sql.Open(driver, dbstring)
	if err != nil {
		t.Fatal(err)
	}

	err = goose.Up(db, "../../migrations")
	if err != nil {
		t.Fatal(err)
	}

	roomRepo := postgres.NewRoomRepository(db, "postgres")
	messageRepo := postgres.NewMessageRepository(db, "postgres")
	memberRepo := postgres.NewMemberRepository(db, "postgres")
	controller := sharechat.NewController(sharechat.NewControllerInput{
		RoomRepo:    roomRepo,
		MessageRepo: messageRepo,
		MemberRepo:  memberRepo,
		Queue:       redis.NewQueue("0.0.0.0:6379", "", ""),
	})

	server := httptest.NewServer(sharechathttp.NewServer(controller).Handler)
	defer server.Close()
	e := httpexpect.New(t, server.URL)

	existingRoom := sharechat.NewRoom("existing")
	err = roomRepo.InsertRoom(context.Background(), existingRoom)
	if err != nil {
		t.Fatal(err)
	}

	ServeRoom(e, existingRoom.ID)
}
