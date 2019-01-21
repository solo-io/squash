package kube

import (
	"os"

	k8models "github.com/solo-io/squash/pkg/platforms/kubernetes/models"
)

type Config struct {
	Attachment k8models.KubeAttachment
	Debugger   string
	Server     bool
}

func GetConfig() Config {
	return Config{
		Attachment: k8models.KubeAttachment{
			Namespace: os.Getenv("SQUASH_NAMESPACE"),
			Pod:       os.Getenv("SQUASH_POD"),
			Container: os.Getenv("SQUASH_CONTAINER"),
		},
		Debugger: os.Getenv("DEBUGGER"),
		Server:   os.Getenv("DEBUGGER_SERVER") == "1",
	}
}
