package local

import (
	"fmt"
	"os"
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
	// TODO - will it work?
	fmt.Printf("Node debug port available on local port %v.\n", localPort)
	// TODO(mitchdraft) - do this in a less hacky way
	cmd := exec.Command("sleep", "200000")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd
}

func (d *NodeJsDebugger) ExpectRunningPlank() bool {
	// TODO
	return false
}

func (n *NodeJsDebugger) WindowsSupportWarning() string {
	return ""
}
