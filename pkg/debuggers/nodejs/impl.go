package nodejs

import (
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/solo-io/squash/pkg/debuggers"
)

type NodeJsInterface struct{}
type NodeJsInterface8 struct{}

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
	err := enableDebugger(pid)
	if err != nil {
		return nil, err
	}
	gds := &nodejsDebugServer{
		port: 5858,
	}
	return gds, nil
}

func (g *NodeJsInterface8) Attach(pid int) (debuggers.DebugServer, error) {
	err := enableDebugger(pid)
	if err != nil {
		return nil, err
	}
	gds := &nodejsDebugServer{
		port: 9229,
	}
	return gds, nil
}

func enableDebugger(pid int) error {
	log.WithField("pid", pid).Debug("AttachToLiveSession called")
	err := syscall.Kill(pid, syscall.SIGUSR1)
	if err != nil {
		log.WithField("err", err).Error("can't send SIGUSR1 to the process")
		return err
	}
	return nil
}
