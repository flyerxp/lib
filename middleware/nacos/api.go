package nacos

import (
	"context"
	"errors"
	config2 "github.com/flyerxp/lib/v2/config"
	"github.com/flyerxp/lib/v2/logger"
	"github.com/flyerxp/lib/v2/utils/json"
	"github.com/flyerxp/lib/v2/utils/stringL"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Client struct {
	BaseOption config2.MidNacos
	HttpPool   *sync.Pool
	Token      *AccessToken
}
type AccessToken struct {
	AccessToken string        `json:"accessToken"`
	TokenTtl    time.Duration `json:"tokenTtl"`
	GlobalAdmin bool          `json:"globalAdmin"`
	Username    string        `json:"username"`
	Expiration  int64         `json:"expiration,omitempty"`
}

var redisClient redis.UniversalClient

func GetEngine(ctx context.Context, name string) (*Client, error) {
	for _, v := range config2.GetConf().Nacos {
		if v.Name == name {
			return newClient(v), nil
		}
	}
	logger.AddError(ctx, zap.Error(errors.New("nacos conf no find "+name)))
	return nil, errors.New("nacos conf no find " + name)
}
func newClient(o config2.MidNacos) *Client {
	if redisClient == nil {
		op := &redis.UniversalOptions{
			Addrs:        o.Redis.Address,
			MasterName:   o.Redis.Master,
			PoolTimeout:  time.Millisecond * time.Duration(500),
			ReadTimeout:  time.Millisecond * time.Duration(500),
			WriteTimeout: time.Millisecond * time.Duration(500),
			DialTimeout:  time.Millisecond * time.Duration(500),
			MaxIdleConns: 30,
			MaxRetries:   3,
			//ConnMaxLifetime: 30 * time.Second,
			ConnMaxIdleTime: 30 * time.Second,
		}
		if o.Redis.User != "" {
			op.Username = o.Redis.User
		}
		if o.Redis.Pwd != "" {
			op.Password = o.Redis.Pwd
		}
		redisClient = redis.NewUniversalClient(op)
	}

	c := &Client{
		o,
		&sync.Pool{
			New: func() any {
				n := newHttpClient(stringL.GetMd5(o.Url))
				return n
			},
		},
		new(AccessToken),
	}
	return c
}
func (n *Client) GetKey(url string) string {
	key := n.BaseOption.Url + "@@" + url
	return stringL.GetMd5(key)
}
func (n *Client) getUrl(url string) string {
	return n.BaseOption.Url + url
}
func (n *Client) getDataFromCache(ctx context.Context, cacheKey string) (*redis.StringCmd, error) {
	rv := redisClient.Get(ctx, cacheKey)
	return rv, nil
}
func (n *Client) DelToken(ctx context.Context) {
	key := n.GetKey("/v1/auth/login")
	redisClient.Del(ctx, key)
}

func (n *Client) redisIsNilErr(e error) bool {
	if e == nil {
		return false
	}
	return e.Error() == "redis: nil"
}

func (n *Client) GetToken(ctx context.Context) (*AccessToken, error) {
	if n.Token != nil && n.Token.Expiration > time.Now().Unix() {
		return n.Token, nil
	}
	key := n.GetKey("/v1/auth/login")
	rv, err := n.getDataFromCache(ctx, key)
	// 从缓存中获取
	if err == nil && !n.redisIsNilErr(rv.Err()) {
		token := new(AccessToken)
		bt, e := rv.Bytes()
		jsonErr := json.Decode(bt, token)
		if jsonErr == nil && token.Expiration > time.Now().Unix() {
			n.Token = token
			return token, e
		}
	}
	s := logger.StartTime("nacos-get-token")
	hc := n.HttpPool.Get().(*httpClient)

	bToken, bErr := hc.SendRequest(ctx, "POST", n.getUrl("/v1/auth/login"), "username="+n.BaseOption.User+"&password="+n.BaseOption.Pwd, 0, 0)
	n.HttpPool.Put(hc)
	s.Stop(ctx)
	if bErr != nil {
		logger.AddError(ctx, zap.String("nacos", " get failed "+n.getUrl("/v1/auth/login")), zap.Error(bErr))
		logger.WriteErr(ctx)
		return nil, errors.New("nacos request fail")
	} else {
		token := new(AccessToken)
		err = json.Decode(bToken, token)
		if err == nil {
			token.TokenTtl -= 10
			token.Expiration = time.Now().Unix() + int64(token.TokenTtl)
			jsonStr, jsonErr := json.Encode(token)
			if jsonErr == nil && token.TokenTtl > 10 {
				redisClient.Set(ctx, key, string(jsonStr), time.Second*token.TokenTtl)
			} else {
				return nil, jsonErr
			}
			n.Token = token
			return token, err
		} else {
			return nil, err
		}
	}
}

func (n *Client) DeleteCache(ctx context.Context, did string, gp string, ns string) string {
	key := n.GetKey("/nacos/v1/cs/configs" + "@@" + did + "@@" + gp + "@@" + ns)
	redisClient.Del(ctx, key)
	return key
}
func (n *Client) GetConfig(ctx context.Context, did string, gp string, ns string) ([]byte, error) {
	start := time.Now()
	key := n.GetKey("/nacos/v1/cs/configs" + "@@" + did + "@@" + gp + "@@" + ns)
	rv, rErr := n.getDataFromCache(ctx, key)
	if rErr == nil && !n.redisIsNilErr(rv.Err()) {
		logger.AddNacosTime(ctx, int(time.Since(start).Microseconds()))
		return rv.Bytes()
	}
	token, err := n.GetToken(ctx)
	//接口报错，返回空
	if err != nil {
		logger.AddNacosTime(ctx, int(time.Since(start).Microseconds()))
		logger.AddError(ctx, zap.Error(err))
		return []byte{}, err
	} else {
		s := logger.StartTime("nacos-get-config")
		hc := n.HttpPool.Get().(*httpClient)
		bYaml, bErr := hc.SendRequest(ctx, "GET", n.getUrl("/v1/cs/configs?accessToken="+token.AccessToken+"&tenant="+ns+"&dataId="+did+"&group="+gp), "", 0, 0)
		n.HttpPool.Put(hc)
		s.Stop(ctx)
		if bErr == nil {
			sYaml := string(bYaml)
			if rv == nil || rv.String() != sYaml {
				redisClient.Set(ctx, key, sYaml, time.Second*86400*2)
			}
			logger.AddNacosTime(ctx, int(time.Since(start).Microseconds()))
			return bYaml, nil
		} else {
			if rErr != nil && rv != nil && rv.Val() != "" {
				logger.AddNacosTime(ctx, int(time.Since(start).Microseconds()))
				return rv.Bytes()
			}
		}
		logger.AddNacosTime(ctx, int(time.Since(start).Microseconds()))
		return []byte{}, bErr
	}
}
