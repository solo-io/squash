package local

import (
	"fmt"
	"os"
	"os/exec"
)

type JavaPortInterface struct{}

func (g *JavaPortInterface) GetRemoteConnectionCmd(plankName, plankNamespace, podName, podNamespace string, localPort, remotePort int) *exec.Cmd {
	// connect directly to the container we are debugging
	return GetPortForwardCmd(podName, podNamespace, localPort, remotePort)
}

func (j *JavaPortInterface) GetEditorRemoteConnectionCmd(plankName, plankNamespace, podName, podNamespace string, remotePort int) string {
	return getPortForwardWithRandomLocalCmd(podName, podNamespace, remotePort)
}

func (d *JavaPortInterface) GetDebugCmd(localPort int) *exec.Cmd {
	fmt.Printf("Java debug port available on local port %v.\n", localPort)
	// TODO(mitchdraft) - do this in a less hacky way
	cmd := exec.Command("sleep", "200000")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd
}

func (d *JavaPortInterface) ExpectRunningPlank() bool {
	return false
}
