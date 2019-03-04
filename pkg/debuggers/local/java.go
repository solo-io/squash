package local

import (
	"fmt"
	"os"
	"os/exec"
)

type JavaInterface struct{}

func (g *JavaInterface) GetRemoteConnectionCmd(plankName, plankNamespace, podName, podNamespace string, localPort, remotePort int) *exec.Cmd {
	// connect directly to the container we are debugging
	return GetPortForwardCmd(podName, podNamespace, localPort, remotePort)
}

func (j *JavaInterface) GetEditorRemoteConnectionCmd(plankName, plankNamespace, podName, podNamespace string, remotePort int) string {
	return getPortForwardWithRandomLocalCmd(podName, podNamespace, remotePort)
}

func (d *JavaInterface) GetDebugCmd(localPort int) *exec.Cmd {
	cmd := exec.Command("jdb", "-attach", fmt.Sprintf("127.0.0.1:%v", localPort))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd
}

func (d *JavaInterface) ExpectRunningPlank() bool {
	return false
}
