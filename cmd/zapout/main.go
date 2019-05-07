package main

import (
	"fmt"
	"time"

	"github.com/solo-io/squash/pkg/urllogger"

	"go.uber.org/zap"
)

func main() {
	core := urllogger.GetSpoolerLoggerCore()
	logger := zap.New(core)
	for i := 0; i < 100; i++ {
		logger.Info(fmt.Sprintf("HEY %v", i))
		time.Sleep(1 * time.Second)
	}
}
