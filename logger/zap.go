package logger

import (
	"context"
	config2 "github.com/flyerxp/lib/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zap log 日志通用格式

func EncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "name",
		TimeKey:        "ts",
		CallerKey:      "caller",
		FunctionKey:    "func",
		StacktraceKey:  "stacktrace",
		LineEnding:     "\n",
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006/01/02 15:04:05.000Z0700"),
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

var AccessLogV *zap.Logger

func AccessLog() {
	cfg := zap.Config{
		Encoding:    "console",
		Level:       zap.NewAtomicLevelAt(zapcore.InfoLevel),
		OutputPaths: GetPath(config2.GetConf().App.Logger.OutputPaths, "access"),
	}
	cfg.EncoderConfig = zapcore.EncoderConfig{MessageKey: "msg"}
	//cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	//cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	//cfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	//cfg.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	AccessLogV = zap.Must(cfg.Build())
	RegistermakeFileEvent(Event{"error", func() {
		AccessLog()
	}})
}
func WriteAccess(ctx context.Context, format string, v ...interface{}) {
	if AccessLogV == nil {
		AccessLog()
	}
	_ = ctx
	_ = v
	AccessLogV.Info(format)
}
