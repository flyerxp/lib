package redisL

import (
	"context"
	"fmt"
	"github.com/flyerxp/lib/logger"
	"github.com/flyerxp/lib/middleware/redisL"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestConf(t *testing.T) {
	ctx := logger.GetContext(context.Background(), "test")
	//s := time.Now()
	logger.AddNotice(ctx, zap.String("a", "cccccccccccccccc"))
	//big1 := logger.StartTime("beg1")
	start := time.Now()
	fmt.Println("i start")
	r, _ := redisL.GetEngine(ctx, "pubRedis")
	fmt.Println(time.Since(start).Seconds(), "=======11111=======")
	l := logger.StartTime("redis-read")
	r.C.Set(ctx, "aaaa", "bb", time.Second*30)
	rv := r.C.Get(ctx, "aaaa")
	l.Stop(ctx)
	fmt.Println(rv.Err(), rv.Val(), time.Since(start).Seconds(), "=======2222=======")
	return
	/*l2 := logger.StartTime("redis-read")
	//time.Sleep(time.Second)
	l2.Stop(ctx)
	big1.Stop(ctx)
	logger.WriteLine(ctx)
	app.Shutdown(ctx)
	redisL.RedisEngine.Reset()
	fmt.Println(time.Since(s).Milliseconds())*/
}
