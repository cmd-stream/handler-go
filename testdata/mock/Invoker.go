package hmock

import (
	"context"
	"time"

	"github.com/cmd-stream/base-go"
	"github.com/ymz-ncnk/mok"
)

type InvokeFn[T any] func(ctx context.Context, seq base.Seq, at time.Time,
	bytesRead int, cmd base.Cmd[T], proxy base.Proxy) (err error)

func NewInvoker[T any]() Invoker[T] {
	return Invoker[T]{
		Mock: mok.New("Invoker"),
	}
}

type Invoker[T any] struct {
	*mok.Mock
}

func (mock Invoker[T]) RegisterInvoke(fn InvokeFn[T]) Invoker[T] {
	mock.Register("Invoke", fn)
	return mock
}

func (mock Invoker[T]) RegisterNInvoke(n int, fn InvokeFn[T]) Invoker[T] {
	mock.RegisterN("Invoke", n, fn)
	return mock
}

func (mock Invoker[T]) Invoke(ctx context.Context, seq base.Seq, at time.Time,
	bytesRead int,
	cmd base.Cmd[T],
	proxy base.Proxy,
) (err error) {
	vals, err := mock.Call("Invoke", mok.SafeVal[context.Context](ctx), seq, at,
		bytesRead, mok.SafeVal[base.Cmd[T]](cmd), mok.SafeVal[base.Proxy](proxy))
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}
