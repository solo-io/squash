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
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/squash/pkg/api/v1"
	lite "github.com/solo-io/squash/pkg/lite/kube"
	"github.com/solo-io/squash/pkg/platforms"
	"github.com/solo-io/squash/pkg/utils"
	"github.com/solo-io/squash/pkg/utils/kubeutils"
)

type DebugController struct {
	debugger  func(string) Debugger
	conttopid platforms.ContainerProcess
	pidLock   sync.Mutex
	pidMap    map[int]bool

	daClient *v1.DebugAttachmentClient
	ctx      context.Context

	debugattachmentsLock sync.Mutex
	debugattachments     map[string]debugAttachmentData

	liteMode      bool
	inClusterMode bool
}

type debugAttachmentData struct {
	debugger DebugServer
	pid      int
}

func NewDebugController(ctx context.Context,
	debugger func(string) Debugger,
	daClient *v1.DebugAttachmentClient,
	conttopid platforms.ContainerProcess,
	liteMode bool,
	inClusterMode bool) *DebugController {
	return &DebugController{
		debugger:  debugger,
		conttopid: conttopid,

		daClient: daClient,

		pidMap: make(map[int]bool),

		debugattachments: make(map[string]debugAttachmentData),

		liteMode:      liteMode,
		inClusterMode: inClusterMode,
	}
}

