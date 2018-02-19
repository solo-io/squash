package server

import (
	"github.com/solo-io/squash/pkg/models"
)

type DataStore interface {
	DeleteAttachment(name string)
	UpdateAttachment(attachment *models.DebugAttachment) *models.DebugAttachment
	DeleteRequest(name string)
	UpdateRequest(request *models.DebugRequest) *models.DebugRequest
}
