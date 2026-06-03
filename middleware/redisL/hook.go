package redisL

import (
	"context"
	"fmt"
	"github.com/flyerxp/lib/v2/logger"
	"github.com/redis/go-redis/v9"
	"net"
	"time"
)

type HookLog struct {
	baseCtx context.Context
}

func (h HookLog) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		if ctx.Value(logger.GetLogIdKey()) == nil && h.baseCtx != nil {
			ctx = logger.GetContext(h.baseCtx, fmt.Sprintf("redis_dial_%s_%d", addr, time.Now().UnixNano()))
		}
		t := time.Now()
		l := logger.StartTime(addr)
		c, e := next(ctx, network, addr)
		l.Stop(ctx)
		logger.AddRedisConnTime(ctx, int(time.Since(t).Microseconds()))
		return c, e
	}
}
func (HookLog) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		t := time.Now()
		e := next(ctx, cmd)
		logger.AddRedisTime(ctx, int(time.Since(t).Microseconds()))
		return e
	}
}
func (HookLog) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		t := time.Now()
		c := next(ctx, cmds)
		logger.AddRedisTime(ctx, int(time.Since(t).Microseconds()))
		return c
	}
}
