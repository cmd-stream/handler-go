package handler

import "time"

// Conf is a configuration of the Handler.
//
// If the Handler does not receive any commands during CmdReceiveTimeout, the
// corresponding connection will be closed. A timeout value of 0 means it will
// wait indefinitely.
// If At is set to true, the Handler will call Invoker.Invoke() with the
// command's reception time.
type Conf struct {
	CmdReceiveTimeout time.Duration
	At                bool
}
