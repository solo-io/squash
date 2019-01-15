package dlv

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/derekparker/delve/service/rpc1"
	log "github.com/sirupsen/logrus"
	"github.com/solo-io/squash/pkg/debuggers"
)

type DLV struct {
}

type DLVLiveDebugSession struct {
	client  *rpc1.RPCClient
	port    int
	process *os.Process
}

func (d *DLVLiveDebugSession) Detach() error {
	d.client.Detach(false)
	d.process.Kill()
	return nil
}

func (d *DLVLiveDebugSession) Port() int {
	return d.port
}

func (d *DLVLiveDebugSession) HostType() debuggers.DebugHostType {
	return debuggers.DebugHostTypeClient
}

func (d *DLV) attachTo(pid int) (*DLVLiveDebugSession, error) {
	cmd, port, err := startDebugServer(pid)

	if cmd != nil {
		go cmd.Wait()
	}
	if err != nil {
		return nil, err
	}
	// use rpc1 client for vscode extension support
	client := rpc1.NewClient(fmt.Sprintf("localhost:%d", port))
	dls := &DLVLiveDebugSession{
		client:  client,
		port:    port,
		process: cmd.Process,
	}
	return dls, nil
}

func (d *DLV) Attach(pid int) (debuggers.DebugServer, error) {
	return d.attachTo(pid)
}

func startDebugServer(pid int) (*exec.Cmd, int, error) {

	log.WithField("pid", pid).Debug("StartDebugServer called")
	cmd := exec.Command("dlv", "attach", fmt.Sprintf("%d", pid), "--listen=127.0.0.1:0", "--accept-multiclient=true", "--api-version=2", "--headless", "--log")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.WithFields(log.Fields{"cmd": cmd, "args": cmd.Args}).Debug("dlv command")

	err := cmd.Start()
	if err != nil {
		log.WithField("err", err).Error("Failed to start dlv")
		return nil, 0, err
	}

	log.Debug("starting headless dlv for user started, trying to get port")
	time.Sleep(2 * time.Second)
	log.Debug("awake now")
	port, err := debuggers.GetPort(cmd.Process.Pid)
	if err != nil {
		log.WithField("err", err).Error("can't get headless dlv port")
		cmd.Process.Kill()
		cmd.Process.Release()
		return cmd, 0, err
	}

	return cmd, port, nil
}
