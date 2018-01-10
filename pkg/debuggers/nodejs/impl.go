package nodejs

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/solo-io/squash/pkg/debuggers"
)

type NodeJsInterface struct{}

type nodejsDebugServer struct {
	port int
}

func (g *nodejsDebugServer) Detach() error {
	return nil
}

func (g *nodejsDebugServer) Port() int {
	return g.port
}

func (g *nodejsDebugServer) HostType() debuggers.DebugHostType {
	return debuggers.DebugHostTypeTarget
}

func (g *NodeJsInterface) Attach(pid int) (debuggers.DebugServer, error) {

	log.WithField("pid", pid).Debug("AttachToLiveSession called")
	err := syscall.Kill(pid, syscall.SIGUSR1)
	if err != nil {
		log.WithField("err", err).Error("can't send SIGUSR1 to the process")
		return nil, err
	}

	vmaj, _, err := nodeVersion(pid)
	if err != nil {
		log.WithField("err", err).Error("can't determine the NodeJS version")
		return nil, err
	}

	// Listening port after sending USR1 to node process.
	nodePort := 9229
	if vmaj < 8 {
		nodePort = 5858
	}

	gds := &nodejsDebugServer{
		port: nodePort,
	}
	return gds, nil
}

func nodeVersion(pid int) (int, int, error) {
	// run node executabl referenced by the process
	cmd := exec.Command(filepath.Join("/proc", fmt.Sprintf("%d", pid), "exe"), "-v")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, 0, err
	}
	if err = cmd.Start(); err != nil {
		return 0, 0, err
	}
	verstr, _ := ioutil.ReadAll(stdout)

	var maj, min int
	fmt.Sscanf(string(verstr), "v%d.%d", &maj, &min)
	log.WithFields(log.Fields{"major": maj, "minor": min}).Debug("NodeJS vsrsion")
	return maj, min, nil
}
