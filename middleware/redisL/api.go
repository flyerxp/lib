package redisL

import (
	"context"
	"errors"
	"github.com/flyerxp/lib/v2/app"
	config2 "github.com/flyerxp/lib/v2/config"
	"github.com/flyerxp/lib/v2/logger"
	"github.com/flyerxp/lib/v2/middleware/nacos"
	yaml2 "github.com/flyerxp/lib/v2/utils/yaml"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"time"
)

type redisClient struct {
	RedisClient cmap.ConcurrentMap[string, *RedisC]
	RedisConf   cmap.ConcurrentMap[string, config2.MidRedisConf]
}

var RedisEngine *redisClient

type RedisC struct {
	C    redis.UniversalClient
	Conf config2.MidRedisConf
}

func GetEngine(ctx context.Context, name string) (*RedisC, error) {
	if RedisEngine == nil {
		RedisEngine = new(redisClient)
		var confList []config2.MidRedisConf
		RedisEngine.RedisConf = cmap.New[config2.MidRedisConf]()
		RedisEngine.RedisClient = cmap.New[*RedisC]()
		conf := config2.GetConf()
		confList = conf.Redis
		//本地文件中获取
		//RedisEngine.Lock.Lock()
		for _, v := range confList {
			if v.Name != "" {
				RedisEngine.RedisConf.Set(v.Name, v)
			}
		}
		//nacos获取
		if conf.RedisNacos.Name != "" {
			var yaml []byte
			redisList := new(config2.RedisConf)
			ns, e := nacos.GetEngine(ctx, conf.RedisNacos.Name)
			if e == nil {
				yaml, e = ns.GetConfig(ctx, conf.RedisNacos.Did, conf.RedisNacos.Group, conf.RedisNacos.Ns)
				if e == nil {
					e = yaml2.DecodeByBytes(yaml, redisList)
					if e == nil {
						for _, v := range redisList.Redis {
							RedisEngine.RedisConf.Set(v.Name, v)
						}
					} else {
						logger.AddError(ctx, zap.Error(errors.New("yaml conver error")))
					}
				} else {
					logger.AddError(ctx, zap.Error(e))
				}
			} else {
				logger.AddError(ctx, zap.Error(e))
			}
		}
		_ = app.RegisterFunc("redis", "redis close", func() {
			RedisEngine.Reset()
		})
	}
	e, ok := RedisEngine.RedisClient.Get(name)
	if ok {
		return e, nil
	}
	o, okC := RedisEngine.RedisConf.Get(name)
	if okC {
		op := &redis.UniversalOptions{
			Addrs:        o.Address,
			MasterName:   o.Master,
			Username:     o.User,
			Password:     o.Pwd,
			PoolTimeout:  time.Millisecond * time.Duration(500),
			ReadTimeout:  time.Millisecond * time.Duration(500),
			WriteTimeout: time.Millisecond * time.Duration(500),
			DialTimeout:  time.Millisecond * time.Duration(500),
			MaxIdleConns: 30,
		}
		if o.WriteTimeout > 0 {
			op.WriteTimeout = time.Millisecond * time.Duration(o.WriteTimeout)
		}
		if o.ReadTimeout > 0 {
			op.ReadTimeout = time.Millisecond * time.Duration(o.ReadTimeout)
		}
		if o.DialTimeout > 0 {
			op.DialTimeout = time.Millisecond * time.Duration(o.DialTimeout)
		}
		if o.PoolTimeout > 0 {
			op.PoolTimeout = time.Millisecond * time.Duration(o.PoolTimeout)
		}
		objRedis := redis.NewUniversalClient(op)

		objRedis.AddHook(HookLog{})
		objRedisC := new(RedisC)
		go func() {
			objRedisC.C.Ping(ctx)
		}()
		objRedisC.C = objRedis
		objRedisC.Conf = o
		RedisEngine.RedisClient.Set(name, objRedisC)
		return objRedisC, nil
	}
	logger.AddError(ctx, zap.Error(errors.New("no find redis config "+name)))
	return nil, errors.New("no find redis config " + name)
}
func (r *RedisC) GetConf(name string) config2.MidRedisConf {
	return r.Conf
}
func (r *redisClient) Reset() {
	if RedisEngine != nil {
		for _, v := range RedisEngine.RedisClient.Items() {
			if v.C != nil {
				_ = v.C.Close()
			}
		}
		RedisEngine = nil
	}
}
func IsNilErr(e error) bool {
	if e == nil {
		return false
	}
	return e.Error() == "redis: nil"
}
