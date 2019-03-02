package main

import (
	"log"

	"github.com/solo-io/go-utils/clidoc"
	"github.com/solo-io/squash/pkg/squashctl"
	"github.com/solo-io/squash/pkg/version"
)

func main() {
	app, err := squashctl.App(version.Version)
	if err != nil {
		log.Fatal(err)
	}
	clidoc.MustGenerateCliDocs(app)
}
