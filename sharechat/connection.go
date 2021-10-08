package sharechat

type Connection interface {
	WriteMessage(Message) error
	ReadBytes() ([]byte, error)
}
