package debuggers

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
	"github.com/solo-io/squash/pkg/client"
	"github.com/solo-io/squash/pkg/client/debugconfig"
	"github.com/solo-io/squash/pkg/client/debugsessions"
	"github.com/solo-io/squash/pkg/models"
	"github.com/solo-io/squash/pkg/platforms"
)

func RunSquashClient(debugger func(string) Debugger, conttopid platforms.Container2Pid) error {
	log.SetLevel(log.DebugLevel)

	customFormatter := new(log.TextFormatter)
	log.SetFormatter(customFormatter)

	log.Info("Squash Client started")

	server := flag.String("server", os.Getenv("SERVERURL"), "")

	flag.Parse()

	log.WithField("server", *server).Info("handleAttachment")
	u, err := url.Parse(*server)
	if err != nil {
		log.WithField("err", err).Error("RunDebugBridge")
		return err

	}
	cfg := &client.TransportConfig{
		BasePath: path.Join(u.Path, client.DefaultBasePath),
		Host:     u.Host,
		Schemes:  []string{u.Scheme},
	}
	log.WithField("cfg", cfg).Debug("creating client")
	client := client.NewHTTPClientWithConfig(nil, cfg)

	return NewDebugHandler(client, debugger, conttopid).handleAttachments()
}

type DebugHandler struct {
	debugger  func(string) Debugger
	conttopid platforms.Container2Pid
	client    *client.Squash
	debugees  map[int]bool
}

func NewDebugHandler(client *client.Squash, debugger func(string) Debugger,
	conttopid platforms.Container2Pid) *DebugHandler {
	return &DebugHandler{
		client:    client,
		debugger:  debugger,
		conttopid: conttopid,
		debugees:  make(map[int]bool),
	}

}

func getNodeName() string {
	return os.Getenv("NODE_NAME")
}

func (d *DebugHandler) handleAttachments() error {
	for {
		err := d.handleAttachment()
		if err != nil {
			log.WithField("err", err).Warn("error watching for attached container")
		}
	}
}

func (d *DebugHandler) handleAttachment() error {
	ci, err := d.watchForAttached()
	if err != nil {
		log.WithField("err", err).Warn("error watching for attached container")
		return err
	}
	err = retry(func() error { return d.tryToAttach(ci) })

	if err != nil {

		// put an empty debug session to make the debug config inactive

		params := debugsessions.NewPutDebugSessionParams()
		params.Body = &models.DebugSession{
			DebugConfigID: &ci.ID,
		}
		params.DebugConfigID = ci.ID

		log.WithFields(log.Fields{"DebugConfigID": params.DebugConfigID}).Debug("Failed to attach... signaling server.")

		_, err := d.client.Debugsessions.PutDebugSession(params)
		if err != nil {
			log.WithField("err", err).Warn("Error adding debug session for failure!")
		} else {
			log.Info("debug config set inactive!")
		}
	}
	return err
}

