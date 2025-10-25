package handler

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/cmd-stream/core-go"
	cmock "github.com/cmd-stream/core-go/testdata/mock"
	dmock "github.com/cmd-stream/delegate-go/server/testdata/mock"
	mock "github.com/cmd-stream/handler-go/testdata/mock"
	asserterror "github.com/ymz-ncnk/assert/error"
	"github.com/ymz-ncnk/mok"
)

func TestHandler(t *testing.T) {
	delta := 100 * time.Millisecond

	t.Run("Handler should be able to handle several cmds and close when ctx done",
		func(t *testing.T) {
			var (
				ctx, cancel            = context.WithCancel(context.Background())
				wantCmdReceiveDuration = time.Second

				wantN            = 1
				wantErr          = context.Canceled
				seq1    core.Seq = 1
				seq2    core.Seq = 2

				cmd1 = cmock.NewCmd()
				cmd2 = cmock.NewCmd()
				cmds = map[cmock.Cmd]struct{}{cmd1: {}, cmd2: {}}

				done      = make(chan struct{})
				starTime  = time.Now()
				transport = dmock.NewTransport().RegisterNSetReceiveDeadline(3,
					func(deadline time.Time) (err error) {
						wantDeadline := starTime.Add(wantCmdReceiveDuration)
						asserterror.SameTime(deadline, wantDeadline, delta, t)
						return
					},
				).RegisterReceive(
					func() (seq core.Seq, cmd core.Cmd[any], n int, err error) {
						return seq1, cmd1, 1, nil
					},
				).RegisterReceive(
					func() (seq core.Seq, cmd core.Cmd[any], n int, err error) {
						return seq2, cmd2, 2, nil
					},
				).RegisterReceive(
					func() (seq core.Seq, cmd core.Cmd[any], n int, err error) {
						<-done
						n = wantN
						err = errors.New("transport closed")
						return
					},
				).RegisterClose(
					func() (err error) {
						defer close(done)
						return nil
					},
				)
				invoker = mock.NewInvoker[any]().RegisterInvoke(
					func(ctx context.Context, seq core.Seq, at time.Time, bytesRead int,
						cmd core.Cmd[any], proxy core.Proxy,
					) (err error) {
						delete(cmds, cmd.(cmock.Cmd))
						return nil
					},
				).RegisterInvoke(
					func(ctx context.Context, seq core.Seq, at time.Time, bytesRead int,
						cmd core.Cmd[any], proxy core.Proxy,
					) (err error) {
						delete(cmds, cmd.(cmock.Cmd))
						return nil
					},
				)
				handler = New(invoker, WithCmdReceiveDuration(wantCmdReceiveDuration))
				mocks   = []*mok.Mock{cmd1.Mock, cmd2.Mock, transport.Mock, invoker.Mock}
			)
			go func() {
				time.Sleep(100 * time.Millisecond)
				cancel()
			}()
			err := handler.Handle(ctx, transport)
			asserterror.EqualError(err, wantErr, t)
			asserterror.Equal(len(cmds), 0, t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("If Transport.SetReceiveDeadline fails with an error, Handle should return it",
		func(t *testing.T) {
			var (
				ctx, cancel = context.WithCancel(context.Background())
				wantErr     = errors.New("Transport.SetReceiveDeadline error")
				transport   = dmock.NewTransport().RegisterSetReceiveDeadline(
					func(deadline time.Time) (err error) { return wantErr },
				).RegisterClose(
					func() (err error) { return nil },
				)
				handler = New[any](nil, WithCmdReceiveDuration(time.Second))
				mocks   = []*mok.Mock{transport.Mock}
			)
			defer cancel()
			err := handler.Handle(ctx, transport)
			asserterror.EqualError(err, wantErr, t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("If Transport.Receive fails with an error, Handle should return it",
		func(t *testing.T) {
			var (
				ctx, cancel = context.WithCancel(context.Background())
				wantCmdSize = 2
				wantErr     = errors.New("Transport.Receive error")
				transport   = dmock.NewTransport().RegisterReceive(
					func() (seq core.Seq, cmd core.Cmd[any], cmdSize int, err error) {
						cmdSize = wantCmdSize
						err = wantErr
						return
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				handler = New[any](nil)
				mocks   = []*mok.Mock{transport.Mock}
			)
			defer cancel()
			err := handler.Handle(ctx, transport)
			asserterror.EqualError(err, wantErr, t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("If Invoker.Invoke fails with an error, Handle should return it",
		func(t *testing.T) {
			var (
				wantN     = 2
				wantErr   = errors.New("Invoker.Invoke error")
				done      = make(chan struct{})
				transport = dmock.NewTransport().RegisterReceive(
					func() (seq core.Seq, cmd core.Cmd[any], n int, err error) {
						return 1, cmock.NewCmd(), wantN, nil
					},
				).RegisterReceive(
					func() (seq core.Seq, cmd core.Cmd[any], n int, err error) {
						<-done
						return 0, nil, 0, errors.New("transport closed")
					},
				).RegisterClose(
					func() (err error) {
						defer close(done)
						return nil
					},
				)
				invoker = mock.NewInvoker[any]().RegisterInvoke(
					func(ctx context.Context, seq core.Seq, at time.Time, bytesRead int,
						cmd core.Cmd[any], proxy core.Proxy,
					) (err error) {
						asserterror.Equal(bytesRead, wantN, t)
						return wantErr
					},
				)
				handler = New(invoker)
				mocks   = []*mok.Mock{transport.Mock, invoker.Mock}
			)
			err := handler.Handle(context.Background(), transport)
			asserterror.EqualError(err, wantErr, t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("If Conf.At == true, Invoker.Invoke should receive not empty 'at' param",
		func(t *testing.T) {
			var (
				ctx, cancel = context.WithCancel(context.Background())
				wantAt      time.Time
				wantN       = 3
				mu          sync.Mutex
				done        = make(chan struct{})
				transport   = dmock.NewTransport().RegisterReceive(
					func() (seq core.Seq, cmd core.Cmd[any], n int, err error) {
						mu.Lock()
						wantAt = time.Now()
						mu.Unlock()
						return 1, cmock.NewCmd(), wantN, nil
					},
				).RegisterReceive(
					func() (seq core.Seq, cmd core.Cmd[any], n int, err error) {
						<-done
						return 0, nil, 0, errors.New("transport closed")
					},
				).RegisterClose(
					func() (err error) {
						defer close(done)
						return nil
					},
				)
				invoker = mock.NewInvoker[any]().RegisterInvoke(
					func(ctx context.Context, seq core.Seq, at time.Time, bytesRead int,
						cmd core.Cmd[any], proxy core.Proxy,
					) (err error) {
						asserterror.Equal(bytesRead, wantN, t)
						defer cancel()
						mu.Lock()
						asserterror.SameTime(at, wantAt, delta, t)
						mu.Unlock()
						return
					},
				)
				handler = New(invoker, WithAt())
				mocks   = []*mok.Mock{transport.Mock, invoker.Mock}
			)
			err := handler.Handle(ctx, transport)
			asserterror.EqualError(err, context.Canceled, t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("We should be able to interupt Handler, while it invoke a Command and is locked on ctx",
		func(t *testing.T) {
			var (
				ctx, cancel = context.WithCancel(context.Background())
				wantN       = 1
				wantErr     = context.Canceled
				cmd         = cmock.NewCmd()
				done        = make(chan struct{})
				transport   = dmock.NewTransport().RegisterNSetReceiveDeadline(2,
					func(deadline time.Time) (err error) { return nil },
				).RegisterReceive(
					func() (seq core.Seq, cmd core.Cmd[any], n int, err error) {
						return 1, cmd, 1, nil
					},
				).RegisterReceive(
					func() (seq core.Seq, cmd core.Cmd[any], n int, err error) {
						<-done
						return 0, nil, wantN, errors.New("transport closed")
					},
				).RegisterClose(
					func() (err error) {
						defer close(done)
						return
					},
				)
				invoker = mock.NewInvoker[any]().RegisterInvoke(
					func(ctx context.Context, seq core.Seq, at time.Time, bytesRead int,
						cmd core.Cmd[any], proxy core.Proxy,
					) (err error) {
						<-ctx.Done()
						return context.Canceled
					},
				)
				handler = New(invoker, WithCmdReceiveDuration(time.Second))
				mocks   = []*mok.Mock{cmd.Mock, transport.Mock, invoker.Mock}
			)
			go func() {
				time.Sleep(100 * time.Millisecond)
				cancel()
			}()
			err := handler.Handle(ctx, transport)
			asserterror.EqualError(err, wantErr, t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})
}
