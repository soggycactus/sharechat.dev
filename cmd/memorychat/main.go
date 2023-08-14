package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/soggycactus/sharechat.dev/sharechat"
	"github.com/soggycactus/sharechat.dev/sharechat/http"
	"github.com/soggycactus/sharechat.dev/sharechat/memory"
)

var (
	upgrader       websocket.Upgrader = websocket.Upgrader{HandshakeTimeout: 5 * time.Second}
	allowedOrigins http.AllowedOrigins
)

func main() {
	flag.Var(&allowedOrigins, "allowed-origin", "allowed origins for CORS")
	flag.Parse()

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

	server := http.NewServer(controller, upgrader, cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "HEAD", "POST", "OPTIONS"},
	})

	log.Print("starting memorychat server on port 8080")
	log.Fatal(server.Server.ListenAndServe())
}
