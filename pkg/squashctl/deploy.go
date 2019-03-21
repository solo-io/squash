package squashctl

import (
	"fmt"
	"strings"

	"github.com/solo-io/squash/pkg/demo"
	"github.com/solo-io/squash/pkg/install"
	"github.com/solo-io/squash/pkg/options"
	"github.com/spf13/cobra"
)

var defaultDemoNamespace = "default"

func (o *Options) DeployCmd() *cobra.Command {
	dOpts := &o.DeployOptions
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "deploy squash or a demo microservice",
	}

	cmd.AddCommand(
		o.deployDemoCmd(&dOpts.DemoOptions),
		o.deploySquashCmd(&dOpts.SquashProcessOptions),
	)

	return cmd
}

func (o *Options) deployDemoCmd(demoOpts *DemoOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "demo",
		Short: "deploy a demo microservice",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.ensureDemoDeployOpts(demoOpts); err != nil {
				return err
			}
			cs, err := o.getKubeClient()
			if err != nil {
				return err
			}
			switch demoOpts.DemoId {
			case demo.DemoGoGo:
				return demo.DeployGoGo(cs, demoOpts.Namespace1, demoOpts.Namespace2)
			case demo.DemoGoJava:
				return demo.DeployGoJava(cs, demoOpts.Namespace1, demoOpts.Namespace2)
			default:
				return fmt.Errorf("Please choose a valid demo option: %v", strings.Join(demo.DemoIds, ", "))
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&demoOpts.Namespace1, "demo-namespace1", "", "namespace in which to install the sample app")
	f.StringVar(&demoOpts.Namespace2, "demo-namespace2", "", "(optional) ns for second app - defaults to 'namespace' flag's value")
	f.StringVar(&demoOpts.DemoId, "demo-id", "", "which sample microservice to deploy. Options: go-go, go-java")
	return cmd
}

func (o *Options) ensureDemoDeployOpts(dOpts *DemoOptions) error {
	if dOpts.Namespace1 == "" {
		if o.Squash.Machine {
			dOpts.Namespace1 = defaultDemoNamespace
		} else {
			o.chooseAllowedNamespace(&dOpts.Namespace1, "Select a namespace for service 1.")
		}
	}
	if dOpts.Namespace2 == "" {
		if o.Squash.Machine {
			dOpts.Namespace2 = dOpts.Namespace1
		} else {
			o.chooseAllowedNamespace(&dOpts.Namespace2, "Select a namespace for service 2.")
		}
	}
	if dOpts.DemoId == "" {
		if o.Squash.Machine {
			dOpts.DemoId = defaultDemoNamespace
		} else {
			o.chooseString("Choose a demo microservice to deploy", &dOpts.DemoId, demo.DemoIds)
		}
	}
	return nil
}

func (o *Options) deploySquashCmd(spOpts *SquashProcessOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "squash",
		Short: "deploy Squash to cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.ensureSquashDeployOpts(spOpts); err != nil {
				return err
			}
			cs, err := o.getKubeClient()
			if err != nil {
				return err
			}
			return install.InstallSquash(cs, spOpts.Namespace, o.Squash.DebugContainerRepo, o.Squash.DebugContainerVersion, spOpts.Preview)
		},
	}
	f := cmd.Flags()
	f.StringVar(&spOpts.Namespace, "squash-namespace", options.SquashNamespace, "namespace in which to install Squash")
	f.BoolVar(&spOpts.Preview, "preview", false, "If set, prints Squash installation yaml without installing Squash.")
	return cmd
}
func (o *Options) ensureSquashDeployOpts(dOpts *SquashProcessOptions) error {
	// TODO(mitchdraft) - interactive mode
	return nil
}
