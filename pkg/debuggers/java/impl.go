package java

import (
	"fmt"
	"os/exec"

	log "github.com/sirupsen/logrus"
	"github.com/solo-io/squash/pkg/debuggers"
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

func (g *javaDebugServer) HostType() debuggers.DebugHostType {
	return debuggers.DebugHostTypeTarget
}

func (d *javaDebugServer) Cmd() *exec.Cmd {
	return nil
}

func (g *JavaInterface) Attach(pid int) (debuggers.DebugServer, error) {

	log.WithField("pid", pid).Debug("AttachToLiveSession called")
	port, err := debuggers.GetPortOfJavaProcess(pid)
	fmt.Println("just got port it is:")
	fmt.Println(port)
	if err != nil {
		fmt.Println("err", err)
		fmt.Println("can't get jsadebugd port")
		log.WithField("err", err).Error("can't get jsadebugd port")
		return nil, err
	}

	gds := &javaDebugServer{
		port: port,
	}
	return gds, nil
}
