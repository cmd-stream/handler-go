package handler

import "time"

// Conf is a configuration of Handler.
//
// If Handler has not received any commands during the ReceiveTimeout, the
// corresponding connection will be closed, if == 0 waits forever.
// If At == true, Handler will call Invoker.Invoke() with the command receive
// time.
type Conf struct {
	ReceiveTimeout time.Duration
	At             bool
}
