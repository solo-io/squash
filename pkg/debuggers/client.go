package debuggers

import (
	"context"
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/platforms"
	"github.com/solo-io/squash/pkg/utils"
)

func RunSquashClient(debugger func(string) Debugger, conttopid platforms.ContainerProcess) error {
	log.SetLevel(log.DebugLevel)

	customFormatter := new(log.TextFormatter)
	log.SetFormatter(customFormatter)

	log.Info("Squash Client started")

	flag.Parse()

	ctx := context.Background()
	daClient, err := utils.GetDebugAttachmentClient(ctx)
	if err != nil {
		log.WithField("err", err).Error("RunDebugBridge")
		return err
	}

	return NewDebugHandler(ctx, daClient, debugger, conttopid).handleAttachments()
}

type DebugHandler struct {
	ctx context.Context

	debugger        func(string) Debugger
	conttopid       platforms.ContainerProcess
	debugController *DebugController
	daClient        *v1.DebugAttachmentClient

	etag        *string
	attachments []*v1.DebugAttachment
}

func NewDebugHandler(ctx context.Context, daClient *v1.DebugAttachmentClient, debugger func(string) Debugger,
	conttopid platforms.ContainerProcess) *DebugHandler {
	dbghandler := &DebugHandler{
		ctx:       ctx,
		daClient:  daClient,
		debugger:  debugger,
		conttopid: conttopid,
	}

	dbghandler.debugController = NewDebugController(ctx, debugger, daClient, conttopid)
	return dbghandler
}

func getNodeName() string {
	return os.Getenv("NODE_NAME")
}

func (d *DebugHandler) handleAttachments() error {
	// setup event loop
	emitter := v1.NewApiEmitter(*d.daClient)
	syncer := d // DebugHandler implements Sync
	el := v1.NewApiEventLoop(emitter, syncer)
	// run event loop
	// TODO(mitchdraft) - use real values
	namespaces := []string{"squash"}
	wOpts := clients.WatchOpts{}
	errs, err := el.Run(namespaces, wOpts)
	if err != nil {
		return err
	}
	for err := range errs {
		contextutils.LoggerFrom(d.ctx).Errorf("error in setup: %v", err)
	}
	return nil
}

// This implements the syncer interface
func (d *DebugHandler) Sync(ctx context.Context, snapshot *v1.ApiSnapshot) error {
	fmt.Println("running sync")
	daMap := snapshot.Debugattachments
	for _, daList := range daMap {
		for _, da := range daList {
			if err := d.syncOne(da); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *DebugHandler) syncOne(da *v1.DebugAttachment) error {
	switch da.State {
	case v1.DebugAttachment_RequestingAttachment:
		log.Debug("handling requesting attachment")
		go d.debugController.handleAttachmentRequest(da)
		return nil
	case v1.DebugAttachment_PendingAttachment:
		log.Debug("handling pending attachment")
		// do nothing, will transition out of this state according to the result of the RequestingAttachment handler
		return nil
	case v1.DebugAttachment_Attached:
		log.Debug("handling attached")
		// do nothing, this is "steady state"
		return nil
	case v1.DebugAttachment_RequestingDelete:
		log.Debug("handling requesting delete")
		log.WithFields(log.Fields{"attachment.Name": da.Metadata.Name}).Debug("Removing attachment")
		d.debugController.removeAttachment(da.Metadata.Name)
		return nil
	case v1.DebugAttachment_PendingDelete:
		log.Debug("handling pending delete")
		// do nothing, will transition out of this state according to the result of the RequestingDelete handler
		return nil
	default:
		return fmt.Errorf("DebugAttachment state not recognized: %v", da.State)
	}
	return nil
}
