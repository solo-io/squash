package local

import (
	"os/exec"
)

type PythonInterface struct{}

func (i *PythonInterface) GetRemoteConnectionCmd(plankName, plankNamespace, podName, podNamespace string, localPort, remotePort int) *exec.Cmd {
	// TODO
	return GetPortForwardCmd(podName, podNamespace, localPort, remotePort)
}

func (p *PythonInterface) GetEditorRemoteConnectionCmd(plankName, plankNamespace, podName, podNamespace string, remotePort int) string {
	return getPortForwardWithRandomLocalCmd(podName, podNamespace, remotePort)
}

func (d *PythonInterface) GetDebugCmd(localPort int) *exec.Cmd {
	// TODO
	return nil
}

func (d *PythonInterface) ExpectRunningPlank() bool {
	// TODO
	return false
}

func (p *PythonInterface) WindowsSupportWarning() string {
	return ""
}
