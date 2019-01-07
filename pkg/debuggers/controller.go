package debuggers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/platforms"
	"github.com/solo-io/squash/pkg/utils"
)

type DebugController struct {
	debugger         func(string) Debugger
	conttopid        platforms.ContainerProcess
	udpateAttachment func(*v1.DebugAttachment) error

	pidLock sync.Mutex
	pidMap  map[int]bool

	debugattachmentsLock sync.Mutex
	debugattachments     map[string]debugAttachmentData
}

type debugAttachmentData struct {
	debugger DebugServer
	pid      int
}

func NewDebugController(debugger func(string) Debugger,
	udpateAttachment func(*v1.DebugAttachment) error,
	conttopid platforms.ContainerProcess) *DebugController {
	return &DebugController{
		debugger:         debugger,
		conttopid:        conttopid,
		udpateAttachment: udpateAttachment,

		pidMap: make(map[int]bool),

		debugattachments: make(map[string]debugAttachmentData),
	}
}

func (d *DebugController) lockProcess(pid int) bool {
	d.pidLock.Lock()
	defer d.pidLock.Unlock()
	oldpid := d.pidMap[pid]
	d.pidMap[pid] = true
	return !oldpid
}

func (d *DebugController) unlockProcess(pid int) {
	d.pidLock.Lock()
	defer d.pidLock.Unlock()
	delete(d.pidMap, pid)
}

func (d *DebugController) addActiveAttachment(attachment *v1.DebugAttachment, pid int, debugger DebugServer) {
	d.debugattachmentsLock.Lock()
	defer d.debugattachmentsLock.Unlock()
	d.debugattachments[attachment.Metadata.Name] = debugAttachmentData{debugger, pid}
}

func (d *DebugController) removeAttachment(name string) {
	d.debugattachmentsLock.Lock()
	data, ok := d.debugattachments[name]
	delete(d.debugattachments, name)
	d.debugattachmentsLock.Unlock()

	if ok {
		log.WithFields(log.Fields{"attachment.Name": name}).Debug("Detaching attachment")
		err := data.debugger.Detach()
		if err != nil {
			log.WithFields(log.Fields{"attachment.Name": name, "err": err}).Debug("Error detaching")
		}
		d.unlockProcess(data.pid)
	}
}

func (d *DebugController) HandleAddedRemovedAttachments(attachments, removedAtachment []*v1.DebugAttachment) error {

	for _, attachment := range removedAtachment {
		log.WithFields(log.Fields{"attachment.Name": attachment.Metadata.Name}).Debug("Removing attachment")
		d.removeAttachment(attachment.Metadata.Name)
	}

	for _, attachment := range attachments {
		// notify the server that we are attaching, so we won't get the same attachment object next time.
		// TODO(mitchdraft) - this should set a flag on the debug attachment object
		// if err := d.notifyState(attachment, v1.DebugAttachmentStatusStateAttaching); err != nil {
		// 	log.WithFields(log.Fields{"attachment.Name": attachment.Metadata.Name, "err": err}).Debug("Failed set state to attaching in squash server. aborting.")

		// 	d.notifyError(attachment)
		// }
		fmt.Println("att:")
		fmt.Println(attachment)
		go d.handleSingleAttachment(attachment)
	}
	return nil
}