func (d *DebugController) lockProcess(pid int) bool {
	log.WithFields(log.Fields{"pidMap": d.pidMap}).Debug("locking process")
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

func (d *DebugController) addActiveAttachment(da *v1.DebugAttachment, pid int, debugger DebugServer) error {
	d.debugattachmentsLock.Lock()
	defer d.debugattachmentsLock.Unlock()
	d.debugattachments[da.Metadata.Name] = debugAttachmentData{debugger, pid}
	d.markAsAttached(da.Metadata.Namespace, da.Metadata.Name)
	return nil
}

func (d *DebugController) removeAttachment(namespace, name string) {
	d.debugattachmentsLock.Lock()
	d.markForDeletion(namespace, name)
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

func (d *DebugController) handleAttachmentRequest(da *v1.DebugAttachment) {

	// Mark attachment as in progress
	da.State = v1.DebugAttachment_PendingAttachment
	_, err := (*d.daClient).Write(da, clients.WriteOpts{OverwriteExisting: true})
	if err != nil {
		log.WithFields(log.Fields{"da.Name": da.Metadata.Name, "da.Namespace": da.Metadata.Namespace}).Warn("Failed to update attachment status.")
	}
	if d.liteMode {
		// TODO - put in a goroutine
		err = d.tryToAttachPod(da)
	} else {
		err = retry(func() error { return d.tryToAttach(da) })
	}
	if err != nil {
		log.WithFields(log.Fields{"da.Name": da.Metadata.Name, "da.Namespace": da.Metadata.Namespace, "error": err}).Warn("Failed to attach debugger, deleting request.")
		d.markForDeletion(da.Metadata.Namespace, da.Metadata.Name)
	}

}

func (d *DebugController) markForDeletion(namespace, name string) {
	log.WithFields(log.Fields{"namespace": namespace, "name": name}).Debug("marking for deletion")
	da, err := (*d.daClient).Read(namespace, name, clients.ReadOpts{Ctx: d.ctx})
	if err != nil {
		// should not happen, but if it does, the CRD was probably already deleted
		log.WithFields(log.Fields{"da.Name": da.Metadata.Name, "da.Namespace": da.Metadata.Namespace}).Warn("Failed to read attachment prior to delete.")
	}

	da.State = v1.DebugAttachment_PendingDelete

	_, err = (*d.daClient).Write(da, clients.WriteOpts{
		Ctx:               d.ctx,
		OverwriteExisting: true,
	})
	if err != nil {
		log.WithFields(log.Fields{"da.Name": da.Metadata.Name, "da.Namespace": da.Metadata.Namespace}).Warn("Failed to mark attachment for deletion.")
	}
}

func (d *DebugController) deleteResource(namespace, name string) {
	err := (*d.daClient).Delete(namespace, name, clients.DeleteOpts{Ctx: d.ctx, IgnoreNotExist: true})
	if err != nil {
		log.WithFields(log.Fields{"name": name, "namespace": namespace, "error": err}).Warn("Failed to delete resource.")
	}
}

func (d *DebugController) markAsAttached(namespace, name string) {
	da, err := (*d.daClient).Read(namespace, name, clients.ReadOpts{Ctx: d.ctx})
	if err != nil {
		log.WithFields(log.Fields{"da.Name": da.Metadata.Name, "da.Namespace": da.Metadata.Namespace}).Warn("Failed to read attachment prior to marking as attached.")
		d.markForDeletion(namespace, name)
	}

	da.State = v1.DebugAttachment_Attached

	_, err = (*d.daClient).Write(da, clients.WriteOpts{
		Ctx:               d.ctx,
		OverwriteExisting: true,
	})
	if err != nil {
		log.WithFields(log.Fields{"da.Name": da.Metadata.Name, "da.Namespace": da.Metadata.Namespace}).Warn("Failed to mark debug attachment as attached.")
		d.markForDeletion(namespace, name)
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

func (d *DebugController) tryToAttach(da *v1.DebugAttachment) error {

	// make sure this is not a duplicate
	ci, err := d.conttopid.GetContainerInfo(context.Background(), da)
	log.WithField("ContainerInfo", ci).Debug("GetContainerInfo output")
	if err != nil {
		log.WithField("err", err).Warn("GetContainerInfo error")
		return err
	}

	pid, err := FindFirstProcess(ci.Pids, da.ProcessName)
	if err != nil {
		log.WithField("err", err).Warn("FindFirstProcess error")
		return err
	}

	log.WithField("app", da).Info("Attaching to live session")

	p, err := os.FindProcess(pid)
	if err != nil {
		log.WithField("err", err).Error("can't find process")
		return err
	}
	if d.lockProcess(pid) {
		log.WithField("pid", pid).Info("starting to debug")
		debugger, err := d.startDebug(da, p, ci.Name)
		if err != nil {
			log.WithFields(log.Fields{"namespace": da.Metadata.Namespace, "name": da.Metadata.Name, "error": err}).Warn("Error on startDebug")
			d.markForDeletion(da.Metadata.Namespace, da.Metadata.Name)
			return nil // no retry
		}
		fmt.Println("preactive")
		if err := d.addActiveAttachment(da, pid, debugger); err != nil {
			return err
		}

	} else {
		log.WithField("pid", pid).Warn("Already debugging pid. ignoring")
	}
	return nil
}

// uses the kubesquash debug approach
func (d *DebugController) tryToAttachPod(da *v1.DebugAttachment) error {
	liteConfig := lite.SquashConfig{
		InClusterMode:  true,
		TimeoutSeconds: 3,
		Debugger:       "dlv", // TODO(mitchdraft) - pass as arg
	}
	clientset, err := kubeutils.NewKubeClientset(d.inClusterMode)
	if err != nil {
		return err
	}
	return lite.StartDebugContainer(liteConfig, clientset)
}

func (d *DebugController) startDebug(da *v1.DebugAttachment, p *os.Process, targetName string) (DebugServer, error) {
	log.Info("start debug called")

	curdebugger := d.debugger(da.Debugger)

	if curdebugger == nil {
		return nil, errors.New("debugger doesn't exist")
	}

	log.WithFields(log.Fields{"curdebugger": da.Debugger}).Info("start debug params")

	log.WithFields(log.Fields{"pid": p.Pid}).Info("starting debug server")
	var err error
	debugServer, err := curdebugger.Attach(p.Pid)

	if err != nil {
		log.WithField("err", err).Error("Starting debug server error")
		return nil, err
	}

	log.WithField("pid", p.Pid).Info("StartDebugServer - posting debug session")

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

	if err != nil {
		log.WithField("err", err).Warn("Error adding debug session - detaching!")
		debugServer.Detach()
		return nil, err
	}
	log.Info("debug session added!")

	latestDa, err := (*d.daClient).Read(da.Metadata.Namespace, da.Metadata.Name, clients.ReadOpts{Ctx: d.ctx})
	latestDa.DebugServerAddress = fmt.Sprintf("%s:%d", hostName, debugServer.Port())
	latestDa.State = v1.DebugAttachment_Attached

	if _, err := (*d.daClient).Write(latestDa, clients.WriteOpts{Ctx: d.ctx, OverwriteExisting: true}); err != nil {
		log.WithField("err", err).Error("Writing attachment")
		return nil, err
	}

	return debugServer, nil
}
