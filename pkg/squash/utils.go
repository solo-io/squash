package squash

import (
	"context"
	"errors"

	"github.com/solo-io/go-utils/contextutils"

	"github.com/solo-io/squash/pkg/utils/socket"
)

func GetPort(pid int) (int, error) {

	ports, err := socket.GetListeningPortsFor(pid)

	logger := contextutils.LoggerFrom(context.TODO())
	if err != nil {
		logger.Errorw("Can't get listening ports", "pid", pid, "err", err)
		return 0, err
	}

	if len(ports) == 0 {
		logger.Errorw("can't get port for pid", "pid", pid, "err", err, "ports", ports)
		return 0, errors.New("number of ports is zero")
	}

	port := ports[0]
	for _, curport := range ports[1:] {
		if port != curport {
			return 0, errors.New("more than one port and they are different")
		}
	}

	logger.Infow("port found", "pid", pid, "port", port)

	return port, nil
}
