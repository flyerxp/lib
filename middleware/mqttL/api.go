package mqttL

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/flyerxp/lib/app"
	config2 "github.com/flyerxp/lib/config"
	"github.com/flyerxp/lib/logger"
	"github.com/flyerxp/lib/middleware/nacos"
	json2 "github.com/flyerxp/lib/utils/json"
	yaml2 "github.com/flyerxp/lib/utils/yaml"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Mqtt 容器
type MqttContainer struct {
	MqttContainer cmap.ConcurrentMap[string, *MqttClient]
	MqttConf      cmap.ConcurrentMap[string, config2.MidMqttConf]
}

var producerQue *mqttProducer

type mqttProducer struct {
	Wg      sync.WaitGroup
	Sending int32
	p       *ants.Pool
}

func init() {
	var err error
	producerQue = new(mqttProducer)
	producerQue.p, err = ants.NewPool(3, ants.WithExpiryDuration(300*time.Second), ants.WithPreAlloc(true), ants.WithDisablePurge(true), ants.WithPanicHandler(func(a any) {
		logger.WarnWithoutCtx(zap.Any("mqtt", a))
	}))
	if err != nil {
		panic(err)
	}
	AsyncInitMqtt()
}

// Mqtt 客户端
type MqttClient struct {
	CurrMqtt mqtt.Client
}

func (m *MqttClient) Producer(ctx context.Context, o *OutMessage) error {
	start := time.Now()
	codeStr := o.TopicStr
	pMessage, err := json2.Encode(getMqttMessage(o, codeStr, ctx))
	if err != nil {
		logger.AddError(ctx, zap.String("mqtt", "json error"), zap.Error(err))
		return err
	}
	err = producerQue.p.Submit(func() {
		producerQue.Wg.Add(1)
		atomic.AddInt32(&producerQue.Sending, 1)
		//qos  1 网络不稳定情况下,有可能重复，
		token := m.CurrMqtt.Publish(codeStr, o.Qos, o.Retained, pMessage)
		ok := token.WaitTimeout(5 * time.Second)
		if !ok {
			logger.ErrWithoutCtx(zap.String("mqttSendTimeout", o.TopicStr), zap.Any(codeStr, pMessage))
		}
		if token.Error() != nil {
			logger.ErrWithoutCtx(zap.String("mqttSendFail", o.TopicStr), zap.Error(token.Error()), zap.Any(codeStr, pMessage))
		}
		atomic.AddInt32(&producerQue.Sending, -1)
		producerQue.Wg.Done()
	})
	if err != nil {
		logger.AddError(ctx, zap.String("matt", "ants errors"), zap.Error(err))
		return err
	}
	logger.AddMqttTime(ctx, int(time.Since(start).Milliseconds()))
	return nil
}

var mqttEngine *MqttContainer

func AsyncInitMqtt() {
	ctx := logger.GetContext(context.Background(), "asyncinitmqtt")
	initEngine(ctx)
	_ = app.RegisterFunc("mqtt", "mqtt", func() {
		Reset(ctx)
	})
}
func initEngine(ctx context.Context) {
	mqttEngine = new(MqttContainer)
	var confList []config2.MidMqttConf
	mqttEngine.MqttConf = cmap.New[config2.MidMqttConf]()
	mqttEngine.MqttContainer = cmap.New[*MqttClient]()
	conf := config2.GetConf()
	confList = conf.Mqtt
	//本地文件中获取
	for _, v := range confList {
		if v.Name != "" {
			mqttEngine.MqttConf.Set(v.Name, v)
		}
	}
	if conf.MqttNacos.Name != "" {
		var yaml []byte
		mqttList := new(config2.MqttConf)
		ns, e := nacos.GetEngine(ctx, conf.MqttNacos.Name)
		if e == nil {
			yaml, e = ns.GetConfig(ctx, conf.MqttNacos.Did, conf.MqttNacos.Group, conf.MqttNacos.Ns)
			if e == nil {
				e = yaml2.DecodeByBytes(yaml, mqttList)
				if e == nil {
					for _, v := range mqttList.List {
						if v.Scheme == "" {
							v.Scheme = "tcp"
						}
						mqttEngine.MqttConf.Set(v.Name, v)
					}
				} else {
					logger.AddError(ctx, zap.Error(errors.New("yaml conver error")))
					logger.WriteErr(ctx)
				}
			} else {
				logger.AddError(ctx, zap.Error(e))
				logger.WriteErr(ctx)
			}
		} else {
			logger.AddError(ctx, zap.Error(e))
			logger.WriteErr(ctx)
		}
	}
}
func GetEngine(ctx context.Context, name string, opsF ...func(ops *mqtt.ClientOptions)) (*MqttClient, error) {
	if mqttEngine == nil {
		initEngine(ctx)
	}
	e, ok := mqttEngine.MqttContainer.Get(name)
	if ok {
		return e, nil
	}
	o, okC := mqttEngine.MqttConf.Get(name)
	if okC {
		objMqtt := newClient(ctx, o, opsF...)
		mqttEngine.MqttContainer.Set(name, objMqtt)
		return objMqtt, nil
	}
	logger.AddError(ctx, zap.Error(errors.New("no find Mqtt config "+name)))
	logger.WriteErr(ctx)
	panic(fmt.Sprintf("no find Mqtt config %s", name))
}

