package cli

import (
	"context"

	gokubeutils "github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/squash/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

func App(version string) (*cobra.Command, error) {
	app := &cobra.Command{
		Use:   "squash",
		Short: "debug microservices with squash",
		Long: `debug microservices with squash
	Find more information at https://solo.io`,
		Version: version,
	}

	opts := Options{}
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
