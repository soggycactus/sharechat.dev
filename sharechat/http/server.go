package http

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/soggycactus/sharechat.dev/sharechat"
)

type AllowedOrigins []string

func (a *AllowedOrigins) String() string {
	return strings.Join(*a, ",")
}

func (a *AllowedOrigins) Set(value string) error {
	*a = append(*a, value)
	return nil
}

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

type GetRoomMessagesResponse struct {
	Messages   []sharechat.Message `json:"messages"`
	NumResults int                 `json:"numResults"`
	Next       string              `json:"next"`
}

func (s *Server) GetRoomMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, ok := vars["room"]
	if !ok {
		log.Printf("room path parameter not provided")
		http.Error(w, "path parameter not provided", http.StatusBadRequest)
		return
	}

	options := sharechat.GetMessageOptions{
		Limit:  0,
		RoomID: roomID,
	}

	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		options.Limit = limit
	}

	if rawCursor := r.URL.Query().Get("after"); rawCursor != "" {
		cursor := sharechat.MessageCursor{}
		if err := cursor.DecodeFromString(rawCursor); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		options.After = cursor
	}

	if rawCursor := r.URL.Query().Get("before"); rawCursor != "" {
		cursor := sharechat.MessageCursor{}
		if err := cursor.DecodeFromString(rawCursor); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		options.Before = cursor
	}

	messages, err := s.controller.GetMessages(r.Context(), options)
	if err != nil {
		if errors.Is(err, sharechat.ErrInvalidOptions) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("failed to get messages for room %s: %v", roomID, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	nextCursor := sharechat.MessageCursor{}
	if len(messages) > 0 {
		nextCursor = sharechat.MessageCursor{
			Sent: messages[len(messages)-1].Sent,
			ID:   messages[len(messages)-1].ID,
		}
	}

	response := GetRoomMessagesResponse{
		Messages:   messages,
		NumResults: len(messages),
		Next:       nextCursor.Encode(),
	}
	_ = json.NewEncoder(w).Encode(response)
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

func NewServer(controller *sharechat.Controller, upgrader websocket.Upgrader, corsOptons cors.Options) *Server {
	router := mux.NewRouter()
	c := cors.New(corsOptons)
	upgrader.CheckOrigin = c.OriginAllowed
	s := Server{
		upgrader:   &upgrader,
		controller: controller,
	}

	router.HandleFunc("/api/room", s.CreateRoom).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/api/room/{room}/messages", s.GetRoomMessages).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/api/room/{room}", s.GetRoom).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/api/serve/{room}", s.ServeRoom).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/api/healthz", s.Health).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Welcome to sharechat.dev! Unfortunately I'm an infrastructure engineer, so I don't have a fancy UI for you :(\n\nI promise, this server still works though!"))
	}).Methods(http.MethodGet, http.MethodOptions)

	handler := c.Handler(router)
	server := http.Server{
		Addr:         "0.0.0.0:8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	s.Server = &server

	return &s
}
