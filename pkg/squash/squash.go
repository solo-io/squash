package squash

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	gokubeutils "github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/debuggers/remote"
	"github.com/solo-io/squash/pkg/utils"
	"github.com/solo-io/squash/pkg/utils/kubeutils"
	"k8s.io/client-go/kubernetes"
)

// we want to use the default kubeconfig for squash since it runs in the cluster
const squashKubeconfigPath = ""

func RunSquash(ctx context.Context, debugger func(string) remote.Remote) error {
	logger := contextutils.LoggerFrom(ctx)
	logger.Info("Squash Client started")

	flag.Parse()

	daClient, err := utils.GetBasicDebugAttachmentClient(ctx, squashKubeconfigPath)
	if err != nil {
		logger.Errorw("RunDebugBridge", "err", err)
		return err
	}

	restCfg, err := gokubeutils.GetConfig("", squashKubeconfigPath)
	if err != nil {
		return err
	}
	kubeResClient, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return err
	}
	watchNamespaces, err := kubeutils.GetNamespaces(kubeResClient)
	if err != nil {
		return err
	}

	return NewDebugHandler(ctx, watchNamespaces, daClient, debugger).handleAttachments()
}

type DebugHandler struct {
	ctx context.Context

	debugger        func(string) remote.Remote
	debugController *DebugController
	daClient        v1.DebugAttachmentClient

	watchNamespaces []string

	etag        *string
	attachments []*v1.DebugAttachment
}

func NewDebugHandler(ctx context.Context, watchNamespaces []string, daClient v1.DebugAttachmentClient, debugger func(string) remote.Remote) *DebugHandler {
	dbghandler := &DebugHandler{
		ctx:             ctx,
		daClient:        daClient,
		debugger:        debugger,
		watchNamespaces: watchNamespaces,
	}

	dbghandler.debugController = NewDebugController(ctx, debugger, daClient)
	return dbghandler
}

func (d *DebugHandler) handleAttachments() error {
	// setup event loop
	emitter := v1.NewApiEmitter(d.daClient)
	syncer := d // DebugHandler implements Sync
	el := v1.NewApiEventLoop(emitter, syncer)
	// run event loop
	wOpts := clients.WatchOpts{}
	logger := contextutils.LoggerFrom(d.ctx)
	logger.Infow("Watching namespaces", "list", d.watchNamespaces)
	errs, err := el.Run(d.watchNamespaces, wOpts)
	if err != nil {
		return err
	}
	for err := range errs {
		logger.Errorf("error in setup: %v", err)
	}
	return nil
}

// This implements the syncer interface
func (d *DebugHandler) Sync(ctx context.Context, snapshot *v1.ApiSnapshot) error {
	contextutils.LoggerFrom(ctx).Debug("running sync")
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
	ctx, _ := context.WithTimeout(d.ctx, 30*time.Second)
	logger := contextutils.LoggerFrom(d.ctx)
	switch da.State {
	case v1.DebugAttachment_RequestingAttachment:
		logger.Debugf("handling requesting attachment %v", da)
		go d.debugController.handleAttachmentRequest(ctx, da)
		return nil
	case v1.DebugAttachment_PendingAttachment:
		logger.Debug("handling pending attachment")
		// do nothing, will transition out of this state according to the result of the RequestingAttachment handler
		return nil
	case v1.DebugAttachment_Attached:
		logger.Debug("handling attached")
		// do nothing, this is "steady state"
		return nil
	case v1.DebugAttachment_RequestingDelete:
		logger.Debug("handling requesting delete")
		// DO NOTHING - Will refactor this
		// log.WithFields(log.Fields{"attachment.Name": da.Metadata.Name}).Debug("Removing attachment")
		// go func() { d.debugController.removeAttachment(da.Metadata.Namespace, da.Metadata.Name) }()
		return nil
	case v1.DebugAttachment_PendingDelete:
		logger.Debug("handling pending delete")
		// DO NOTHING - Will refactor this
		// d.debugController.deleteResource(da.Metadata.Namespace, da.Metadata.Name)
		// do nothing, will transition out of this state according to the result of the RequestingDelete handler
		return nil
	default:
		return fmt.Errorf("DebugAttachment state not recognized: %v", da.State)
	}
	return nil
}
