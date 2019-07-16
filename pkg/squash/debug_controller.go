package squash

import (
	"context"
	"os"
	"sync"

	"github.com/solo-io/go-utils/contextutils"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/config"
	"github.com/solo-io/squash/pkg/debuggers/remote"
	sqOpts "github.com/solo-io/squash/pkg/options"
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

	logger := contextutils.LoggerFrom(d.ctx)
	if ok {
		logger.Debugw("Detaching attachment", "attachment.Name", name)
		err := data.debugger.Detach()
		if err != nil {
			logger.Debugw("Error detaching", "attachment.Name", name, "err", err)
		}
	}
}

func (d *DebugController) handleAttachmentRequest(ctx context.Context, da *v1.DebugAttachment) {

	// Mark attachment as in progress
	da.State = v1.DebugAttachment_PendingAttachment
	_, err := d.daClient.Write(da, clients.WriteOpts{OverwriteExisting: true})
	logger := contextutils.LoggerFrom(d.ctx)
	if err != nil {
		logger.Warnw("Failed to update attachment status.", "da.Name", da.Metadata.Name, "da.Namespace", da.Metadata.Namespace)
	}
	// TODO - put in a goroutine
	err = d.tryToAttachPod(ctx, da)
	if err != nil {
		logger.Warnw("Failed to attach debugger, deleting request.", "da.Name", da.Metadata.Name, "da.Namespace", da.Metadata.Namespace, "error", err)
		d.markForDeletion(da.Metadata.Namespace, da.Metadata.Name)
	}

}

func (d *DebugController) setState(namespace, name string, state v1.DebugAttachment_State) {
	logger := contextutils.LoggerFrom(d.ctx)
	logger.Debugw("marking state", "namespace", namespace, "name", name, "state", state)
	da, err := d.daClient.Read(namespace, name, clients.ReadOpts{Ctx: d.ctx})
	if err != nil {
		// should not happen, but if it does, the CRD was probably already deleted
		logger.Warnw("Failed to read attachment.", "da.Name", da.Metadata.Name, "da.Namespace", da.Metadata.Namespace)
	}

	da.State = state

	_, err = d.daClient.Write(da, clients.WriteOpts{
		Ctx:               d.ctx,
		OverwriteExisting: true,
	})
	if err != nil {
		logger.Warnw("Failed to set attachment state.", "da.Name", da.Metadata.Name, "da.Namespace", da.Metadata.Namespace)
	}
}
func (d *DebugController) markForDeletion(namespace, name string) {
	logger := contextutils.LoggerFrom(d.ctx)
	logger.Debug("called mark for deletion from squash - skipping for now")
	return
	logger.Debugw("marking for deletion", "namespace", namespace, "name", name)
	da, err := d.daClient.Read(namespace, name, clients.ReadOpts{Ctx: d.ctx})
	if err != nil {
		// should not happen, but if it does, the CRD was probably already deleted
		logger.Debugw("Failed to read attachment prior to delete.", "da.Name", da.Metadata.Name, "da.Namespace", da.Metadata.Namespace)
	}

	da.State = v1.DebugAttachment_PendingDelete

	_, err = d.daClient.Write(da, clients.WriteOpts{
		Ctx:               d.ctx,
		OverwriteExisting: true,
	})
	if err != nil {
		logger.Warnw("Failed to mark attachment for deletion.", "da.Name", da.Metadata.Name, "da.Namespace", da.Metadata.Namespace)
	}
}

func (d *DebugController) deleteResource(namespace, name string) {
	err := d.daClient.Delete(namespace, name, clients.DeleteOpts{Ctx: d.ctx, IgnoreNotExist: true})
	if err != nil {
		contextutils.LoggerFrom(d.ctx).Warnw("Failed to delete resource.", "name", name, "namespace", namespace, "error", err)
	}
}

func (d *DebugController) markAsAttached(namespace, name string) {
	da, err := d.daClient.Read(namespace, name, clients.ReadOpts{Ctx: d.ctx})
	logger := contextutils.LoggerFrom(d.ctx)
	if err != nil {
		logger.Warnw("Failed to read attachment prior to marking as attached.", "da.Name", da.Metadata.Name, "da.Namespace", da.Metadata.Namespace)
		d.markForDeletion(namespace, name)
	}

	da.State = v1.DebugAttachment_Attached

	_, err = d.daClient.Write(da, clients.WriteOpts{
		Ctx:               d.ctx,
		OverwriteExisting: true,
	})
	if err != nil {
		logger.Warnw("Failed to mark debug attachment as attached.", "da.Name", da.Metadata.Name, "da.Namespace", da.Metadata.Namespace)
		d.markForDeletion(namespace, name)
	}
}

func (d *DebugController) tryToAttachPod(ctx context.Context, da *v1.DebugAttachment) error {
	s := config.NewSquashConfig(&d.daClient)
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

	s.SquashNamespace = os.Getenv(sqOpts.PlankEnvDebugSquashNamespace)

	dbt := config.DebugTarget{}
	// if err := s.ExpectToGetUniqueDebugTargetFromSpec(&dbt); err != nil {
	// 	return err
	// }
	_, err := config.StartDebugContainer(ctx, s, dbt)
	if err != nil {
		return err
	}
	d.markAsAttached(da.Metadata.Namespace, da.Metadata.Name)
	return nil
}
