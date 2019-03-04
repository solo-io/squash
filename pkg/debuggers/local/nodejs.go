package local

import (
	"os/exec"
)

type NodeJsDebugger struct{}

func (g *NodeJsDebugger) GetRemoteConnectionCmd(plankName, plankNamespace, podName, podNamespace string, localPort, remotePort int) *exec.Cmd {
	// TODO
	return GetPortForwardCmd(podName, podNamespace, localPort, remotePort)
}

func (n *NodeJsDebugger) GetEditorRemoteConnectionCmd(plankName, plankNamespace, podName, podNamespace string, remotePort int) string {
	return getPortForwardWithRandomLocalCmd(podName, podNamespace, remotePort)
}

func (d *NodeJsDebugger) GetDebugCmd(localPort int) *exec.Cmd {
	// TODO
	return nil
}

func (d *NodeJsDebugger) ExpectRunningPlank() bool {
	// TODO
	return false
}
