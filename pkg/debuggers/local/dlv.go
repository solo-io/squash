package local

import (
	"fmt"
	"os"
	"os/exec"
)

type DLV struct {
}

func (d *DLV) GetRemoteConnectionCmd(plankName, plankNamespace, podName, podNamespace string, localPort, remotePort int) *exec.Cmd {
	// for dlv, we proxy through the debug container
	return GetPortForwardCmd(plankName, plankNamespace, localPort, remotePort)
}

func (d *DLV) GetEditorRemoteConnectionCmd(plankName, plankNamespace, podName, podNamespace string, remotePort int) string {
	// for dlv, we proxy through the debug container
	return getPortForwardWithRandomLocalCmd(plankName, plankNamespace, remotePort)
}

func (d *DLV) GetDebugCmd(localPort int) *exec.Cmd {
	cmd := exec.Command("dlv", "connect", fmt.Sprintf("127.0.0.1:%v", localPort))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd
}

func (d *DLV) ExpectRunningPlank() bool {
	return true
}
