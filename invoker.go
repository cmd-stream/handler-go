package handler

import (
	"context"
	"time"

	"github.com/cmd-stream/base-go"
)

// Invoker executes commands.
//
// At parameter can contain (if configured) the command receive time.
type Invoker[T any] interface {
	Invoke(ctx context.Context, at time.Time, seq base.Seq, cmd base.Cmd[T],
		proxy base.Proxy) error
}
