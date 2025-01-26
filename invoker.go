package handler

import (
	"context"
	"time"

	"github.com/cmd-stream/base-go"
)

// Invoker is used by the server to execute Commands.
//
// A single Invoker instance is used to execute Commands in separate goroutines,
// so it must be thread-safe.
//
// Invoke method can serve as a central point to handle operations that are
// common across multiple Commands (such as logging).
type Invoker[T any] interface {
	Invoke(ctx context.Context, at time.Time, seq base.Seq, cmd base.Cmd[T],
		proxy base.Proxy) error
}

// InvokerFn functional implementation of the Invoker interface.
type InvokerFn[T any] func(ctx context.Context, at time.Time, seq base.Seq,
	cmd base.Cmd[T], proxy base.Proxy) error

func (i InvokerFn[T]) Invoke(ctx context.Context, at time.Time, seq base.Seq,
	cmd base.Cmd[T], proxy base.Proxy) error {
	return i(ctx, at, seq, cmd, proxy)
}
