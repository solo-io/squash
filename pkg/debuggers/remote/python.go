package remote

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
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

func (d *ptvsdDebugServer) HostType() DebugHostType {
	return DebugHostTypeTarget
}

func (d *ptvsdDebugServer) Cmd() *exec.Cmd {
	return nil
}

func (i *PythonInterface) Attach(pid int) (DebugServer, error) {

	logger := contextutils.LoggerFrom(context.TODO())
	logger.Debugw("AttachToLiveSession called", "pid", pid)
	port, err := getPtvsdPort(pid)
	if err != nil {
		logger.Errorw("can't get ptvsd port", "err", err)
		return nil, err
	}

	logger.Debugw("Found python debug port", "pid", pid, "port", port)
	ds := &ptvsdDebugServer{
		port: port,
	}
	return ds, nil
}

// Search /proc/{PID}/cwd for file with "ptvsd.enable_attach" string and extracts port from it
// Example of config: ptvsd.enable_attach("my_secret", address = ('0.0.0.0', 3000)), returns 3000
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
	contextutils.LoggerFrom(context.TODO()).Debugw("searching root", "root", root)
	re := regexp.MustCompile(PtvsdSearchString)

	werr := filepath.Walk(root+"/", func(p string, fi os.FileInfo, err error) error {
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

		// String was found in a file - all errors from now on are reported
		f.Seek(int64(idxs[0]), io.SeekStart)
		b := make([]byte, idxs[1]-idxs[0])
		_, err = f.Read(b)
		if err != nil {
			return err
		}
		args := strings.Split(string(b), ",")
		if len(args) > 2 {
			_, err = fmt.Sscanf(args[2], "%d", &port)
			if err != nil {
				return err
			}
		}
		// Terminate walk
		return errors.New("")
	})

	if port == 0 {
		if werr != nil {
			return 0, fmt.Errorf("%s is not found. Error: %s", PtvsdSearchString, werr)
		}
		return 0, fmt.Errorf("%s is not found in python sources", PtvsdSearchString)
	}
	return port, nil
}
