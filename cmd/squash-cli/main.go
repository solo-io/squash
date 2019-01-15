package main

import (
	"fmt"
	"os"
	"time"

	check "github.com/solo-io/go-checkpoint"
	"github.com/solo-io/squash/pkg/cmd/cli"
	"github.com/solo-io/squash/pkg/version"
)

// var RootCmd = &cobra.Command{
// 	Use:   "squash",
// 	Short: "squash",
// }

var serverurl string
var jsonoutput bool

type Error struct {
	Type string
	Info string
}

func main() {
	start := time.Now()
	defer check.CallReport("squash", version.Version, start)

	app, err := cli.App(version.Version)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := app.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// RootCmd.PersistentFlags().StringVar(&serverurl, "url", os.Getenv("SQUASH_SERVER_URL"), "url for app server. probably a kubernetes service url. Default is the env variable SQUASH_SERVER_URL")
	// RootCmd.PersistentFlags().BoolVar(&jsonoutput, "json", false, "output json format")
	// if err := RootCmd.Execute(); err != nil {
	// 	log.Fatal(err)
	// }
}

func toptr(s string) *string {
	return &s
}
