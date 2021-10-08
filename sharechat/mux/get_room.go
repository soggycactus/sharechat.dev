package mux

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/soggycactus/sharechat.dev/sharechat"
)

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

		_ = json.NewEncoder(w).Encode(response)
	}
}
