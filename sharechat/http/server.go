package http

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/soggycactus/sharechat.dev/sharechat"
)

var upgrader = websocket.Upgrader{HandshakeTimeout: 5 * time.Second}

func NewServer(controller *sharechat.Controller) *http.Server {
	createRoom := NewCreateRoomHandler(controller)
	serveRoom := NewServeRoomHandler(controller, upgrader)
	getRoom := NewGetRoomHandler(controller)
	getMessages := NewGetRoomMessagesHandler(controller)

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

	return &server
}
