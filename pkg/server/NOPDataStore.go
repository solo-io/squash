package server

import (
	"github.com/solo-io/squash/pkg/models"
)

func NewNOPDataStore() DataStore {
	return &NOPDataStore{}
}

type NOPDataStore struct{}

func (*NOPDataStore) DeleteAttachment(name string) {}

func (*NOPDataStore) UpdateAttachment(attachment *models.DebugAttachment) *models.DebugAttachment {
	return attachment
}

func (*NOPDataStore) DeleteRequest(name string) {}

func (*NOPDataStore) UpdateRequest(request *models.DebugRequest) *models.DebugRequest {
	return request
}
