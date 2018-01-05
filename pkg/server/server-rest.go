package server

import (
	"bytes"
	"context"
	"io"
	"math/rand"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-openapi/runtime/middleware"
	"github.com/solo-io/squash/pkg/models"
	"github.com/solo-io/squash/pkg/platforms"
	"github.com/solo-io/squash/pkg/restapi/operations/debugattachment"
	"github.com/solo-io/squash/pkg/restapi/operations/debugrequest"
)

type RestHandler struct {
	debugAttachments        map[string]*models.DebugAttachment
	debugAttachmentsMapLock sync.RWMutex

	debugRequests        map[string]*models.DebugRequest
	debugRequestsMapLock sync.RWMutex

	containerLocator platforms.ContainerLocator

	attachmentlisteners     []chan int
	attachmentlistenersLock sync.Mutex
}

func NewRestHandler(containerLocator platforms.ContainerLocator) *RestHandler {
	return &RestHandler{
		debugAttachments: make(map[string]*models.DebugAttachment),
		debugRequests:    make(map[string]*models.DebugRequest),
		containerLocator: containerLocator,
	}
}

func verify_or_generate(s string) string {
	if s != "" {
		return s
	}
	return randomString()
}

func (r *RestHandler) DebugattachmentAddDebugAttachmentHandler(params debugattachment.AddDebugAttachmentParams) middleware.Responder {
	// if then else!
	// TODO generate name if needed
	log.Info("DebugattachmentAddDebugAttachmentHandler called!")

	// validate the attachment
	dbgattachment := params.Body

	attachment, container, err := r.containerLocator.Locate(params.HTTPRequest.Context(), dbgattachment.Spec.Attachment)
	if err != nil {
		log.WithError(err).Warn("DebugattachmentAddDebugAttachmentHandler can't locate container")
		return debugattachment.NewAddDebugAttachmentNotFound()
	}

	if dbgattachment.Spec == nil {
		dbgattachment.Spec = &models.DebugAttachmentSpec{}
	}
	dbgattachment.Spec.Attachment = attachment
	if dbgattachment.Spec.Image == "" {
		dbgattachment.Spec.Image = container.Image
	} else if dbgattachment.Spec.Image != container.Image {
		return debugattachment.NewAddDebugAttachmentBadRequest()
	}

	dbgattachment.Spec.Node = container.Node

	if dbgattachment.Metadata == nil {
		dbgattachment.Metadata = &models.ObjectMeta{}
	}
	dbgattachment.Metadata.Name = verify_or_generate(dbgattachment.Metadata.Name)

	if dbgattachment.Spec.MatchRequest {
		log.WithField("dbgattachment", dbgattachment).Debug("trying to match a debug request for debug attachment")

		// find a matching request for the same image
		dr := r.findUnboundDebugRequest(dbgattachment)
		if dr == nil {
			// error!
			return debugattachment.NewAddDebugAttachmentNotFound()
		}

		// copy the requested debugger if needed
		if dbgattachment.Spec.Debugger == "" && dr.Spec.Debugger != nil {
			dbgattachment.Spec.Debugger = *dr.Spec.Debugger
		}

		if dbgattachment.Spec.ProcessName != "" && dbgattachment.Spec.ProcessName != dr.Spec.ProcessName {
			log.WithFields(
				log.Fields{
					"dbgattachment":                  dbgattachment,
					"dbgattachment.Spec.ProcessName": dbgattachment.Spec.ProcessName,
					"dr.Spec.ProcessName":            dr.Spec.ProcessName,
				}).Warning("debug attachment processName conflict")
		}
		dbgattachment.Spec.ProcessName = dr.Spec.ProcessName

		// we found a matching request - we can save now.
		r.saveDebugAttachment(dbgattachment)

		go func(dr models.DebugRequest) {
			dr.Status.DebugAttachmentRef = dbgattachment.Metadata.Name
			// place the debug attachment
			// update  the debug request
			// release all locks
			r.updateDebugRequest(dr)
		}(*dr)

	} else {
		log.WithField("dbgattachment", dbgattachment).Debug("DebugattachmentAddDebugAttachmentHandler match not needed, done.")

		r.saveDebugAttachment(dbgattachment)
	}

	return debugattachment.NewAddDebugAttachmentCreated().WithPayload(dbgattachment)
}

func (r *RestHandler) findUnboundDebugRequest(dbgattachment *models.DebugAttachment) *models.DebugRequest {
	r.debugRequestsMapLock.RLock()
	defer r.debugRequestsMapLock.RUnlock()
	for _, dr := range r.debugRequests {
		if dr.Status.DebugAttachmentRef != "" {
			continue
		}

		if dr.Spec.Image == nil || *dr.Spec.Image != dbgattachment.Spec.Image {
			continue
		}

		// logical NOT XOR
		if (dr.Spec.Debugger == nil) == (dbgattachment.Spec.Debugger == "") {
			continue
		}

		// Found a match, return it!

		return dr
	}
	return nil
}

