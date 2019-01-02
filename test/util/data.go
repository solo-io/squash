package util_test

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/squash/pkg/api/v1"
)

func generateDebugAttachment(name, namespace, dbgger, image, pod, container, processName string) v1.DebugAttachment {

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
func generateDebugAttachmentDlv1(name, namespace string) v1.DebugAttachment {
	dbgger := "dlv"
	image := "mk"
	pod := "somepod"
	container := "somecontainer"
	processName := "pcsnm"
	return generateDebugAttachment(name, namespace, dbgger, image, pod, container, processName)
}
