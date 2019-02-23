package local

import (
	"os/exec"
)

type NodeJsDebugger struct{}

func (g *NodeJsDebugger) GetRemoteConnectionCmd(plankName, plankNamespace, podName, podNamespace string, localPort, remotePort int) *exec.Cmd {
	// TODO
	return GetPortForwardCmd(podName, podNamespace, localPort, remotePort)
}
