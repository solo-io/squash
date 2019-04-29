package squashctl

import (
	"fmt"
	"strings"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/squash/pkg/api/v1"
	sqOpts "github.com/solo-io/squash/pkg/options"
	squashutils "github.com/solo-io/squash/pkg/utils"
	"github.com/solo-io/squash/pkg/utils/kubeutils"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (o *Options) UtilsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "utils",
		Short:   "call various squash utils",
		Example: "squash utils list-attachments",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(
		o.listAttachmentsCmd(),
		o.deletePermissionsCmd(),
		o.deletePlankPodsCmd(),
		o.deleteAttachmentsCmd(),
		o.registerResourcesCmd(),
	)

	return cmd
}

func (o *Options) listAttachmentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-attachments",
		Short: "list all existing debug attachments",
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := o.getKubeClient()
			if err != nil {
				return err
			}
			nsList, err := kubeutils.GetNamespaces(cs)
			if err != nil {
				return err
			}
			daClient, err := o.getDAClient()
			if err != nil {
				return err
			}
			das, err := squashutils.ListDebugAttachments(o.ctx, daClient, nsList)
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

func (o *Options) deleteAttachmentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-attachments",
		Short: "delete all existing debug attachments and plank pods",
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := o.getKubeClient()
			if err != nil {
				return err
			}
			nsList, err := kubeutils.GetNamespaces(cs)
			if err != nil {
				return err
			}
			daClient, err := o.getDAClient()
			if err != nil {
				return err
			}
			das, err := squashutils.GetAllDebugAttachments(o.ctx, daClient, nsList)
			if err != nil {
				return err
			}

			fmt.Printf("Found %v debug attachments\n", len(das))
			if err := o.deleteAttachmentList(das, true); err != nil {
				fmt.Println(err)
			}
			return o.deletePlankPods()
		},
	}
	return cmd
}

func (o *Options) registerResourcesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-resources",
		Short: "register the custom resource definitions (CRDs) needed by squash",
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := o.getKubeClient()
			if err != nil {
				return err
			}
			nsList, err := kubeutils.GetNamespaces(cs)
			if err != nil {
				return err
			}
			daClient, err := squashutils.GetDebugAttachmentClientWithRegistration(o.ctx)
			if err != nil {
				return err
			}
			// do a trivial operation with the client to ensure that it is expressed
			_, err = squashutils.GetAllDebugAttachments(o.ctx, daClient, nsList)
			if err != nil {
				return err
			}

			fmt.Println("Registered DebugAttachment CRD")
			return nil
		},
	}
	return cmd
}

func (o *Options) deleteAttachmentList(das v1.DebugAttachmentList, continueOnError bool) error {
	daClient, err := o.getDAClient()
	if err != nil {
		return err
	}
	for _, da := range das {
		if err := daClient.Delete(da.Metadata.Namespace, da.Metadata.Name, clients.DeleteOpts{}); err != nil {
			if continueOnError {
				fmt.Println(err)
			} else {
				return err
			}
		}
	}
	return nil
}

func (o *Options) deletePermissionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-permissions",
		Short: "remove all service accounts, roles, and role bindings created by Squash.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.deleteSquashPermissions()
		},
	}
	return cmd
}

func (o *Options) deleteSquashPermissions() error {
	cs, err := o.getKubeClient()
	if err != nil {
		return err
	}
	namespace := o.Squash.SquashNamespace

	if err := cs.CoreV1().ServiceAccounts(namespace).Delete(sqOpts.PlankServiceAccountName, &metav1.DeleteOptions{}); err != nil {
		fmt.Println(err)
	}
	if err := cs.Rbac().ClusterRoles().Delete(sqOpts.PlankClusterRoleName, &metav1.DeleteOptions{}); err != nil {
		fmt.Println(err)
	}
	if err := cs.Rbac().ClusterRoleBindings().Delete(sqOpts.PlankClusterRoleBindingName, &metav1.DeleteOptions{}); err != nil {
		fmt.Println(err)
	}

	if err := cs.CoreV1().ServiceAccounts(namespace).Delete(sqOpts.SquashServiceAccountName, &metav1.DeleteOptions{}); err != nil {
		fmt.Println(err)
	}
	if err := cs.Rbac().ClusterRoles().Delete(sqOpts.SquashClusterRoleName, &metav1.DeleteOptions{}); err != nil {
		fmt.Println(err)
	}
	if err := cs.Rbac().ClusterRoleBindings().Delete(sqOpts.SquashClusterRoleBindingName, &metav1.DeleteOptions{}); err != nil {
		fmt.Println(err)
	}
	return nil
}

func (o *Options) deletePlankPodsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-planks",
		Short: "remove all plank debugger pods created by Squash.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.deletePlankPods()
		},
	}
	return cmd
}

// TODO(mitchdraft) - should exclude squash pod from this, add labels to squash and plank pods so they can be distinguished
func (o *Options) deletePlankPods() error {
	cs, err := o.getKubeClient()
	if err != nil {
		return err
	}
	namespace := o.Squash.SquashNamespace
	planks, err := cs.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: sqOpts.PlankLabelSelectorString})
	if err != nil {
		return err
	}
	fmt.Printf("Found %v plank pods in namespace %v\n", len(planks.Items), namespace)
	for _, plank := range planks.Items {
		name := plank.ObjectMeta.Name
		if err := cs.CoreV1().Pods(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
			return err
		}
		fmt.Printf("Deleted plank pod %v.\n", name)
	}

	return nil
}
