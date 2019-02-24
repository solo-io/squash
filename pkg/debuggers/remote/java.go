package remote

import (
	"fmt"
	"os"
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

// TODO - this should just return nil and exit, refactor with update to squashctl.getCreatedPod
func (d *javaDebugServer) Cmd() *exec.Cmd {
	cmd := exec.Command("sleep", "360000")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.WithFields(log.Fields{"cmd": cmd, "args": cmd.Args}).Debug("tmp sleep command")

	err := cmd.Start()
	fmt.Println(err)
	return cmd
	// return nil
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
