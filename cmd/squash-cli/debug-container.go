package main

import (
	"encoding/json"
	"errors"
	"os"

	"fmt"

	"github.com/solo-io/squash/pkg/client/debugattachment"
	"github.com/solo-io/squash/pkg/platforms/kubernetes"

	"github.com/solo-io/squash/pkg/models"
	"github.com/spf13/cobra"
)

func init() {

	namespace := "default"
	processName := ""

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

			c, err := getClient()
			if err != nil {
				return err
			}

			dbgattchment := models.DebugAttachment{
				Spec: &models.DebugAttachmentSpec{
					Attachment: &kubernetes.KubeAttachment{
						Namespace: namespace,
						Pod:       pod,
						Container: container,
					},
					ProcessName: processName,
					Image:       image,
					Debugger:    debuggertype,
				},
			}

			params := debugattachment.NewAddDebugAttachmentParams()
			params.Body = &dbgattchment
			res, err := c.Debugattachment.AddDebugAttachment(params)

			if err != nil {
				if !jsonoutput {
					fmt.Println("Failed adding container - check parameter names match container on the platform. error:", err)
				} else {
					json.NewEncoder(os.Stdout).Encode(Error{Type: "unknown", Info: err.Error()})
				}
				os.Exit(1)
			}

			dbgattchment = *res.Payload

			if !jsonoutput {
				fmt.Println("Debug config id:", dbgattchment.Metadata.Name)
			} else {
				json.NewEncoder(os.Stdout).Encode(dbgattchment)
			}

			return nil
		},
	}

	debugContainerCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace the pod belongs to")
	debugContainerCmd.Flags().StringVarP(&processName, "processName", "p", "", "Process name to debug (defaults to the first running process)")

	RootCmd.AddCommand(debugContainerCmd)

}
