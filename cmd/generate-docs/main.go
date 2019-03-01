package main

import (
	"log"

	TODOclidoc "github.com/solo-io/squash/cmd/internal/clidoc" // will move to go-utils
	"github.com/solo-io/squash/pkg/squashctl"
	"github.com/solo-io/squash/pkg/version"
)

func main() {
	app, err := squashctl.App(version.Version)
	if err != nil {
		log.Fatal(err)
	}
	TODOclidoc.MustGenerateCliDocs(app)
}
