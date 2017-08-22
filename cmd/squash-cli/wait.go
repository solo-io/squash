package main

import (
	"encoding/json"
	"os"

	"fmt"

	"github.com/solo-io/squash/pkg/client/debugsessions"

	"github.com/spf13/cobra"
)

func init() {

	sessionWaitTimeout := 1.0
	var waitForSessionCmd = &cobra.Command{
		Use:   "wait dbgconfigid",
		Short: "wait for a debug session to appear for a debug config",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := getClient()
			if err != nil {
				return err
			}

			if len(args) != 1 {
				return fmt.Errorf("Invalid number of arguments: %v", args)
			}

			id := args[0]

			params := debugsessions.NewPopDebugSessionParams()
			params.DebugConfigID = id
			params.XTimeout = &sessionWaitTimeout

			res, err := c.Debugsessions.PopDebugSession(params)

			if err != nil {

				if !jsonoutput {
					fmt.Println("Error:", err)
				} else {
					errjson := Error{
						Type: "Unknown",
						Info: err.Error(),
					}
					if _, ok := err.(*debugsessions.PopDebugSessionRequestTimeout); ok {
						errjson.Type = "Timeout"
					}
					json.NewEncoder(os.Stderr).Encode(errjson)
				}
				os.Exit(-1)
			}

			session := *res.Payload

			if !jsonoutput {
				fmt.Println("Debug session started! debug server is at:", session.URL)
			} else {
				json.NewEncoder(os.Stdout).Encode(session)
			}

			return nil
		},
	}
	waitForSessionCmd.Flags().Float64VarP(&sessionWaitTimeout, "timeout", "t", 1.0, "wait timeout")

	RootCmd.AddCommand(waitForSessionCmd)

}
