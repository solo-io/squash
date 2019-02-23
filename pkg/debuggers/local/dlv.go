package local

import (
	"os/exec"
)

type DLV struct {
}

func (d *DLV) GetRemoteConnectionCmd(plankName, plankNamespace, podName, podNamespace string, localPort, remotePort int) *exec.Cmd {
	// for dlv, we proxy through the debug container
	return GetPortForwardCmd(plankName, plankNamespace, localPort, remotePort)
}
