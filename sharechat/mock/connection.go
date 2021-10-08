package mock

import (
	"sync"

	"github.com/soggycactus/sharechat.dev/sharechat"
)

type Connection struct {
	readResult  readBytesResult
	writeResult error
	mu          sync.Mutex
	sent        bool
	inbound     map[string]*sharechat.Message
}

func NewConnection() *Connection {
	return &Connection{sent: false, inbound: make(map[string]*sharechat.Message)}
}

type readBytesResult struct {
	bytes []byte
	err   error
}

func (c *Connection) WriteMessage(v sharechat.Message) error {
	c.mu.Lock()
	c.inbound[v.ID] = &v
	c.mu.Unlock()
	return c.writeResult
}

func (c *Connection) ReadBytes() ([]byte, error) {
	c.mu.Lock()
	if !c.sent {
		c.sent = true
		c.mu.Unlock()
		return c.readResult.bytes, c.readResult.err
	}
	c.mu.Unlock()
	// block forever so we don't keep sending messages
	select {}
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
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.inbound
}
