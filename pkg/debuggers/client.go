package debuggers

import (
	"context"
	"errors"
	"flag"
	"fmt"
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
	attachments, err := d.watchForAttached()

	if err != nil {
		log.WithField("err", err).Warn("error watching for attached container")
		return err
	}

	for _, attachment := range attachments {
		// notify the server that we are attaching, so we won't get the same attachment object next time.
		if err := d.notifyState(attachment, models.DebugAttachmentStatusStateAttaching); err != nil {
			log.WithFields(log.Fields{"attachment.Name": attachment.Metadata.Name, "err": err}).Debug("Failed set state to attaching in squash server. aborting.")

			d.notifyError(attachment)
		}
		go d.handleSingleAttachment(attachment)
	}
	return nil
}

func (d *DebugHandler) handleSingleAttachment(attachment *models.DebugAttachment) {

	err := retry(func() error { return d.tryToAttach(attachment) })

	if err != nil {
		log.WithFields(log.Fields{"attachment.Name": attachment.Metadata.Name}).Debug("Failed to attach... signaling server.")
		d.notifyError(attachment)
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

func (d *DebugHandler) tryToAttach(attachment *models.DebugAttachment) error {

	// make sure this is not a duplicate

	pid, err := d.conttopid.GetPid(context.Background(), attachment.Spec.Attachment)

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
	if !d.debugees[pid] {
		log.WithField("pid", pid).Info("starting to debug")
		d.debugees[pid] = true
		err := d.startDebug(attachment, p)
		if err != nil {
			d.notifyError(attachment)
		}
	} else {
		log.WithField("pid", pid).Warn("Already debugging pid. ignoring")
		d.notifyError(attachment)
	}
	return nil
}

func (d *DebugHandler) notifyError(attachment *models.DebugAttachment) {
	d.notifyState(attachment, models.DebugAttachmentStatusStateError)
}

func (d *DebugHandler) notifyState(attachment *models.DebugAttachment, newstate string) error {

	attachmentCopy := *attachment

	params := debugattachment.NewPatchDebugAttachmentParams()
	if attachmentCopy.Status == nil {
		attachmentCopy.Status = &models.DebugAttachmentStatus{}
	}
	attachmentCopy.Status.State = newstate
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

func (d *DebugHandler) startDebug(attachment *models.DebugAttachment, p *os.Process) error {
	log.Info("start debug called")

	curdebugger := d.debugger(attachment.Spec.Debugger)

	if curdebugger == nil {
		return errors.New("debugger doesn't exist")
	}

	log.WithFields(log.Fields{"curdebugger": attachment.Spec.Debugger}).Info("start debug params")

	log.WithFields(log.Fields{"pid": p.Pid}).Info("starting debug server")
	var err error
	debugServer, err := curdebugger.Attach(p.Pid)

	if err != nil {
		log.WithField("err", err).Error("Starting debug server error")
		return err
	}

	log.WithField("pid", p.Pid).Info("StartDebugServer - posting debug session")

	attachmentPatch := &models.DebugAttachment{
		Metadata: attachment.Metadata,
		Spec:     attachment.Spec,
	}

	podName := ""
	switch debugServer.PodType() {
	case DebugPodTypeTarget:
		att, ok := attachment.Spec.Attachment.(map[string]interface{})
		if ok {
			podName, _ = att["pod"].(string)
		}
	case DebugPodTypeClient:
		podName = os.Getenv("HOST_ADDR")
	}

	if len(podName) == 0 {
		err = fmt.Errorf("Cannot find POD name for type: %d", debugServer.PodType())
		log.WithField("err", err).Error("Starting debug server error")
		return err
	}

	attachmentPatch.Status = &models.DebugAttachmentStatus{
		DebugServerAddress: fmt.Sprintf("%s:%d", podName, debugServer.Port()),
		State:              models.DebugAttachmentStatusStateAttached,
	}
	params := debugattachment.NewPatchDebugAttachmentParams()
	params.Body = attachmentPatch
	params.DebugAttachmentID = attachment.Metadata.Name

	log.WithFields(log.Fields{"patchDebugAttachment": params.Body, "DebugAttachmentID": params.DebugAttachmentID}).Debug("Notifying server of attachment to debug config object")
	_, err = d.client.Debugattachment.PatchDebugAttachment(params)

	if err != nil {
		log.WithField("err", err).Warn("Error adding debug session - detaching!")
		debugServer.Detach()
	} else {
		log.Info("debug session added!")
	}
	return nil
}

func (d *DebugHandler) watchForAttached() ([]*models.DebugAttachment, error) {
	for {
		params := debugattachment.NewGetDebugAttachmentsParams()
		nodename := getNodeName()
		params.Node = &nodename
		t := true
		params.Wait = &t
		none := models.DebugAttachmentStatusStateNone
		params.State = &none
		log.WithField("params", params).Debug("watchForAttached - calling PopContainerToDebug")

		resp, err := d.client.Debugattachment.GetDebugAttachments(params)

		if _, ok := err.(*debugattachment.GetDebugAttachmentsRequestTimeout); ok {
			continue
		}

		if err != nil {
			log.WithField("err", err).Warn("watchForAttached - error calling function:")
			time.Sleep(time.Second)
			continue
		}

		attachment := resp.Payload

		log.WithField("attachment", spew.Sdump(attachment)).Info("watchForAttached - got debug attachment!")

		return attachment, nil
	}
}
