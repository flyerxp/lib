package elastic

import (
	"context"
	"github.com/flyerxp/lib/app"
	"github.com/flyerxp/lib/config"
	"github.com/flyerxp/lib/logger"
	"github.com/flyerxp/lib/middleware/nacos"
	yaml2 "github.com/flyerxp/lib/utils/yaml"
	"github.com/jmoiron/sqlx"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"sync"
)

type EsContainer struct {
	EsContainer cmap.ConcurrentMap[string, *ElasticClient]
	EsConf      cmap.ConcurrentMap[string, config.MidEsConf]
}
type ElasticClient struct {
	Poll   *sync.Pool
	CurrEs *QEngine
}

var EsEngine *EsContainer

func GetEngine(ctx context.Context, name string) (*ElasticClient, error) {
	if EsEngine == nil {
		EsEngine = new(EsContainer)
		var confList []config.MidEsConf
		EsEngine.EsConf = cmap.New[config.MidEsConf]()
		EsEngine.EsContainer = cmap.New[*ElasticClient]()
		conf := config.GetConf()
		confList = conf.Elastic
		//本地文件中获取
		for _, v := range confList {
			if v.Name != "" {
				EsEngine.EsConf.Set(v.Name, v)
			}
		}
		//nacos获取
		if conf.ElasticNacos.Name != "" {
			var yaml []byte
			elasticList := new(config.ElasticConf)
			ns, e := nacos.GetEngine(ctx, conf.ElasticNacos.Name)
			if e == nil {
				yaml, e = ns.GetConfig(ctx, conf.ElasticNacos.Did, conf.ElasticNacos.Group, conf.ElasticNacos.Ns)

				if e == nil {
					e = yaml2.DecodeByBytes(yaml, elasticList)
					if e == nil {
						for _, v := range elasticList.List {
							EsEngine.EsConf.Set(v.Name, v)
						}
					} else {
						logger.AddError(ctx, zap.Error(errors.New("yaml conver error")))
					}
				}
			}
		}
		_ = app.RegisterFunc("elastic", "elastic close", func() {
			EsEngine.Reset()
		})
	}
	e, ok := EsEngine.EsContainer.Get(name)
	if ok {
		return e, nil
	}
	o, okC := EsEngine.EsConf.Get(name)
	if okC {
		objelastic := newClient(ctx, o)
		EsEngine.EsContainer.Set(name, objelastic)
		return objelastic, nil
	}
	logger.AddError(ctx, zap.Error(errors.New("no find elastic config "+name)))
	return nil, errors.New("no find elastic config " + name)
}
func newClient(ctx context.Context, o config.MidEsConf) *ElasticClient {
	c := &sync.Pool{
		New: func() any {
			q, _ := newEngine(ctx, o)
			return q
		},
	}
	return &ElasticClient{c, nil}
}
func (m *ElasticClient) GetElastic() *QEngine {
	if m.CurrEs == nil {
		m.CurrEs = m.Poll.Get().(*QEngine)
	}
	return m.CurrEs
}
func (m *ElasticClient) PutDb(a *sqlx.DB) {
	m.Poll.Put(a)
}
func (m *EsContainer) Reset() {
	if EsEngine != nil {
		for _, v := range EsEngine.EsContainer.Items() {
			if v.CurrEs != nil {
				_ = v.CurrEs.Client.Close
			}
		}
		EsEngine = nil
	}
}