func (r *RestHandler) updateDebugRequest(dr models.DebugRequest) {
	r.debugRequestsMapLock.Lock()
	defer r.debugRequestsMapLock.Unlock()
	r.debugRequests[dr.Metadata.Name] = &dr
}

func (r *RestHandler) saveDebugAttachment(da *models.DebugAttachment) {
	if da.Status == nil {
		da.Status = &models.DebugAttachmentStatus{
			State: models.DebugAttachmentStatusStateNone,
		}
	}
	func() {
		r.debugAttachmentsMapLock.Lock()
		defer r.debugAttachmentsMapLock.Unlock()
		r.debugAttachments[da.Metadata.Name] = da
	}()

	log.WithField("dbgattachment", da).Debug("saveDebugAttachment - notifying waiters")
	r.notify()
}

func (r *RestHandler) getDebugAttachment(name string) *models.DebugAttachment {
	r.debugAttachmentsMapLock.RLock()
	defer r.debugAttachmentsMapLock.RUnlock()
	return r.debugAttachments[name]
}

func (r *RestHandler) DebugrequestCreateDebugRequestHandler(params debugrequest.CreateDebugRequestParams) middleware.Responder {
	dr := params.Body

	if dr.Metadata == nil {
		dr.Metadata = &models.ObjectMeta{}
	}
	dr.Metadata.Name = verify_or_generate(dr.Metadata.Name)
	if dr.Status == nil {
		dr.Status = &models.DebugRequestStatus{}
	}
	r.debugRequestsMapLock.Lock()
	defer r.debugRequestsMapLock.Unlock()

	r.debugRequests[dr.Metadata.Name] = dr

	return debugrequest.NewCreateDebugRequestCreated().WithPayload(dr)
}

func (r *RestHandler) DebugattachmentPatchDebugAttachmentHandler(params debugattachment.PatchDebugAttachmentParams) middleware.Responder {
	newDa := params.Body
	oldDa := r.getDebugAttachment(newDa.Metadata.Name)
	if oldDa == nil {
		return debugattachment.NewPatchDebugAttachmentNotFound()
	}

	log.WithFields(log.Fields{
		"oldDa": spew.Sdump(oldDa), "newDa": spew.Sdump(newDa),
	}).Warn("DebugattachmentPatchDebugAttachmentHandler")

	oldDaCopy := *oldDa
	if newDa.Status != nil {
		if oldDaCopy.Status == nil {
			oldDaCopy.Status = &models.DebugAttachmentStatus{}
		}

		if newDa.Status.State != "" {
			if canUpdateState(oldDaCopy.Status.State, newDa.Status.State) {
				oldDaCopy.Status.State = newDa.Status.State
			} else {
				log.WithFields(log.Fields{"attachment": oldDaCopy,
					"oldstate": oldDaCopy.Status.State, "newstate": newDa.Status.State,
				}).Warn("Conflict - trying to update to an old state")
				return debugattachment.NewPatchDebugAttachmentConflict()
			}
		}
		if newDa.Status.DebugServerAddress != "" {
			if oldDaCopy.Status.DebugServerAddress == "" {
				oldDaCopy.Status.DebugServerAddress = newDa.Status.DebugServerAddress
			} else {
				log.WithFields(log.Fields{"attachment": oldDaCopy,
					"old": oldDaCopy.Status.DebugServerAddress, "new": newDa.Status.DebugServerAddress,
				}).Warn("Conflict - trying to update to an existing debug server address")
				return debugattachment.NewPatchDebugAttachmentConflict()
			}
		}
	}

	r.saveDebugAttachment(&oldDaCopy)
	return debugattachment.NewPatchDebugAttachmentOK().WithPayload(&oldDaCopy)
}
func canUpdateState(oldstate, newstate string) bool {
	states := map[string]int{models.DebugAttachmentStatusStateNone: 0,
		models.DebugAttachmentStatusStateAttaching: 1,
		models.DebugAttachmentStatusStateAttached:  2,
		models.DebugAttachmentStatusStateError:     3,
	}

	return states[newstate] > states[oldstate]
}

func (r *RestHandler) DebugattachmentDeleteDebugAttachmentHandler(params debugattachment.DeleteDebugAttachmentParams) middleware.Responder {
	r.debugAttachmentsMapLock.Lock()
	defer r.debugAttachmentsMapLock.Unlock()
	delete(r.debugAttachments, params.DebugAttachmentID)

	return debugattachment.NewDeleteDebugAttachmentOK()
}

func (r *RestHandler) DebugrequestDeleteDebugRequestHandler(params debugrequest.DeleteDebugRequestParams) middleware.Responder {
	r.debugRequestsMapLock.Lock()
	defer r.debugRequestsMapLock.Unlock()
	delete(r.debugRequests, params.DebugRequestID)

	return debugrequest.NewDeleteDebugRequestOK()
}

