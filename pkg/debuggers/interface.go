package debuggers

type DebugServer interface {
	Detach() error
	Port() int
}

/// Debugger interface. implement this to add a new debugger support to squash.
type Debugger interface {

	/// Attach a debugger to pid and return the port that the debug server listens on.
	Attach(pid int) (DebugServer, error)
}
