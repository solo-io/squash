package kube

import (
	"os"

	"github.com/solo-io/squash/pkg/api/v1"
)

type Config struct {
	Attachment v1.DebugAttachment
	Debugger   string
	Server     bool
}

func GetConfig() Config {
	return Config{
		Attachment: v1.DebugAttachment{
			DebugNamespace: os.Getenv("SQUASH_NAMESPACE"),
			Pod:            os.Getenv("SQUASH_POD"),
			Container:      os.Getenv("SQUASH_CONTAINER"),
		},
		Debugger: os.Getenv("DEBUGGER"),
		Server:   os.Getenv("DEBUGGER_SERVER") == "1",
	}
}
