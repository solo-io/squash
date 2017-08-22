package gdb

import (
	"os/exec"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/solo-io/squash/pkg/debuggers"

	"context"
	"fmt"
	"strings"
	"time"

	"errors"

	"github.com/solo-io/gdb"
)

type GdbError map[string]interface{}

func (e GdbError) Error() string {

	if msg, ok := e["msg"]; ok {
		if smsg, ok := msg.(string); ok {
			return smsg
		}
	}
	return fmt.Sprintf("%#v", e)
}

func retToErr(ret Event) error {

	if ret.Class() != "error" {
		return nil
	}
	payload := ret.Payload()
	if payload == nil {
		return errors.New("unknown gdb error")
	}
	return GdbError(payload)
}

type Event map[string]interface{}

func (e Event) Class() string {
	return e.getStringProp("class")
}

func (e Event) Payload() map[string]interface{} {
	if p, ok := e["payload"]; ok {
		if p1, ok := p.(map[string]interface{}); ok {
			return p1
		}
	}
	return nil
}

func (e Event) getStringProp(s string) string {
	if p, ok := e[s]; ok {
		if sp, ok := p.(string); ok {
			return sp
		}
	}
	return ""
}

type GDB struct {
	C         <-chan Event
	debugger  *gdb.Gdb
	logger    *log.Entry
	eventchan chan debuggers.Event
}

func attachTo(pid int) (debuggers.LiveDebugSession, error) {
	c := make(chan Event, 1000)
	g := &GDB{
		C:         c,
		logger:    log.WithField("pid", pid),
		eventchan: make(chan debuggers.Event),
	}
	d, err := gdb.New(g.notify(c))
	if err != nil {
		g.logger.WithField("err", err).Error("gdb new")
		return nil, err
	}
	g.debugger = d

	_, err = g.SendShort("target-attach", fmt.Sprintf("%d", pid))
	if err != nil {
		log.WithField("err", err).Error("target attach error")
		g.debugger.Exit()
		return nil, err
	}

	go g.waitForEventWorker()

	return g, nil
}

func (g *GDB) IntoDebugServer() (debuggers.DebugServer, error) {
	return nil, errors.New("gdb doesn't support reconnecting. detach and re-attach instead")
}

func (g *GDB) SendShort(operation string, arguments ...string) (Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return g.Send(ctx, operation, arguments...)
}

func (g *GDB) Send(ctx context.Context, operation string, arguments ...string) (Event, error) {
	g.logger.WithFields(log.Fields{"operation": operation, "arguments": arguments}).Info("Sending command to gdb")

	r1, err := g.debugger.SendWithContext(ctx, operation, arguments...)

	r := Event(r1)
	if err != nil {
		g.logger.WithField("err", err).WithField("cmd", operation).WithField("args", arguments).Error("Send command error")
		return r, err
	}

	return r, retToErr(r)
}

type Reason string

func (r Reason) HasExited() bool {
	// see here: https://www-zeuthen.desy.de/dv/documentation/unixguide/infohtml/gdb/GDB_002fMI-Async-Records.html
	return strings.HasPrefix(string(r), "exited")
}

func (g *GDB) Detach() error {
	_, err := g.SendShort("target-detach")
	if err != nil {
		log.WithField("err", err).Error("target detach error")
		g.debugger.Exit()
		return err
	}
	return nil
}

func (g *GDB) SetBreakpoint(bp string) error {
	_, err := g.Break(bp)
	return err
}

func (g *GDB) waitForEventWorker() {

	g.logger.Infof("waiting for gdb to stop")
	for e := range g.C {
		g.logger.WithField("event", e).Info("Got debug event")
		switch e.Class() {
		case "stopped":
			exited := false
			if r, ok := e.Payload()["reason"]; ok {
				if rs, ok := r.(string); ok {
					reason := Reason(rs)
					exited = reason.HasExited()
				}
			}
			g.eventchan <- debuggers.Event{Exited: exited}
			close(g.eventchan)

		default:
		}

	}
}

func (g *GDB) waitForStop() (Event, Reason) {
	g.logger.Infof("waiting for gdb to stop")
	for e := range g.C {
		g.logger.WithField("event", e).Info("Got debug event")
		if e.Class() == "stopped" {
			if r, ok := e.Payload()["reason"]; ok {
				if rs, ok := r.(string); ok {
					return e, Reason(rs)
				}
			}
			return e, Reason("")

		}
	}
	return nil, Reason("")
}

func (g *GDB) start() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	_, err := g.Send(ctx, "exec-run")
	if err != nil {
		return err
	}
	return nil
}

func (g *GDB) Continue() (<-chan debuggers.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	_, err := g.Send(ctx, "exec-continue")
	if err != nil {
		return nil, err
	}
	return g.eventchan, nil
}

func (g *GDB) Break(linespec string) (string, error) {
	event, err := g.SendShort("break-insert", linespec)
	if err != nil {
		return "", err
	}
	bkpt := (event.Payload()["bkpt"]).(map[string]interface{})

	return bkpt["number"].(string), nil
}

func (g *GDB) notify(c chan Event) func(map[string]interface{}) {
	return func(notification map[string]interface{}) {
		g.logger.WithField("notification", notification).Info("got notification")
		c <- notification
	}
}

type GdbInterface struct{}

func (g *GdbInterface) AttachTo(pid int) (debuggers.LiveDebugSession, error) {
	return attachTo(pid)
}

type gdbDebugServer struct {
	port int
	cmd  *exec.Cmd
}

func (g *gdbDebugServer) Detach() error {
	g.cmd.Process.Signal(syscall.SIGINT)
	return nil
}

func (g *gdbDebugServer) Port() int {
	return g.port
}

func (g *GdbInterface) StartDebugServer(pid int) (debuggers.DebugServer, error) {

	log.WithField("pid", pid).Debug("AttachToLiveSession called")
	cmd := exec.Command("gdbserver", "--attach", ":0", fmt.Sprintf("%d", pid))
	cmd.Start()
	log.Debug("starting gdbserver for user started, trying to get port")
	time.Sleep(time.Second)
	port, err := debuggers.GetPort(cmd.Process.Pid)
	if err != nil {
		log.WithField("err", err).Error("can't get gdbserver port")
		cmd.Process.Kill()
		cmd.Process.Release()
		return nil, err
	}

	// be polite and wait
	go cmd.Wait()
	gds := &gdbDebugServer{
		port: port,
		cmd:  cmd,
	}
	return gds, nil
}
