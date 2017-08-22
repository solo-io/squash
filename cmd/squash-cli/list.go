package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/solo-io/squash/pkg/client/debugconfig"
	"github.com/solo-io/squash/pkg/models"

	"github.com/spf13/cobra"
)

func init() {

	var listCmd = &cobra.Command{
		Use:     "list [id]",
		Short:   "lists debug configs",
		Aliases: []string{"ps"},
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) > 1 {
				return errors.New("Too many args")
			}
			id := ""
			if len(args) == 1 {
				id = args[0]
			}

			c, err := getClient()
			if err != nil {
				return err
			}
			if id == "" {

				params := debugconfig.NewGetDebugConfigsParams()
				res, err := c.Debugconfig.GetDebugConfigs(params)

				if err != nil {
					panic(err)
				}

				dbgcfgs := res.Payload

				if !jsonoutput {
					printDebugCfgs(dbgcfgs)
				} else {
					json.NewEncoder(os.Stdout).Encode(dbgcfgs)
				}

			} else {

				params := debugconfig.NewGetDebugConfigParams()
				params.DebugConfigID = id
				res, err := c.Debugconfig.GetDebugConfig(params)

				if err != nil {
					panic(err)
				}

				dbgcfg := res.Payload

				if !jsonoutput {
					printDebugCfgs([]*models.DebugConfig{dbgcfg})
				} else {
					json.NewEncoder(os.Stdout).Encode(dbgcfg)
				}

			}
			return nil
		},
	}
	RootCmd.AddCommand(listCmd)
}

func printDebugCfgs(debugconfigs []*models.DebugConfig) {
	table := []string{"Active\tID\tAttachment.Name\tAttachment.Type\tDebugger\tImage\tImmediately\n"}
	for _, cfg := range debugconfigs {
		table = append(table, fmt.Sprintf("%v\t%s\t%s\t%s\t%s\t%s\t%v\n", cfg.Active, cfg.ID, *cfg.Attachment.Name, *cfg.Attachment.Type, cfg.Debugger, *cfg.Image, cfg.Immediately))
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	for _, r := range table {
		w.Write([]byte(r))
	}
	w.Flush()

}
