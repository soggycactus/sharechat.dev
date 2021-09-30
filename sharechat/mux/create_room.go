package mux

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/soggycactus/sharechat.dev/sharechat"
)

type CreateRoomResponse struct {
	RoomID string `json:"room_id"`
}

func NewCreateRoomHandler(roomRepo sharechat.RoomRepository, memberRepo sharechat.MemberRepository, messageRepo sharechat.MessageRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		room := sharechat.NewRoom("test", memberRepo, messageRepo)
		if err := roomRepo.InsertRoom(room); err != nil {
			log.Printf("failed to insert room %s: %v", room.ID, err)
			http.Error(w, "failed to create room", http.StatusInternalServerError)
			return
		}

		go room.Start()

		_ = json.NewEncoder(w).Encode(CreateRoomResponse{RoomID: room.ID})
	}
}
