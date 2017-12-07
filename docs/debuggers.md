# Debuggers

Currently Squash support the following debuggers:
 - [gdb](docs/debuggers/gdb.md)
 - [dlv](https://github.com/derekparker/delve)

Future planes:
  - [Nodejs](https://nodejs.org/api/debugger.html)
  - [Java](http://docs.oracle.com/javase/7/docs/technotes/guides/jpda/jdwp-spec.html)

<BR>
=> We are looking for community help to add support for more debuggers.

##

**Debuggers** conform to the interface:

```go
type Debugger interface {

	/// Attach a debugger to pid and return the a debug server object
	Attach(pid int) (DebugServer, error)
}
```

Where `DebugServer` consists of the following:

```go
type DebugServer interface {
	/// Detach from the process we are debugging (allowing it to resume normal execution).
	Detach() error
	///  Return the port that the debug server listens on.
	Port() int
}
```

To add debugger support to squash, implement the functions above and add it to the squash client [main file](../../cmd/squash-client/platforms/kubernetes/main.go).

```go
func getDebugger(dbgtype string) debuggers.Debugger {
	
	var g gdb.GdbInterface
	var d dlv.DLV
	
	switch dbgtype {
	case "dlv":
		return &d
	case "gdb":
		return &g
	default:
		return nil
	}
}

```
