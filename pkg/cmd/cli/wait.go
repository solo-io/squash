package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/models"
	"github.com/solo-io/squash/pkg/utils"
	"github.com/spf13/cobra"
)

// WaitCmd
// TODO - make this more useful, read until you find one in the Attached state
func WaitCmd(namespace string, name string, timeout float64) (v1.DebugAttachment, error) {
	ctx := context.Background()
	daClient, err := utils.GetDebugAttachmentClient(ctx)
	if err != nil {
		return v1.DebugAttachment{}, err
	}
	reOptions := clients.ReadOpts{
		Ctx: ctx,
	}
	da, err := (*daClient).Read(namespace, name, reOptions)
	if err != nil {
		// TODO(mitchdraft) implement a periodic check instead of waiting the full timeout duration
		time.Sleep(time.Duration(int32(timeout)) * time.Second)
		o := Options{
			daClient: daClient,
			ctx:      ctx,
		}
		da, err = o.getNamedDebugAttachment(name)
		if err != nil {
			return v1.DebugAttachment{}, err
		}
	}
	return *da, nil

}

func WaitAttCmd(o *Options) *cobra.Command {
	// attachmentWaitTimeout := 1.0
	waitForAttachmentCmd := &cobra.Command{
		Use:   "wait dbgattachmentid",
		Short: "wait for a debug config to have a debug server url appear",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 1 {
				return fmt.Errorf("Invalid number of arguments: %v", args)
			}

			id := args[0]

			timedOut := false
			da, err := o.getNamedDebugAttachment(id)
			if err != nil || da.State != v1.DebugAttachment_Attached {
				// TODO(mitchdraft) implement a periodic check instead of waiting the full timeout duration
				const overrideTimeoutTODO = 3 // TODO(mitchdraft) - unhardcode
				// time.Sleep(time.Duration(int32(o.Wait.Timeout)) * time.Second)
				time.Sleep(overrideTimeoutTODO * time.Second)
				da, err = o.getNamedDebugAttachment(id)
				if err != nil {
					o.exitOnError(timedOut, err)
				}
			}

			if !o.Json {
				fmt.Println("Debug session started! debug server is at:", da.DebugServerAddress)
			} else {
				// TODO - update the ide plugins to use the new format then remove this
				modelFormat := models.ConvertDebugAttachment(da)
				json.NewEncoder(os.Stdout).Encode(modelFormat)
			}

			return nil
		},
	}
	waitForAttachmentCmd.Flags().Float64VarP(&(o.Wait).Timeout, "timeout", "t", 1.0, "wait timeout")
	return waitForAttachmentCmd
}

// TODO(mitchdraft) - standardize error handling esp. wrt json output
func (o *Options) exitOnError(timedOut bool, err error) {

	if !o.Json {
		fmt.Println("Error:", err)
	} else {
		errjson := Error{
			Type: "Unknown",
			Info: err.Error(),
		}
		if timedOut {
			errjson.Type = "Timeout"
		}
		json.NewEncoder(os.Stderr).Encode(errjson)
	}
	os.Exit(-1)
}
