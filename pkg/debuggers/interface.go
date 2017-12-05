package debuggers

type Detachable interface {
	Detach() error
}

type DebugServer interface {
	Detachable
	Port() int
}

/// Debugger interface. implement this to add a new debugger support to squash.
type Debugger interface {

	/// Attach a debugger to pid and return the port that the debug server listens on.
	StartDebugServer(pid int) (DebugServer, error)
}
