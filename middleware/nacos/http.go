package nacos

import (
	"context"
	"errors"
	"fmt"
	"github.com/flyerxp/lib/v2/logger"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/valyala/fasthttp"
	zap "go.uber.org/zap"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type httpClient struct {
	readTimeout  time.Duration
	writeTimeout time.Duration
	maxIdleConn  time.Duration
	conntTmeout  time.Duration
	http         *fasthttp.Client
	errReturn    map[string]interface{}
	Token        string
}

var drivers = cmap.New[*httpClient]()

type lastMessage struct {
	RequestMethod string `json:"request_method"`
	RequestUrl    string `json:"request_url"`
	RequestBody   string `json:"request_body"`
	Timeout       string `json:"timeout"`
	CurlError     string `json:"curl_error"`
	ResponseBody  string `json:"response_body"`
	Cluster       string `json:"cluster"`
}

func newHttpClient(cName string) *httpClient {
	if h, ok := drivers.Get(cName); ok {
		return h
	}
	c := new(httpClient)
	c.initHttp()
	drivers.Set(cName, c)
	return c
}

func (c *httpClient) initHttp() {
	if c.http == nil {
		if c.readTimeout == 0 {
			c.readTimeout = time.Duration(600) * time.Millisecond
		}
		if c.writeTimeout == 0 {
			c.writeTimeout = time.Duration(800) * time.Millisecond
		}
		if c.maxIdleConn == 0 {
			c.maxIdleConn = time.Duration(60) * time.Minute
		}
		c.http = &fasthttp.Client{
			Name:                          "lang go nacos",
			ReadTimeout:                   c.readTimeout,
			WriteTimeout:                  c.writeTimeout,
			MaxIdleConnDuration:           c.maxIdleConn,
			NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
			DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
			DisablePathNormalizing:        true,
			// increase DNS cache time to an hour instead of default minute
			Dial: (&fasthttp.TCPDialer{
				Concurrency: 10,
			}).Dial,
		}
	}
}
func (c *httpClient) reset() {
	c.http.CloseIdleConnections()
	c.http = nil
	c.initHttp()
}

func (c *httpClient) httpConnError(err error) (string, bool) {
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
func (c *httpClient) SendRequest(ctx context.Context, method string, url string, sendData string, timeout time.Duration, tryNums uint) ([]byte, error) {
	if c.http == nil {
		c.initHttp()
	}
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	//req.SetBodyRaw([]byte{})
	req.Header.SetMethod(method)
	req.Header.SetContentTypeBytes([]byte("application/x-www-form-urlencoded"))
	//req.Header.Add("Accept-Encoding", "gzip")
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
	respBody := resp.Body()
	if err == nil {
		if statusCode == http.StatusOK {
			return respBody, nil
		} else {
			lm := lastMessage{
				RequestMethod: string(req.Header.Method()),
				RequestUrl:    string(req.Header.Host()) + string(req.Header.RequestURI()),
				RequestBody:   sendData,
				CurlError:     http.StatusText(statusCode),
				ResponseBody:  string(respBody),
				Timeout:       timeoutDesc,
			}
			if strings.Contains(url, "auth") {
				lm.RequestBody = ""
			}
			err = errors.New(fmt.Sprintf("ERR invalid HTTP response code: %d body %s", statusCode, string(respBody)))
			logger.AddWarn(ctx, zap.Any("nacos request fail info1", lm))
			return nil, err
		}
	} else {
		errName, known := c.httpConnError(err)
		errMsg := ""
		if known {
			errMsg = fmt.Sprintf("WARN conn error: %v\n", errName)
		} else {
			errMsg = fmt.Sprintf("ERR conn failure:%s %v\n", errName, err)
		}
		logger.AddWarn(ctx, zap.Error(err), zap.Any("nacos request error info2", errMsg))
		// || errName == "timeout"
		if tryNums < 1 && (errName == "conn_close") {
			return c.SendRequest(ctx, method, url, sendData, timeout, tryNums+1)
		}
		lm := lastMessage{
			RequestMethod: string(req.Header.Method()),
			RequestUrl:    string(req.Header.Host()) + string(req.Header.RequestURI()),
			RequestBody:   sendData,
			CurlError:     errName,
			ResponseBody:  string(respBody),
			Timeout:       timeoutDesc,
		}
		if strings.Contains(url, "auth") {
			lm.RequestBody = ""
		}
		logger.AddWarn(ctx, zap.Error(err), zap.Any("nacos request fail info", lm))
		return nil, errors.New(errMsg)
	}
}
