package debuggers

// DebugHostType - type of host to connect debugger
type DebugHostType int

const (
	// DebugHostTypeClient - debugger needs to connect to squash-client
	DebugHostTypeClient = iota
	// DebugHostTypeTarget - debugger needs to connect to target
	DebugHostTypeTarget
)

type DebugServer interface {
	/// Detach from the process we are debugging (allowing it to resume normal execution).
	Detach() error
	/// Returns either DebugHostTypeClient or DebugHostTypeTarget
	HostType() DebugHostType
	/// Return the port that the debug server listens on.
	Port() int
}

/// Debugger interface. implement this to add a new debugger support to squash.
type Debugger interface {

	/// Attach a debugger to pid and return the a debug server object
	Attach(pid int) (DebugServer, error)
}
