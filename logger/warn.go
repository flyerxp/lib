package logger

import (
	json2 "encoding/json"
	config2 "github.com/flyerxp/lib/v2/config"
	"github.com/flyerxp/lib/v2/utils/json"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"sync"
	"time"
)

type warnLog struct {
	ZapLog   *zap.Logger
	once     sync.Once
	isInitEd bool
}
type warMetrics struct {
	Warn          []zap.Field
	TotalExecTime int
	Expire        time.Time
}

var warnLogV = new(warnLog)

func getWarnLog() {
	warnLogV.once.Do(func() {
		rawJSON, _ := json.Encode(config2.GetConf().App.ErrLog)
		var cfg zap.Config
		if err := json2.Unmarshal(rawJSON, &cfg); err != nil {
			log.Print(err)
		}
		cfg.OutputPaths = GetPath(cfg.OutputPaths, "warn")
		cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		cfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
		cfg.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder
		warnLogV.ZapLog = zap.Must(cfg.Build())
		warnLogV.isInitEd = true
		RegistermakeFileEvent(Event{"error", func() {
			warnLogV = new(warnLog)
			getWarnLog()
		}})
	})
}
