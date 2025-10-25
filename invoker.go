package handler

import (
	"context"
	"time"

	"github.com/cmd-stream/core-go"
)

// Invoker executes Commands on the server.
//
// A single Invoker instance handles Commands concurrently in separate goroutines,
// so it must be thread-safe. The Invoke method can act as a central handler for
// common operations across multiple Commands, such as logging.
type Invoker[T any] interface {
	Invoke(ctx context.Context, seq core.Seq, at time.Time, bytesRead int,
		cmd core.Cmd[T], proxy core.Proxy) error
}

// InvokerFn is a functional implementation of the Invoker interface.
type InvokerFn[T any] func(ctx context.Context, seq core.Seq, at time.Time,
	bytesRead int, cmd core.Cmd[T], proxy core.Proxy) error

func (i InvokerFn[T]) Invoke(ctx context.Context, seq core.Seq, at time.Time,
	bytesRead int, cmd core.Cmd[T], proxy core.Proxy,
) error {
	return i(ctx, seq, at, bytesRead, cmd, proxy)
}
