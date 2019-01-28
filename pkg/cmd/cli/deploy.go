package cli

import (
	"fmt"
	"strings"

	"github.com/solo-io/squash/pkg/demo"
	"github.com/solo-io/squash/pkg/install"
	"github.com/spf13/cobra"
)

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
			if err := ensureDemoDeployOpts(demoOpts); err != nil {
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
	f.StringVar(&demoOpts.Namespace1, "demoNamespace1", "default", "namespace in which to install the sample app")
	f.StringVar(&demoOpts.Namespace2, "demoNamespace2", "", "(optional) ns for second app - defaults to 'namespace' flag's value")
	f.StringVar(&demoOpts.DemoId, "demoId", "default", "which sample microservice to deploy. Options: go-go, go-java")
	return cmd
}

func ensureDemoDeployOpts(dOpts *DemoOptions) error {
	// TODO(mitchdraft) - interactive mode
	return nil
}

func (top *Options) deployAgentCmd(agentOpts *AgentOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "deploy a squash agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureAgentDeployOpts(agentOpts); err != nil {
				return err
			}
			return install.InstallAgent(top.KubeClient, agentOpts.Namespace)
		},
	}
	f := cmd.Flags()
	f.StringVar(&agentOpts.Namespace, "agentNamespace", install.DefaultNamespace, "namespace in which to install the sample app")
	return cmd
}
func ensureAgentDeployOpts(dOpts *AgentOptions) error {
	// TODO(mitchdraft) - interactive mode
	return nil
}
