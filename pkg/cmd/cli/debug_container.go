package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	soloutils "github.com/solo-io/go-utils/v1/common"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/cmd/cli/options"
	defaults "github.com/solo-io/squash/pkg/options"
	"github.com/solo-io/squash/pkg/utils"
	"github.com/spf13/cobra"
)

// TODO(mitchdraft) - remove this or put it in a shared location
type Error struct {
	Type string
	Info string
}

func DebugContainerCmd(opts *options.Options) *cobra.Command {
	dcOpts := &opts.DebugContainer
	var debugContainerCmd = &cobra.Command{
		Use:   "debug-container image pod container [debugger]",
		Short: "debug-container adds a container type debug config",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureDebugContainerOpts(dcOpts, args); err != nil {
				return err
			}
			da := debugAttachmentFromOpts(*dcOpts)
			dbgattchment, err := debugContainer(da)
			if err != nil {
				return err
			}

			if err != nil {
				if !opts.Json {
					fmt.Println("Failed adding container - check parameter names match container on the platform. error:", err)
				} else {
					json.NewEncoder(os.Stdout).Encode(Error{Type: "unknown", Info: err.Error()})
				}
				return err
			}

			if !opts.Json {
				fmt.Println("Debug config id:", dbgattchment.Metadata.Name)
			} else {
				json.NewEncoder(os.Stdout).Encode(dbgattchment)
			}

			return nil
		},
	}

	debugContainerCmd.Flags().StringVarP(&dcOpts.Namespace, "namespace", "n", "default", "Namespace the pod belongs to")
	debugContainerCmd.Flags().StringVarP(&dcOpts.ProcessName, "processName", "p", "", "Process name to debug (defaults to the first running process)")

	return debugContainerCmd
}

func ensureDebugContainerOpts(dcOpts *options.DebugContainer, args []string) error {
	var err error
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
	dcOpts.Name = soloutils.RandStringBytes(6)
	return nil
}

func debugAttachmentFromOpts(o options.DebugContainer) v1.DebugAttachment {
	return v1.DebugAttachment{
		Metadata: core.Metadata{
			Name:      o.Name,
			Namespace: defaults.SquashNamespace,
		},
		Pod:            o.Pod,
		Container:      o.Container,
		ProcessName:    o.ProcessName,
		Image:          o.Image,
		Debugger:       o.DebuggerType,
		DebugNamespace: o.Namespace,
	}
}

func debugContainer(da v1.DebugAttachment) (*v1.DebugAttachment, error) {
	ctx := context.Background()
	daClient, err := utils.GetDebugAttachmentClient(ctx)
	if err != nil {
		return &v1.DebugAttachment{}, err
	}
	writeOpts := clients.WriteOpts{
		Ctx:               ctx,
		OverwriteExisting: false,
	}

	return (*daClient).Write(&da, writeOpts)
}
