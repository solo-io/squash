package server

import (
	"sync"

	"github.com/solo-io/squash/pkg/models"
)

type ServerData struct {
	debugAttachments        map[string]*models.DebugAttachment
	debugAttachmentsMapLock sync.RWMutex
	debugRequests           map[string]*models.DebugRequest
	debugRequestsMapLock    sync.RWMutex
}

func NewServerData() *ServerData {
	return &ServerData{
		debugAttachments: make(map[string]*models.DebugAttachment),
		debugRequests:    make(map[string]*models.DebugRequest),
	}
}

func (r *ServerData) GetAttachments() *map[string]*models.DebugAttachment { return &r.debugAttachments }
func (r *ServerData) GetRequests() *map[string]*models.DebugRequest       { return &r.debugRequests }
func (r *ServerData) GetAttachmentsLock() *sync.RWMutex                   { return &r.debugAttachmentsMapLock }
func (r *ServerData) GetRequestsLock() *sync.RWMutex                      { return &r.debugRequestsMapLock }

func (r *ServerData) GetDebugRequestNoLock(name string) *models.DebugRequest {
	return r.debugRequests[name]
}

func (r *ServerData) GetDebugAttachmentNoLock(name string) *models.DebugAttachment {
	return r.debugAttachments[name]
}

func (r *ServerData) FindUnboundDebugRequest(dbgattachment *models.DebugAttachment) *models.DebugRequest {
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

func (r *ServerData) UpdateDebugRequest(dr *models.DebugRequest, store DataStore) {
	r.debugRequestsMapLock.Lock()
	defer r.debugRequestsMapLock.Unlock()
	if store != nil {
		r.debugRequests[dr.Metadata.Name] = store.UpdateRequest(dr)
	} else {
		r.debugRequests[dr.Metadata.Name] = dr
	}
}

func (r *ServerData) GetDebugRequest(name string) *models.DebugRequest {
	r.debugRequestsMapLock.RLock()
	defer r.debugRequestsMapLock.RUnlock()

	return r.debugRequests[name]
}

func (r *ServerData) DeleteDebugRequest(name string, store DataStore) {
	r.debugRequestsMapLock.Lock()
	defer r.debugRequestsMapLock.Unlock()
	delete(r.debugRequests, name)
	if store != nil {
		store.DeleteRequest(name)
	}
}

func (r *ServerData) UpdateDebugAttachment(a *models.DebugAttachment, store DataStore) {
	r.debugAttachmentsMapLock.Lock()
	defer r.debugAttachmentsMapLock.Unlock()
	if store != nil {
		r.debugAttachments[a.Metadata.Name] = store.UpdateAttachment(a)
	} else {
		r.debugAttachments[a.Metadata.Name] = a
	}
}

func (r *ServerData) GetDebugAttachment(name string) *models.DebugAttachment {
	r.debugAttachmentsMapLock.RLock()
	defer r.debugAttachmentsMapLock.RUnlock()
	return r.debugAttachments[name]
}

func (r *ServerData) DeleteDebugAttachment(name string, store DataStore) {
	r.debugAttachmentsMapLock.Lock()
	defer r.debugAttachmentsMapLock.Unlock()
	delete(r.debugAttachments, name)
	if store != nil {
		store.DeleteAttachment(name)
	}
}
