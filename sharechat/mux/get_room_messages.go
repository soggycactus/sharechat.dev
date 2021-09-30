package mux

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/soggycactus/sharechat.dev/sharechat"
)

func NewGetRoomMessagesHandler(repo sharechat.MessageRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		roomID, ok := vars["room"]
		if !ok {
			log.Printf("room path parameter not provided")
			http.Error(w, "path parameter not provided", http.StatusBadRequest)
			return
		}

		messages, err := repo.GetMessagesByRoom(roomID)
		if err != nil {
			log.Printf("failed to get messages for room %s: %v", roomID, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(messages)
	}
}
