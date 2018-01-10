package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/solo-io/squash/pkg/client/debugrequest"
	"github.com/solo-io/squash/pkg/models"
	"github.com/spf13/cobra"
)

func init() {

	processName := ""
	var debugServiceCmd = &cobra.Command{
		Use:   "debug-request image debugger [processName]",
		Short: "debug-request adds a debug request.",
		RunE: func(cmd *cobra.Command, args []string) error {
			image := ""
			debugger := ""

			switch len(args) {
			case 3:
				processName = args[2]
				fallthrough
			case 2:
				image, debugger = args[0], args[1]
			default:
				return errors.New("invalid number of arguments")
			}
			// parse
			dbgrequest := models.DebugRequest{
				Spec: &models.DebugRequestSpec{
					Debugger:    &debugger,
					ProcessName: processName,
					Image:       &image,
				},
			}

			c, err := getClient()
			if err != nil {
				return err
			}
			params := debugrequest.NewCreateDebugRequestParams()
			params.Body = &dbgrequest
			res, err := c.Debugrequest.CreateDebugRequest(params)

			if err != nil {
				if !jsonoutput {
					fmt.Println("Failed adding service - check parameter names match service name on the platform")
				} else {
					json.NewEncoder(os.Stdout).Encode(Error{Type: "unkown", Info: err.Error()})
				}
			}

			dbgrequest = *res.Payload

			if !jsonoutput {
				fmt.Println("Debug config id:", dbgrequest.Metadata.Name)
			} else {
				json.NewEncoder(os.Stdout).Encode(dbgrequest)
			}

			return nil
		},
	}

	debugServiceCmd.Flags().StringVarP(&processName, "processName", "p", "", "Process name to debug (defaults to the first running process)")

	RootCmd.AddCommand(debugServiceCmd)

}
