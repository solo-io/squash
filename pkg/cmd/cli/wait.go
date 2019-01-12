package cli

import (
	"context"
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/options"
	"github.com/solo-io/squash/pkg/utils"
)

// WaitCmd
// TODO - make this more useful, read until you find one in the Attached state
func WaitCmd(namespace string, name string, timeout float64) (v1.DebugAttachment, error) {
	ctx := context.Background()
	daClient, err := utils.GetDebugAttachmentClient(ctx)
	if err != nil {
		return v1.DebugAttachment{}, err
	}
	reOptions := clients.ReadOpts{
		Ctx: ctx,
	}
	da, err := (*daClient).Read(namespace, name, reOptions)
	if err != nil {
		// TODO(mitchdraft) implement a periodic check instead of waiting the full timeout duration
		time.Sleep(time.Duration(int32(timeout)) * time.Second)
		da, err = (*daClient).Read(options.SquashClientNamespace, name, reOptions)
		if err != nil {
			return v1.DebugAttachment{}, err
		}
	}
	return *da, nil

}
