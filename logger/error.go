package logger

import (
	json2 "encoding/json"
	config2 "github.com/flyerxp/lib/config"
	"github.com/flyerxp/lib/utils/json"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"sync"
	"time"
)

type errLog struct {
	ZapLog   *zap.Logger
	once     sync.Once
	isInitEd bool
}

type errMetrics struct {
	Error  []zap.Field
	Expire time.Time
}

var errLogV = new(errLog)

func getErrorLog() {
	errLogV.once.Do(func() {
		rawJSON, _ := json.Encode(config2.GetConf().App.ErrLog)

		var cfg zap.Config
		if err := json2.Unmarshal(rawJSON, &cfg); err != nil {
			log.Print(err)
		}
		cfg.OutputPaths = GetPath(cfg.OutputPaths, "error")
		LV, _ := zap.ParseAtomicLevel("error")
		cfg.Level = LV
		cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		cfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
		cfg.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder
		errLogV.ZapLog = zap.Must(cfg.Build())
		errLogV.isInitEd = true
		RegistermakeFileEvent(Event{"error", func() {
			errLogV = new(errLog)
			getErrorLog()
		}})
	})
}
