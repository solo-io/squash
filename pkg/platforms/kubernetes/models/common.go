package models

import (
	"encoding/json"

	v1 "github.com/solo-io/squash/pkg/api/v1"
)

type KubeAttachment struct {
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Container string `json:"container"`
}

func GenericToKubeAttachment(attachment interface{}) (*KubeAttachment, error) {
	jsonbytes, err := json.Marshal(attachment)
	if err != nil {
		return nil, err
	}
	ka := &KubeAttachment{}
	err = json.Unmarshal(jsonbytes, ka)
	if err != nil {
		return nil, err
	}
	return ka, nil

}

// TODO - refactor this because it should never error
func DebugAttachmentToKubeAttachment(da *v1.DebugAttachment) (*KubeAttachment, error) {
	ka := &KubeAttachment{
		Namespace: da.DebugNamespace,
		Pod:       da.Pod,
		Container: da.Container,
	}
	return ka, nil

}
