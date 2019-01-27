package cli

import (
	"context"
	"fmt"

	gokubeutils "github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/squash/pkg/kscmd"
	"github.com/solo-io/squash/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

/*
Notes on CLI design

An options struct is populated by a combination of:
- user input args
- user input flags
- env variables
- config file
- defaults

A specific command is specified by a chain of strings

The options struct is interpreted according to the command
Ideally, the options struct's format should follow the command tree's format

All commands should have an interactive mode.
Interactive mode and option validation can be implemented with this pattern:
```
if err := top.ensureParticularCmdOption(po *particularOption); err != nil {
    return err
}
```
- Methods should be built off of the root of the options tree (the "top" var in the example above). This allows sub commands to share common values.
- Sub commands should only modify their portion of the options tree. (This makes it easier to move sub commands around if we want a different organization later).

*/

func App(version string) (*cobra.Command, error) {
	opts := Options{}
	app := &cobra.Command{
		Use:   "squash",
		Short: "debug microservices with squash",
		Long: `debug microservices with squash
	Find more information at https://solo.io`,
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			// when no sub commands are specified, run in squash lite mode
			fmt.Println("Attaching debugger")
			_, err := kscmd.StartDebugContainer(opts.LiteOptions)
			return err
		},
	}

	if err := initializeOptions(&opts); err != nil {
		return &cobra.Command{}, err
	}

	app.SuggestionsMinimumDistance = 1
	app.AddCommand(
		DebugContainerCmd(&opts),
		DebugRequestCmd(&opts),
		ListCmd(&opts),
		WaitAttCmd(&opts),
		opts.DeployCmd(&opts),
	)

	app.PersistentFlags().BoolVar(&opts.Json, "json", false, "output json format")
	applyLiteFlags(&opts.LiteOptions, app.PersistentFlags())

	return app, nil
}

func initializeOptions(o *Options) error {
	ctx := context.Background()
	daClient, err := utils.GetDebugAttachmentClient(ctx)
	if err != nil {
		return err
	}
	o.ctx = ctx
	o.daClient = daClient

	restCfg, err := gokubeutils.GetConfig("", "")
	if err != nil {
		return err
	}
	kubeClient, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return err
	}
	o.KubeClient = kubeClient

	o.DeployOptions = defaultDeployOptions()
	return nil
}
