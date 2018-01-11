package nodejs

import (
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/solo-io/squash/pkg/debuggers"
)

const (
	nodeJSDebuggerPort  = 5858
	nodeJSInspectorPort = 9229
)

type NodeJsDebugger struct {
	isInspectorEnabled bool
}

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

func (g *NodeJsDebugger) Attach(pid int) (debuggers.DebugServer, error) {

	log.WithField("pid", pid).Debug("AttachToLiveSession called")
	err := syscall.Kill(pid, syscall.SIGUSR1)
	if err != nil {
		log.WithField("err", err).Error("can't send SIGUSR1 to the process")
		return nil, err
	}
	gds := &nodejsDebugServer{}

	if g.isInspectorEnabled {
		gds.port = nodeJSInspectorPort
	} else {
		gds.port = nodeJSDebuggerPort
	}
	return gds, nil
}

func (g *NodeJsDebugger) EnableInspector(i bool) {
	g.isInspectorEnabled = i
}

func enableDebugger(pid int) error {
	return nil
}
