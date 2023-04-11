package TCP_protocol

/*
 */

import (
	"bufio"
	"context"
	"godis/src/lib/logger"
	"godis/src/lib/sync/atomic"
	"godis/src/lib/sync/wait"
	"io"
	"net"
	"sync"
	"time"
)

// 客户端处理结构
type EchoHandler struct {
	activeConn sync.Map
	closing    atomic.Boolean
}

// 生成实例
func MakeEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

// 客户端
type EchoClient struct {
	Conn    net.Conn
	Waiting wait.Wait
}

// 关闭单个用户连接 （用到了wait下的超时处理）
func (c *EchoClient) Close() error {
	c.Waiting.WaitWithTimeout(10 * time.Second)
	c.Conn.Close()
	return nil
}

// 客户端处理核心 用于处理数据传输给客户端（换行截至）
func (h *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Get() {
		_ = conn.Close()
		return
	}

	client := &EchoClient{
		Conn: conn,
	}
	h.activeConn.Store(client, struct{}{})

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.Info("connection close")
				h.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}
		client.Waiting.Add(1)
		b := []byte(msg)
		_, _ = conn.Write(b)
		client.Waiting.Done()
	}
}

// 关闭所有客户端连接
func (h *EchoHandler) Close() error {
	logger.Info("handler shutting down...")
	h.closing.Set(true)
	h.activeConn.Range(func(key interface{}, value interface{}) bool {
		client := key.(*EchoClient)
		_ = client.Close()
		return true
	})
	return nil
}
