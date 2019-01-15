package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	soloutils "github.com/solo-io/go-utils/v1/common"
	"github.com/spf13/cobra"
)

func DebugRequestCmd(o *Options) *cobra.Command {
	processName := ""
	drOpts := &o.DebugRequest
	debugServiceCmd := &cobra.Command{
		Use:   "debug-request image debugger",
		Short: "debug-request adds a debug request.",
		RunE: func(cmd *cobra.Command, args []string) error {

			if err := ensureDebugRequestOpts(drOpts, args); err != nil {
				return err
			}
			da := debugAttachmentFromOpts(*drOpts)
			dbgattchment, err := o.debugContainer(da)
			if err != nil {
				return err
			}

			if err != nil {
				if !o.Json {
					fmt.Println("Failed adding service - check parameter names match service name on the platform")
				} else {
					json.NewEncoder(os.Stdout).Encode(Error{Type: "unkown", Info: err.Error()})
				}
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

	debugServiceCmd.Flags().StringVarP(&processName, "processName", "p", "", "Process name to debug (defaults to the first running process)")

	return debugServiceCmd

}

func ensureDebugRequestOpts(drOpts *DebugContainer, args []string) error {
	image := ""
	debugger := ""

	switch len(args) {
	case 2:
		image, debugger = args[0], args[1]
	default:
		return errors.New("invalid number of arguments")
	}

	drOpts.Image = image
	drOpts.DebuggerType = debugger
	drOpts.Name = soloutils.RandKubeNameBytes(6)

	return nil
}
