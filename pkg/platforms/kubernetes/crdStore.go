package kubernetes

import (
	log "github.com/Sirupsen/logrus"
	"github.com/solo-io/squash/pkg/models"
	"github.com/solo-io/squash/pkg/platforms/kubernetes/crd"
	"github.com/solo-io/squash/pkg/server"
	"k8s.io/client-go/rest"
)

func NewCRDDataStore(cfg *rest.Config, data *server.ServerData) (*CRDStore, error) {
	c, err := crd.NewCrdClient(cfg, NewCRDStoreCallback(data))
	if err != nil {
		return nil, err
	}

	return &CRDStore{
		cfg:    cfg,
		data:   data,
		client: c,
	}, nil
}

type CRDStore struct {
	cfg    *rest.Config
	client *crd.CrdClient
	data   *server.ServerData
}

type CRDStoreCallback struct {
	data *server.ServerData
}

func NewCRDStoreCallback(data *server.ServerData) *CRDStoreCallback {
	return &CRDStoreCallback{data: data}
}

func (r *CRDStore) DeleteAttachment(name string) {
	r.client.DeleteAttachment(name)
}

func (r *CRDStore) UpdateAttachment(attachment *models.DebugAttachment) *models.DebugAttachment {
	a := attachment
	var err error
	if r.data.GetDebugAttachmentNoLock(attachment.Metadata.Name) != nil {
		// Update
		a, err = r.client.UpdateAttachment(a)
		if err != nil {
			log.WithField("error", err).Error("Update attachment failed")
			return attachment
		}
	} else {
		// Add
		a, err = r.client.CreateAttachment(a, true)
		if err != nil {
			log.WithField("error", err).Error("Add attachment failed")
			return attachment
		}
	}
	return a
}

func (r *CRDStore) DeleteRequest(name string) {
	r.client.DeleteRequest(name)
}

func (r *CRDStore) UpdateRequest(request *models.DebugRequest) *models.DebugRequest {
	a := request
	var err error
	if r.data.GetDebugRequestNoLock(request.Metadata.Name) != nil {
		// Update
		a, err = r.client.UpdateRequest(a)
		if err != nil {
			log.WithField("error", err).Error("Update request failed")
			return request
		}
	} else {
		// Add
		a, err = r.client.CreateRequest(a, true)
		if err != nil {
			log.WithField("error", err).Error("Add request failed")
			return request
		}
	}
	return a
}

// SyncCrdChanges is a callback from the CRD controller to sync chnages from CRDs back to Squash internal storage
func (r *CRDStoreCallback) SyncCRDChanges(objtype crd.ObjectType, action crd.Action, obj interface{}) {
	switch objtype {
	case crd.TypeDebugAttachment:
		att := obj.(*models.DebugAttachment)
		r.syncDebugAttachment(att, action)
	case crd.TypeDebugRequest:
		req := obj.(*models.DebugRequest)
		r.syncDebugRequest(req, action)
	default:
		log.WithField("type", objtype).Error("SyncCrdChanges - Unknown object type")
	}
}

func (r *CRDStoreCallback) syncDebugAttachment(att *models.DebugAttachment, action crd.Action) {
	// TODO: check object version and make sure to execute action only if incoming version is newer
	switch action {
	case crd.ActionAdd:
		r.data.UpdateDebugAttachment(att, nil)
	case crd.ActionUpdate:
		r.data.UpdateDebugAttachment(att, nil)
	case crd.ActionDelete:
		r.data.DeleteDebugAttachment(att.Metadata.Name, nil)
	default:
		log.WithField("action", action).Error("syncDebugAttachment - Unknown action type")
	}
}

func (r *CRDStoreCallback) syncDebugRequest(req *models.DebugRequest, action crd.Action) {
	// TODO: check object version and make sure to execute action only if incoming version is newer
	switch action {
	case crd.ActionAdd:
		r.data.UpdateDebugRequest(req, nil)
	case crd.ActionUpdate:
		r.data.UpdateDebugRequest(req, nil)
	case crd.ActionDelete:
		r.data.DeleteDebugRequest(req.Metadata.Name, nil)
	default:
		log.WithField("action", action).Error("syncDebugRequest - Unknown action type")
	}
}
