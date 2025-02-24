package handler

import "time"

// Conf configures the Handler.
//
//   - If the Handler does not receive any Commands within the CmdReceiveDuration
//     period, the corresponding connection will be closed. A CmdReceiveDuration
//     value of 0 means the Handler will wait indefinitely for Commands.
//   - If the At flag is set to true, the Handler will invoke Invoker.Invoke()
//     with the timestamp of when the Command was received.
type Conf struct {
	CmdReceiveDuration time.Duration
	At                 bool
}
