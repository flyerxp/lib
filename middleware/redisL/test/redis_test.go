package redisL

import (
	"context"
	"fmt"
	"github.com/flyerxp/lib/app"
	"github.com/flyerxp/lib/logger"
	"github.com/flyerxp/lib/middleware/redisL"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestConf(t *testing.T) {

	s := time.Now()
	logger.AddNotice(zap.String("a", "cccccccccccccccc"))
	big1 := logger.StartTime("beg1")
	r, _ := redisL.GetEngine("pubRedis", context.Background())

	l := logger.StartTime("redis-read")
	r.C.Get(context.Background(), "aaaa")
	l.Stop()
	l2 := logger.StartTime("redis-read")
	//time.Sleep(time.Second)
	l2.Stop()
	big1.Stop()
	logger.WriteLine()
	app.Shutdown(context.Background())
	redisL.RedisEngine.Reset()
	fmt.Println(time.Since(s).Milliseconds())
}
