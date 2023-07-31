package handler

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/cmd-stream/base-go"
	base_mock "github.com/cmd-stream/base-go/testdata/mock"
	delegate_mock "github.com/cmd-stream/delegate-go/testdata/mock"
)

func TestProxy(t *testing.T) {

	t.Run("Send should work correctly", func(t *testing.T) {
		var (
			wantSeq    base.Seq = 1
			wantResult          = base_mock.NewResult()
			transport           = delegate_mock.NewServerTransport().RegisterSend(
				func(seq base.Seq, result base.Result) (err error) {
					if seq != wantSeq {
						err = fmt.Errorf("unexpectd seq, want '%v' actual '%v'", wantSeq, seq)
					}
					if result != wantResult {
						err = fmt.Errorf("unexpectd result, want '%v' actual '%v'",
							wantResult,
							result)
					}
					return
				},
			).RegisterFlush(
				func() (err error) { return nil },
			)
			proxy = NewProxy[any](transport)
			err   = proxy.Send(wantSeq, wantResult)
		)
		if err != nil {
			t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
		}
	})

	t.Run("If Transport.Send fails with an error, Send should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("Transport.Send error")
				transport = delegate_mock.NewServerTransport().RegisterSend(
					func(seq base.Seq, result base.Result) (err error) {
						return wantErr
					},
				)
				proxy = NewProxy[any](transport)
				err   = proxy.Send(1, base_mock.NewResult())
			)
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
		})

	t.Run("If Transport.Flush fails with an error, Send should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("Transport.Flush error")
				transport = delegate_mock.NewServerTransport().RegisterSend(
					func(seq base.Seq, result base.Result) (err error) {
						return nil
					},
				).RegisterFlush(
					func() (err error) { return wantErr },
				)
				proxy = NewProxy[any](transport)
				err   = proxy.Send(1, base_mock.NewResult())
			)
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
		})

	t.Run("SendWithDeadline should work correctly", func(t *testing.T) {
		var (
			wantDeadline          = time.Now()
			wantSeq      base.Seq = 1
			wantResult            = base_mock.NewResult()
			transport             = delegate_mock.NewServerTransport().RegisterSetSendDeadline(
				func(deadline time.Time) (err error) {
					if deadline != wantDeadline {
						err = fmt.Errorf("unexpectd deadline, want '%v' actual '%v'",
							wantDeadline,
							deadline)
					}
					return
				},
			).RegisterSend(
				func(seq base.Seq, result base.Result) (err error) {
					if seq != wantSeq {
						return fmt.Errorf("unexpectd seq, want '%v' actual '%v'", wantSeq, seq)
					}
					if result != wantResult {
						return fmt.Errorf("unexpectd result, want '%v' actual '%v'",
							wantResult,
							result)
					}
					return nil
				},
			).RegisterFlush(
				func() (err error) { return nil },
			)
			proxy = NewProxy[any](transport)
			err   = proxy.SendWithDeadline(wantDeadline, wantSeq, wantResult)
		)
		if err != nil {
			t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
		}
	})

	t.Run("If Transport.SetSendDeadline fails with an error, SendWithDeadline should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("Transport.SetSendDeadline error")
				transport = delegate_mock.NewServerTransport().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) {
						return wantErr
					},
				)
				proxy = NewProxy[any](transport)
				err   = proxy.SendWithDeadline(time.Now(), 1, base_mock.NewResult())
			)
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
		})

	t.Run("If Transport.Send fails with an error, SendWithDeadline should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("Transport.Send error")
				transport = delegate_mock.NewServerTransport().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) { return nil },
				).RegisterSend(
					func(seq base.Seq, result base.Result) (err error) { return wantErr },
				)
				proxy = NewProxy[any](transport)
				err   = proxy.SendWithDeadline(time.Now(), 1, base_mock.NewResult())
			)
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
		})

	t.Run("If Transport.Flush fails with an error, SendWithDeadline should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("Transport.Flush error")
				transport = delegate_mock.NewServerTransport().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) { return nil },
				).RegisterSend(
					func(seq base.Seq, result base.Result) (err error) { return nil },
				).RegisterFlush(
					func() (err error) { return wantErr },
				)
				proxy = NewProxy[any](transport)
				err   = proxy.SendWithDeadline(time.Now(), 1, base_mock.NewResult())
			)
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
		})

}
