package local

import (
	"os/exec"
)

type GdbInterface struct{}

func (g *GdbInterface) GetRemoteConnectionCmd(plankName, plankNamespace, podName, podNamespace string, localPort, remotePort int) *exec.Cmd {
	// proxy through the debug container
	return GetPortForwardCmd(plankName, plankNamespace, localPort, remotePort)
}

func (g *GdbInterface) GetEditorRemoteConnectionCmd(plankName, plankNamespace, podName, podNamespace string, remotePort int) string {
	return getPortForwardWithRandomLocalCmd(plankName, plankNamespace, remotePort)
}

func (d *GdbInterface) GetDebugCmd(localPort int) *exec.Cmd {
	// TODO
	return nil
}

func (d *GdbInterface) ExpectRunningPlank() bool {
	// TODO
	return false
}

func (g *GdbInterface) WindowsSupportWarning() string {
	return "Squash does not currently support the gdb interactive terminal on Windows. Please use the vscode extension or pass the --machine flag to squashctl."
}
