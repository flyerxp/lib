package gormL

import (
	"context"
	"fmt"
	"github.com/flyerxp/lib/v2/config"
	"github.com/flyerxp/lib/v2/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"time"
)

type GormPlugin struct {
	IsPrintSQLDuration bool
	DbName             string
	LogLevel           gormLogger.LogLevel
	baseCtx            context.Context
}

func (l *GormPlugin) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	l.LogLevel = level
	return l
}

func (l *GormPlugin) Info(ctx context.Context, msg string, args ...interface{}) {
}

func (l *GormPlugin) Warn(ctx context.Context, msg string, args ...interface{}) {
}

func (l *GormPlugin) Error(ctx context.Context, msg string, args ...interface{}) {
}

func (l *GormPlugin) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if ctx.Value(logger.GetLogIdKey()) == nil {
		if l.baseCtx != nil {
			ctx = logger.GetContext(l.baseCtx, fmt.Sprintf("gorm_%s_%d", l.DbName, time.Now().UnixNano()))
		} else {
			if config.GetConf().Env != "product" {
				panic("no find logid in context please use have context method")
			} else {
				logger.ErrWithoutCtx(zap.String("gorm", "no find logid in context, please use have context method"))
			}
			return
		}
	}

	sqlKey := logger.RegisterGormCounter(ctx, l.DbName)
	sqlKey.Add()

	elapsed := int(time.Since(begin).Microseconds())
	logger.AddGormSqlTime(ctx, elapsed)

	sql, rows := fc()

	key := sqlKey.GetString("query")
	agKey := sqlKey.GetString("args")

	if err != nil && err != gorm.ErrRecordNotFound {
		logger.AddError(ctx, zap.String(key, sql), zap.Int64(agKey, rows), zap.Error(err))
		if l.IsPrintSQLDuration {
			logger.AddNotice(ctx, zap.String(key, sql), zap.Int64(agKey, rows), zap.Int(sqlKey.GetString("execTime"), elapsed))
		}
		return
	}

	if elapsed > 2000000 {
		logger.AddWarn(ctx, zap.String(key, sql), zap.Int64(agKey, rows))
	}

	if l.IsPrintSQLDuration {
		logger.AddNotice(ctx, zap.String(key, sql), zap.Int64(agKey, rows), zap.Float32(sqlKey.GetString("execTime"), float32(elapsed)/1000))
	}
}
