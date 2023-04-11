package TCP_protocol

/*TCP服务端
 */

import (
	"context"
	"fmt"
	"godis/src/interface/tcp"
	"godis/src/lib/logger"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// tcp 服务属性
type Config struct {
	Address    string        `yaml:"address"`
	Maxconnect uint32        `yaml:"max-connect"`
	Timeout    time.Duration `yaml:"timeout"`
}

// 服务器中客户端计数器
var ClientCounter int

// 监听中断信号，收到后通知服务器关闭
func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {
	closeChan := make(chan struct{})
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("bind: %s, start listening...", cfg.Address))
	ListenAndServe(listener, handler, closeChan)
	return nil
}

// 监听、提供服务 closeChan有消息就关闭
func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	errCh := make(chan error, 1)
	defer close(errCh)
	go func() {
		select {
		case <-closeChan:
			logger.Info("get exit signal")
		case er := <-errCh:
			logger.Info(fmt.Sprintf("accept error: %s", er.Error()))
		}
		logger.Info("shutting down...")
		_ = listener.Close()
		_ = handler.Close()
	}()

	ctx := context.Background()
	var waitDone sync.WaitGroup
	for {
		conn, err := listener.Accept()
		if err != nil {
			errCh <- err
			break
		}
		logger.Info("accept link")
		ClientCounter++
		waitDone.Add(1)
		go func() {
			defer func() {
				waitDone.Done()
				ClientCounter--
			}()
			handler.Handle(ctx, conn)
		}()
	}
	waitDone.Wait()
}
