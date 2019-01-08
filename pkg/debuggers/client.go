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

	dbghandler.debugController = NewDebugController(debugger, dbghandler.notifyState, conttopid)
	return dbghandler
}

func getNodeName() string {
	return os.Getenv("NODE_NAME")
}

func (d *DebugHandler) handleAttachments() error {
	// setup event loop
	emitter := v1.NewApiEmitter(*d.daClient)
	syncer := d
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
		d.debugController.HandleAddedRemovedAttachments(daList, v1.DebugAttachmentList{})
	}
	return nil
}

func (d *DebugHandler) notifyState(attachment *v1.DebugAttachment) error {

	log.WithFields(log.Fields{"patchDebugAttachment": attachment, "DebugAttachmentID": attachment.Metadata.Name}).Debug("Notifying server of attachment to debug config object")

	// _, err := d.client.Debugattachment.PatchDebugAttachment(params)
	// if err != nil {
	// 	log.WithField("err", err).Warn("Error notifing debug session attachment - detaching!")
	// } else {
	// 	log.Info("debug attachment notified of attachment!")
	// }
	// return err
	return nil
}

// func (d *DebugHandler) watchForAttached() ([]*v1.DebugAttachment, []*v1.DebugAttachment, error) {
// 	for {
// 		params := debugattachment.NewGetDebugAttachmentsParams()
// 		nodename := getNodeName()
// 		params.Node = &nodename
// 		params.IfNoneMatch = d.etag
// 		t := true
// 		params.Wait = &t
// 		log.WithField("params", spew.Sdump(params)).Debug("watchForAttached - calling GetDebugAttachments")

// 		// TODO(mitchdraft) - implement
// 		// resp, err := d.client.Debugattachment.GetDebugAttachments(params)

// 		// We need to find\get events for deleted attachments. to sync them.
// 		// similar to the control loop in kubelet

// 		// if _, ok := err.(*debugattachment.GetDebugAttachmentsRequestTimeout); ok {
// 		// 	continue
// 		// }

// 		// if err != nil {
// 		// 	log.WithField("err", err).Warn("watchForAttached - error calling function:")
// 		// 	time.Sleep(time.Second)
// 		// 	continue
// 		// }

// 		// attachments := resp.Payload
// 		// d.etag = &resp.ETag
// 		// log.WithField("attachments", spew.Sdump(attachments)).Debug("getUpdatedSnapshots - filtering new attachments")

// 		return d.getUpdatedSnapshots(attachments)
// 	}
// }

// func (d *DebugHandler) getUpdatedSnapshots(attachments []*v1.DebugAttachment) ([]*v1.DebugAttachment, []*v1.DebugAttachment, error) {
// 	prevttachments := d.attachments
// 	d.attachments = attachments

// 	if len(d.attachments) == 0 {
// 		log.WithFields(log.Fields{"prevttachments": spew.Sdump(prevttachments)}).Info("no current attachments.")
// 		return attachments, prevttachments, nil
// 	}

// 	// find all attachments that are deleted
// 	var deletedAttachments []*v1.DebugAttachment
// 	for _, attch := range prevttachments {
// 		if !contains(attch, attachments) {
// 			deletedAttachments = append(deletedAttachments, attch)
// 		}
// 	}

// 	// ignore all the state none from the list. we don't do it earlier to not mistake them for
// 	// deleted
// 	var newattachments []*v1.DebugAttachment
// 	for _, attch := range attachments {
// 		if attch.Status.State != "none" {
// 			continue
// 		}
// 		if !contains(attch, prevttachments) {
// 			newattachments = append(newattachments, attch)
// 		}
// 	}

// 	// find all attachments that are new and in state none

// 	log.WithFields(log.Fields{"newattachments": spew.Sdump(newattachments), "deletedAttachments": spew.Sdump(deletedAttachments)}).Info("watchForAttached - got debug attachment!")
// 	return newattachments, deletedAttachments, nil
// }

// func contains(attachment *v1.DebugAttachment, attachments []*v1.DebugAttachment) bool {
// 	name := func(a *v1.DebugAttachment) string {
// 		if a.Metadata != nil {
// 			return a.Metadata.Name
// 		}
// 		return ""
// 	}
// 	attachmentName := name(attachment)
// 	for _, a := range attachments {

// 		if name(a) == attachmentName {
// 			return true
// 		}
// 	}
// 	return false
// }
