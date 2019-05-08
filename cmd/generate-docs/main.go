package main

import (
	"context"

	"github.com/solo-io/go-utils/clidoc"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/squash/pkg/squashctl"
	"github.com/solo-io/squash/pkg/version"
)

func main() {
	app, err := squashctl.App(version.Version)
	if err != nil {
		contextutils.LoggerFrom(context.TODO()).Fatal(err)
	}
	clidoc.MustGenerateCliDocs(app)
}
