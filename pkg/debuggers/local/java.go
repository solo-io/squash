package local

import (
	"os/exec"
)

type JavaInterface struct{}

func (g *JavaInterface) GetRemoteConnectionCmd(plankName, plankNamespace, podName, podNamespace string, localPort, remotePort int) *exec.Cmd {
	// connect directly to the container we are debugging
	return GetPortForwardCmd(podName, podNamespace, localPort, remotePort)
}