func newClient(ctx context.Context, o config2.MidMqttConf, opsF ...func(ops *mqtt.ClientOptions)) *MqttClient {
	conf := config2.GetConf().App
	opts := mqtt.NewClientOptions()
	if o.Scheme == "ssl" {
		opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: false, ClientAuth: tls.NoClientCert})
	}
	opts.SetClientID(conf.Type + "_" + conf.Name + "_" + time.Now().Format("01_02_150405") + "_" + strconv.Itoa(rand.Intn(1000)))
	for i := range o.Address {
		opts.AddBroker(fmt.Sprintf(o.Scheme+"://%s", o.Address[i]))
	}
	if o.User != "" {
		opts.SetUsername(o.User)
	}
	if o.Pwd != "" {
		opts.SetPassword(o.Pwd)
	}
	//opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: false, ClientAuth: tls.NoClientCert})
	opts.SetKeepAlive(30 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetClientID(conf.Type + "_" + conf.Name + "_" + time.Now().Format("01_02_150405") + "_" + strconv.Itoa(rand.Intn(1000)))
	opts.SetPingTimeout(1 * time.Second)
	opts.SetConnectRetry(true)
	opts.SetAutoReconnect(true)
	opts.SetConnectTimeout(1 * time.Second)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetWriteTimeout(time.Second * 1)
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		logger.WarnWithoutCtx(zap.String("mqtt", "mqtt connect lost"), zap.Error(err))
	})
	opts.SetReconnectingHandler(func(c mqtt.Client, op *mqtt.ClientOptions) {
		logger.WarnWithoutCtx(zap.String("mqtt", "mqtt mqttReConnect"))
	})
	for i := range opsF {
		opsF[i](opts)
	}
	client := mqtt.NewClient(opts)
	start := time.Now()
	token := client.Connect()
	ok := token.WaitTimeout(500 * time.Millisecond)
	if !ok {
		logger.ErrWithoutCtx(zap.String(o.Name+" timeout", strings.Join(o.Address, ",")))
		panic("mqtt 连接超时" + o.Name)
	}
	if token.Error() != nil {
		logger.AddError(ctx, zap.Error(token.Error()))
		logger.WriteErr(ctx)
		panic(token.Error())
	}
	ctime := time.Since(start).Milliseconds()
	logger.AddMqttConnTime(ctx, int(ctime))
	return &MqttClient{client}
}
func (m *MqttClient) GetMqtt() mqtt.Client {
	return m.CurrMqtt
}
func (p *MqttContainer) Reset() {
	for k := range mqttEngine.MqttContainer.Items() {
		if v, ok := mqttEngine.MqttContainer.Pop(k); ok {
			v.CurrMqtt.Disconnect(2000)
		}
	}
	mqttEngine = nil
}
func Reset(ctx context.Context) {
	Flush(ctx)
	if !producerQue.p.IsClosed() {
		err := producerQue.p.ReleaseTimeout(time.Second * 60)
		if err != nil {
			logger.ErrWithoutCtx(zap.Error(err))
		}
	}
	if mqttEngine != nil {
		for _, v := range mqttEngine.MqttContainer.Items() {
			v.CurrMqtt.Disconnect(2000)
		}
		mqttEngine = nil
	}
}
