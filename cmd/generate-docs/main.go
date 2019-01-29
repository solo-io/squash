package main

import (
	"log"

	"github.com/solo-io/squash/pkg/squashctl"
	"github.com/solo-io/squash/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func main() {
	app, err := squashctl.App(version.Version)
	if err != nil {
		log.Fatal(err)
	}

	disableAutoGenTag(app)

	err = doc.GenMarkdownTree(app, "./docs/cli")
	if err != nil {
		log.Fatal(err)
	}
}

func disableAutoGenTag(c *cobra.Command) {
	c.DisableAutoGenTag = true
	for _, c := range c.Commands() {
		disableAutoGenTag(c)
	}
}
