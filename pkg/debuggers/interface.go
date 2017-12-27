package debuggers

// DebugPodType - type of pod to connect debugger
type DebugPodType int

const (
	// DebugPodTypeClient - debugger needs to connect to squash-client pod
	DebugPodTypeClient = 0
	// DebugPodTypeTarget - debugger needs to connect to target pod
	DebugPodTypeTarget = 1
)

type DebugServer interface {
	/// Detach from the process we are debugging (allowing it to resume normal execution).
	Detach() error
	/// Returns either DebugPodTypeClient or DebugPodTypeTarget
	PodType() DebugPodType
	/// Return the port that the debug server listens on.
	Port() int
}

/// Debugger interface. implement this to add a new debugger support to squash.
type Debugger interface {

	/// Attach a debugger to pid and return the a debug server object
	Attach(pid int) (DebugServer, error)
}
