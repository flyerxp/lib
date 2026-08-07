package pulsarL

import (
	"sync"

	"github.com/apache/pulsar-client-go/pulsar/log"
	config2 "github.com/flyerxp/lib/v2/config"
	"github.com/flyerxp/lib/v2/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"strings"
)

// L 对外导出的通用日志实例，调用方式与原 logrus 完全兼容
var (
	L    *zap.SugaredLogger
	once sync.Once
)

// pulsarZapLogger 适配器：同时实现 pulsar 的 log.Logger 和 log.Entry 接口
type pulsarZapLogger struct {
	sugar *zap.SugaredLogger
}

// 基础日志方法
func (l *pulsarZapLogger) Debug(args ...interface{})                 { l.sugar.Debug(args...) }
func (l *pulsarZapLogger) Info(args ...interface{})                  { l.sugar.Info(args...) }
func (l *pulsarZapLogger) Warn(args ...interface{})                  { l.sugar.Warn(args...) }
func (l *pulsarZapLogger) Error(args ...interface{})                 { l.sugar.Error(args...) }
func (l *pulsarZapLogger) Debugf(format string, args ...interface{}) { l.sugar.Debugf(format, args...) }
func (l *pulsarZapLogger) Infof(format string, args ...interface{})  { l.sugar.Infof(format, args...) }
func (l *pulsarZapLogger) Warnf(format string, args ...interface{})  { l.sugar.Warnf(format, args...) }
func (l *pulsarZapLogger) Errorf(format string, args ...interface{}) { l.sugar.Errorf(format, args...) }

// 字段扩展方法（完整实现 Entry 接口）
func (l *pulsarZapLogger) WithField(name string, value interface{}) log.Entry {
	return &pulsarZapLogger{sugar: l.sugar.With(name, value)}
}

func (l *pulsarZapLogger) WithFields(fields log.Fields) log.Entry {
	kv := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		kv = append(kv, k, v)
	}
	return &pulsarZapLogger{sugar: l.sugar.With(kv...)}
}

// 本次补全的缺失方法：WithError
func (l *pulsarZapLogger) WithError(err error) log.Entry {
	return &pulsarZapLogger{sugar: l.sugar.With(zap.Error(err))}
}

// 子日志方法（实现 Logger 接口独有方法）
func (l *pulsarZapLogger) SubLogger(fields log.Fields) log.Logger {
	kv := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		kv = append(kv, k, v)
	}
	return &pulsarZapLogger{sugar: l.sugar.With(kv...)}
}

// GetLogger 返回可直接赋值给 pulsar.ClientOptions.Logger 的实例
func GetLogger() log.Logger {
	return &pulsarZapLogger{sugar: getLog()}
}

func getLog() *zap.SugaredLogger {
	once.Do(func() {
		paths := logger.GetPath(config2.GetConf().App.Logger.OutputPaths, "pulsar")
		logPath := ""
		for _, v := range paths {
			if strings.Contains(v, "/") {
				logPath = filepath.Dir(v) + "/pulsar.log"
			}
		}

		var writeSyncer zapcore.WriteSyncer
		if logPath == "" {
			writeSyncer = zapcore.AddSync(os.Stdout)
		} else {
			writeSyncer = zapcore.AddSync(&lumberjack.Logger{
				Filename:   logPath,
				MaxSize:    1024000,
				MaxBackups: 5,
				MaxAge:     48,
				Compress:   true,
				LocalTime:  true,
			})
		}

		// JSON 字段与原 logrus 默认输出完全对齐，日志采集无感知
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "time"
		encoderConfig.LevelKey = "level"
		encoderConfig.MessageKey = "msg"
		encoder := zapcore.NewJSONEncoder(encoderConfig)

		// 仅输出 Warn 及以上级别，与原 logrus.WarnLevel 完全一致
		core := zapcore.NewCore(encoder, writeSyncer, zap.WarnLevel)
		L = zap.New(core).Sugar()
	})
	return L
}
