package main

import (
	"context"
	"fmt"

	"github.com/solo-io/squash/pkg/version"

	"github.com/solo-io/squash/pkg/urllogger"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/squash/pkg/plank"
	"go.uber.org/zap"
)

func main() {
	//logger, _ := zap.NewProduction()
	logger := zap.New(urllogger.GetSpoolerLoggerCoreDefault())
	defer logger.Sync()
	contextutils.SetFallbackLogger(logger.Sugar())
	ctx := context.Background()
	ctx = contextutils.WithLogger(ctx, "squash-plank")
	logger.Sugar().Infof("plank %v, %v", version.Version, version.TimeStamp)

	err := plank.Debug(ctx)
	if err != nil {
		fmt.Println(err)
		logger.With(zap.Error(err)).Fatal("debug failed!")

	}
}
