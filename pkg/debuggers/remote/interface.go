package remote

import (
	"os/exec"
)

// DebugHostType - type of host to connect debugger
type DebugHostType int

const (
	// DebugHostTypeClient - debugger needs to connect to squash-client
	DebugHostTypeClient DebugHostType = iota
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
	// Return the cmd representing the debugger process, if any
	Cmd() *exec.Cmd
}

/// Debugger interface. implement this to add a new debugger support to squash.
type Remote interface {

	/// Attach a debugger to pid and return the a debug server object
	Attach(pid int, env map[string]string) (DebugServer, error)
}
