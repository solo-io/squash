package kube

import (
	"fmt"
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

	return Config{
		Attachment: v1.DebugAttachment{
			Metadata: core.Metadata{
				Name:      fmt.Sprintf("%v-%v", pod, container),
				Namespace: debugNamespace,
			},
			DebugNamespace: debugNamespace,
			Pod:            pod,
			Container:      container,
			// This is the debugger specified by the user
			// options are dlv, gdb, java, nodejs, python, etc.
			Debugger: os.Getenv("DEBUGGER_NAME"),
		},
		// This is the debugger installed in the container
		// Options are dlv or gdb
		Debugger: os.Getenv("DEBUGGER"),
	}
}
