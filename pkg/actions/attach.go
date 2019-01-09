package actions

import (
	"fmt"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/squash/pkg/api/v1"
)

func (uc *UserController) Attach(name, namespace, image, pod, container, processName, dbgger string) (*v1.DebugAttachment, error) {
	fmt.Println("attaching...................")
	fmt.Println(pod, container, image, processName, dbgger)
	da := v1.DebugAttachment{
		Metadata: core.Metadata{
			Name:      name,
			Namespace: namespace,
		},
		Debugger:       dbgger,
		Image:          image,
		Pod:            pod,
		Container:      container,
		DebugNamespace: namespace,
		State:          v1.DebugAttachment_PendingAttachment,
	}
	if processName != "" {
		da.ProcessName = processName
	}
	writeOpts := clients.WriteOpts{
		Ctx:               uc.ctx,
		OverwriteExisting: false,
	}
	return (*uc.daClient).Write(&da, writeOpts)
}
