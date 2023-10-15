package pulsarL

import (
	config2 "github.com/flyerxp/lib/config"
	"github.com/flyerxp/lib/logger"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var L *logrus.Logger
var once sync.Once

func init() {
	go getLog()
}
func getLog() *logrus.Logger {
	once.Do(func() {
		paths := logger.GetPath(config2.GetConf().App.Logger.OutputPaths, "pulsar")
		logPath := ""
		for _, v := range paths {
			if strings.Contains(v, "/") {
				logPath = filepath.Dir(v) + "/pulsar.log"
			}
		}
		pLog := &logrus.Logger{
			Formatter:    new(logrus.JSONFormatter),
			Hooks:        make(logrus.LevelHooks),
			Level:        logrus.WarnLevel,
			ExitFunc:     os.Exit,
			ReportCaller: false,
		}

		if logPath == "" {
			pLog.Out = os.Stdout
		} else {
			pLog.Out = &lumberjack.Logger{
				Filename:   logPath,
				MaxSize:    1024000,
				MaxBackups: 5,
				MaxAge:     48,
				Compress:   true,
				LocalTime:  true,
			}
		}
		L = pLog
	})
	return L
}
