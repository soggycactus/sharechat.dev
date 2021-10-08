package mux

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/soggycactus/sharechat.dev/sharechat"
)

type CreateRoomResponse struct {
	RoomID string `json:"room_id"`
}

func NewCreateRoomHandler(controller *sharechat.Controller) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		room, err := controller.CreateRoom(context.Background(), uuid.NewString())
		if err != nil {
			log.Printf("failed to create room: %v", err)
			http.Error(w, "failed to create room", http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(CreateRoomResponse{RoomID: room.ID})
	}
}
