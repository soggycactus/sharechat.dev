package http

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/soggycactus/sharechat.dev/sharechat"
)

func NewCreateRoomHandler(controller *sharechat.Controller) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		room, err := controller.CreateRoom(context.Background(), uuid.NewString())
		if err != nil {
			log.Printf("failed to create room: %v", err)
			http.Error(w, "failed to create room", http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(room)
	}
}

func NewGetRoomHandler(controller *sharechat.Controller) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		roomID, ok := vars["room"]
		if !ok {
			log.Print("room path parameter not provided")
			http.Error(w, "path parameter not provided", http.StatusBadRequest)
			return
		}

		response, err := controller.GetRoom(r.Context(), roomID)
		if err != nil {
			log.Printf("failed to get room: %v", err)
			http.Error(w, "failed to get room", http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}

func NewGetRoomMessagesHandler(controller *sharechat.Controller) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		roomID, ok := vars["room"]
		if !ok {
			log.Printf("room path parameter not provided")
			http.Error(w, "path parameter not provided", http.StatusBadRequest)
			return
		}

		messages, err := controller.GetMessagesByRoom(r.Context(), roomID)
		if err != nil {
			log.Printf("failed to get messages for room %s: %v", roomID, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(messages)
	}
}

func NewServeRoomHandler(controller *sharechat.Controller, upgrader websocket.Upgrader) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		roomID, ok := vars["room"]
		if !ok {
			log.Printf("room path parameter not provided")
			http.Error(w, "path parameter not provided", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("could not upgrade connection: %v", err)
			return
		}

		if err := controller.ServeRoom(context.Background(), roomID, &Connection{conn: conn}); err != nil {
			var publishErr *sharechat.ErrFailedToPublish
			if errors.As(err, &publishErr) {
				log.Printf("room %s is serving, but member failed to publish: %v", roomID, publishErr)
				return
			}
			log.Printf("failed to serve room %s: %v", roomID, err)
			return
		}
	}
}
