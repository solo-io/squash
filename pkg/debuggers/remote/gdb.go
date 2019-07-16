package remote

import (
	"context"
	"os/exec"
	"syscall"

	"github.com/solo-io/go-utils/contextutils"

	"fmt"
	"time"
)

type GdbInterface struct{}

type gdbDebugServer struct {
	port int
	cmd  *exec.Cmd
}

func (g *gdbDebugServer) Detach() error {
	g.cmd.Process.Signal(syscall.SIGINT)
	return nil
}

func (g *gdbDebugServer) Port() int {
	return g.port
}

func (g *gdbDebugServer) HostType() DebugHostType {
	return DebugHostTypeClient
}

func (g *gdbDebugServer) Cmd() *exec.Cmd {
	return g.cmd
}

func (g *GdbInterface) Attach(pid int) (DebugServer, error) {

	logger := contextutils.LoggerFrom(context.TODO())
	logger.Debugw("AttachToLiveSession called", "pid", pid)
	cmd := exec.Command("gdbserver", "--attach", ":0", fmt.Sprintf("%d", pid))
	cmd.Start()
	logger.Debug("starting gdbserver for user started, trying to get port")
	time.Sleep(time.Second)
	port, err := GetPort(cmd.Process.Pid)
	if err != nil {
		logger.Errorw("can't get gdbserver port", "err", err)
		cmd.Process.Kill()
		cmd.Process.Release()
		return nil, err
	}

	// be polite and wait
	go cmd.Wait()
	gds := &gdbDebugServer{
		port: port,
		cmd:  cmd,
	}
	return gds, nil
}
