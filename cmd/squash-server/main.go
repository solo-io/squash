package main

import (
	"context"
	"fmt"
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/squash/pkg/options"
	"github.com/solo-io/squash/pkg/utils"
)

func main() {
	fmt.Println("running squash server")
	ctx := context.Background()
	daClient, err := utils.GetDebugAttachmentClient(ctx)
	wOpts := clients.WatchOpts{
		Ctx:         ctx,
		RefreshRate: time.Second,
	}

	das, dErrs, err := (*daClient).Watch(options.SquashNamespace, wOpts)
	if err != nil {
		fmt.Println(err)
		// return nil, errors.Wrapf(err, "beginning debug attachments watch")
	}
	fmt.Println("not in the loop")
	go func() {
		fmt.Println("in the loop")
		var cancel context.CancelFunc = func() {}
		defer func() { cancel() }()
		for {
			select {
			case err, ok := <-dErrs:
				if !ok {
					fmt.Println(err)
					return
				}
				// publsherr(err)
			case daList, ok := <-das:
				if !ok {
					return
				}
				cancel()
				ctx, cancel = context.WithCancel(ctx)

				if len(daList) == 0 {
					continue
				}
				fmt.Printf("found %v das\n", len(daList))

			}
		}
	}()
	time.Sleep(4 * time.Second)
}
