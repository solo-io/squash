package server

import (
	"context"
	"fmt"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/squash/pkg/api/v1"
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

			fmt.Printf("found %v das\n", len(daList))
			// err := sync(daList)
			if err != nil {
				// TODO(mitchdraft) move this into an event loop
				fmt.Println(err)
			}
			// if len(daList) == 0 {
			// 	continue
			// }
			// fmt.Printf("found %v das\n", len(daList))
		}
	}
}

// func sync(daList v1.DebugAttachmentList) error {
// 	fmt.Println("running sync")
// 	debuggers.HandleAddedRemovedAttachments(daList, []*v1.DebugAttachmentList{})
// 	// for _, d := range daList {
// 	// 	fmt.Println(d)
// 	// 	fmt.Println(d.Status.State)
// 	// 	if d.Status.State == core.Status_Pending {
// 	// 		if err := handlePendingDebugAttachment(d); err != nil {
// 	// 			return err
// 	// 		}
// 	// 	}
// 	// }
// 	return nil
// }

func handlePendingDebugAttachment(d *v1.DebugAttachment) error {
	// try to connect to the debugger
	// if successful, write "Accepted" to the status
	ctx := context.TODO()
	daClient, err := utils.GetDebugAttachmentClient(ctx)
	if err != nil {
		return err
	}
	// TODO - WIP Only do this if the debugger is actually connected
	d.Status.State = core.Status_Accepted
	o, err := (*daClient).Write(d, clients.WriteOpts{
		Ctx:               ctx,
		OverwriteExisting: true,
	})
	if err != nil {
		return err
	}
	fmt.Println(o)

	return nil
}
