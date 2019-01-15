package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/models"
	"github.com/solo-io/squash/pkg/options"

	"github.com/spf13/cobra"
)

func ListCmd(o *Options) *cobra.Command {
	listCmd := &cobra.Command{
		Use:     "list what [id]",
		Short:   "lists debug requests or attachments",
		Aliases: []string{"ps"},
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) > 2 {
				return errors.New("too many args")
			}
			if len(args) == 0 {
				return errors.New("no type provided")
			}
			objType := args[0]
			id := ""
			if len(args) == 2 {
				id = args[1]
			}

			switch objType {

			case "debugattachments", "attachments", "a":
				return o.listattachments(id)
			case "debugrequests", "requests", "r":
				// TODO(mitchdraft) - implement debugrequests
				return o.listattachments(id)
				// return listrequests(daClient, id)
			default:
				return errors.New("invalid type provided")
			}
			return nil
		},
	}
	return listCmd
}

func (o *Options) listattachments(name string) error {

	if name == "" {

		das, err := (*o.daClient).List(options.SquashClientNamespace, clients.ListOpts{Ctx: o.ctx})
		if err != nil {
			return err
		}

		if o.Json {
			printDebugAttachments(das)
		} else {
			// TODO - update the ide plugins to use the new format then remove this
			modelFormat := models.ConvertDebugAttachments(das)
			json.NewEncoder(os.Stdout).Encode(modelFormat)
		}

	} else {

		da, err := (*o.daClient).Read(options.SquashClientNamespace, name, clients.ReadOpts{Ctx: o.ctx})
		if err != nil {
			return err
		}

		if !o.Json {
			// TODO(mitchdraft) need to tmp convert this to the old format
			printDebugAttachments([]*v1.DebugAttachment{da})
		} else {
			json.NewEncoder(os.Stdout).Encode(da)
		}

	}
	return nil
}

// func listrequests(c *client.Squash, name string) error {
// 	return fmt.Errorf("TODO")

// 	if name == "" {

// 		params := debugrequest.NewGetDebugRequestsParams()
// 		res, err := c.Debugrequest.GetDebugRequests(params)

// 		if err != nil {
// 			panic(err)
// 		}

// 		dbgrequests := res.Payload

// 		if !jsonoutput {
// 			printDebugRequests(dbgrequests)
// 		} else {
// 			json.NewEncoder(os.Stdout).Encode(dbgrequests)
// 		}

// 	} else {

// 		params := debugrequest.NewGetDebugRequestParams()
// 		params.DebugRequestID = name
// 		res, err := c.Debugrequest.GetDebugRequest(params)

// 		if err != nil {
// 			panic(err)
// 		}

// 		dbgrequest := res.Payload

// 		if !jsonoutput {
// 			printDebugRequests([]*models.DebugRequest{dbgrequest})
// 		} else {
// 			json.NewEncoder(os.Stdout).Encode(dbgrequest)
// 		}

// 	}
// }

func printDebugAttachments(das []*v1.DebugAttachment) {
	table := []string{"State\tID\tDebugger\tImage\tDebugger Address\n"}
	for _, da := range das {
		table = append(table, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\n", da.State, da.Metadata.Name, da.Debugger, da.Image, da.DebugServerAddress))
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	for _, r := range table {
		w.Write([]byte(r))
	}
	w.Flush()

}

// func printDebugRequests(debugconfigs []*models.DebugRequest) {
// 	table := []string{"ID\tDebugger\tImage\tBound Attachment name\n"}
// 	for _, rqst := range debugconfigs {
// 		table = append(table, fmt.Sprintf("%s\t%s\t%s\t%s\n", rqst.Metadata.Name, nilToEmpty(rqst.Spec.Debugger), nilToEmpty(rqst.Spec.Image), rqst.Status.DebugAttachmentRef))
// 	}

// 	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
// 	for _, r := range table {
// 		w.Write([]byte(r))
// 	}
// 	w.Flush()

// }
func nilToEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