func retry(f func() error) error {
	tries := 3
	for i := 0; i < (tries - 1); i++ {
		if err := f(); err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return f()
}

func (d *DebugHandler) tryToAttach(ci *models.DebugConfig) error {

	// make sure this is not a duplicate

	pid, err := d.conttopid.GetPid(context.Background(), *ci.Attachment.Name)

	if err != nil {
		log.WithField("err", err).Warn("FindFirstProcess error")
		return err
	}

	log.WithField("app", ci).Info("Attaching to live session")

	p, err := os.FindProcess(pid)
	if err != nil {
		log.WithField("err", err).Error("can't find process")
		return err
	}
	if !d.debugees[pid] {
		log.WithField("pid", pid).Info("starting to debug")
		d.debugees[pid] = true
		go d.startDebug(ci, p)
	} else {
		log.WithField("pid", pid).Warn("Already debugging pid. ignoring")
	}
	return nil
}

func (d *DebugHandler) waitForErrorAndStop(ci *models.DebugConfig, curdebugger Debugger, p *os.Process) (DebugServer, error) {
	logger := log.WithField("pid", p.Pid)
	logger.Info("Waiting for error")
	session, err := curdebugger.AttachTo(p.Pid)
	if err != nil {
		log.WithField("err", err).Error("can't attach process")
		return nil, err
	}
	var port DebugServer
	defer func() {
		if port == nil {
			session.Detach()
		}
	}()

	logger.Info("Setting breakpoints")

	for _, bp := range ci.Breakpoints {
		err := session.SetBreakpoint(*bp.Location)
		if err != nil {
			return nil, err
		}
		logger.WithField("bp", *bp.Location).Info("Breakpoint set")
	}

	logger.Info("Continuing")
	stopch, err := session.Continue()
	if err != nil {
		log.WithField("err", err).Error("can't continue process")
		return nil, err
	}
	logger.Info("Waiting for process to stop")

	e := <-stopch
	logger.Info("Process stopped!")
	if e.Exited {
		return nil, errors.New("process exited")
	}

	port, err = session.IntoDebugServer()
	if err != nil {
		p.Signal(syscall.SIGSTOP)
	}

	// detach also delete all breakpoints in gdb, so that's good.
	return port, nil
}

func (d *DebugHandler) startDebug(ci *models.DebugConfig, p *os.Process) {
	log.Info("start debug called")

	shouldWaitForError := !ci.Immediately

	curdebugger := d.debugger(ci.Debugger)

	log.WithFields(log.Fields{"curdebugger": ci.Debugger, "shouldWaitForError": shouldWaitForError}).Info("start debug params")

	shouldthaw := false
	var port DebugServer
	if shouldWaitForError {
		var err error
		port, err = d.waitForErrorAndStop(ci, curdebugger, p)
		if err != nil {
			log.WithField("err", err).Warn("Error waiting for error!")
			return // err
		}
		if port == nil {
			shouldthaw = true
		}
	}

	if port == nil {
		log.WithFields(log.Fields{"pid": p.Pid}).Info("starting debug server")
		var err error
		port, err = curdebugger.StartDebugServer(p.Pid)
		if err != nil {
			log.WithField("err", err).Error("Starting debug server error")
			return // err
		}
		log.WithField("port", port).Info("StartDebugServer return dbg server port")
		if shouldthaw {

			log.WithField("pid", p.Pid).Info("StartDebugServer - should thaw")
			p.Signal(syscall.SIGCONT)
		}
	}
	log.WithField("pid", p.Pid).Info("StartDebugServer - posting debug session")

	params := debugsessions.NewPutDebugSessionParams()
	params.Body = &models.DebugSession{
		URL:           fmt.Sprintf("%s:%d", os.Getenv("HOST_ADDR"), port.Port()),
		DebugConfigID: &ci.ID,
	}
	params.DebugConfigID = ci.ID

	log.WithFields(log.Fields{"URL": params.Body.URL, "DebugConfigID": params.DebugConfigID}).Debug("Trying to add debug session!")

	_, err := d.client.Debugsessions.PutDebugSession(params)
	if err != nil {
		log.WithField("err", err).Warn("Error adding debug session - detaching!")
		port.Detach()
	} else {
		log.Info("debug session added!")
	}
}

func (d *DebugHandler) watchForAttached() (*models.DebugConfig, error) {
	for {
		params := debugconfig.NewPopContainerToDebugParams()
		params.Node = getNodeName()
		log.WithField("params", params).Debug("watchForAttached - calling PopContainerToDebug")

		resp, err := d.client.Debugconfig.PopContainerToDebug(params)

		if _, ok := err.(*debugconfig.PopContainerToDebugRequestTimeout); ok {
			continue
		}

		if err != nil {
			log.WithField("err", err).Warn("watchForAttached - error calling function:")
			time.Sleep(time.Second)
			continue
		}

		ci := resp.Payload

		if *ci.Attachment.Type != models.AttachmentTypeContainer {
			log.WithField("ci", spew.Sdump(ci)).Warn("watchForAttached - recieved bad attachment")
			continue
		}
		log.WithField("ci", spew.Sdump(ci)).Info("watchForAttached - got debug config!")

		return ci, nil
	}
}
