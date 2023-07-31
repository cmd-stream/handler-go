package handler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cmd-stream/base-go"
	base_mock "github.com/cmd-stream/base-go/testdata/mock"
	delegate_mock "github.com/cmd-stream/delegate-go/testdata/mock"
	"github.com/cmd-stream/handler-go/testdata/mock"
	"github.com/ymz-ncnk/mok"
)

const Delta = 100 * time.Millisecond

func TestHandler(t *testing.T) {

	t.Run("Handler should be able to handle several cmds and close when ctx done", func(t *testing.T) {
		var (
			ctx, cancel = context.WithCancel(context.Background())
			conf        = Conf{ReceiveTimeout: time.Second}

			wantErr          = context.Canceled
			seq1    base.Seq = 1
			seq2    base.Seq = 2

			cmd1 = base_mock.NewCmd()
			cmd2 = base_mock.NewCmd()
			cmds = map[base_mock.Cmd]struct{}{cmd1: {}, cmd2: {}}

			starTime  = time.Now()
			transport = func() (transport delegate_mock.ServerTransport) {
				done := make(chan struct{})
				transport = delegate_mock.NewServerTransport().RegisterNSetReceiveDeadline(3,
					func(deadline time.Time) (err error) {
						wantDeadline := starTime.Add(conf.ReceiveTimeout)
						if !SameTime(deadline, wantDeadline) {
							err = fmt.Errorf("Transport.Receive(), unepxected deadline, want '%v' actual '%v'",
								deadline,
								wantDeadline)
						}
						return
					},
				).RegisterReceive(
					func() (seq base.Seq, cmd base.Cmd[any], err error) {
						return seq1, cmd1, nil
					},
				).RegisterReceive(
					func() (seq base.Seq, cmd base.Cmd[any], err error) {
						return seq2, cmd2, nil
					},
				).RegisterReceive(
					func() (seq base.Seq, cmd base.Cmd[any], err error) {
						<-done
						err = errors.New("transport closed")
						return
					},
				).RegisterClose(
					func() (err error) {
						defer close(done)
						return nil
					},
				)
				return
			}()

			invoker = mock.NewInvoker[any]().RegisterInvoke(
				func(ctx context.Context, at time.Time, seq base.Seq, cmd base.Cmd[any],
					proxy base.Proxy) (err error) {
					delete(cmds, cmd.(base_mock.Cmd))
					return nil
				},
			).RegisterInvoke(
				func(ctx context.Context, at time.Time, seq base.Seq, cmd base.Cmd[any],
					proxy base.Proxy) (err error) {
					delete(cmds, cmd.(base_mock.Cmd))
					return nil
				},
			)
			handler = Handler[any]{conf, invoker}
			mocks   = []*mok.Mock{cmd1.Mock, cmd2.Mock, transport.Mock, invoker.Mock}
		)
		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()
		testHandler(ctx, handler, transport, wantErr, mocks, t)
		if len(cmds) > 0 {
			t.Error("not all cmds handled")
		}
	})

	t.Run("If Transport.SetReceiveDeadline fails with an error, Handle should return it",
		func(t *testing.T) {
			var (
				ctx, cancel = context.WithCancel(context.Background())
				wantErr     = errors.New("Transport.SetReceiveDeadline error")
				transport   = delegate_mock.NewServerTransport().RegisterSetReceiveDeadline(
					func(deadline time.Time) (err error) { return wantErr },
				).RegisterClose(
					func() (err error) { return nil },
				)
				handler = Handler[any]{Conf{ReceiveTimeout: time.Second}, nil}
				mocks   = []*mok.Mock{transport.Mock}
			)
			defer cancel()
			testHandler(ctx, handler, transport, wantErr, mocks, t)
		})

	t.Run("If Transport.Receive fails with an error, Handle should return it",
		func(t *testing.T) {
			var (
				ctx, cancel = context.WithCancel(context.Background())
				wantErr     = errors.New("Transport.Receive error")
				transport   = delegate_mock.NewServerTransport().RegisterReceive(
					func() (seq base.Seq, cmd base.Cmd[any], err error) {
						err = wantErr
						return
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				handler = Handler[any]{Conf{}, nil}
				mocks   = []*mok.Mock{transport.Mock}
			)
			defer cancel()
			testHandler(ctx, handler, transport, wantErr, mocks, t)
		})

	t.Run("If Invoker.Invoke fails with an error, Handle should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("Invoker.Invoke error")
				transport = func() (transport delegate_mock.ServerTransport) {
					done := make(chan struct{})
					transport = delegate_mock.NewServerTransport().RegisterReceive(
						func() (seq base.Seq, cmd base.Cmd[any], err error) {
							return 1, base_mock.NewCmd(), nil
						},
					).RegisterReceive(
						func() (seq base.Seq, cmd base.Cmd[any], err error) {
							<-done
							return 0, nil, errors.New("transport closed")
						},
					).RegisterClose(
						func() (err error) {
							defer close(done)
							return nil
						},
					)
					return
				}()
				invoker = mock.NewInvoker[any]().RegisterInvoke(
					func(ctx context.Context, at time.Time, seq base.Seq,
						cmd base.Cmd[any], proxy base.Proxy) (err error) {
						return wantErr
					},
				)
				handler = Handler[any]{Conf{}, invoker}
				mocks   = []*mok.Mock{transport.Mock, invoker.Mock}
			)
			testHandler(context.Background(), handler, transport, wantErr, mocks, t)
		})

	t.Run("If Conf.At == true, Invoker.Invoke should receive not empty 'at' param",
		func(t *testing.T) {
			var (
				ctx, cancel = context.WithCancel(context.Background())
				wantAt      time.Time
				mu          sync.Mutex
				transport   = func() (transport delegate_mock.ServerTransport) {
					done := make(chan struct{})
					transport = delegate_mock.NewServerTransport().RegisterReceive(
						func() (seq base.Seq, cmd base.Cmd[any], err error) {
							mu.Lock()
							wantAt = time.Now()
							mu.Unlock()
							return 1, base_mock.NewCmd(), nil
						},
					).RegisterReceive(
						func() (seq base.Seq, cmd base.Cmd[any], err error) {
							<-done
							return 0, nil, errors.New("transport closed")
						},
					).RegisterClose(
						func() (err error) {
							defer close(done)
							return nil
						},
					)
					return
				}()
				invoker = mock.NewInvoker[any]().RegisterInvoke(
					func(ctx context.Context, at time.Time, seq base.Seq,
						cmd base.Cmd[any], proxy base.Proxy) (err error) {
						defer cancel()
						mu.Lock()
						if !SameTime(at, wantAt) {
							err = fmt.Errorf("unexpected 'at' param, want '%v' actual '%v'",
								wantAt,
								at)
						}
						mu.Unlock()
						return
					},
				)
				handler = Handler[any]{Conf{At: true}, invoker}
				mocks   = []*mok.Mock{transport.Mock, invoker.Mock}
			)

			testHandler(ctx, handler, transport, context.Canceled, mocks, t)

		})

	t.Run("We should be able to interupt Handler, while it invoke a command and is locked on ctx",
		func(t *testing.T) {
			var (
				conf = Conf{
					ReceiveTimeout: time.Second,
				}
				ctx, cancel = context.WithCancel(context.Background())
				wantErr     = context.Canceled
				cmd         = base_mock.NewCmd()
				transport   = func() (transport delegate_mock.ServerTransport) {
					done := make(chan struct{})
					transport = delegate_mock.NewServerTransport().RegisterNSetReceiveDeadline(2,
						func(deadline time.Time) (err error) { return nil },
					).RegisterReceive(
						func() (seq base.Seq, cmd base.Cmd[any], err error) { return 1, cmd, nil },
					).RegisterReceive(
						func() (seq base.Seq, cmd base.Cmd[any], err error) {
							<-done
							return 0, nil, errors.New("transport closed")
						},
					).RegisterClose(
						func() (err error) {
							defer close(done)
							return
						},
					)
					return
				}()
				invoker = mock.NewInvoker[any]().RegisterInvoke(
					func(ctx context.Context, at time.Time, seq base.Seq, cmd base.Cmd[any], proxy base.Proxy) (err error) {
						<-ctx.Done()
						return context.Canceled
					},
				)
				handler = Handler[any]{conf, invoker}
				mocks   = []*mok.Mock{cmd.Mock, transport.Mock, invoker.Mock}
			)
			go func() {
				time.Sleep(100 * time.Millisecond)
				cancel()
			}()
			testHandler(ctx, handler, transport, wantErr, mocks, t)
		})

}

func SameTime(t1, t2 time.Time) bool {
	return !(t1.Before(t2.Truncate(Delta)) || t1.After(t2.Add(Delta)))
}

func testHandler(ctx context.Context, handler Handler[any],
	transport delegate_mock.ServerTransport,
	wantErr error,
	mocks []*mok.Mock,
	t *testing.T,
) {
	err := handler.Handle(ctx, transport)
	if err != wantErr {
		t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
	}
	if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
		t.Error(infomap)
	}
}
