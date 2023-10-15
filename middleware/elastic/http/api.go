package http

import (
	"context"
	"fmt"
	"github.com/flyerxp/lib/config"
	"github.com/flyerxp/lib/logger"
	"github.com/flyerxp/lib/middleware/elastic/result"
	json2 "github.com/flyerxp/lib/utils/json"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

type HttpClient struct {
	cluster      config.MidEsConf
	gateway      config.MidEsConf
	active       string
	readTimeout  time.Duration
	writeTimeout time.Duration
	maxIdleConn  time.Duration
	conntTmeout  time.Duration
	http         *fasthttp.Client
	requestIndex uint8
	errReturn    map[string]interface{}
	clusterName  string
	Queue        int64
	FailFun      func(message *EKError)
}
type LastSendInfo struct {
	RequestMethod string `json:"request_method"`
	RequestUrl    string `json:"request_url"`
	RequestBody   string `json:"request_body"`
}
type LastMessage struct {
	RequestMethod string `json:"request_method"`
	RequestUrl    string `json:"request_url"`
	RequestBody   string `json:"request_body"`
	Timeout       string `json:"timeout"`
	CurlError     string `json:"curl_error"`
	ResponseBody  string `json:"response_body"`
	Cluster       string `json:"cluster"`
}
type LastError struct {
	File         string       `json:"file"`
	Line         int          `json:"line"`
	Code         int          `json:"code"`
	LastSendInfo LastSendInfo `json:"last_send_info"`
	Message      *LastMessage `json:"message"`
	Track        error        `json:"track"`
}
type EKError map[string]interface {
}

var defaultReturn result.ErrReturn

func NewHttpClient(ctx context.Context, MidEsConf config.MidEsConf) *HttpClient {
	c := new(HttpClient)
	c.initCluster(MidEsConf)
	c.activeF(ctx)
	c.initHttp(ctx)
	return c
}

func (c *HttpClient) initCluster(MidEsConf config.MidEsConf) {
	c.clusterName = MidEsConf.Name
	c.cluster = MidEsConf
	c.gateway = MidEsConf
}
func (c *HttpClient) initHttp(ctx context.Context) {
	if c.http == nil {
		if c.gateway.ReadTimeout == "" {
			c.readTimeout = time.Duration(600) * time.Millisecond
		} else if v, e := time.ParseDuration(c.gateway.ReadTimeout); e == nil {
			c.readTimeout = v
		} else {
			logger.AddWarn(ctx, zap.String("elastic", " read timeout error"+c.gateway.ReadTimeout))
		}
		if c.gateway.WriteTimeout == "" {
			c.writeTimeout = time.Duration(800) * time.Millisecond
		} else if v, e := time.ParseDuration(c.gateway.WriteTimeout); e == nil {
			c.writeTimeout = v
		} else {
			logger.AddWarn(ctx, zap.String("elastic", " write timeout error"+c.gateway.WriteTimeout))
		}
		if c.gateway.MaxIdleConn == "" {
			c.maxIdleConn = time.Duration(60) * time.Minute
		} else if v, e := time.ParseDuration(c.gateway.MaxIdleConn); e == nil {
			c.maxIdleConn = v
		} else {
			logger.AddWarn(ctx, zap.String("elastic", " max idle timeout error"+c.gateway.MaxIdleConn))
		}
		c.http = &fasthttp.Client{
			Name:                          "flyerxp elastic",
			ReadTimeout:                   c.readTimeout,
			WriteTimeout:                  c.writeTimeout,
			MaxIdleConnDuration:           c.maxIdleConn,
			NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
			DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
			DisablePathNormalizing:        true,
			// increase DNS cache time to an hour instead of default minute
			Dial: (&fasthttp.TCPDialer{
				Concurrency: 4096,
			}).Dial,
		}
	}
}
func (c *HttpClient) activeF(ctx context.Context) {
	if len(c.cluster.Host) == 0 {
		logger.AddError(ctx, zap.String("elastic", c.clusterName+" cluster no address"))
		return
	}
	if c.gateway.AutoDetect {
		c.populateNodes(ctx, 0)
	}
	i := rand.Intn(len(c.cluster.Host))
	c.active = c.cluster.Host[i]
}

func (c *HttpClient) populateNodes(ctx context.Context, tryNums uint) {
	var tmpCluster []string
	//如果节点数大于3 从节点里重新选择一个
	if len(c.cluster.Host) >= 1 && c.active != "" {
		for _, t := range c.cluster.Host {
			if t != c.active {
				tmpCluster = append(tmpCluster, t)
			}
		}
		if len(tmpCluster) == 0 {
			tmpCluster = c.gateway.Host
		}
		i := rand.Intn(len(tmpCluster))
		c.active = tmpCluster[i]
	} else {
		i := rand.Intn(len(c.cluster.Host))
		c.active = c.cluster.Host[i]
	}
	r, e := c.SendRequest(ctx, "GET", "/_nodes/_all/http", "", c.readTimeout, 1000, false)
	if e != nil {
		if tryNums == 0 {
			c.populateNodes(ctx, tryNums+1)
			return
		}
		if tryNums == 1 {
			c.cluster = c.gateway
			if atomic.AddInt64(&c.Queue, +1) == 1 {
				time.AfterFunc(time.Minute*1, func() {
					c.reset(ctx)
					c.clearResetQue()
				})
			}
		}
		return
	}
	resp := new(result.SearchNodes)
	e = json2.Decode(r, &resp)
	if e != nil {
		c.cluster = c.gateway
	} else if resp.NodesT.Total > 0 {
		tmpCluster = []string{}
		pa := ""
		for _, t := range resp.Nodes {
			pa = ""
			if t.Http.PublishAddress != "" {
				_, err := url.Parse("http://" + t.Http.PublishAddress)
				if err == nil {
					if strings.Contains(t.Http.PublishAddress, "/") {
						//k8s 特殊路径处理
						ts := strings.Split(t.Http.PublishAddress, ":")
						pa = t.Host + ":" + ts[len(ts)-1]
					} else {
						//实体机路径
						pa = t.Http.PublishAddress
					}
					tmpCluster = append(tmpCluster, pa)
				}
			}
		}
		if len(tmpCluster) > 0 {
			c.cluster.Host = tmpCluster
		}
	} else {
		c.cluster = c.gateway
	}
}

func (c *HttpClient) clearResetQue() {
	c.Queue = 0
}
func (c *HttpClient) reset(ctx context.Context) {
	c.http.CloseIdleConnections()
	c.http = nil
	c.activeF(ctx)
	c.initHttp(ctx)
}
func (c *HttpClient) Close() {
	c.http.CloseIdleConnections()
}
func (c *HttpClient) httpConnError(err error) (string, bool) {
	errName := ""
	known := false
	if err == fasthttp.ErrTimeout {
		errName = "timeout"
		known = true
	} else if err == fasthttp.ErrNoFreeConns {
		errName = "conn_limit"
		known = true
	} else if err == fasthttp.ErrConnectionClosed {
		errName = "conn_close"
		known = true
	} else {
		errName = reflect.TypeOf(err).String()
		if errName == "*net.OpError" {
			// Write and Read errors are not so often and in fact they just mean timeout problems
			errName = "timeout"
			known = true
		}
	}
	return errName, known
}

func (c *HttpClient) SendRequest(ctx context.Context, method string, url string, sendData string, timeout time.Duration, tryNums uint, isRetryConn bool) ([]byte, error) {
	if c.http == nil {
		c.initHttp(ctx)
	}
	if c.active == "" {
		c.activeF(ctx)
	}
	if c.active == "" {
		defaultReturn.Aggregations = make([]interface{}, 0)
		tmpJson, _ := json2.Encode(defaultReturn)
		return tmpJson, errors.New("cluster no find" + c.clusterName)
	}
	req := fasthttp.AcquireRequest()
	req.SetRequestURI("http://" + c.active + url)
	if c.gateway.User != "" {
		req.URI().SetUsername(c.gateway.User)
		req.URI().SetPassword(c.gateway.Pwd)
	}
	req.Header.SetMethod(method)
	req.Header.SetContentTypeBytes([]byte("application/json"))
	req.Header.Add("Accept-Encoding", "gzip")
	if method == fasthttp.MethodPost || method == fasthttp.MethodDelete || method == fasthttp.MethodPut {
		req.SetBodyString(sendData)
	}
	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()
	var err error
	timeoutDesc := ""
	if timeout > 0 {
		err = c.http.DoTimeout(req, resp, timeout)
		timeoutDesc = timeout.String()
	} else {
		err = c.http.DoTimeout(req, resp, c.readTimeout)
		timeoutDesc = c.readTimeout.String()
	}

	statusCode := resp.StatusCode()
	respBody, errZip := resp.BodyGunzip()
	if err == nil {
		if statusCode == http.StatusOK {
			if errZip == nil {
				return respBody, nil
			} else {
				return nil, errZip
			}
		} else {
			lm := LastMessage{
				RequestMethod: string(req.Header.Method()),
				RequestUrl:    string(req.Header.Host()) + "/" + string(req.Header.RequestURI()),
				RequestBody:   sendData,
				CurlError:     http.StatusText(statusCode),
				ResponseBody:  string(respBody),
				Cluster:       c.gateway.Name,
				Timeout:       timeoutDesc,
			}
			err = errors.New(fmt.Sprintf("ERR invalid HTTP response code: %d body %s", statusCode, string(respBody)))
			logger.AddWarn(ctx, zap.Any("elastic", lm))
			if c.FailFun != nil {
				c.FailFun(c.getError(statusCode, &lm, errors.Wrap(err, "elastic")))
			}
			defaultReturn.Aggregations = make([]interface{}, 0)
			tmpJson, _ := json2.Encode(defaultReturn)
			return tmpJson, err
		}
	} else {
		errName, known := c.httpConnError(err)
		errMsg := ""
		if known {
			errMsg = fmt.Sprintf("WARN conn error: %v\n", errName)
		} else {
			errMsg = fmt.Sprintf("ERR conn failure:%s %v\n", errName, err)
		}
		logger.AddWarn(ctx, zap.String("elastic", errMsg))
		if isRetryConn {
			c.reset(ctx)
		}
		// || errName == "timeout"
		if tryNums < 1 && (errName == "conn_close") {
			return c.SendRequest(ctx, method, url, sendData, timeout, tryNums+1, isRetryConn)
		}
		lm := LastMessage{
			RequestMethod: string(req.Header.Method()),
			RequestUrl:    string(req.Header.Host()) + string(req.Header.RequestURI()),
			RequestBody:   sendData,
			CurlError:     errName,
			ResponseBody:  string(respBody),
			Cluster:       c.active,
			Timeout:       timeoutDesc,
		}
		if c.FailFun != nil {
			c.FailFun(c.getError(statusCode, &lm, errors.Wrap(err, "elastic")))
		}
		defaultReturn.Aggregations = make([]interface{}, 0)
		tmpJson, _ := json2.Encode(defaultReturn)
		return tmpJson, errors.New(errMsg)
	}
}
func (c *HttpClient) getError(code int, lm *LastMessage, err error) *EKError {
	_, codePath, codeLine, _ := runtime.Caller(5)
	le := LastError{
		File: codePath,
		Line: codeLine,
		Code: code,
		LastSendInfo: LastSendInfo{
			RequestMethod: lm.RequestMethod,
			RequestUrl:    lm.RequestUrl,
			RequestBody:   lm.RequestBody,
		},
		Message: lm,
		Track:   err,
	}
	return &EKError{
		"LastError": le,
		"ErrorTime": time.Now(),
		"Lang":      "GO",
	}
}
