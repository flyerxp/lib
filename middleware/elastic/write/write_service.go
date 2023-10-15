package write

import (
	"context"
	"github.com/flyerxp/lib/middleware/elastic/http"
	"github.com/flyerxp/lib/middleware/elastic/table"
	"time"
)

type WriteService struct {
	httpClient   *http.HttpClient
	table        table.Table
	version      float32
	timeOut      time.Duration
	source       []string
	context      context.Context
	IsError      bool
	ErrorDetails error
	Dsl          string
	writeDo      *writeDo
}
type QEngine interface {
	GetClient() *http.HttpClient
	GetVersion() float32
	GetTable() table.Table
}

// 获取一个serche服务
func GetWriteService(engine QEngine, ctx context.Context) *WriteService {
	s := new(WriteService)
	s.InitService(engine)
	s.context = ctx
	s.writeDo.ContextF(ctx)
	return s
}
func (s *WriteService) InitService(e QEngine) {
	s.httpClient = e.GetClient()
	s.table = e.GetTable()
	s.version = e.GetVersion()
	s.writeDo = new(writeDo)
}

type Write struct {
	Index         string      `json:"index"`
	PrimaryInt    int         `json:"primary_int"`
	PrimaryString string      `json:"primary_string"`
	Routing       string      `json:"routing"`
	Data          interface{} `json:"data"`
}

func (s *WriteService) Insert(v []*Write) (*WriteResult, error) {
	return s.writeDo.insert(s, v)
}
func (s *WriteService) Update(v []*Write) (*WriteResult, error) {
	return s.writeDo.update(s, v)
}
func (s *WriteService) CEsFields() *Write {
	return new(Write)
}
func (s *WriteService) setError(e error) {
	if e != nil {
		s.IsError = true
		s.ErrorDetails = e
	}
}
