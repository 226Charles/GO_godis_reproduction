package TCP_protocol

import (
	"net"
	"sync"
	"sync/atomic"
)

type EchoHandler struct {
	activeConn sync.Map
	closing    atomic.Bool
}

func MakeEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

type EchoClient struct {
	Conn    net.Conn
	Waiting wait.wait
}

func (c *EchoClient) Close() error {
	c.Waiting.WaitWithTimeout
}
