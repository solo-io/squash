package squash

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/config"
	"github.com/solo-io/squash/pkg/debuggers/remote"
	"github.com/solo-io/squash/pkg/options"
	"github.com/solo-io/squash/pkg/version"
)

type DebugController struct {
	debugger func(string) remote.Remote
	pidLock  sync.Mutex
	pidMap   map[int]bool

	daClient v1.DebugAttachmentClient
	ctx      context.Context

	debugattachmentsLock sync.Mutex
	debugattachments     map[string]debugAttachmentData
}

type debugAttachmentData struct {
	debugger remote.DebugServer
	pid      int
}

func NewDebugController(ctx context.Context,
	debugger func(string) remote.Remote,
	daClient v1.DebugAttachmentClient) *DebugController {
	return &DebugController{
		debugger: debugger,

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
	_, err := d.daClient.Write(da, clients.WriteOpts{OverwriteExisting: true})
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
	da, err := d.daClient.Read(namespace, name, clients.ReadOpts{Ctx: d.ctx})
	if err != nil {
		// should not happen, but if it does, the CRD was probably already deleted
		log.WithFields(log.Fields{"da.Name": da.Metadata.Name, "da.Namespace": da.Metadata.Namespace}).Warn("Failed to read attachment.")
	}

	da.State = state

	_, err = d.daClient.Write(da, clients.WriteOpts{
		Ctx:               d.ctx,
		OverwriteExisting: true,
	})
	if err != nil {
		log.WithFields(log.Fields{"da.Name": da.Metadata.Name, "da.Namespace": da.Metadata.Namespace}).Warn("Failed to set attachment state.")
	}
}
func (d *DebugController) markForDeletion(namespace, name string) {
	log.Debug("called mark for deletion from squash - skipping for now")
	return
	log.WithFields(log.Fields{"namespace": namespace, "name": name}).Debug("marking for deletion")
	da, err := d.daClient.Read(namespace, name, clients.ReadOpts{Ctx: d.ctx})
	if err != nil {
		// should not happen, but if it does, the CRD was probably already deleted
		log.WithFields(log.Fields{"da.Name": da.Metadata.Name, "da.Namespace": da.Metadata.Namespace}).Warn("Failed to read attachment prior to delete.")
	}

	da.State = v1.DebugAttachment_PendingDelete

	_, err = d.daClient.Write(da, clients.WriteOpts{
		Ctx:               d.ctx,
		OverwriteExisting: true,
	})
	if err != nil {
		log.WithFields(log.Fields{"da.Name": da.Metadata.Name, "da.Namespace": da.Metadata.Namespace}).Warn("Failed to mark attachment for deletion.")
	}
}

func (d *DebugController) deleteResource(namespace, name string) {
	err := d.daClient.Delete(namespace, name, clients.DeleteOpts{Ctx: d.ctx, IgnoreNotExist: true})
	if err != nil {
		log.WithFields(log.Fields{"name": name, "namespace": namespace, "error": err}).Warn("Failed to delete resource.")
	}
}

func (d *DebugController) markAsAttached(namespace, name string) {
	da, err := d.daClient.Read(namespace, name, clients.ReadOpts{Ctx: d.ctx})
	if err != nil {
		log.WithFields(log.Fields{"da.Name": da.Metadata.Name, "da.Namespace": da.Metadata.Namespace}).Warn("Failed to read attachment prior to marking as attached.")
		d.markForDeletion(namespace, name)
	}

	da.State = v1.DebugAttachment_Attached

	_, err = d.daClient.Write(da, clients.WriteOpts{
		Ctx:               d.ctx,
		OverwriteExisting: true,
	})
	if err != nil {
		log.WithFields(log.Fields{"da.Name": da.Metadata.Name, "da.Namespace": da.Metadata.Namespace}).Warn("Failed to mark debug attachment as attached.")
		d.markForDeletion(namespace, name)
	}
}

func (d *DebugController) tryToAttachPod(da *v1.DebugAttachment) error {
	s := config.NewSquashConfig()
	s.TimeoutSeconds = 300
	s.Machine = true
	s.NoClean = true

	s.DebugContainerVersion = version.ImageVersion
	s.DebugContainerRepo = version.ImageRepo

	s.CRISock = "/var/run/dockershim.sock"

	s.Debugger = da.Debugger
	s.Namespace = da.Metadata.Namespace
	s.Pod = da.Pod
	s.Container = da.Image

	// TODO(mitchdraft) - set this from an os.Env read
	s.SquashNamespace = options.SquashNamespace

	dbt := config.DebugTarget{}
	// if err := s.ExpectToGetUniqueDebugTargetFromSpec(&dbt); err != nil {
	// 	return err
	// }
	_, err := config.StartDebugContainer(s, dbt)
	if err != nil {
		return err
	}
	d.markAsAttached(da.Metadata.Namespace, da.Metadata.Name)
	return nil
}
