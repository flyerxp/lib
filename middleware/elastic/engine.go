package elastic

import (
	"context"
	"github.com/flyerxp/lib/config"
	"github.com/flyerxp/lib/logger"
	"github.com/flyerxp/lib/middleware/elastic/http"
	"github.com/flyerxp/lib/middleware/elastic/query"
	"github.com/flyerxp/lib/middleware/elastic/table"
	"github.com/flyerxp/lib/middleware/elastic/write"
)

type QEngine struct {
	Client                 *http.HttpClient
	Tables                 table.Table
	Version                float32
	MaxWindowResult        map[string]int
	TrackTotalHits         map[string]int
	DefaultMaxWindowResult int
	DefaultTrackTotalHits  int
}

//lint:ignore U1000 Ignore unused function temporarily for debugging
func newEngine(ctx context.Context, c config.MidEsConf) (*QEngine, error) {
	engine := &QEngine{
		Client:  http.NewHttpClient(ctx, c),
		Tables:  table.Table{},
		Version: 7,
	}
	engine.MaxWindowResult = c.MaxWindowResult
	engine.TrackTotalHits = c.TrackTotalHits
	if c.DefaultTrackTotalHits == 0 {
		engine.DefaultTrackTotalHits = 30000
	} else {
		engine.DefaultTrackTotalHits = c.DefaultTrackTotalHits
	}
	if c.DefaultMaxWindowResult == 0 {
		engine.DefaultMaxWindowResult = 100000
	} else {
		engine.DefaultMaxWindowResult = c.DefaultMaxWindowResult
	}
	return engine, nil
}
func (e *QEngine) SetVersion(v float32) {
	e.Version = v
}
func (e *QEngine) SetAllowLocalCurd(v bool) {
	e.Tables.IsAllowLocalCurd = v
}
func (e *QEngine) SetMaxWindowResult(v int) {
	e.Tables.MaxWindowResult = v
}
func (e *QEngine) SetRouting(v string) {
	e.Tables.Routing = v
}
func (e *QEngine) SetTable(tName string) {
	e.Tables.Name = tName
	if e.MaxWindowResult != nil && len(e.MaxWindowResult) > 0 {
		if n, ok := e.MaxWindowResult[tName]; ok {
			e.Tables.MaxWindowResult = n
		}
	}
	if e.Tables.MaxWindowResult == 0 {
		e.Tables.MaxWindowResult = e.DefaultMaxWindowResult
	}
	if e.TrackTotalHits != nil && len(e.TrackTotalHits) > 0 {
		if n, ok := e.TrackTotalHits[tName]; ok {
			e.Tables.TrackTotalHits = uint64(n)
		}
	}
	if e.Tables.TrackTotalHits == 0 {
		e.Tables.TrackTotalHits = uint64(e.DefaultTrackTotalHits)
	}
}
func (e *QEngine) GetTable() table.Table {
	return e.Tables
}
func (e *QEngine) GetClient() *http.HttpClient {
	return e.Client
}
func (e *QEngine) GetVersion() float32 {
	return e.Version
}
func (e *QEngine) GetSearchService(ctx context.Context) *query.SearchSevice {
	if ctx.Value(logger.GetLogIdKey()) == nil {
		panic("context no " + logger.GetLogIdKey())
	}
	return query.GetSearchService(ctx, e)
}
func (e *QEngine) GetWriteService(ctx context.Context) *write.WriteService {
	if ctx.Value(logger.GetLogIdKey()) == nil {
		panic("context no " + logger.GetLogIdKey())
	}
	return write.GetWriteService(e, ctx)
}
