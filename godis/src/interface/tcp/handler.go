package tcp

/*tcp处理接口
 */

import (
	"context"
	"net"
)

type HandleFunc func(ctx context.Context, conn net.Conn)

type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}
