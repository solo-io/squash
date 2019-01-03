package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/cmd/cli/options"
	"github.com/solo-io/squash/pkg/utils"
	"github.com/spf13/cobra"
)

// TODO(mitchdraft) - remove this or put it in a shared location
type Error struct {
	Type string
	Info string
}

func DebugContainerCmd(opts *options.Options) *cobra.Command {
	var debugContainerCmd = &cobra.Command{
		Use:   "debug-container image pod container [debugger]",
		Short: "debug-container adds a container type debug config",
		RunE: func(cmd *cobra.Command, args []string) error {
			var image, pod, container, debuggertype string
			debuggertype = "gdb"
			var err error
			switch len(args) {
			case 4:
				debuggertype = args[3]
				fallthrough
			case 3:
				image, pod, container = args[0], args[1], args[2]
			default:
				err = errors.New("invalid number of arguments")
			}
			if err != nil {
				return err
			}

			da := v1.DebugAttachment{
				Metadata: core.Metadata{
					Name:      "tmpname111", // TODO(mitchdraft) - implement a naming scheme
					Namespace: opts.Namespace,
				},
				Pod:         pod,
				Container:   container,
				ProcessName: opts.ProcessName,
				Image:       image,
				Debugger:    debuggertype,
			}
			fmt.Println("about to trigger")
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

	debugContainerCmd.Flags().StringVarP(&opts.Namespace, "namespace", "n", "default", "Namespace the pod belongs to")
	debugContainerCmd.Flags().StringVarP(&opts.ProcessName, "processName", "p", "", "Process name to debug (defaults to the first running process)")

	return debugContainerCmd
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
	fmt.Println("writing")
	return (*daClient).Write(&da, writeOpts)

}
