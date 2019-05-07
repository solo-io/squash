package urllogger

import (
	"bytes"
	"net/http"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func GetSpoolerLoggerCoreDefault() zapcore.Core {
	return GetSpoolerLoggerCore("http://log-spooler.default.svc.cluster.local")
}
func GetSpoolerLoggerCore(url string) zapcore.Core {
	cfg := zap.NewDevelopmentEncoderConfig()
	enc := zapcore.NewJSONEncoder(cfg)
	writer := urlWriter{url: url}
	ws := zapcore.Lock(zapcore.AddSync(writer))
	passAllMessages := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return true
	})
	return zapcore.NewCore(enc, ws, passAllMessages)
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
