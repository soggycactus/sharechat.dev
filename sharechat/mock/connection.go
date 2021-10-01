package mock

import "github.com/soggycactus/sharechat.dev/sharechat"

type Connection struct {
	readResult  readBytesResult
	writeResult error
	inbound     map[string]*sharechat.Message
}

func NewConnection() *Connection {
	return &Connection{inbound: make(map[string]*sharechat.Message)}
}

type readBytesResult struct {
	bytes []byte
	err   error
}

func (c *Connection) WriteMessage(v sharechat.Message) error {
	c.inbound[v.ID] = &v
	return c.writeResult
}

func (c *Connection) ReadBytes() ([]byte, error) {
	return c.readResult.bytes, c.readResult.err
}

func (c *Connection) WithReadBytesResult(bytes []byte, err error) *Connection {
	c.readResult = readBytesResult{bytes: bytes, err: err}
	return c
}

func (c *Connection) WithWriteMessageResult(err error) *Connection {
	c.writeResult = err
	return c
}

func (c *Connection) InboundMessages() map[string]*sharechat.Message {
	return c.inbound
}
