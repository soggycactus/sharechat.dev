package main

import (
	"context"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/soggycactus/sharechat.dev/sharechat"
	"github.com/soggycactus/sharechat.dev/sharechat/http"
	"github.com/soggycactus/sharechat.dev/sharechat/memory"
)

var upgrader websocket.Upgrader = websocket.Upgrader{HandshakeTimeout: 5 * time.Second}

func main() {
	roomRepo := memory.NewRoomRepo()
	messageRepo := memory.NewMessageRepo()
	memberRepo := memory.NewMemberRepo(messageRepo)
	controller := sharechat.NewController(sharechat.NewControllerInput{
		RoomRepo:    roomRepo,
		MessageRepo: messageRepo,
		MemberRepo:  memberRepo,
		Queue:       memory.NewQueue(),
		Healthcheck: func(c context.Context) error { return nil },
	})

	server := http.NewServer(controller, upgrader)

	log.Print("starting memorychat server on port 8080")
	log.Fatal(server.Server.ListenAndServe())
}
