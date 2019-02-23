package local

import "os/exec"

/// Debugger interface. implement this to add a new debugger support to squash.
type Local interface {

	// TODO - refactor this to use v2 DA api
	// (since all of these args belong in the DA spec)
	GetRemoteConnectionCmd(plankName, plankNamespace, podName, podNamespace string, localPort, remotePort int) *exec.Cmd

	GetDebugCmd(localPort int) *exec.Cmd

	// ExpectRunningPod indicates if this local debugger should be paired with an active plank pod
	ExpectRunningPlank() bool
}

// candidate alternative:
type PortSpec struct {
	PlankName      string
	PlankNamespace string
	PodName        string
	PodNamespace   string
	LocalPort      int
	PlankPort      int
	PodPort        int
}
