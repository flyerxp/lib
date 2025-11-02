package pulsarL

import (
	"context"
	"errors"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/apache/pulsar-client-go/pulsar/log"
	"github.com/flyerxp/lib/v2/app"
	config2 "github.com/flyerxp/lib/v2/config"
	"github.com/flyerxp/lib/v2/logger"
	"github.com/flyerxp/lib/v2/middleware/nacos"
	yaml2 "github.com/flyerxp/lib/v2/utils/yaml"
	cmap "github.com/orcaman/concurrent-map/v2"
	"go.uber.org/zap"
	"strings"
	"sync"
	"time"
)

// Pulsar 容器
type PulsarContainer struct {
	PulsarContainer cmap.ConcurrentMap[string, *PulsarClient]
	PulsarConf      cmap.ConcurrentMap[string, config2.MidPulsarConf]
	MyLock          *sync.Mutex
}

func init() {
	AsyncInitPulsar()
}

// Pulsar 客户端
type PulsarClient struct {
	CurrPulsar pulsar.Client
}

var pulsarEngine *PulsarContainer

func AsyncInitPulsar() {
	ctx := logger.GetContext(context.Background(), "asyncinitpulsar")
	go func() {
		initEngine(ctx)
		InitTopic(ctx)
		ProducerPre(ctx)
	}()
	_ = app.RegisterFunc("pulsar", "pulsar", func() {
		Reset(ctx)
	})
}
func initEngine(ctx context.Context) {
	pulsarEngine = new(PulsarContainer)
	var confList []config2.MidPulsarConf
	pulsarEngine.PulsarConf = cmap.New[config2.MidPulsarConf]()
	pulsarEngine.PulsarContainer = cmap.New[*PulsarClient]()
	conf := config2.GetConf()
	confList = conf.Pulsar
	//本地文件中获取
	for _, v := range confList {
		if v.Name != "" {
			pulsarEngine.PulsarConf.Set(v.Name, v)
		}
	}
	if conf.PulsarNacos.Name != "" {
		var yaml []byte
		pulsarList := new(config2.PulsarConf)

		ns, e := nacos.GetEngine(ctx, conf.PulsarNacos.Name)
		if e == nil {
			yaml, e = ns.GetConfig(ctx, conf.PulsarNacos.Did, conf.PulsarNacos.Group, conf.PulsarNacos.Ns)
			if e == nil {
				e = yaml2.DecodeByBytes(yaml, pulsarList)
				if e == nil {
					for _, v := range pulsarList.List {
						pulsarEngine.PulsarConf.Set(v.Name, v)
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
func GetEngine(ctx context.Context, name string) (*PulsarClient, error) {
	if pulsarEngine == nil {
		if pulsarEngine.MyLock == nil {
			pulsarEngine.MyLock = new(sync.Mutex)
		}
		pulsarEngine.MyLock.Lock()
		defer pulsarEngine.MyLock.Unlock()
		if pulsarEngine == nil {
			initEngine(ctx)
		}
	}
	e, ok := pulsarEngine.PulsarContainer.Get(name)
	if ok {
		return e, nil
	}
	o, okC := pulsarEngine.PulsarConf.Get(name)
	if okC {
		objPulsar := newClient(ctx, o)
		pulsarEngine.PulsarContainer.Set(name, objPulsar)
		return objPulsar, nil
	}
	logger.AddError(ctx, zap.Error(errors.New("no find Pulsar config "+name)))
	logger.WriteErr(ctx)
	panic(fmt.Sprintf("no find Pulsar config %s", name))
}

// https://github.com/golang-migrate/migrate/blob/master/database/Pulsar/README.md

func newClient(ctx context.Context, o config2.MidPulsarConf) *PulsarClient {
	op := pulsar.ClientOptions{
		URL:               "pulsar://" + strings.Join(o.Address, ","),
		Logger:            log.NewLoggerWithLogrus(getLog()),
		ConnectionTimeout: time.Second * 1,
		OperationTimeout:  time.Second * 1,
	}
	if o.Token != "" {
		op.Authentication = pulsar.NewAuthenticationToken(o.Token)
	}
	if o.User != "" && o.Pwd != "" {
		var e error
		op.Authentication, e = pulsar.NewAuthenticationBasic(o.User, o.Pwd)
		logger.AddError(ctx, zap.String("pulsar", "pulsar user auth err"), zap.Error(e))
	}
	client, err := pulsar.NewClient(op)
	if err != nil {
		logger.AddError(ctx, zap.Error(err))
		logger.WriteErr(ctx)
	}
	return &PulsarClient{client}
}

func (m *PulsarClient) GetPulsar() pulsar.Client {
	return m.CurrPulsar
}
func (p *PulsarContainer) Reset() {
	for k := range pulsarEngine.PulsarContainer.Items() {
		if v, ok := pulsarEngine.PulsarContainer.Pop(k); ok {
			v.CurrPulsar.Close()
		}
	}
	pulsarEngine = nil
}
func Reset(ctx context.Context) {
	Flush(ctx)
	if pulsarEngine != nil {
		for _, v := range pulsarEngine.PulsarContainer.Items() {
			v.CurrPulsar.Close()
		}
		initProducer()
		pulsarEngine = nil
		_, _ = topicDistributionF(ctx)
	}
}
