package squashctl

import (
	"fmt"
	"strings"

	squashutils "github.com/solo-io/squash/pkg/utils"
	"github.com/solo-io/squash/pkg/utils/kubeutils"
	"github.com/spf13/cobra"
)

func (top *Options) UtilsCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "utils",
		Short:   "call various squash utils",
		Example: "squash utils list-attachments",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(
		top.listAttachmentsCmd(),
	)

	return cmd
}

func (top *Options) listAttachmentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-attachments",
		Short: "list all existing debug attachments",
		RunE: func(cmd *cobra.Command, args []string) error {
			nsList, err := kubeutils.GetNamespaces(top.KubeClient)
			if err != nil {
				return err
			}
			das, err := squashutils.ListDebugAttachments(top.ctx, top.daClient, nsList)
			if err != nil {
				return err
			}

			if len(das) == 0 {
				fmt.Println("Found no debug attachments")
				return nil
			}
			fmt.Printf("Existing debug attachments:\n")
			fmt.Println(strings.Join(das, "\n"))
			return nil
		},
	}
	return cmd
}
