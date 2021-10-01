package mux

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/soggycactus/sharechat.dev/sharechat"
)

func NewServeRoomHandler(repo sharechat.RoomRepository, upgrader websocket.Upgrader) func(w http.ResponseWriter, r *http.Request) {
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
			log.Printf("room %s does not exist", roomID)
			http.Error(w, "room does not exist", http.StatusNotFound)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("could not upgrade connection: %v", err)
			http.Error(w, "could not upgrade connection", http.StatusInternalServerError)
			return
		}

		sub := sharechat.NewMember(uuid.NewString(), room, &Connection{conn: conn})
		go sub.Listen()
		go sub.Broadcast()
	}
}
