package debuggers

import (
	"flag"
	"net/url"
	"os"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
	"github.com/solo-io/squash/pkg/client"
	"github.com/solo-io/squash/pkg/client/debugattachment"
	"github.com/solo-io/squash/pkg/models"
	"github.com/solo-io/squash/pkg/platforms"
)

func RunSquashClient(debugger func(string) Debugger, conttopid platforms.ContainerProcess) error {
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
	debugger        func(string) Debugger
	conttopid       platforms.ContainerProcess
	client          *client.Squash
	debugController *DebugController

	etag        *string
	attachments []*models.DebugAttachment
}

func NewDebugHandler(client *client.Squash, debugger func(string) Debugger,
	conttopid platforms.ContainerProcess) *DebugHandler {
	dbghandler := &DebugHandler{
		client:    client,
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
	for {
		err := d.handleAttachment()
		if err != nil {
			log.WithField("err", err).Warn("error watching for attached container")
		}
	}
}

func (d *DebugHandler) handleAttachment() error {
	attachments, removedAtachment, err := d.watchForAttached()

	if err != nil {
		log.WithField("err", err).Warn("error watching for attached container")
		return err
	}
	return d.debugController.HandleAddedRemovedAttachments(attachments, removedAtachment)
}

func (d *DebugHandler) notifyState(attachment *models.DebugAttachment) error {

	attachmentCopy := *attachment
	params := debugattachment.NewPatchDebugAttachmentParams()
	if attachmentCopy.Status == nil {
		attachmentCopy.Status = &models.DebugAttachmentStatus{}
	}
	params.Body = &attachmentCopy
	params.DebugAttachmentID = attachment.Metadata.Name

	log.WithFields(log.Fields{"patchDebugAttachment": params.Body, "DebugAttachmentID": params.DebugAttachmentID}).Debug("Notifying server of attachment to debug config object")

	_, err := d.client.Debugattachment.PatchDebugAttachment(params)
	if err != nil {
		log.WithField("err", err).Warn("Error notifing debug session attachment - detaching!")
	} else {
		log.Info("debug attachment notified of attachment!")
	}
	return err
}

func (d *DebugHandler) watchForAttached() ([]*models.DebugAttachment, []*models.DebugAttachment, error) {
	for {
		params := debugattachment.NewGetDebugAttachmentsParams()
		nodename := getNodeName()
		params.Node = &nodename
		params.IfNoneMatch = d.etag
		t := true
		params.Wait = &t
		log.WithField("params", spew.Sdump(params)).Debug("watchForAttached - calling GetDebugAttachments")

		resp, err := d.client.Debugattachment.GetDebugAttachments(params)

		// We need to find\get events for deleted attachments. to sync them.
		// similar to the control loop in kubelet

		if _, ok := err.(*debugattachment.GetDebugAttachmentsRequestTimeout); ok {
			continue
		}

		if err != nil {
			log.WithField("err", err).Warn("watchForAttached - error calling function:")
			time.Sleep(time.Second)
			continue
		}

		attachments := resp.Payload
		d.etag = &resp.ETag
		log.WithField("attachments", spew.Sdump(attachments)).Debug("getUpdatedSnapshots - filtering new attachments")

		return d.getUpdatedSnapshots(attachments)
	}
}

func (d *DebugHandler) getUpdatedSnapshots(attachments []*models.DebugAttachment) ([]*models.DebugAttachment, []*models.DebugAttachment, error) {
	prevttachments := d.attachments
	d.attachments = attachments

	if len(d.attachments) == 0 {
		log.WithFields(log.Fields{"prevttachments": spew.Sdump(prevttachments)}).Info("no current attachments.")
		return attachments, prevttachments, nil
	}

	// find all attachments that are deleted
	var deletedAttachments []*models.DebugAttachment
	for _, attch := range prevttachments {
		if !contains(attch, attachments) {
			deletedAttachments = append(deletedAttachments, attch)
		}
	}

	// ignore all the state none from the list. we don't do it earlier to not mistake them for
	// deleted
	var newattachments []*models.DebugAttachment
	for _, attch := range attachments {
		if attch.Status.State != "none" {
			continue
		}
		if !contains(attch, prevttachments) {
			newattachments = append(newattachments, attch)
		}
	}

	// find all attachments that are new and in state none

	log.WithFields(log.Fields{"newattachments": spew.Sdump(newattachments), "deletedAttachments": spew.Sdump(deletedAttachments)}).Info("watchForAttached - got debug attachment!")
	return newattachments, deletedAttachments, nil
}

func contains(attachment *models.DebugAttachment, attachments []*models.DebugAttachment) bool {
	name := func(a *models.DebugAttachment) string {
		if a.Metadata != nil {
			return a.Metadata.Name
		}
		return ""
	}
	attachmentName := name(attachment)
	for _, a := range attachments {

		if name(a) == attachmentName {
			return true
		}
	}
	return false
}
