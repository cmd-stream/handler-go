package handler

import (
	"context"
	"sync"
	"time"

	"github.com/cmd-stream/base-go"
	"github.com/cmd-stream/delegate-go"
)

// New creates a new Handler.
func New[T any](invoker Invoker[T], ops ...SetOption) *Handler[T] {
	h := Handler[T]{invoker: invoker}
	Apply(ops, &h.options)
	return &h
}

// Handler implements the delegate.ServerTransportHandler interface.
//
// It receives Commands sequentially and executes each in a separate goroutine
// using the Invoker.
//
// If an error occurs, the Handler closes the transport connection.
type Handler[T any] struct {
	invoker Invoker[T]
	options Options
}

func (h *Handler[T]) Handle(ctx context.Context,
	transport delegate.ServerTransport[T]) (err error) {
	var (
		wg             = &sync.WaitGroup{}
		ownCtx, cancel = context.WithCancel(ctx)
		errs           = make(chan error, 1)
	)
	wg.Add(1)
	go receiveCmdAndInvoke(ownCtx, transport, h.invoker, errs, wg, h.options)

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

func receiveCmdAndInvoke[T any](ctx context.Context,
	transport delegate.ServerTransport[T],
	invoker Invoker[T],
	errs chan<- error,
	wg *sync.WaitGroup,
	options Options,
) {
	var (
		seq   base.Seq
		cmd   base.Cmd[T]
		err   error
		at    time.Time
		proxy = NewProxy[T](transport)
	)
	for {
		if options.CmdReceiveDuration != 0 {
			deadline := time.Now().Add(options.CmdReceiveDuration)
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
		if options.At {
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
