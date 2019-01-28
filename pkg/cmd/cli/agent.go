package cli

import (
	"fmt"
	"strings"

	squashutils "github.com/solo-io/squash/pkg/utils"
	"github.com/solo-io/squash/pkg/utils/kubeutils"
	"github.com/spf13/cobra"
)

const squashAgentDescription = `Squash agent allows you to debug with RBAC enabled.
This may be desired for shared clusters. Squash supports RBAC through its agent
mode. In this mode, a squash agent runs in your cluster as a server. When a
user initiates a debug session with squash, the agent will create the debugging
pod, rather than the person who initiated the debug session. The agent will
only open debug session on pods in namespaces where you have CRD write access.
You can configure squash to use agent mode by setting the ENABLE_RBAC_MODE value
in your .squash config file.
`

func (top *Options) AgentCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "manage the squash agent",
		Long:  squashAgentDescription,
	}

	cmd.AddCommand(
		top.agentStatusCmd(),
		top.agentDeleteCmd(),
	)

	return cmd
}

func (top *Options) agentStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "list status of squash agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			nsList, err := kubeutils.GetNamespaces(top.KubeClient)
			if err != nil {
				return err
			}
			fmt.Printf("looking for squash agent in namespaces %v\n", strings.Join(nsList, ", "))
			squashDeployments, err := squashutils.ListSquashDeployments(top.KubeClient, nsList)
			if err != nil {
				return err
			}

			switch len(squashDeployments) {
			case 0:
				fmt.Println("Found no squash agent deployments")
				return nil
			case 1:
				fmt.Printf("Squash agent is deployed in namespace %v.\n", squashDeployments[0].ObjectMeta.Namespace)
				return nil
			}
			fmt.Printf("Found %v squash agents across these namespaces:\n", len(squashDeployments))
			matchNsList := []string{}
			for _, dep := range squashDeployments {
				matchNsList = append(matchNsList, dep.ObjectMeta.Namespace)
			}
			fmt.Println(strings.Join(matchNsList, ", "))
			return nil
		},
	}
	return cmd
}

func (top *Options) agentDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete squash agents by namespace",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("Please specify one namespace")
			}
			ns := args[0]
			fmt.Printf("Looking for squash agent in namespace %v\n", ns)
			squashDeployments, err := squashutils.ListSquashDeployments(top.KubeClient, []string{ns})
			if err != nil {
				return err
			}

			switch len(squashDeployments) {
			case 0:
				fmt.Println("Found no squash agent deployments")
				return nil
			default:
				count, err := squashutils.DeleteSquashDeployments(top.KubeClient, squashDeployments)
				if err != nil {
					return fmt.Errorf("Deleted %v deployments: %v", count, err)
				}
				fmt.Printf("Deleted %v deployments\n", count)
				return nil
			}
		},
	}
	return cmd
}
