package main

import (
	"errors"

	"fmt"

	"github.com/solo-io/squash/pkg/client/debugconfig"

	"github.com/spf13/cobra"
)

func init() {

	var rmCmd = &cobra.Command{
		Use:     "delete [id]",
		Short:   "delete a debug config",
		Aliases: []string{"rm"},
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) == 0 {
				return errors.New("Debug config id required")
			}

			c, err := getClient()
			if err != nil {
				return err
			}

			for _, id := range args {

				params := debugconfig.NewDeleteDebugConfigParams()
				params.DebugConfigID = id
				_, err := c.Debugconfig.DeleteDebugConfig(params)

				if !jsonoutput && (err != nil) {
					fmt.Print("Error deleting id: ", id, err)
				}

			}

			if jsonoutput {
				fmt.Println("{}")
			}
			return nil
		},
	}
	RootCmd.AddCommand(rmCmd)

}
