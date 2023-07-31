package handler

import "time"

// Conf is a configuration of the Handler.
//
// If the Handler has not received any commands during the ReceiveTimeout, the
// corresponding connection will be closed, if == 0 waits forever.
//
// If At == true, the Handler will call the Invoker.Invoke() with the command
// receive time.
type Conf struct {
	ReceiveTimeout time.Duration
	At             bool
}
