package plank

import (
	"os"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/squash/pkg/api/v1"
)

type Config struct {
	Attachment v1.DebugAttachment
	Debugger   string
}

func GetConfig() Config {
	debugNamespace := os.Getenv("SQUASH_NAMESPACE")
	pod := os.Getenv("SQUASH_POD")
	container := os.Getenv("SQUASH_CONTAINER")
	daName := v1.GenDebugAttachmentName(pod, container)

	return Config{
		// TODO(mitchdraft) - read these values from k8s, don't re-generate them
		Attachment: v1.DebugAttachment{
			Metadata: core.Metadata{
				Name:      daName,
				Namespace: debugNamespace,
			},
			DebugNamespace: debugNamespace,
			Pod:            pod,
			Container:      container,
			// This is the debugger specified by the user
			// options are dlv, gdb, java, nodejs, python, etc.
			Debugger:   os.Getenv("DEBUGGER_NAME"),
			Attachment: os.Getenv("HOSTNAME"),
		},
		// This is the debugger installed in the container
		// Options are dlv or gdb
		Debugger: os.Getenv("DEBUGGER"),
	}
}
