package debuggers

/// Event from the debugger indicating that the inferior has stopped.
/// If Exited is true, means that the inferior process exited.
type Event struct {
	Exited bool
}

type Detachable interface {
	Detach() error
}

type DebugServer interface {
	Detachable
	Port() int
}

/// Debugger interface. implement this to add a new debugger support to squash.
type Debugger interface {
	/// Attach to process, and return a debug session object,that allows us to set breakpoints, etc..
	/// This is needed to debug in cluster mode.
	AttachTo(pid int) (LiveDebugSession, error)

	/// Attach a debugger to pid and return the port that the debug server listens on.
	StartDebugServer(pid int) (DebugServer, error)
}

/// Live debug session allowes us to set breakpoints so we can stop when an interesting event happens.
type LiveDebugSession interface {

	/// Set a breakpoint. Should be called before Continue
	SetBreakpoint(bp string) error

	/// Continue until the debugger stops. this can be due to a previously set breakpoint, crash, exit.
	/// The debugger should only stop if something usual that requires attention has happened.
	Continue() (<-chan Event, error)

	/// Only one of IntoDebugServer, Detach needs to work. preferably IntoDebugServer.
	/// If you implement one, the other can return error.

	/// Get a debug server, just as if Debugger.StartDebugServer() was called.
	/// Can't use this object after this call.
	IntoDebugServer() (DebugServer, error)

	/// Detach the debugger from the process.
	Detachable
}
