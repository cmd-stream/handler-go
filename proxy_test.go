package handler

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/cmd-stream/core-go"
	cmock "github.com/cmd-stream/core-go/testdata/mock"
	dmock "github.com/cmd-stream/delegate-go/server/testdata/mock"
	asserterror "github.com/ymz-ncnk/assert/error"
)

func TestProxy(t *testing.T) {

	t.Run("Send should work correctly", func(t *testing.T) {
		var (
			wantSeq    core.Seq = 1
			wantResult          = cmock.NewResult()
			wantN      int      = 1
			wantErr    error    = nil
			transport           = dmock.NewTransport().RegisterSend(
				func(seq core.Seq, result core.Result) (n int, err error) {
					asserterror.Equal(seq, wantSeq, t)
					asserterror.EqualDeep(result, wantResult, t)
					return wantN, wantErr
				},
			).RegisterFlush(
				func() (err error) { return wantErr },
			)
			proxy  = NewProxy(transport)
			n, err = proxy.Send(wantSeq, wantResult)
		)
		asserterror.Equal(n, wantN, t)
		asserterror.Equal(err, wantErr, t)
	})

	t.Run("If Transport.Send fails with an error, Send should return it",
		func(t *testing.T) {
			var (
				wantN     int = 2
				wantErr       = errors.New("Transport.Send error")
				transport     = dmock.NewTransport().RegisterSend(
					func(seq core.Seq, result core.Result) (n int, err error) {
						return wantN, wantErr
					},
				)
				proxy  = NewProxy(transport)
				n, err = proxy.Send(1, cmock.NewResult())
			)
			asserterror.Equal(n, wantN, t)
			asserterror.Equal(err, wantErr, t)
		})

	t.Run("If Transport.Flush fails with an error, Send should return it",
		func(t *testing.T) {
			var (
				wantN     int = 3
				wantErr       = errors.New("Transport.Flush error")
				transport     = dmock.NewTransport().RegisterSend(
					func(seq core.Seq, result core.Result) (n int, err error) {
						return wantN, nil
					},
				).RegisterFlush(
					func() (err error) { return wantErr },
				)
				proxy  = NewProxy(transport)
				n, err = proxy.Send(1, cmock.NewResult())
			)
			asserterror.Equal(n, wantN, t)
			asserterror.Equal(err, wantErr, t)
		})

	t.Run("SendWithDeadline should work correctly", func(t *testing.T) {
		var (
			wantDeadline          = time.Now()
			wantSeq      core.Seq = 1
			wantResult            = cmock.NewResult()
			wantN        int      = 4
			wantErr      error    = nil
			transport             = dmock.NewTransport().RegisterSetSendDeadline(
				func(deadline time.Time) (err error) {
					if deadline != wantDeadline {
						err = fmt.Errorf("unexpectd deadline, want '%v' actual '%v'",
							wantDeadline,
							deadline)
					}
					return
				},
			).RegisterSend(
				func(seq core.Seq, result core.Result) (n int, err error) {
					asserterror.Equal(seq, wantSeq, t)
					asserterror.EqualDeep(result, wantResult, t)
					return wantN, nil
				},
			).RegisterFlush(
				func() (err error) { return wantErr },
			)
			proxy  = NewProxy(transport)
			n, err = proxy.SendWithDeadline(wantSeq, wantResult, wantDeadline)
		)
		asserterror.Equal(n, wantN, t)
		asserterror.Equal(err, wantErr, t)
	})

	t.Run("If Transport.SetSendDeadline fails with an error, SendWithDeadline should return it",
		func(t *testing.T) {
			var (
				wantN     int = 0
				wantErr       = errors.New("Transport.SetSendDeadline error")
				transport     = dmock.NewTransport().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) {
						return wantErr
					},
				)
				proxy  = NewProxy(transport)
				n, err = proxy.SendWithDeadline(1, cmock.NewResult(), time.Now())
			)
			asserterror.Equal(n, wantN, t)
			asserterror.Equal(err, wantErr, t)
		})

	t.Run("If Transport.Send fails with an error, SendWithDeadline should return it",
		func(t *testing.T) {
			var (
				wantN     int = 5
				wantErr       = errors.New("Transport.Send error")
				transport     = dmock.NewTransport().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) { return nil },
				).RegisterSend(
					func(seq core.Seq, result core.Result) (n int, err error) { return wantN, wantErr },
				)
				proxy  = NewProxy(transport)
				n, err = proxy.SendWithDeadline(1, cmock.NewResult(), time.Now())
			)
			asserterror.Equal(n, wantN, t)
			asserterror.Equal(err, wantErr, t)
		})

	t.Run("If Transport.Flush fails with an error, SendWithDeadline should return it",
		func(t *testing.T) {
			var (
				wantN     int = 6
				wantErr       = errors.New("Transport.Flush error")
				transport     = dmock.NewTransport().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) { return nil },
				).RegisterSend(
					func(seq core.Seq, result core.Result) (n int, err error) { return wantN, nil },
				).RegisterFlush(
					func() (err error) { return wantErr },
				)
				proxy  = NewProxy(transport)
				n, err = proxy.SendWithDeadline(1, cmock.NewResult(), time.Now())
			)
			asserterror.Equal(n, wantN, t)
			asserterror.Equal(err, wantErr, t)
		})

}
