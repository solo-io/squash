package python

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/solo-io/squash/pkg/debuggers"
)

const (
	PtvsdPortEnvVariable  = "PTVSD_PORT_NUMBER"
	PtvsdSearchString     = `ptvsd\.enable_attach.*`
	PtvsdMaxFileSize      = 1024 * 1024
	PtvsdMaxNumberOfFiles = 1000
)

type PythonInterface struct{}

type ptvsdDebugServer struct {
	port int
}

func (d *ptvsdDebugServer) Detach() error {
	return nil
}

func (d *ptvsdDebugServer) Port() int {
	return d.port
}

func (d *ptvsdDebugServer) HostType() debuggers.DebugHostType {
	return debuggers.DebugHostTypeTarget
}

func (i *PythonInterface) Attach(pid int) (debuggers.DebugServer, error) {

	log.WithField("pid", pid).Debug("AttachToLiveSession called")
	port, err := getPtvsdPort(pid)
	if err != nil {
		log.WithField("err", err).Error("can't get ptvsd port")
		return nil, err
	}

	log.WithFields(log.Fields{"pid": pid, "port": port}).Debug("Found python debug port")
	ds := &ptvsdDebugServer{
		port: port,
	}
	return ds, nil
}

// Search /proc/{PID}/cwd for file with "ptvsd.enable_attach" string and extracts port from it
func getPtvsdPort(pid int) (int, error) {
	port := 0
	fileNum := 0
	// Try environment var first
	pe := os.Getenv(PtvsdPortEnvVariable)
	if len(pe) >= 1 {
		fmt.Sscanf(pe, "%d", &port)
		return port, nil
	}

	root := filepath.Join("/proc", fmt.Sprintf("%d", pid), "cwd")
	log.WithField("root", root).Debug("searching root")
	re := regexp.MustCompile(PtvsdSearchString)

	filepath.Walk(root+"/", func(p string, fi os.FileInfo, err error) error {
		if err != nil || fi == nil || fi.IsDir() {
			return nil
		}

		if filepath.Ext(p) != ".py" {
			return nil
		}
		f, err := os.Open(p)
		if err != nil {
			return nil
		}
		defer f.Close()

		idxs := re.FindReaderIndex(bufio.NewReader(io.LimitReader(f, PtvsdMaxFileSize)))
		if idxs == nil {
			fileNum++
			if fileNum >= PtvsdMaxNumberOfFiles {
				return errors.New("File limit reached")
			}
			return nil
		}
		f.Seek(int64(idxs[0]), 0)
		b := make([]byte, idxs[1]-idxs[0])
		_, err = f.Read(b)
		if err != nil {
			return nil
		}
		args := strings.Split(string(b), ",")
		if len(args) > 2 {
			fmt.Sscanf(args[2], "%d", &port)
		}
		return errors.New("")
	})

	if port == 0 {
		return 0, fmt.Errorf("%s is not found in python sources", PtvsdSearchString)
	}
	return port, nil
}
