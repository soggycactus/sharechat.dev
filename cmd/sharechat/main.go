package main

import (
	"log"

	"github.com/soggycactus/sharechat.dev/sharechat"
	"github.com/soggycactus/sharechat.dev/sharechat/memory"
	"github.com/soggycactus/sharechat.dev/sharechat/mux"
	"github.com/soggycactus/sharechat.dev/sharechat/redis"
)

func main() {
	roomRepo := memory.NewRoomRepo()
	messageRepo := memory.NewMessageRepo()
	memberRepo := memory.NewMemberRepo(messageRepo)
	controller := sharechat.NewController(sharechat.NewControllerInput{
		RoomRepo:    roomRepo,
		MessageRepo: messageRepo,
		MemberRepo:  memberRepo,
		Queue:       redis.NewQueue("0.0.0.0:6379", "", ""),
	})

	server := mux.NewServer(controller)

	log.Print("starting server on port 8080")
	log.Fatal(server.ListenAndServe())
}
