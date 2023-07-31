package handler

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/cmd-stream/base-go"
	"github.com/cmd-stream/delegate-go"
)

// NewProxy creates a new Proxy.
func NewProxy[T any](transport delegate.ServerTransport[T]) Proxy[T] {
	var flushFlag uint32
	return Proxy[T]{transport, &flushFlag, &sync.Mutex{}}
}

// Proxy is an implementation of the base.Proxy interface.
type Proxy[T any] struct {
	transport delegate.ServerTransport[T]
	flushFlag *uint32
	mu        *sync.Mutex
}

func (p Proxy[T]) Send(seq base.Seq, result base.Result) (err error) {
	p.mu.Lock()
	err = p.transport.Send(seq, result)
	p.mu.Unlock()
	if err != nil {
		return
	}
	return p.flush()
}

func (p Proxy[T]) SendWithDeadline(deadline time.Time, seq base.Seq,
	result base.Result) (err error) {
	p.mu.Lock()
	err = p.transport.SetSendDeadline(deadline)
	if err != nil {
		p.mu.Unlock()
		return
	}
	err = p.transport.Send(seq, result)
	p.mu.Unlock()
	if err != nil {
		return
	}
	return p.flush()
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
