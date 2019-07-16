package remote

import (
	"context"
	"os/exec"
	"syscall"

	"github.com/solo-io/go-utils/contextutils"
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

func (g *nodejsDebugServer) HostType() DebugHostType {
	return DebugHostTypeTarget
}

func (d *nodejsDebugServer) Cmd() *exec.Cmd {
	return nil
}

func (g *nodejsDebugServer) Attach(pid int) (DebugServer, error) {

	logger := contextutils.LoggerFrom(context.TODO())
	logger.Debugw("AttachToLiveSession called", "pid", pid)
	err := syscall.Kill(pid, syscall.SIGUSR1)
	if err != nil {
		logger.Errorw("can't send SIGUSR1 to the process", "err", err)
		return nil, err
	}
	return g, nil
}
