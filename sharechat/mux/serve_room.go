package mux

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/soggycactus/sharechat.dev/sharechat"
)

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
			http.Error(w, "could not upgrade connection", http.StatusInternalServerError)
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
