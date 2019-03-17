package main

import (
	"context"
	"os"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/squash/ci/internal/extconfig"
)

func main() {

	ctx := context.TODO()
	if len(os.Args) != 2 {
		contextutils.LoggerFrom(ctx).Fatal("Must pass a single argument ( version )")
	}
	version := os.Args[1]

	extconfig.MustCreateDevResources(ctx, version)
}
