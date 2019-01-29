package main

import (
	"fmt"
	"os"
	"time"

	check "github.com/solo-io/go-checkpoint"
	"github.com/solo-io/squash/pkg/cmd/cli"
	"github.com/solo-io/squash/pkg/version"
)

func main() {
	start := time.Now()
	defer check.CallReport("squashctl", version.Version, start)

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
