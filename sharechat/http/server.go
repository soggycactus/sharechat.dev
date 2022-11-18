package http

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/soggycactus/sharechat.dev/sharechat"
)

type Server struct {
	upgrader   *websocket.Upgrader
	controller *sharechat.Controller
	Server     *http.Server
}

func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	if err := s.controller.Healthcheck(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		log.Printf("healthcheck failed: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) CreateRoom(w http.ResponseWriter, r *http.Request) {
	room, err := s.controller.CreateRoom(context.Background())
	if err != nil {
		log.Printf("failed to create room: %v", err)
		http.Error(w, "failed to create room", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(room)
}

func (s *Server) GetRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, ok := vars["room"]
	if !ok {
		log.Print("room path parameter not provided")
		http.Error(w, "path parameter not provided", http.StatusBadRequest)
		return
	}

	response, err := s.controller.GetRoom(r.Context(), roomID)
	if err != nil {
		log.Printf("failed to get room: %v", err)
		http.Error(w, "failed to get room", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (s *Server) GetRoomMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, ok := vars["room"]
	if !ok {
		log.Printf("room path parameter not provided")
		http.Error(w, "path parameter not provided", http.StatusBadRequest)
		return
	}

	messages, err := s.controller.GetMessagesByRoom(r.Context(), roomID)
	if err != nil {
		log.Printf("failed to get messages for room %s: %v", roomID, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(messages)
}

func (s *Server) ServeRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, ok := vars["room"]
	if !ok {
		log.Printf("room path parameter not provided")
		http.Error(w, "path parameter not provided", http.StatusBadRequest)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("could not upgrade connection: %v", err)
		return
	}

	if err := s.controller.ServeRoom(context.Background(), roomID, &Connection{conn: conn}); err != nil {
		var publishErr *sharechat.ErrFailedToPublish
		if errors.As(err, &publishErr) {
			log.Printf("room %s is serving, but member failed to publish: %v", roomID, publishErr)
			return
		}
		log.Printf("failed to serve room %s: %v", roomID, err)
		return
	}
}

func NewServer(controller *sharechat.Controller, upgrader websocket.Upgrader) *Server {
	router := mux.NewRouter()
	s := Server{
		upgrader:   &upgrader,
		controller: controller,
	}

	router.HandleFunc("/api/room", s.CreateRoom).Methods(http.MethodPost)
	router.HandleFunc("/api/room/{room}/messages", s.GetRoomMessages).Methods(http.MethodGet)
	router.HandleFunc("/api/room/{room}", s.GetRoom).Methods(http.MethodGet)
	router.HandleFunc("/api/serve/{room}", s.ServeRoom).Methods(http.MethodGet)
	router.HandleFunc("/api/healthz", s.Health).Methods(http.MethodGet)

	server := http.Server{
		Addr:         "0.0.0.0:8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	s.Server = &server

	return &s
}
