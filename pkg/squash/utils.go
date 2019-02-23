package squash

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/solo-io/squash/pkg/utils/socket"
)

func GetPort(pid int) (int, error) {

	ports, err := socket.GetListeningPortsFor(pid)

	if err != nil {
		log.WithFields(log.Fields{"pid": pid, "err": err}).Error("Can't get listening ports")
		return 0, err
	}

	if len(ports) == 0 {
		log.WithFields(log.Fields{"pid": pid, "err": err, "ports": ports}).Error("can't get port for pid")
		return 0, errors.New("Number of ports is zero")
	}

	port := ports[0]
	for _, curport := range ports[1:] {
		if port != curport {
			return 0, errors.New("More than one port and they are different")
		}
	}

	log.WithFields(log.Fields{"pid": pid, "port": port}).Info("port found")

	return port, nil
}
