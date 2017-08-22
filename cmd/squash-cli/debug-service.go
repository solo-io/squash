package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/solo-io/squash/pkg/client/debugconfig"

	"github.com/solo-io/squash/pkg/models"
	"github.com/spf13/cobra"
)

func init() {

	var inputFile string
	var breakpoints []string

	var debugServiceCmd = &cobra.Command{
		Use:   "debug-service name image debugger OR add -f debugconfig.json ",
		Short: "debug-service adds a service type debug config",
		RunE: func(cmd *cobra.Command, args []string) error {
			var input io.ReadCloser
			var err error
			name := ""
			image := ""
			debugger := "gdb"
			switch len(args) {
			case 0:
				if inputFile == "-" {
					input, err = os.Stdin, nil
				} else if inputFile != "" {
					input, err = os.Open(inputFile)
				}
			case 3:
				debugger = args[2]
				fallthrough
			case 2:
				name, image = args[0], args[1]
			default:
				err = errors.New("invalid number of arguments")
			}
			if err != nil {
				return err
			}
			// parse
			dbgcfg := models.DebugConfig{}
			if input != nil {
				defer input.Close()
				json.NewDecoder(input).Decode(&dbgcfg)

			} else if name != "" {
				tmp := models.AttachmentTypeService
				dbgcfg.Attachment = &models.Attachment{
					Name: &name,
					Type: &tmp,
				}
				dbgcfg.Image = &image
				dbgcfg.Debugger = debugger
			} else {
				return errors.New("Please provide name and image, or input file!")
			}

			for _, b := range breakpoints {
				bp := b
				dbgcfg.Breakpoints = append(dbgcfg.Breakpoints, &models.Breakpoint{
					Location: &bp,
				})
			}

			c, err := getClient()
			if err != nil {
				return err
			}
			params := debugconfig.NewAddDebugConfigParams()
			params.Body = &dbgcfg
			res, err := c.Debugconfig.AddDebugConfig(params)

			if err != nil {
				if !jsonoutput {
					fmt.Println("Failed adding service - check parameter names match service name on the platform")
				} else {
					json.NewEncoder(os.Stdout).Encode(Error{Type: "unkown", Info: err.Error()})
				}
			}

			dbgcfg = *res.Payload

			if !jsonoutput {
				fmt.Println("Debug config id:", dbgcfg.ID)
			} else {
				json.NewEncoder(os.Stdout).Encode(dbgcfg)
			}

			return nil
		},
	}

	debugServiceCmd.Flags().StringVarP(&inputFile, "file", "f", "", "app file")
	debugServiceCmd.Flags().StringSliceVarP(&breakpoints, "breakpoint", "b", nil, "Breakpoint to place e.g. 'main.go:34'")
	RootCmd.AddCommand(debugServiceCmd)

}
