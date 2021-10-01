package mux

import (
	"errors"

	"github.com/gorilla/websocket"
	"github.com/soggycactus/sharechat.dev/sharechat"
)

type Connection struct {
	conn *websocket.Conn
}

func (w *Connection) ReadBytes() ([]byte, error) {
	_, bytes, err := w.conn.ReadMessage()
	if err != nil {
		var closeErr *websocket.CloseError
		if errors.As(err, &closeErr) {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				return bytes, err
			}
			return bytes, sharechat.ExpectedCloseError
		}
	}
	return bytes, err
}

func (w *Connection) WriteMessage(v sharechat.Message) error {
	return w.conn.WriteJSON(v)
}
