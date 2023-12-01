package mysqlL

import (
	"context"
	"github.com/flyerxp/lib/config"
	"github.com/flyerxp/lib/logger"
	"go.uber.org/zap"
	"time"
)

type Hooks struct {
	*zap.Logger
	SqlKey             *logger.Counter
	IsPrintSQLDuration bool
	DbName             string
}

// Before hook will print the query with it's args and return the context with the timestamp
func (h *Hooks) Before(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	if ctx.Value(logger.GetLogIdKey()) == nil {
		if config.GetConf().Env != "product" {
			panic("no find logid in context please use have context method")
		} else {
			logger.ErrWithoutCtx(zap.String("mysql", "no find logid in context, please use have context method"))
		}
		return ctx, nil
	}
	h.SqlKey = logger.RegisterMysqlCounter(ctx, h.DbName)
	h.SqlKey.Add()
	key := h.SqlKey.GetString("query")
	return context.WithValue(ctx, key, time.Now()), nil
}

// After hook will get the timestamp registered on the Before hook and print the elapsed time
func (h *Hooks) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	key := h.SqlKey.GetString("query")
	agKey := h.SqlKey.GetString("args")
	begin, ok := ctx.Value(key).(time.Time)
	var runTime = 1000
	if ok {
		timeout := int(time.Since(begin).Milliseconds())
		runTime = timeout
		logger.AddMysqlTime(ctx, timeout)
		if timeout > 2000 {
			logger.AddWarn(ctx, zap.String(key, query), zap.Any(agKey, args))
		}
	}
	if h.IsPrintSQLDuration {
		logger.AddNotice(ctx, zap.String(key, query), zap.Any(agKey, args), zap.Int(h.SqlKey.GetString("execTime"), runTime))
	}
	return ctx, nil
}
func (h *Hooks) OnError(ctx context.Context, err error, query string, args ...interface{}) error {
	/*if err.Error() == "driver: bad connection" || err.Error() == "invalid connection" {
		e, _ := GetEngine(ctx, h.DbName)
		e.CloseDb()
	}*/
	key := h.SqlKey.GetString("query")
	agKey := h.SqlKey.GetString("args")
	var runTime = 1000
	if begin, ok := ctx.Value(key).(time.Time); ok {
		runTime = int(time.Since(begin).Milliseconds())
		logger.AddMysqlTime(ctx, runTime)
	}
	logger.AddError(ctx, zap.String(key, query), zap.Any(agKey, args), zap.Error(err))
	if h.IsPrintSQLDuration {
		logger.AddNotice(ctx, zap.String(key, query), zap.Any(agKey, args), zap.Int(h.SqlKey.GetString("execTime"), runTime))
	}
	return nil
}
