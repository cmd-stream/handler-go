# handler-go
handler-go was designed to handle client connections on the cmd-stream server 
for Golang. 

It contains a `delegate.ServerTransportHandler` implementation, which executes 
each command, with help of the `Invoker`, in its own gorountine.

# Tests
Test coverage is about 96%.