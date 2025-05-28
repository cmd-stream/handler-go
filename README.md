# handler-go

[![Go Reference](https://pkg.go.dev/badge/github.com/cmd-stream/handler-go.svg)](https://pkg.go.dev/github.com/cmd-stream/handler-go)
[![GoReportCard](https://goreportcard.com/badge/cmd-stream/handler-go)](https://goreportcard.com/report/github.com/cmd-stream/handler-go)
[![codecov](https://codecov.io/gh/cmd-stream/handler-go/graph/badge.svg?token=04UEO65CLJ)](https://codecov.io/gh/cmd-stream/handler-go)

handler-go provides the connection handler for the cmd-stream server.

It implements the `server.TransportHandler` interface from the delegate-go 
module and executes each command in its own goroutine using the `Invoker`.
