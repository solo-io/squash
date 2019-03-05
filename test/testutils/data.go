package testutils

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/squash/pkg/api/v1"
)

func GenerateDebugAttachment(name, namespace, dbgger, image, pod, container, processName string) v1.DebugAttachment {

	da := v1.DebugAttachment{
		Metadata: core.Metadata{
			Name:      name,
			Namespace: namespace,
		},
		Debugger:  dbgger,
		Image:     image,
		Pod:       pod,
		Container: container,
	}
	if processName != "" {
		da.ProcessName = processName
	}
	return da
}
func GenerateDebugAttachmentDlv1(name, namespace string) v1.DebugAttachment {
	dbgger := "dlv"
	image := "mk"
	pod := "somepod"
	container := "somecontainer"
	processName := "pcsnm"
	return GenerateDebugAttachment(name, namespace, dbgger, image, pod, container, processName)
}
