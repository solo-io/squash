package main

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/squash/pkg/urllogger"
	"go.uber.org/zap"

	"github.com/solo-io/squash/pkg/squash"

	"github.com/solo-io/squash/pkg/debuggers/remote"
	"github.com/solo-io/squash/pkg/platforms/kubernetes"
	"github.com/solo-io/squash/pkg/version"
)

func main() {
	//logger, _ := zap.NewProduction()
	logger := zap.New(urllogger.GetSpoolerLoggerCoreDefault())
	defer logger.Sync()

	logger.Sugar().Infof("squash started %v, %v", version.Version, version.TimeStamp)

	ctx := context.Background()
	ctx = contextutils.WithLogger(ctx, "squash")
	mustGetContainerProcessLocator(ctx)
	err := squash.RunSquash(ctx, remote.GetParticularDebugger)
	if err != nil {
		fmt.Println(err)
		logger.With(zap.Error(err)).Fatal("Error running debug bridge")

	}

}

// The debugging pod needs to be able to get a container process
// This function is a way to fail early (from the squash pod) if the running
// version of Kubernetes does not support the needed API.
func mustGetContainerProcessLocator(ctx context.Context) {
	_, err := kubernetes.NewContainerProcess()
	if err != nil {
		_, err := kubernetes.NewCRIContainerProcessAlphaV1()
		if err != nil {
			contextutils.LoggerFrom(ctx).With(zap.Error(err)).Fatal("Cannot get container process locator")
		}
	}
}
