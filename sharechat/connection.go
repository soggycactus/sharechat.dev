package sharechat

import (
	"errors"
)

type Connection interface {
	WriteMessage(Message) error
	ReadBytes() ([]byte, error)
}

// ExpectedCloseError is a wrapper error
// that signifies the connection closed expectedly.
// We do not need to log these errors.
var ExpectedCloseError = errors.New("expected close error")
