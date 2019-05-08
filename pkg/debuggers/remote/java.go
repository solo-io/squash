package remote

import (
	"context"
	"os/exec"

	"github.com/solo-io/go-utils/contextutils"
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

	logger := contextutils.LoggerFrom(context.TODO())
	logger.Debugw("AttachToLiveSession called", "pid", pid)
	port, err := GetPortOfJavaProcess(pid)
	if err != nil {
		logger.Errorw("can't get java debug port", "err", err)
		return nil, err
	}

	gds := &javaDebugServer{
		port: port,
	}
	return gds, nil
}
