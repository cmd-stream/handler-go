package mock

import (
	"context"
	"reflect"
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
	var ctxVal reflect.Value
	if ctx == nil {
		ctxVal = reflect.Zero(reflect.TypeOf((*context.Context)(nil)).Elem())
	} else {
		ctxVal = reflect.ValueOf(ctx)
	}
	var cmdVal reflect.Value
	if cmd == nil {
		cmdVal = reflect.Zero(reflect.TypeOf((*base.Cmd[any])(nil)).Elem())
	} else {
		cmdVal = reflect.ValueOf(cmd)
	}
	var proxyVal reflect.Value
	if proxy == nil {
		proxyVal = reflect.Zero(reflect.TypeOf((*base.Proxy)(nil))).Elem()
	} else {
		proxyVal = reflect.ValueOf(proxy)
	}
	vals, err := mock.Call("Invoke", ctxVal, at, seq, cmdVal, proxyVal)
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}
