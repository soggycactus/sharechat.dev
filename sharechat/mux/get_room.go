package mux

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/soggycactus/sharechat.dev/sharechat"
)

type GetRoomResponse struct {
	RoomID   string             `json:"room_id"`
	RoomName string             `json:"room_name"`
	Members  []sharechat.Member `json:"members"`
}

func NewGetRoomHandler(repo sharechat.RoomRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		roomID, ok := vars["room"]
		if !ok {
			log.Printf("room path parameter not provided")
			http.Error(w, "path parameter not provided", http.StatusBadRequest)
			return
		}

		room, err := repo.GetRoom(roomID)
		if err != nil {
			http.Error(w, "room not found", http.StatusNotFound)
			return
		}

		members, err := repo.GetRoomMembers(room.ID)
		if err != nil {
			http.Error(w, "failed to get room members", http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(&GetRoomResponse{RoomID: room.ID, RoomName: room.Name, Members: *members})
	}
}
