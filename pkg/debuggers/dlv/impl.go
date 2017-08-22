package dlv

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/derekparker/delve/service/api"
	"github.com/derekparker/delve/service/rpc1"
	"github.com/solo-io/squash/pkg/debuggers"
)

type DLV struct {
}

type DLVLiveDebugSession struct {
	client *rpc1.RPCClient
	events chan interface{}
	port   int
}

func (d *DLVLiveDebugSession) Events() <-chan interface{} {
	return d.events
}

func (d *DLVLiveDebugSession) SetBreakpoint(bp string) error {
	scope := api.EvalScope{
		Frame:       0,
		GoroutineID: -1,
	}
	l, err := d.client.FindLocation(scope, bp)

	if err != nil {
		return err
	}
	if len(l) != 1 {
		return errors.New("abigous location")
	}
	location := l[0]
	bpapi := &api.Breakpoint{
		Addr: location.PC,
	}
	// TODO: should we save and clear the breakpoints on detach?
	_, err = d.client.CreateBreakpoint(bpapi)

	if err != nil {
		return err
	}
	return nil
}

func (d *DLVLiveDebugSession) Continue() (<-chan debuggers.Event, error) {
	ch := make(chan debuggers.Event)
	go func() {
		ch <- debuggers.Event{Exited: (<-d.client.Continue()).Exited}
		close(ch)
	}()

	return ch, nil
}

func (d *DLVLiveDebugSession) Detach() error {
	d.client.Detach(false)
	return nil
}

func (d *DLVLiveDebugSession) IntoDebugServer() (debuggers.DebugServer, error) {
	return d, nil
}

func (d *DLVLiveDebugSession) Port() int {
	return d.port
}

func (d *DLV) AttachTo(pid int) (debuggers.LiveDebugSession, error) {
	return d.attachTo(pid)
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
		client: client,
		events: make(chan interface{}),
		port:   port,
	}
	return dls, nil
}

func (d *DLV) StartDebugServer(pid int) (debuggers.DebugServer, error) {
	return d.attachTo(pid)
}

func startDebugServer(pid int) (*exec.Cmd, int, error) {

	log.WithField("pid", pid).Debug("StartDebugServer called")
	cmd := exec.Command("dlv", "attach", fmt.Sprintf("%d", pid), "--listen=127.0.0.1:0", "--accept-multiclient=true", "--headless", "--log")
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
	port, err := debuggers.GetPort(cmd.Process.Pid)
	if err != nil {
		log.WithField("err", err).Error("can't get headless dlv port")
		cmd.Process.Kill()
		cmd.Process.Release()
		return cmd, 0, err
	}

	return cmd, port, nil
}
