package main

import (
	"encoding/json"
	"errors"
	"os"

	"fmt"

	"github.com/solo-io/squash/pkg/client/debugconfig"

	"github.com/solo-io/squash/pkg/models"
	"github.com/spf13/cobra"
)

func init() {
	var breakpoints []string

	var debugContainerCmd = &cobra.Command{
		Use:   "debug-container image pod container [type]",
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

			dbgconfig := models.DebugConfig{
				Attachment: &models.Attachment{
					Type: toptr(models.AttachmentTypeContainer),
					Name: toptr(fmt.Sprintf("%s:%s", pod, container)),
				},
				Image:       &image,
				Immediately: true,
				Debugger:    debuggertype,
			}
			immediately := true
			for _, b := range breakpoints {
				bp := b
				dbgconfig.Breakpoints = append(dbgconfig.Breakpoints, &models.Breakpoint{
					Location: &bp,
				})
				immediately = false
			}
			dbgconfig.Immediately = immediately

			params := debugconfig.NewAddDebugConfigParams()
			params.Body = &dbgconfig
			res, err := c.Debugconfig.AddDebugConfig(params)

			if err != nil {
				if !jsonoutput {
					fmt.Println("Failed adding container - check parameter names match container on the platform. error:", err)
				} else {
					json.NewEncoder(os.Stdout).Encode(Error{Type: "unkown", Info: err.Error()})
				}
				os.Exit(1)
			}

			dbgconfig = *res.Payload

			if !jsonoutput {
				fmt.Println("Debug config id:", dbgconfig.ID)
			} else {
				json.NewEncoder(os.Stdout).Encode(dbgconfig)
			}

			return nil
		},
	}
	debugContainerCmd.Flags().StringSliceVarP(&breakpoints, "breakpoint", "b", nil, "Breakpoint to place e.g. 'main.go:34'")

	RootCmd.AddCommand(debugContainerCmd)

}
