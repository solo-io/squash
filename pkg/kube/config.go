package kube

import (
	"os"

	v1 "github.com/solo-io/squash/pkg/api/v1"
)

type Config struct {
	Attachment v1.DebugAttachment
	Debugger   string
}

func GetConfig() Config {
	return Config{
		Attachment: v1.DebugAttachment{
			DebugNamespace: os.Getenv("SQUASH_NAMESPACE"),
			Pod:            os.Getenv("SQUASH_POD"),
			Container:      os.Getenv("SQUASH_CONTAINER"),
			// This is the debugger specified by the user
			// options are dlv, gdb, java, nodejs, python, etc.
			Debugger: os.Getenv("DEBUGGER_NAME"),
		},
		// This is the debugger installed in the container
		// Options are dlv or gdb
		Debugger: os.Getenv("DEBUGGER"),
	}
}
