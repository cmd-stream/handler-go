package handler

import (
	"context"
	"time"

	"github.com/cmd-stream/base-go"
)

// Invoker executes Commands on the server.
//
// A single Invoker instance handles Commands concurrently in separate goroutines,
// so it must be thread-safe.
//
// The Invoke method can act as a central handler for common operations across
// multiple Commands, such as logging.
type Invoker[T any] interface {
	Invoke(ctx context.Context, at time.Time, seq base.Seq, cmd base.Cmd[T],
		proxy base.Proxy) error
}

// InvokerFn is a functional implementation of the Invoker interface.
type InvokerFn[T any] func(ctx context.Context, at time.Time, seq base.Seq,
	cmd base.Cmd[T], proxy base.Proxy) error

func (i InvokerFn[T]) Invoke(ctx context.Context, at time.Time, seq base.Seq,
	cmd base.Cmd[T], proxy base.Proxy) error {
	return i(ctx, at, seq, cmd, proxy)
}
