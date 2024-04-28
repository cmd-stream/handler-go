package mock

import (
	"context"
	"time"

	"github.com/cmd-stream/base-go"
	"github.com/ymz-ncnk/mok"
)

func NewInvoker[T any]() Invoker[T] {
	return Invoker[T]{
		Mock: mok.New("Invoker"),
	}
}

type Invoker[T any] struct {
	*mok.Mock
}

func (mock Invoker[T]) RegisterInvoke(
	fn func(ctx context.Context, at time.Time, seq base.Seq, cmd base.Cmd[T],
		proxy base.Proxy) (err error)) Invoker[T] {
	mock.Register("Invoke", fn)
	return mock
}

func (mock Invoker[T]) RegisterNInvoke(n int,
	fn func(ctx context.Context, at time.Time, seq base.Seq, cmd base.Cmd[T],
		proxy base.Proxy) (err error)) Invoker[T] {
	mock.RegisterN("Invoke", n, fn)
	return mock
}

func (mock Invoker[T]) Invoke(ctx context.Context, at time.Time, seq base.Seq,
	cmd base.Cmd[T],
	proxy base.Proxy,
) (err error) {
	vals, err := mock.Call("Invoke", mok.SafeVal[context.Context](ctx), at, seq,
		mok.SafeVal[base.Cmd[T]](cmd), mok.SafeVal[base.Proxy](proxy))
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}
