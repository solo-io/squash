package remote

import (
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type JavaInterface struct{}

type javaDebugServer struct {
	port int
}

func (g *javaDebugServer) Detach() error {
	return nil
}

func (g *javaDebugServer) Port() int {
	return g.port
}

func (g *javaDebugServer) HostType() DebugHostType {
	return DebugHostTypeTarget
}

func (d *javaDebugServer) Cmd() *exec.Cmd {
	return nil
}

func (g *JavaInterface) Attach(pid int) (DebugServer, error) {

	log.WithField("pid", pid).Debug("AttachToLiveSession called")
	port, err := GetPortOfJavaProcess(pid)
	if err != nil {
		log.WithField("err", err).Error("can't get java debug port")
		return nil, err
	}

	gds := &javaDebugServer{
		port: port,
	}
	return gds, nil
}
