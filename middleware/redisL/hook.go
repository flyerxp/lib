package redisL

import (
	"context"
	"github.com/flyerxp/lib/config"
	"github.com/flyerxp/lib/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"net"
	"time"
)

type HookLog struct{}

func (HookLog) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		if ctx.Value(logger.GetLogIdKey()) == nil {
			if config.GetConf().Env != "product" {
				panic("no find logid in context please use have context method")
			} else {
				logger.ErrWithoutCtx(zap.String("mysql", "no find logid in context, please use have context method"))
				c, e := next(ctx, network, addr)
				return c, e
			}
		}
		t := time.Now()
		l := logger.StartTime(addr)
		c, e := next(ctx, network, addr)
		l.Stop(ctx)
		logger.AddRedisConnTime(ctx, int(time.Since(t).Milliseconds()))
		return c, e
	}
}
func (HookLog) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		t := time.Now()
		e := next(ctx, cmd)
		logger.AddRedisTime(ctx, int(time.Since(t).Milliseconds()))
		return e
	}
}
func (HookLog) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		t := time.Now()
		c := next(ctx, cmds)
		logger.AddRedisTime(ctx, int(time.Since(t).Milliseconds()))
		return c
	}
}
