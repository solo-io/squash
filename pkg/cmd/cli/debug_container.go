package cli

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/utils"
)

// DebugContainerCmd
func DebugContainerCmd(da v1.DebugAttachment) (*v1.DebugAttachment, error) {
	ctx := context.Background()
	daClient, err := utils.GetDebugAttachmentClient(ctx)
	if err != nil {
		return &v1.DebugAttachment{}, err
	}
	writeOpts := clients.WriteOpts{
		Ctx:               ctx,
		OverwriteExisting: false,
	}
	return (*daClient).Write(&da, writeOpts)

}
