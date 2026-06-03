package mysqlL

import (
	"context"
	"fmt"
	"github.com/flyerxp/lib/v2/config"
	"github.com/flyerxp/lib/v2/logger"
	"go.uber.org/zap"
	"strings"
	"time"
)

type Hooks struct {
	*zap.Logger
	IsPrintSQLDuration bool
	DbName             string
	baseCtx            context.Context
}

// Before hook will print the query with it's args and return the context with the timestamp
func (h *Hooks) Before(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	if ctx.Value(logger.GetLogIdKey()) == nil {
		if h.baseCtx != nil {
			ctx = logger.GetContext(h.baseCtx, fmt.Sprintf("mysql_%s_%d", h.DbName, time.Now().UnixNano()))
		} else {
			if config.GetConf().Env != "product" {
				panic("no find logid in context please use have context method")
			} else {
				logger.ErrWithoutCtx(zap.String("mysql", "no find logid in context, please use have context method"))
			}
			return ctx, nil
		}
	}
	sqlKey := logger.RegisterMysqlCounter(ctx, h.DbName)
	sqlKey.Add()
	key := sqlKey.GetString("query")
	return context.WithValue(ctx, key, time.Now()), nil
}

// After hook will get the timestamp registered on the Before hook and print the elapsed time
func (h *Hooks) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	sqlKey := logger.RegisterMysqlCounter(ctx, h.DbName)
	key := sqlKey.GetString("query")
	agKey := sqlKey.GetString("args")
	begin, ok := ctx.Value(key).(time.Time)
	var runTime int
	if ok {
		timeout := int(time.Since(begin).Microseconds())
		runTime = timeout
		logger.AddMysqlTime(ctx, timeout)
		if timeout > 2000000 {
			logger.AddWarn(ctx, zap.String(key, query), zap.Any(agKey, args))
		}
	}
	if h.IsPrintSQLDuration {
		logger.AddNotice(ctx, zap.String(key, query), zap.Any(agKey, args), zap.Float32(sqlKey.GetString("execTime"), float32(runTime)/1000))
	}
	return ctx, nil
}
func (h *Hooks) OnError(ctx context.Context, err error, query string, args ...interface{}) error {
	if err.Error() == "driver: bad connection" || err.Error() == "invalid connection" || strings.Contains(err.Error(), "Server shutdown in progress") {
		e, _ := GetEngine(ctx, h.DbName)
		e.CloseDb()
	}
	sqlKey := logger.RegisterMysqlCounter(ctx, h.DbName)
	key := sqlKey.GetString("query")
	agKey := sqlKey.GetString("args")
	var runTime int
	if begin, ok := ctx.Value(key).(time.Time); ok {
		runTime = int(time.Since(begin).Microseconds())
		logger.AddMysqlTime(ctx, runTime)
	}
	logger.AddError(ctx, zap.String(key, query), zap.Any(agKey, args), zap.Error(err))
	if h.IsPrintSQLDuration {
		logger.AddNotice(ctx, zap.String(key, query), zap.Any(agKey, args), zap.Int(sqlKey.GetString("execTime"), runTime))
	}
	return err
}
