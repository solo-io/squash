package kubernetes

import "encoding/json"

type KubeAttachment struct {
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Container string `json:"container"`
}

func genericToKubeAttachment(attachment interface{}) (*KubeAttachment, error) {
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
