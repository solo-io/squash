package main

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/squash/pkg/kube"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	contextutils.SetFallbackLogger(logger.Sugar())
	ctx := context.Background()
	ctx = contextutils.WithLogger(ctx, "squash")

	err := kube.Debug(ctx)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("debug failed!")

	}
}
