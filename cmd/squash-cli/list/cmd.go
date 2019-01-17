package list

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/solo-io/squash/pkg/client"
	"github.com/solo-io/squash/pkg/client/debugattachment"
	"github.com/solo-io/squash/pkg/client/debugrequest"
	"github.com/solo-io/squash/pkg/models"

	"github.com/spf13/cobra"
)

func init() {

	var listCmd = &cobra.Command{
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

			c, err := getClient()
			if err != nil {
				return err
			}
			switch objType {

			case "debugattachments", "attachments", "a":
				listattachments(c, id)
			case "debugrequests", "requests", "r":
				listrequests(c, id)
			default:
				return errors.New("invalid type provided")
			}
			return nil
		},
	}
	RootCmd.AddCommand(listCmd)
}

func listattachments(c *client.Squash, name string) {

	if name == "" {

		params := debugattachment.NewGetDebugAttachmentsParams()

		// be explicit
		f := false
		params.Wait = &f

		res, err := c.Debugattachment.GetDebugAttachments(params)

		if err != nil {
			panic(err)
		}

		dbgattachments := res.Payload

		if !jsonoutput {
			printDebugAttachments(dbgattachments)
		} else {
			json.NewEncoder(os.Stdout).Encode(dbgattachments)
		}

	} else {

		params := debugattachment.NewGetDebugAttachmentParams()
		params.DebugAttachmentID = name
		res, err := c.Debugattachment.GetDebugAttachment(params)

		if err != nil {
			panic(err)
		}

		dbgattachment := res.Payload

		if !jsonoutput {
			printDebugAttachments([]*models.DebugAttachment{dbgattachment})
		} else {
			json.NewEncoder(os.Stdout).Encode(dbgattachment)
		}

	}
}
func listrequests(c *client.Squash, name string) {

	if name == "" {

		params := debugrequest.NewGetDebugRequestsParams()
		res, err := c.Debugrequest.GetDebugRequests(params)

		if err != nil {
			panic(err)
		}

		dbgrequests := res.Payload

		if !jsonoutput {
			printDebugRequests(dbgrequests)
		} else {
			json.NewEncoder(os.Stdout).Encode(dbgrequests)
		}

	} else {

		params := debugrequest.NewGetDebugRequestParams()
		params.DebugRequestID = name
		res, err := c.Debugrequest.GetDebugRequest(params)

		if err != nil {
			panic(err)
		}

		dbgrequest := res.Payload

		if !jsonoutput {
			printDebugRequests([]*models.DebugRequest{dbgrequest})
		} else {
			json.NewEncoder(os.Stdout).Encode(dbgrequest)
		}

	}
}

func printDebugAttachments(debugconfigs []*models.DebugAttachment) {
	table := []string{"State\tID\tDebugger\tImage\tDebugger Address\n"}
	for _, atch := range debugconfigs {
		table = append(table, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\n", atch.Status.State, atch.Metadata.Name, atch.Spec.Debugger, atch.Spec.Image, atch.Status.DebugServerAddress))
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	for _, r := range table {
		w.Write([]byte(r))
	}
	w.Flush()

}

func printDebugRequests(debugconfigs []*models.DebugRequest) {
	table := []string{"ID\tDebugger\tImage\tBound Attachment name\n"}
	for _, rqst := range debugconfigs {
		table = append(table, fmt.Sprintf("%s\t%s\t%s\t%s\n", rqst.Metadata.Name, nilToEmpty(rqst.Spec.Debugger), nilToEmpty(rqst.Spec.Image), rqst.Status.DebugAttachmentRef))
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	for _, r := range table {
		w.Write([]byte(r))
	}
	w.Flush()

}
func nilToEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
