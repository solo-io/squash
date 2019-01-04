package server

import (
	"context"
	"fmt"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/squash/pkg/options"
	"github.com/solo-io/squash/pkg/utils"
)

// ServerCmd is for testing only, at this point
func ServerCmd() error {

	fmt.Println("watching for changes to debug attachments")
	ctx := context.Background()
	daClient, err := utils.GetDebugAttachmentClient(ctx)
	wOpts := clients.WatchOpts{
		Ctx: ctx,
	}

	das, dErrs, err := (*daClient).Watch(options.SquashNamespace, wOpts)
	if err != nil {
		fmt.Println(err)
		return err
	}

	var cancel context.CancelFunc = func() {}
	defer func() { cancel() }()
	for {
		select {
		case err, ok := <-dErrs:
			if !ok {
				return err
			}
		case daList, ok := <-das:
			if !ok {
				return err
			}
			cancel()
			ctx, cancel = context.WithCancel(ctx)

			if len(daList) == 0 {
				continue
			}
			fmt.Printf("found %v das\n", len(daList))
		}
	}
}
