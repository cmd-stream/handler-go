package handler

import "time"

// Deprecated: use github.com/cmd-stream/cmd-stream-go instead.
type Options struct {
	CmdReceiveDuration time.Duration
	At                 bool
}

// WithCmdReceiveDuration sets the maximum time the Handler will wait for a
// command. If no command arrives within this duration, the connection is
// closed. A duration of 0 means the Handler will wait indefinitely.
//
// Deprecated: use github.com/cmd-stream/cmd-stream-go instead.
func WithCmdReceiveDuration(d time.Duration) SetOption {
	return func(o *Options) { o.CmdReceiveDuration = d }
}

// WithAt enables the "at" flag. When enabled, the Handler passes the command's
// received timestamp to Invoker.Invoke().
//
// Deprecated: use github.com/cmd-stream/cmd-stream-go instead.
func WithAt() SetOption {
	return func(o *Options) { o.At = true }
}

// Deprecated: use github.com/cmd-stream/cmd-stream-go instead.
type SetOption func(o *Options)

// Deprecated: use github.com/cmd-stream/cmd-stream-go instead.
func Apply(ops []SetOption, o *Options) {
	for i := range ops {
		if ops[i] != nil {
			ops[i](o)
		}
	}
}
