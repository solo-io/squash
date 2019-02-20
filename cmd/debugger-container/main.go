package main

import (
	"context"
	"fmt"
	"time"

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
		fmt.Println(err)
		time.Sleep(1000 * time.Second)
		logger.With(zap.Error(err)).Fatal("debug failed!")

	}
}
