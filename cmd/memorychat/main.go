package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/soggycactus/sharechat.dev/sharechat/memory"
	muxHandlers "github.com/soggycactus/sharechat.dev/sharechat/mux"
)

var upgrader = websocket.Upgrader{HandshakeTimeout: 1024}

func main() {
	// instantiate in-memory repos
	roomRepo := memory.NewRoomRepo()
	messageRepo := memory.NewMessageRepo()
	memberRepo := memory.NewMemberRepo(&messageRepo)

	createRoom := muxHandlers.NewCreateRoomHandler(&roomRepo, &memberRepo, &messageRepo)
	serveRoom := muxHandlers.NewServeRoomHandler(&roomRepo, upgrader)
	getRoom := muxHandlers.NewGetRoomHandler(&roomRepo)
	getMessages := muxHandlers.NewGetRoomMessagesHandler(&messageRepo)

	router := mux.NewRouter()
	router.HandleFunc("/api/room", createRoom).Methods(http.MethodPost)
	router.HandleFunc("/api/room/{room}/messages", getMessages).Methods(http.MethodGet)
	router.HandleFunc("/api/room/{room}", getRoom).Methods(http.MethodGet)
	router.HandleFunc("/api/serve/{room}", serveRoom).Methods(http.MethodGet)

	server := http.Server{
		Addr:         "0.0.0.0:8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Print("starting server on port 8080")
	log.Fatal(server.ListenAndServe())
}
