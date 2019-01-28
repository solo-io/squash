package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	gokubeutils "github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/utils/kubeutils"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

func DebugContainerCmd(o *Options) *cobra.Command {
	dcOpts := &o.DebugContainer
	var debugContainerCmd = &cobra.Command{
		Use:   "debug-container image pod container [debugger]",
		Short: "debug-container adds a container type debug config",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureDebugContainerOpts(dcOpts, args); err != nil {
				return err
			}
			da, err := debugAttachmentFromOpts(*dcOpts)
			if err != nil {
				return err
			}

			dbgattchment, err := o.debugContainer(da)
			if err != nil {
				return err
			}

			if err != nil {
				if !o.Json {
					fmt.Println("Failed adding container - check parameter names match container on the platform. error:", err)
				} else {
					json.NewEncoder(os.Stdout).Encode(Error{Type: "unknown", Info: err.Error()})
				}
				return err
			}

			if !o.Json {
				fmt.Println("Debug config id:", dbgattchment.Metadata.Name)
			} else {
				// TODO - convert
				json.NewEncoder(os.Stdout).Encode(dbgattchment)
			}

			return nil
		},
	}

	debugContainerCmd.PersistentFlags().StringVarP(&dcOpts.Namespace, "namespace", "n", "default", "Namespace the pod belongs to")
	debugContainerCmd.PersistentFlags().StringVarP(&dcOpts.ProcessName, "processName", "p", "", "Process name to debug (defaults to the first running process)")

	return debugContainerCmd
}

func ensureDebugContainerOpts(dcOpts *DebugContainer, args []string) error {
	var err error
	dcOpts.DebuggerType = "gdb"
	switch len(args) {
	case 4:
		dcOpts.DebuggerType = args[3]
		fallthrough
	case 3:
		dcOpts.Image = args[0]
		dcOpts.Pod = args[1]
		dcOpts.Container = args[2]
	default:
		err = errors.New("invalid number of arguments")
	}
	if err != nil {
		return err
	}
	dcOpts.Name = RandKubeNameBytes(6)
	return nil
}

func debugAttachmentFromOpts(dc DebugContainer) (v1.DebugAttachment, error) {
	restCfg, err := gokubeutils.GetConfig("", "")
	if err != nil {
		return v1.DebugAttachment{}, err
	}
	kubeResClient, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return v1.DebugAttachment{}, err
	}
	ns, err := kubeutils.GetPodNamespace(kubeResClient, dc.Pod)
	if err != nil {
		return v1.DebugAttachment{}, err
	}
	return v1.DebugAttachment{
		Metadata: core.Metadata{
			Name:      dc.Name,
			Namespace: ns,
		},
		Debugger:       dc.DebuggerType,
		Image:          dc.Image,
		Pod:            dc.Pod,
		Container:      dc.Container,
		DebugNamespace: ns,
		State:          v1.DebugAttachment_RequestingAttachment,
		ProcessName:    dc.ProcessName,
	}, nil
}

func (o *Options) debugContainer(da v1.DebugAttachment) (*v1.DebugAttachment, error) {
	writeOpts := clients.WriteOpts{
		Ctx:               o.ctx,
		OverwriteExisting: false,
	}

	return (*o.daClient).Write(&da, writeOpts)
}
