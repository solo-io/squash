package debuggers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/solo-io/squash/pkg/utils"
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

func GetPortOfJavaProcess(pid int) (int, error) {

	args, err := utils.GetCmdArgsByPid(pid)
	if err != nil {
		log.WithFields(log.Fields{"pid": pid, "err": err}).Error("Can't get command line arguments")
		return 0, err
	}
	for _, arg := range args {
		if strings.HasPrefix(arg, "-agentlib") || strings.HasPrefix(arg, "-Xrunjdwp") {
			ss := strings.Split(arg, ",")
			for _, s := range ss {
				if strings.HasPrefix(s, "address") {
					a := strings.Split(s, "=")
					if len(a) > 1 {
						port, err := strconv.Atoi(a[1])
						if err == nil {
							// Got the port number
							return port, nil
						}
					}
					break
				}
			}
			break
		}
	}
	return 0, fmt.Errorf("Can't find port in java command line arguments for PID: %d", pid)
}
