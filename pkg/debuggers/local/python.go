package local

import (
	"os/exec"
)

type PythonInterface struct{}

func (i *PythonInterface) GetRemoteConnectionCmd(plankName, plankNamespace, podName, podNamespace string, localPort, remotePort int) *exec.Cmd {
	// TODO
	return GetPortForwardCmd(podName, podNamespace, localPort, remotePort)
}