func (r *RestHandler) DebugattachmentGetDebugAttachmentHandler(params debugattachment.GetDebugAttachmentParams) middleware.Responder {
	r.debugAttachmentsMapLock.RLock()
	defer r.debugAttachmentsMapLock.RUnlock()

	da := r.debugAttachments[params.DebugAttachmentID]
	if da != nil {
		return debugattachment.NewGetDebugAttachmentOK().WithPayload(da)
	}
	return debugattachment.NewGetDebugAttachmentNotFound()
}

func contains(s string, sa []string) bool {
	for _, si := range sa {
		if s == si {
			return true
		}
	}
	return false
}

func (r *RestHandler) DebugattachmentGetDebugAttachmentsHandler(params debugattachment.GetDebugAttachmentsParams) middleware.Responder {
	node := params.Node
	state := params.State
	states := params.States
	if state != nil {
		states = append(states, *state)
	}
	wait := false
	if params.Wait != nil {
		wait = *params.Wait
	}
	names := params.Names

	ctx := params.HTTPRequest.Context()
	if params.XTimeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration((*params.XTimeout)*float64(time.Second)))
		defer cancel()
	}

	log.Info("GetDebugAttachmentsHandler called!")

	var listener chan int

	var debugattachments []*models.DebugAttachment
	filter := func() {
		r.debugAttachmentsMapLock.RLock()
		defer r.debugAttachmentsMapLock.RUnlock()

		for _, attachment := range r.debugAttachments {
			if node != nil && attachment.Spec != nil && *node != attachment.Spec.Node {
				continue
			}

			if len(states) > 0 && attachment.Status != nil {
				statusstate := attachment.Status.State
				cont := true
				for _, state := range states {
					if statusstate == state {
						cont = false
					}
				}
				if cont {
					continue
				}
			}
			if len(names) != 0 && !contains(attachment.Metadata.Name, names) {
				continue
			}

			debugattachments = append(debugattachments, attachment)
		}
	}

	filter()

	if wait && len(debugattachments) == 0 {
		// wait!
		// wait!
		// while the table is locked add a channel to listener list.
		listener = make(chan int, 1)
		r.addListener(listener)
		defer r.removeListener(listener)

		for {
			select {
			case <-listener:
				filter()
			case <-ctx.Done():
				// return timeout!
				log.Debug("GetDebugAttachmentsHandler timing out!")

				return debugattachment.NewGetDebugAttachmentsRequestTimeout()
			}

			if len(debugattachments) != 0 {
				break
			}

		}
	}

	return debugattachment.NewGetDebugAttachmentsOK().WithPayload(debugattachments)
}

func (r *RestHandler) addListener(listener chan int) {
	r.attachmentlistenersLock.Lock()
	defer r.attachmentlistenersLock.Unlock()
	r.attachmentlisteners = append(r.attachmentlisteners, listener)
}

func (r *RestHandler) notify() {
	r.attachmentlistenersLock.Lock()
	defer r.attachmentlistenersLock.Unlock()
	for _, l := range r.attachmentlisteners {
		select {
		case l <- 0:
		default:
		}
	}
}

func (r *RestHandler) removeListener(listener chan int) {
	r.attachmentlistenersLock.Lock()
	defer r.attachmentlistenersLock.Unlock()
	for i := range r.attachmentlisteners {
		if r.attachmentlisteners[i] == listener {
			r.attachmentlisteners[i] = nil
			r.attachmentlisteners[i] = r.attachmentlisteners[len(r.attachmentlisteners)-1]
			r.attachmentlisteners = r.attachmentlisteners[:len(r.attachmentlisteners)-1]
			return
		}
	}
}

func (r *RestHandler) DebugrequestGetDebugRequestsHandler(params debugrequest.GetDebugRequestsParams) middleware.Responder {
	r.debugRequestsMapLock.RLock()
	defer r.debugRequestsMapLock.RUnlock()
	debugrequests := make([]*models.DebugRequest, 0, len(r.debugRequests))
	for _, dr := range r.debugRequests {
		debugrequests = append(debugrequests, dr)
	}
	return debugrequest.NewGetDebugRequestsOK().WithPayload(debugrequests)
}

func (r *RestHandler) DebugrequestGetDebugRequestHandler(params debugrequest.GetDebugRequestParams) middleware.Responder {
	r.debugRequestsMapLock.RLock()
	defer r.debugRequestsMapLock.RUnlock()

	dr := r.debugRequests[params.DebugRequestID]
	if dr != nil {
		return debugrequest.NewGetDebugRequestOK().WithPayload(dr)
	}
	return debugrequest.NewGetDebugRequestNotFound()
}

type randReader struct {
	letters string
}

func newRandReader() *randReader {
	return &randReader{
		letters: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
	}
}

func (r *randReader) randByte() byte {
	return r.letters[rand.Int()%len(r.letters)]
}

func (r *randReader) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = r.randByte()
	}
	return len(p), nil
}

func randomString() string {
	r := newRandReader()
	var buf bytes.Buffer
	_, err := io.CopyN(&buf, r, 10)
	if err != nil {
		// should never happen
		panic(err)
	}

	return buf.String()
}
