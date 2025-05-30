package mock

import (
	"context"
	"time"

	"github.com/cmd-stream/core-go"
	"github.com/ymz-ncnk/mok"
)

type InvokeFn[T any] func(ctx context.Context, seq core.Seq, at time.Time,
	bytesRead int, cmd core.Cmd[T], proxy core.Proxy) (err error)

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

func (mock Invoker[T]) Invoke(ctx context.Context, seq core.Seq, at time.Time,
	bytesRead int,
	cmd core.Cmd[T],
	proxy core.Proxy,
) (err error) {
	vals, err := mock.Call("Invoke", mok.SafeVal[context.Context](ctx), seq, at,
		bytesRead, mok.SafeVal[core.Cmd[T]](cmd), mok.SafeVal[core.Proxy](proxy))
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}
