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

	AttachTo(pid int) (LiveDebugSession, error)
	StartDebugServer(pid int) (DebugServer, error)
}
```

Where `DebugServer` consists of the following:

```go
type DebugServer interface {

	Detachable
	Port() int
}
```

and `Detachable` consists of the following:

```go
type Detachable interface {

	Detach() error
}
```

and `LiveDebugSession` consists of the following:

```go
type LiveDebugSession interface {

	SetBreakpoint(bp string) error
  	Continue() (<-chan Event, error)
  	IntoDebugServer() (DebugServer, error)
  	Detachable
}
```

To add debugger support to squash, implement the functions and add it to the squash client [main file](../../cmd/squash-client/platforms/kubernetes/main.go).

```go
func getDebugger(dbgtype string) debuggers.Debugger {
	
	var g gdb.GdbInterface
	var d dlv.DLV
	
	switch dbgtype {
	case "dlv":
		return &d
	case "gdb":
		fallthrough
	default:
		return &g
	}
}

```
