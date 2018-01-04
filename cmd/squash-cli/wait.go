package main

import (
	"encoding/json"
	"os"

	"fmt"

	"github.com/solo-io/squash/pkg/client/debugattachment"
	"github.com/solo-io/squash/pkg/models"
	"github.com/spf13/cobra"
)

func init() {

	attachmentWaitTimeout := 1.0
	var waitForAttachmentCmd = &cobra.Command{
		Use:   "wait dbgattachmentid",
		Short: "wait for a debug config to have a debug server url appear",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := getClient()
			if err != nil {
				return err
			}

			if len(args) != 1 {
				return fmt.Errorf("Invalid number of arguments: %v", args)
			}

			id := args[0]

			params := debugattachment.NewGetDebugAttachmentsParams()
			params.States = []string{models.DebugAttachmentStatusStateAttached, models.DebugAttachmentStatusStateError}
			params.Names = []string{id}
			t := true
			params.Wait = &t
			params.XTimeout = &attachmentWaitTimeout

			res, err := c.Debugattachment.GetDebugAttachments(params)

			if err != nil {

				if !jsonoutput {
					fmt.Println("Error:", err)
				} else {
					errjson := Error{
						Type: "Unknown",
						Info: err.Error(),
					}
					if _, ok := err.(*debugattachment.GetDebugAttachmentsRequestTimeout); ok {
						errjson.Type = "Timeout"
					}
					json.NewEncoder(os.Stderr).Encode(errjson)
				}
				os.Exit(-1)
			}

			attachments := res.Payload
			if len(attachments) != 1 {
				panic(fmt.Sprintf("error getting attachments - successfull call ambiguous %v", attachments))
			}
			attachment := attachments[0]

			if !jsonoutput {
				fmt.Println("Debug session started! debug server is at:", attachment.Status.DebugServerAddress)
			} else {
				json.NewEncoder(os.Stdout).Encode(attachment)
			}

			return nil
		},
	}
	waitForAttachmentCmd.Flags().Float64VarP(&attachmentWaitTimeout, "timeout", "t", 1.0, "wait timeout")

	RootCmd.AddCommand(waitForAttachmentCmd)

}
