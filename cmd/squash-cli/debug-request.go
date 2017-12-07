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

	var debugServiceCmd = &cobra.Command{
		Use:   "debug-request image debugger",
		Short: "debug-request adds a debug request.",
		RunE: func(cmd *cobra.Command, args []string) error {
			image := ""
			debugger := ""
			switch len(args) {
			case 2:
				image, debugger = args[0], args[1]
			default:
				return errors.New("invalid number of arguments")
			}
			// parse
			dbgrequest := models.DebugRequest{
				Spec: &models.DebugRequestSpec{
					Debugger: &debugger,
					Image:    &image,
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

	RootCmd.AddCommand(debugServiceCmd)

}
