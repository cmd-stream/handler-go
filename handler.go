package handler

import (
	"context"
	"sync"
	"time"

	"github.com/cmd-stream/base-go"
	dser "github.com/cmd-stream/delegate-go/server"
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

func (h *Handler[T]) Handle(ctx context.Context, transport dser.Transport[T]) (
	err error) {
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

func receiveCmdAndInvoke[T any](ctx context.Context, transport dser.Transport[T],
	invoker Invoker[T],
	errs chan<- error,
	wg *sync.WaitGroup,
	options Options,
) {
	var (
		seq   base.Seq
		cmd   base.Cmd[T]
		n     int
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
		seq, cmd, n, err = transport.Receive()
		if err != nil {
			queueErr(err, errs)
			wg.Done()
			return
		}
		if options.At {
			at = time.Now()
		}
		wg.Add(1)
		go invokeCmd(ctx, seq, at, n, cmd, invoker, proxy, errs, wg)
	}
}

func invokeCmd[T any](ctx context.Context, seq base.Seq, at time.Time,
	bytesRead int,
	cmd base.Cmd[T],
	invoker Invoker[T],
	proxy base.Proxy,
	errs chan<- error,
	wg *sync.WaitGroup,
) {
	err := invoker.Invoke(ctx, seq, at, bytesRead, cmd, proxy)
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