func (d *DebugController) handleSingleAttachment(attachment *v1.DebugAttachment) {

	fmt.Println("att2:")
	fmt.Println(attachment)
	err := retry(func() error { return d.tryToAttach(attachment) })

	if err != nil {
		log.WithFields(log.Fields{"attachment.Name": attachment.Metadata.Name}).Debug("Failed to attach... signaling server.")
		// TODO replace swagger functionality
		fmt.Println("swagger functionality update")
		// d.notifyError(attachment)
	}
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

func FindFirstProcess(pids []int, processName string) (int, error) {
	log.WithField("processName", processName).Debug("Finding process to debug")
	minpid := 0
	var mintime *time.Time
	for _, pid := range pids {
		p := filepath.Join("/proc", fmt.Sprintf("%d", pid))
		n, err := os.Stat(p)
		if err != nil {
			log.WithFields(log.Fields{"pid": pid, "err": err}).Info("Failed to stat the process, skipping")
			continue
		}
		ss, err := utils.GetCmdArgsByPid(pid)
		if err != nil || len(ss) < 1 {
			log.WithFields(log.Fields{"pid": pid, "err": err}).Info("Failed to get command args for the process, skipping")
			continue
		}

		currentProcessName := ss[0]

		// take the base name
		currentProcessName = filepath.Base(currentProcessName)
		log.WithFields(log.Fields{"pid": pid, "currentProcessName": currentProcessName}).Debug("checking")

		if processName == "" || strings.EqualFold(currentProcessName, processName) {
			t := n.ModTime()
			if (mintime == nil) || t.Before(*mintime) {
				mintime = &t
				minpid = pid
			}
		}
	}

	if minpid == 0 {
		return 0, errors.New("no process found")
	}
	return minpid, nil
}

func (d *DebugController) tryToAttach(attachment *v1.DebugAttachment) error {

	fmt.Println("att3:")
	fmt.Println(attachment)
	// make sure this is not a duplicate
	ci, err := d.conttopid.GetContainerInfo(context.Background(), attachment)
	fmt.Println("ci")
	fmt.Println(ci)
	if err != nil {
		log.WithField("err", err).Warn("GetContainerInfo error")
		return err
	}

	pid, err := FindFirstProcess(ci.Pids, attachment.ProcessName)
	if err != nil {
		log.WithField("err", err).Warn("FindFirstProcess error")
		return err
	}

	log.WithField("app", attachment).Info("Attaching to live session")

	p, err := os.FindProcess(pid)
	if err != nil {
		log.WithField("err", err).Error("can't find process")
		return err
	}
	if d.lockProcess(pid) {
		log.WithField("pid", pid).Info("starting to debug")
		debugger, err := d.startDebug(attachment, p, ci.Name)
		if err != nil {
			// TODO - replace swagger functionality
			fmt.Println("swagger functionality update")
			// d.notifyError(attachment)
			return nil // no retry
		}
		d.addActiveAttachment(attachment, pid, debugger)

	} else {
		log.WithField("pid", pid).Warn("Already debugging pid. ignoring")
		// TODO
		fmt.Println("swagger functionality update")
		// d.notifyError(attachment)
	}
	return nil
}

// func (d *DebugController) notifyError(attachment *v1.DebugAttachment) {
// 	d.notifyState(attachment, v1.DebugAttachmentStatusStateError)
// }

// func (d *DebugController) notifyState(attachment *v1.DebugAttachment, newstate string) error {

// 	attachmentCopy := *attachment
// 	attachmentCopy.Status.State = newstate
// 	return d.udpateAttachment(&attachmentCopy)
// }

func (d *DebugController) startDebug(attachment *v1.DebugAttachment, p *os.Process, targetName string) (DebugServer, error) {
	fmt.Println("att4:")
	fmt.Println(attachment)
	log.Info("start debug called")

	curdebugger := d.debugger(attachment.Debugger)

	if curdebugger == nil {
		return nil, errors.New("debugger doesn't exist")
	}

	log.WithFields(log.Fields{"curdebugger": attachment.Debugger}).Info("start debug params")

	log.WithFields(log.Fields{"pid": p.Pid}).Info("starting debug server")
	var err error
	debugServer, err := curdebugger.Attach(p.Pid)

	if err != nil {
		log.WithField("err", err).Error("Starting debug server error")
		return nil, err
	}

	log.WithField("pid", p.Pid).Info("StartDebugServer - posting debug session")

	attachmentPatch := &v1.DebugAttachment{
		Metadata:   attachment.Metadata,
		Attachment: attachment.Attachment,
		Debugger:   attachment.Debugger,
	}

	hostName := ""
	switch debugServer.HostType() {
	case DebugHostTypeTarget:
		hostName = targetName
	case DebugHostTypeClient:
		hostName = os.Getenv("HOST_ADDR")
	}

	if len(hostName) == 0 {
		err = fmt.Errorf("Cannot find Host name for type: %d", debugServer.HostType())
		log.WithField("err", err).Error("Starting debug server error")
		return nil, err
	}

	attachmentPatch.DebugServerAddress = fmt.Sprintf("%s:%d", hostName, debugServer.Port())
	attachmentPatch.State = "attached" // TODO(mitchdraft) make this an enum in the api

	log.WithFields(log.Fields{"newattachment": attachmentPatch}).Debug("Notifying server of attachment to debug config object")
	err = d.udpateAttachment(attachmentPatch)

	if err != nil {
		log.WithField("err", err).Warn("Error adding debug session - detaching!")
		debugServer.Detach()
	} else {
		log.Info("debug session added!")
	}
	return debugServer, nil
}
