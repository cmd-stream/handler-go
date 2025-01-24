package handler

import (
	"context"
	"time"

	"github.com/cmd-stream/base-go"
)

// Invoker executes commands.
//
// At parameter defines (if Conf.At == true) the command reception time.
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
