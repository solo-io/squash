package squashctl

import (
	"fmt"
	"strings"

	"github.com/solo-io/squash/pkg/demo"
	"github.com/solo-io/squash/pkg/install"
	"github.com/spf13/cobra"
)

var defaultDemoNamespace = "default"

func (top *Options) DeployCmd(o *Options) *cobra.Command {
	dOpts := &o.DeployOptions
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "deploy the squash agent or a demo microservice",
	}

	cmd.AddCommand(
		top.deployDemoCmd(&dOpts.DemoOptions),
		top.deployAgentCmd(&dOpts.AgentOptions),
	)

	return cmd
}

func (top *Options) deployDemoCmd(demoOpts *DemoOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "demo",
		Short: "deploy a demo microservice",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := top.ensureDemoDeployOpts(demoOpts); err != nil {
				return err
			}
			switch demoOpts.DemoId {
			case "default":
				fallthrough
			case demo.DemoGoGo:
				fmt.Println("using go-go")
				return demo.DeployGoGo(top.KubeClient, demoOpts.Namespace1, demoOpts.Namespace2)
			case demo.DemoGoJava:
				fmt.Println("using go-java")
				return demo.DeployGoJava(top.KubeClient, demoOpts.Namespace1, demoOpts.Namespace2)
			default:
				return fmt.Errorf("Please choose a valid demo option: %v", strings.Join(demo.DemoIds, ", "))
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&demoOpts.Namespace1, "demoNamespace1", "", "namespace in which to install the sample app")
	f.StringVar(&demoOpts.Namespace2, "demoNamespace2", "", "(optional) ns for second app - defaults to 'namespace' flag's value")
	f.StringVar(&demoOpts.DemoId, "demoId", "default", "which sample microservice to deploy. Options: go-go, go-java")
	return cmd
}

func (top *Options) ensureDemoDeployOpts(dOpts *DemoOptions) error {
	if dOpts.Namespace1 == "" {
		if top.Squash.Machine {
			dOpts.Namespace1 = defaultDemoNamespace
		} else {
			top.chooseAllowedNamespace(&dOpts.Namespace1, "Select a namespace for service 1.")
		}
	}
	if dOpts.Namespace2 == "" {
		if top.Squash.Machine {
			dOpts.Namespace2 = dOpts.Namespace1
		} else {
			top.chooseAllowedNamespace(&dOpts.Namespace2, "Select a namespace for service 2.")
		}
	}
	return nil
}

func (top *Options) deployAgentCmd(agentOpts *AgentOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "deploy a squash agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := top.ensureAgentDeployOpts(agentOpts); err != nil {
				return err
			}
			return install.InstallAgent(top.KubeClient, agentOpts.Namespace)
		},
	}
	f := cmd.Flags()
	f.StringVar(&agentOpts.Namespace, "agentNamespace", install.DefaultNamespace, "namespace in which to install the sample app")
	return cmd
}
func (top *Options) ensureAgentDeployOpts(dOpts *AgentOptions) error {
	// TODO(mitchdraft) - interactive mode
	return nil
}
