package handler

import (
	"context"
	"sync"
	"time"

	"github.com/cmd-stream/base-go"
	"github.com/cmd-stream/delegate-go"
)

// New creates a new Handler.
func New[T any](conf Conf, invoker Invoker[T]) *Handler[T] {
	return &Handler[T]{conf, invoker}
}

// Handler is an implementation of the delegate.ServerTransportHandler
// interface.
//
// It sequentially receives commands and executes each of them in a separate
// gorountine using Invoker.
//
// On any error it closes the transport.
type Handler[T any] struct {
	conf    Conf
	invoker Invoker[T]
}

func (h *Handler[T]) Handle(ctx context.Context,
	transport delegate.ServerTransport[T]) (err error) {
	var (
		wg             = &sync.WaitGroup{}
		ownCtx, cancel = context.WithCancel(ctx)
		errs           = make(chan error, 1)
	)
	wg.Add(1)
	go receiveCmdAndInvoke(ownCtx, h.conf, transport, h.invoker, errs, wg)

	select {
	case <-ownCtx.Done():
		err = context.Canceled
	case err = <-errs:
	}

	cancel()
	if err := transport.Close(); err != nil {
		panic(err)
	}

	wg.Wait()
	return
}

func receiveCmdAndInvoke[T any](ctx context.Context, conf Conf,
	transport delegate.ServerTransport[T],
	invoker Invoker[T],
	errs chan<- error,
	wg *sync.WaitGroup,
) {
	var (
		seq   base.Seq
		cmd   base.Cmd[T]
		err   error
		at    time.Time
		proxy = NewProxy[T](transport)
	)
	for {
		if conf.ReceiveTimeout != 0 {
			deadline := time.Now().Add(conf.ReceiveTimeout)
			if err = transport.SetReceiveDeadline(deadline); err != nil {
				queueErr(err, errs)
				wg.Done()
				return
			}
		}
		seq, cmd, err = transport.Receive()
		if err != nil {
			queueErr(err, errs)
			wg.Done()
			return
		}
		if conf.At {
			at = time.Now()
		}
		wg.Add(1)
		go invokeCmd(ctx, seq, cmd, at, invoker, proxy, errs, wg)
	}
}

func invokeCmd[T any](ctx context.Context, seq base.Seq, cmd base.Cmd[T],
	at time.Time,
	invoker Invoker[T],
	proxy base.Proxy,
	errs chan<- error,
	wg *sync.WaitGroup,
) {
	err := invoker.Invoke(ctx, at, seq, cmd, proxy)
	if err != nil {
		queueErr(err, errs)
	}
	wg.Done()
}

func queueErr(err error, errs chan<- error) {
	select {
	case errs <- err:
	default:
	}
}
