# handler-go

[![Go Reference](https://pkg.go.dev/badge/github.com/cmd-stream/handler-go.svg)](https://pkg.go.dev/github.com/cmd-stream/handler-go)
[![GoReportCard](https://goreportcard.com/badge/cmd-stream/handler-go)](https://goreportcard.com/report/github.com/cmd-stream/handler-go)
[![codecov](https://codecov.io/gh/cmd-stream/handler-go/graph/badge.svg?token=04UEO65CLJ)](https://codecov.io/gh/cmd-stream/handler-go)

handler-go was designed for cmd-stream-go to handle client connections on the
server.

It contains an implementation of the `delegate.ServerTransportHandler` 
interface, which executes each command in its own gorountine using `Invoker`.
