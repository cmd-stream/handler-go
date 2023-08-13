# handler-go
handler-go was designed for cmd-stream-go to handle client connections on the
server.

It contains an implementation of the `delegate.ServerTransportHandler` 
interface, which executes each command in its own gorountine using `Invoker`.

# Tests
Test coverage is about 96%.