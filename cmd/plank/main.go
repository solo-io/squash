package main

import (
	"context"
	"fmt"

	"github.com/solo-io/squash/pkg/urllogger"

	log "github.com/sirupsen/logrus"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/squash/pkg/plank"
	"github.com/solo-io/squash/pkg/version"
	"go.uber.org/zap"
)

func main() {
	// TODO - switch to zap throughout
	log.SetLevel(log.DebugLevel)
	log.Infof("plank %v, %v", version.Version, version.TimeStamp)
	//logger, _ := zap.NewProduction()
	logger := zap.New(urllogger.GetSpoolerLoggerCoreDefault())
	defer logger.Sync()
	contextutils.SetFallbackLogger(logger.Sugar())
	ctx := context.Background()
	ctx = contextutils.WithLogger(ctx, "squash")

	err := plank.Debug(ctx)
	if err != nil {
		fmt.Println(err)
		logger.With(zap.Error(err)).Fatal("debug failed!")

	}
}
