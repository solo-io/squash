package debuggers

type DebugServer interface {
	/// Detach from the process we are debugging (allowing it to resume normal execution).
	Detach() error
	///  Return the port that the debug server listens on.
	Port() int
}

/// Debugger interface. implement this to add a new debugger support to squash.
type Debugger interface {

	/// Attach a debugger to pid and return the a debug server object
	Attach(pid int) (DebugServer, error)
}
