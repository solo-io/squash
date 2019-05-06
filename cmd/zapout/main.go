package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	cfg := zap.NewDevelopmentEncoderConfig()
	enc := zapcore.NewJSONEncoder(cfg)
	writer := urlWriter{url: "http://log-spooler.default.svc.cluster.local"}
	ws := zapcore.Lock(zapcore.AddSync(writer))
	passAllMessages := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return true
	})
	core := zapcore.NewCore(enc, ws, passAllMessages)
	logger := zap.New(core)
	for i := 0; i < 100; i++ {
		logger.Info(fmt.Sprintf("HEY %v", i))
		time.Sleep(1 * time.Second)
	}
}

type urlWriter struct {
	url string
}

func (u urlWriter) Write(p []byte) (n int, err error) {
	req, err := http.NewRequest("POST", u.url, bytes.NewBuffer(p))
	if err != nil {
		return 0, err
	}
	_, err = http.DefaultClient.Do(req)
	return len(p), err
}
