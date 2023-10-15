package logger

import (
	"context"
	"errors"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/flyerxp/lib/logger"
	hertzzap "github.com/hertz-contrib/logger/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"testing"
	"time"
)

func TestEncode(t *testing.T) {
	/*dynamicLevel := zap.NewAtomicLevel()
	dynamicLevel.SetLevel(zap.DebugLevel)*/
	l := hertzzap.NewLogger(
		hertzzap.WithCores([]hertzzap.CoreConfig{
			{
				Enc: zapcore.NewConsoleEncoder(logger.EncoderConfig()),
				Ws:  zapcore.AddSync(os.Stdout),
				Lvl: zap.NewAtomicLevelAt(zapcore.DebugLevel),
			},
			{
				Enc: zapcore.NewJSONEncoder(logger.EncoderConfig()),
				Ws:  getWriteSyncer("hertz.log"),
				Lvl: zap.NewAtomicLevelAt(zapcore.DebugLevel),
				/*Lvl: zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
					return lev == zap.DebugLevel
				}),*/
			},
		}...),
	)
	defer l.Sync()
	hlog.SetLogger(l)
	hlog.Notice("notice log")
	hlog.Notice("notice log2")
	hlog.Infof("hello %s", "hertz")
	hlog.Info("hertz")
	hlog.Warn("hertz")
	hlog.Error("error")
	hlog.Debugf("xxxxxxxxxxxxxxx")
}
func getWriteSyncer(file string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   file,
		MaxSize:    1000,
		MaxBackups: 5,
		MaxAge:     48,
		Compress:   true,
		LocalTime:  true,
	}
	return zapcore.AddSync(lumberJackLogger)
}
func Test2Encode(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.GetLogIdKey(), "test")
	logger.RegisterNewLog(ctx)
	logger.AddMongoTime(ctx, 1)
	logger.AddMysqlTime(ctx, 1)
	logger.AddPulsarTime(ctx, 1)
	logger.AddKafkaTime(ctx, 1)
	logger.AddRpcTime(ctx, 1)
	logger.AddEsTime(ctx, 1)
	logger.AddRocketTime(ctx, 1)
	logger.AddRocketConnTime(ctx, 1)
	logger.AddRedisConnTime(ctx, 1)
	logger.AddMongoConnTime(ctx, 1)
	logger.AddMysqlConnTime(ctx, 1)

	logger.AddPulsarConnTime(ctx, 1)
	logger.AddKafkaConnTime(ctx, 1)
	logger.AddRpcConnTime(ctx, 1)
	logger.AddEsConnTime(ctx, 1)
	logger.AddNotice(ctx, zap.Int("cccc", 1111))
	logger.AddRedisTime(ctx, 10)
	logger.SetExecTime(ctx, 12)
	logger.AddNotice(ctx, zap.String("cccc", "add add add"))
	logger.AddError(ctx, zap.Error(errors.New("error error error")))
	logger.AddError(ctx, zap.Error(errors.New("error error error")))
	logger.AddWarn(ctx, zap.Error(errors.New("warn warn warn")))
	logger.WriteErr(ctx)
	logger.WriteLine(ctx)
	time.Sleep(time.Second * 35)
	logger.AddNotice(ctx, zap.String("cccc", "add add add"))

	logger.WriteLine(ctx)
}
func TestSync(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.GetLogIdKey(), "test")
	//在
	//defer app.Shutdown(context.Background())
	logger.AddError(ctx, zap.Error(errors.New("aaaaaaaa")))
	logger.AddWarn(ctx, zap.Error(errors.New("bbbbb")))

	logger.AddNotice(ctx, zap.String("a", "bbbbbbbbbbbb"))
	logger.WriteLine(ctx)
	logger.WriteAccess(ctx, "aaaa")
	logger.WriteLine(ctx)
	//logger.WriteAccess(context.Background(), "xxx bbbb cccc")
	time.Sleep(time.Second)

	//app.Shutdown(context.Background())
	//logger.WriteErr()  //立即写入错误

}
