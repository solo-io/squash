package nodejs

import (
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/solo-io/squash/pkg/debuggers"
)

const (
	DebuggerPort  = 5858
	InspectorPort = 9229
)

type NodeJsDebugger struct{}

type nodejsDebugServer struct {
	port int
}

func NewNodeDebugger(p int) *nodejsDebugServer {
	return &nodejsDebugServer{port: p}
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

func (g *nodejsDebugServer) Attach(pid int) (debuggers.DebugServer, error) {

	log.WithField("pid", pid).Debug("AttachToLiveSession called")
	err := syscall.Kill(pid, syscall.SIGUSR1)
	if err != nil {
		log.WithField("err", err).Error("can't send SIGUSR1 to the process")
		return nil, err
	}
	return g, nil
}
