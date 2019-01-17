package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/squash/pkg/options"
	"github.com/solo-io/squash/pkg/utils"
)

// this program exists to support the Squash development cycle
func main() {
	namespace := options.SquashClientNamespace
	if err := ServerCmd(namespace); err != nil {
		fmt.Println(err)
	}
}

func usage() {
	use := `This is a helper utility for watching changes to squash resources.
Suggestion: run this in a side terminal while you are making changes.
Edit the print loop to show the stats you care about.
`
	fmt.Println(use)
}

func ServerCmd(namespace string) error {

	usage()
	fmt.Printf("watching for changes to debug attachments in ns: %v\n", namespace)
	ctx := context.Background()
	daClient, err := utils.GetDebugAttachmentClient(ctx)
	wOpts := clients.WatchOpts{
		Ctx: ctx,
	}

	das, dErrs, err := (*daClient).Watch(namespace, wOpts)
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

			fmt.Println("-----")
			strs := []string{fmt.Sprintf("found %v das", len(daList))}
			for _, da := range daList {
				strs = append(strs, fmt.Sprintf("name: %v, stat: %v", da.Metadata.Name, da.State))
			}
			for _, da := range daList {
				strs = append(strs, fmt.Sprintf("name: %v, durl: %v", da.Metadata.Name, da.DebugServerAddress))
			}
			fmt.Printf(strings.Join(strs, "\n"))
			if err != nil {
				// TODO(mitchdraft) move this into an event loop
				fmt.Println(err)
			}
			fmt.Println("")
			fmt.Println("=====")
		}
	}
}
