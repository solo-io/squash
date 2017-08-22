package main

import (
	"log"
	"net/url"
	"os"

	"github.com/solo-io/squash/pkg/client"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "squash",
	Short: "squash",
}

var serverurl string
var jsonoutput bool

type Error struct {
	Type string
	Info string
}

func getClient() (*client.Squash, error) {

	url, err := url.Parse(serverurl)
	if err != nil {
		return nil, err
	}

	cfg := &client.TransportConfig{
		BasePath: url.Path,
		Host:     url.Host,
		Schemes:  []string{url.Scheme},
	}
	client := client.NewHTTPClientWithConfig(nil, cfg)

	return client, nil
}

func main() {

	RootCmd.PersistentFlags().StringVar(&serverurl, "url", os.Getenv("SQUASH_SERVER_URL"), "url for app server. probably a kubernetes service url. Default is the env variable SQUASH_SERVER_URL")
	RootCmd.PersistentFlags().BoolVar(&jsonoutput, "json", false, "output json format")
	if err := RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func toptr(s string) *string {
	return &s
}
