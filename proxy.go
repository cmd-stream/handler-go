package handler

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cmd-stream/core-go"
	dsrv "github.com/cmd-stream/delegate-go/server"
)

// NewProxy creates a new Proxy.
func NewProxy[T any](transport dsrv.Transport[T]) Proxy[T] {
	var flushFlag uint32
	return Proxy[T]{transport, &flushFlag, &sync.Mutex{}}
}

// Proxy implemets the core.Proxy interface.
type Proxy[T any] struct {
	transport dsrv.Transport[T]
	flushFlag *uint32
	mu        *sync.Mutex
}

func (p Proxy[T]) LocalAddr() net.Addr {
	return p.transport.LocalAddr()
}

func (p Proxy[T]) RemoteAddr() net.Addr {
	return p.transport.RemoteAddr()
}

func (p Proxy[T]) Send(seq core.Seq, result core.Result) (n int, err error) {
	p.mu.Lock()
	n, err = p.transport.Send(seq, result)
	p.mu.Unlock()
	if err != nil {
		return
	}
	return n, p.flush()
}

func (p Proxy[T]) SendWithDeadline(seq core.Seq, result core.Result,
	deadline time.Time) (n int, err error) {
	p.mu.Lock()
	err = p.transport.SetSendDeadline(deadline)
	if err != nil {
		p.mu.Unlock()
		return
	}
	n, err = p.transport.Send(seq, result)
	if err != nil {
		p.mu.Unlock()
		return
	}
	p.mu.Unlock()
	return n, p.flush()
}

func (p Proxy[T]) flush() (err error) {
	if swapped := atomic.CompareAndSwapUint32(p.flushFlag, 0, 1); swapped {
		p.mu.Lock()
		atomic.CompareAndSwapUint32(p.flushFlag, 1, 0)
		err = p.transport.Flush()
		p.mu.Unlock()
	}
	return
}
