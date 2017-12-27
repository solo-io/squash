package java

import (
	log "github.com/Sirupsen/logrus"
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

func (g *javaDebugServer) PodType() debuggers.DebugPodType {
	return debuggers.DebugPodTypeTarget
}

func (g *JavaInterface) Attach(pid int) (debuggers.DebugServer, error) {

	log.WithField("pid", pid).Debug("AttachToLiveSession called")
	port, err := debuggers.GetPortOfJavaProcess(pid)
	if err != nil {
		log.WithField("err", err).Error("can't get jsadebugd port")
		return nil, err
	}

	gds := &javaDebugServer{
		port: port,
	}
	return gds, nil
}
