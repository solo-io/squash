package main

import (
	"fmt"
	"os"
	"time"

	check "github.com/solo-io/go-checkpoint"
	"github.com/solo-io/squash/pkg/cmd/cli"
	"github.com/solo-io/squash/pkg/version"
)

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

}

func toptr(s string) *string {
	return &s
}
