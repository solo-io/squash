package squashctl

import (
	"fmt"
	"strings"

	squashutils "github.com/solo-io/squash/pkg/utils"
	"github.com/solo-io/squash/pkg/utils/kubeutils"
	"github.com/spf13/cobra"
)

const squashCommandDescription = `Squash allows you to debug with RBAC enabled.
This may be desired for shared clusters. Squash supports RBAC through its secure.
mode. In this mode, a Squash runs in your cluster as a server. When a
user initiates a debug session with Squash, Squash will create the debugging
pod, rather than the person who initiated the debug session. Squash will
only open debug session on pods in namespaces where you have CRD write access.
You can configure squash to use secure mode by setting the secure_mode value
in your .squash config file.
`

func (o *Options) SquashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "squash",
		Short: "manage the squash",
		Long:  squashCommandDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(
		o.squashStatusCmd(),
		o.squashDeleteCmd(),
	)

	return cmd
}

func (o *Options) squashStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "list status of Squash process",
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := o.getKubeClient()
			if err != nil {
				return err
			}
			nsList, err := kubeutils.GetNamespaces(cs)
			if err != nil {
				return err
			}
			fmt.Printf("looking for Squash process in namespaces %v\n", strings.Join(nsList, ", "))
			squashDeployments, err := squashutils.ListSquashDeployments(cs, nsList)
			if err != nil {
				return err
			}

			switch len(squashDeployments) {
			case 0:
				fmt.Println("Found no Squash deployments")
				return nil
			case 1:
				fmt.Printf("Squash is deployed in namespace %v.\n", squashDeployments[0].ObjectMeta.Namespace)
				return nil
			}
			fmt.Printf("Found %v Squash processes across these namespaces:\n", len(squashDeployments))
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

func (o *Options) squashDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete Squash processes from your cluster by namespace",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("Please specify one namespace")
			}
			ns := args[0]
			fmt.Printf("Looking for Squash process in namespace %v\n", ns)
			cs, err := o.getKubeClient()
			if err != nil {
				return err
			}
			squashDeployments, err := squashutils.ListSquashDeployments(cs, []string{ns})
			if err != nil {
				return err
			}

			switch len(squashDeployments) {
			case 0:
				fmt.Println("Found no Squash deployments")
				return nil
			default:
				count, err := squashutils.DeleteSquashDeployments(cs, squashDeployments)
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
