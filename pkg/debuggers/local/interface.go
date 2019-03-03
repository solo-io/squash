package local

import "os/exec"

/// Debugger interface. implement this to add a new debugger support to squash.
type Local interface {

	// TODO - refactor this to use v2 DA api
	// (since all of these args belong in the DA spec)
	GetRemoteConnectionCmd(plankName, plankNamespace, podName, podNamespace string, localPort, remotePort int) *exec.Cmd

	// returns the kubectl port-forward command to be called by the editor extension
	GetEditorRemoteConnectionCmd(plankName, plankNamespace, podName, podNamespace string, remotePort int) string

	GetDebugCmd(localPort int) *exec.Cmd

	// ExpectRunningPod indicates if this local debugger should be paired with an active plank pod
	ExpectRunningPlank() bool
}
