package debuggers

import (
	"context"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/kscmd"
	"github.com/solo-io/squash/pkg/platforms"
	"github.com/solo-io/squash/pkg/version"
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
}

type debugAttachmentData struct {
	debugger DebugServer
	pid      int
}

func NewDebugController(ctx context.Context,
	debugger func(string) Debugger,
	daClient *v1.DebugAttachmentClient,
	conttopid platforms.ContainerProcess) *DebugController {
	return &DebugController{
		debugger:  debugger,
		conttopid: conttopid,

		daClient: daClient,

		pidMap: make(map[int]bool),

		debugattachments: make(map[string]debugAttachmentData),
	}
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
	}
}

func (d *DebugController) handleAttachmentRequest(da *v1.DebugAttachment) {

	// Mark attachment as in progress
	da.State = v1.DebugAttachment_PendingAttachment
	_, err := (*d.daClient).Write(da, clients.WriteOpts{OverwriteExisting: true})
	if err != nil {
		log.WithFields(log.Fields{"da.Name": da.Metadata.Name, "da.Namespace": da.Metadata.Namespace}).Warn("Failed to update attachment status.")
	}
	// TODO - put in a goroutine
	err = d.tryToAttachPod(da)
	if err != nil {
		log.WithFields(log.Fields{"da.Name": da.Metadata.Name, "da.Namespace": da.Metadata.Namespace, "error": err}).Warn("Failed to attach debugger, deleting request.")
		d.markForDeletion(da.Metadata.Namespace, da.Metadata.Name)
	}

}

func (d *DebugController) setState(namespace, name string, state v1.DebugAttachment_State) {
	log.WithFields(log.Fields{"namespace": namespace, "name": name, "state": state}).Debug("marking state")
	da, err := (*d.daClient).Read(namespace, name, clients.ReadOpts{Ctx: d.ctx})
	if err != nil {
		// should not happen, but if it does, the CRD was probably already deleted
		log.WithFields(log.Fields{"da.Name": da.Metadata.Name, "da.Namespace": da.Metadata.Namespace}).Warn("Failed to read attachment.")
	}

	da.State = state

	_, err = (*d.daClient).Write(da, clients.WriteOpts{
		Ctx:               d.ctx,
		OverwriteExisting: true,
	})
	if err != nil {
		log.WithFields(log.Fields{"da.Name": da.Metadata.Name, "da.Namespace": da.Metadata.Namespace}).Warn("Failed to set attachment state.")
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

func (d *DebugController) markAsAttached(namespace, name, debUrl string) {
	da, err := (*d.daClient).Read(namespace, name, clients.ReadOpts{Ctx: d.ctx})
	if err != nil {
		log.WithFields(log.Fields{"da.Name": da.Metadata.Name, "da.Namespace": da.Metadata.Namespace}).Warn("Failed to read attachment prior to marking as attached.")
		d.markForDeletion(namespace, name)
	}

	da.State = v1.DebugAttachment_Attached
	da.DebugServerAddress = debUrl

	_, err = (*d.daClient).Write(da, clients.WriteOpts{
		Ctx:               d.ctx,
		OverwriteExisting: true,
	})
	if err != nil {
		log.WithFields(log.Fields{"da.Name": da.Metadata.Name, "da.Namespace": da.Metadata.Namespace}).Warn("Failed to mark debug attachment as attached.")
		d.markForDeletion(namespace, name)
	}
}

// uses the kubesquash debug approach
func (d *DebugController) tryToAttachPod(da *v1.DebugAttachment) error {
	ksConfig := kscmd.SquashConfig{
		TimeoutSeconds: 300,
		Machine:        true,
		InCluster:      true,
		DebugServer:    true,

		DebugContainerVersion: version.ImageVersion,
		DebugContainerRepo:    version.ImageRepo,

		CRISock: "/var/run/dockershim.sock",

		Debugger:  da.Debugger,
		Namespace: da.Metadata.Namespace,
		Pod:       da.Pod,
		Container: da.Image,
	}
	debPod, err := kscmd.StartDebugContainer(ksConfig)
	// TODO(mitchdraft) - refactor once old squash functionality is removed
	if err != nil {
		return err
	}
	debUrl := fmt.Sprintf("%v:%v", debPod.ObjectMeta.Name, 1235)
	d.markAsAttached(da.Metadata.Namespace, da.Metadata.Name, debUrl)
	return nil
}
