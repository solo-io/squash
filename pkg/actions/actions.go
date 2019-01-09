package actions

import (
	"fmt"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/squash/pkg/api/v1"
)

// Attach creates a DebugAttachment with a state of PendingAttachment
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

// Remove sets the DebugAttachment state to PendingDelete
func (uc *UserController) Delete(namespace, name string) (*v1.DebugAttachment, error) {
	fmt.Println("removing...................")
	fmt.Println(name, namespace)

	da, err := (*uc.daClient).Read(namespace, name, clients.ReadOpts{Ctx: uc.ctx})
	if err != nil {
		return &v1.DebugAttachment{}, err
	}
	da.State = v1.DebugAttachment_PendingDelete

	writeOpts := clients.WriteOpts{
		Ctx:               uc.ctx,
		OverwriteExisting: true,
	}
	return (*uc.daClient).Write(da, writeOpts)
}

// Counts returns the number of debug attachments by type
func (uc *UserController) Counts(namespace, name string) (map[v1.DebugAttachment_State]int, error) {
	counts := make(map[v1.DebugAttachment_State]int)

	das, err := (*uc.daClient).List(namespace, clients.ListOpts{Ctx: uc.ctx})
	if err != nil {
		return counts, err
	}

	for _, da := range das {
		if current, ok := counts[da.State]; ok {
			counts[da.State] = current + 1
		} else {
			counts[da.State] = 1
		}
	}
	return counts, nil
}
