package query

import (
	"context"
	"github.com/flyerxp/lib/logger"
	"github.com/flyerxp/lib/middleware/elastic/http"
	"github.com/flyerxp/lib/middleware/elastic/table"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type SearchSevice struct {
	httpClient      *http.HttpClient
	table           table.Table
	version         float32
	routing         string
	scroll          string
	timeOut         time.Duration
	constantScore   bool
	maxWindowResult int
	source          []string
	sort            Sort
	query           interface{}
	from            int
	size            int
	match           ConnMatchAnd
	agge            Agge
	getOtherInfo    bool
	trackTotalHits  uint64
	context         context.Context
	where           ConnAnd
	IsError         bool
	ErrorDetails    error
	Dsl             string
	searchDo        *searchDo
}

func (s *SearchSevice) SetConstantScore(v bool) *SearchSevice {
	s.constantScore = v
	return s
}

type QEngine interface {
	GetClient() *http.HttpClient
	GetVersion() float32
	GetTable() table.Table
}

// 获取一个serche服务
func GetSearchService(ctx context.Context, engine QEngine) *SearchSevice {
	s := new(SearchSevice)
	s.InitService(engine)
	s.context = ctx
	s.searchDo.ContextF(ctx)
	if s.table.Routing != "" {
		s.routing = s.table.Routing
	}
	if s.table.MaxWindowResult > 0 {
		s.maxWindowResult = s.table.MaxWindowResult
	}
	if s.table.TrackTotalHits > 0 {
		s.trackTotalHits = s.table.TrackTotalHits
	}
	return s
}
func (s *SearchSevice) InitService(e QEngine) {
	s.httpClient = e.GetClient()
	s.table = e.GetTable()
	s.version = e.GetVersion()
	s.constantScore = true
	s.searchDo = new(searchDo)
}

// 获取单条记录 return 成功标记,是否找到标记
func (s *SearchSevice) Find(bean interface{}) (error, bool) {
	s.clearError()
	return s.searchDo.Find(s, bean)
}

// 返回查询语句
func (s *SearchSevice) GetQeuryBody() (string, error) {
	return s.searchDo.GetRequestBody(s)
}

// 是否存在 返回 成功标记 是否存在标记
func (s *SearchSevice) IsExist() (error, bool) {
	s.clearError()
	return s.searchDo.isExist(s)
}

// 是否存在 返回 成功标记 数量
func (s *SearchSevice) Count() (error, int64) {
	s.clearError()
	return s.searchDo.Count(s)
}

// 是否成功  统计信息
func (s *SearchSevice) Rows(bean interface{}) (error, *SearchRows) {
	s.clearError()
	return s.searchDo.Rows(s, bean)
}

// 是否成功  统计信息
func (s *SearchSevice) GroupRows() (error, *SearchResultGroupBy) {
	s.clearError()
	return s.searchDo.GroupRows(s)
}

// 是否成功  批量查询游标 t 时间游标 1m 5m 10m
func (s *SearchSevice) Batch(t string) *ResultScroll {
	s.clearError()
	if t == "" {
		t = "5m"
	}
	s.scroll = t
	return s.searchDo.Batch(s)
}

// 汇总方法 返回 SearchAggeStats
func (s *SearchSevice) Stats(f string) (error, *SearchAggeStats) {
	s.clearError()
	return s.searchDo.Stats(s, f)
}
func (s *SearchSevice) RequestApi(method string, url string, body string) ([]byte, error) {
	s.clearError()
	return s.searchDo.RequestApi(s, method, url, body)
}

// 汇总方法 返回 SearchResultGroupBy
func (s *SearchSevice) GroupBy(a *AggeSearch) *SearchSevice {
	s.agge.AddAgge(a)
	return s
}

// 汇总方法 返回 SearchResultGroupBy
func (s *SearchSevice) GroupBodyBy(a *map[string]interface{}) *SearchSevice {
	s.agge.setBody(a)
	return s
}
func (s *SearchSevice) WhereM(and []*ConnAnd) *SearchSevice {
	e := s.where.WhereM(and)
	s.setError(e)
	return s
}
func (s *SearchSevice) OrWhereM(and []*ConnAnd) *SearchSevice {
	e := s.where.OrWhereM(and)
	s.setError(e)
	return s
}
func (s *SearchSevice) NotWhereM(and []*ConnAnd) *SearchSevice {
	e := s.where.NotWhereM(and)
	s.setError(e)
	return s
}
func (s *SearchSevice) GetNewWhere() *ConnAnd {
	return new(ConnAnd)
}

func (s *SearchSevice) Where(f string, v *EsFields) *SearchSevice {
	if f == s.routing {
		if v.DataType == "string" {
			s.routing = f + "_" + v.StringData
		} else {
			s.routing = f + "_" + strconv.Itoa(v.IntData)
		}
		e := s.where.Where(s.context, "_routing", s.FieldString(s.routing))
		if e != nil {
			logger.AddWarn(s.context, zap.Error(e))
		}
	}
	e := s.where.Where(s.context, f, v)
	s.setError(e)
	return s
}
func (s *SearchSevice) WhereGeo(v *EsFields) *SearchSevice {
	e := s.where.WhereGeo(v)
	s.setError(e)
	return s
}
func (s *SearchSevice) WhereIn(f string, v *EsFields) *SearchSevice {
	e := s.where.WhereIn(s.context, f, v)
	s.setError(e)
	return s
}
func (s *SearchSevice) NotWhere(f string, v *EsFields) *SearchSevice {
	e := s.where.NotWhere(s.context, f, v)
	s.setError(e)
	return s
}
func (s *SearchSevice) NotWhereIn(f string, v *EsFields) *SearchSevice {
	e := s.where.NotWhereIn(s.context, f, v)
	s.setError(e)
	return s
}
func (s *SearchSevice) WhereMatch(v *Matchs) *SearchSevice {
	e := s.match.Where(s.context, v)
	s.setError(e)
	return s
}
func (s *SearchSevice) NotWhereMatch(v *Matchs) *SearchSevice {
	e := s.match.NotWhere(s.context, v)
	s.setError(e)
	return s
}
func (s *SearchSevice) OrWhereMatchM(and []*ConnMatchAnd) *SearchSevice {
	e := s.match.OrWhereM(and)
	s.setError(e)
	return s
}
func (s *SearchSevice) WhereMatchM(and []*ConnMatchAnd) *SearchSevice {
	e := s.match.OrWhereM(and)
	s.setError(e)
	return s
}
func (s *SearchSevice) GetNewWhereMatch() *ConnMatchAnd {
	return new(ConnMatchAnd)
}
func (s *SearchSevice) Between(f string, min int, max int) *SearchSevice {
	s.where.Between(f, min, max)

	return s
}
func (s *SearchSevice) Compare(f string, compare string, v int) *SearchSevice {
	e := s.where.Compare(f, compare, v)
	s.setError(e)
	return s
}
func (s *SearchSevice) Cols(t []string) *SearchSevice {
	s.source = t
	return s
}
func (s *SearchSevice) LimitF(from int, size int) *SearchSevice {
	s.from, s.size = from, size
	return s
}
func (s *SearchSevice) TimeoutF(t time.Duration) *SearchSevice {
	s.timeOut = t
	return s
}
func (s *SearchSevice) FromF(from int) *SearchSevice {
	s.from = from
	return s
}
func (s *SearchSevice) SizeF(size int) *SearchSevice {
	s.size = size
	return s
}
func (s *SearchSevice) SourceF(v []string) *SearchSevice {
	s.source = v
	return s
}

func (s *SearchSevice) TrackTotalHitsF(v uint64) *SearchSevice {
	s.trackTotalHits = v
	return s
}
func (s *SearchSevice) FieldInt(value int) *EsFields {
	e := new(EsFields)
	e.setType("int")
	e.SetIntValue(value)
	return e
}
func (s *SearchSevice) FieldString(value string) *EsFields {
	e := new(EsFields)
	e.setType("string")
	e.SetStringValue(value)
	return e
}
func (s *SearchSevice) FieldIntArray(value []int) *EsFields {
	e := new(EsFields)
	e.setType("intArray")
	e.SetIntValueForArray(value)
	return e
}
func (s *SearchSevice) FieldStringArray(value []string) *EsFields {
	e := new(EsFields)
	e.setType("stringArray")
	e.SetStringValueForArray(value)
	return e
}

// 汇总对象
func (s *SearchSevice) FieldGroupBy(f string, a string) *AggeSearch {
	e := &AggeSearch{
		Field: f,
		Alias: a,
		Hint:  "map",
	}
	return e
}
func (s *SearchSevice) FieldGeo(f string, r int, lat float32, lon float32) *EsFields {
	e := new(EsFields)
	e.setType("geo")
	e.SetGeo(FilterGeo{
		Field: f,
		Lat:   lat,
		Lon:   lon,
		Unit:  strconv.Itoa(r) + "km",
	})
	return e
}
func (s *SearchSevice) FieldMatch(f string, v string) *Matchs {
	return &Matchs{
		Field:  []string{f},
		Values: v,
	}
}
func (s *SearchSevice) FieldMatchPrefix(f string, v string) *Matchs {
	return &Matchs{
		Field:  []string{f},
		Values: v,
		Type:   "prefix",
	}
}

// 增加脚本排序
func (s *SearchSevice) AddScriptSort(script string, parma map[string]interface{}, sort string) *SearchSevice {
	s.sort.AddScriptSort(SortScript{
		Script: SortScriptDetails{Source: script,
			Lang:   "painless",
			Params: parma},
		Order: sort,
	})
	return s
}

// 增加字段排序
func (s *SearchSevice) AddFieldSort(f string, sort string) *SearchSevice {
	s.sort.AddFieldSort(SortField{
		Field: f,
		Order: sort,
	})
	return s
}

// 增加坐标排序
func (s *SearchSevice) AddGeoSort(f string, lat float32, lon float32, sort string) *SearchSevice {
	s.sort.AddGeoSort(SortGeo{
		Field:        f,
		Lat:          lat,
		Lon:          lon,
		Order:        sort,
		Unit:         "km",
		DistanceType: "plane",
	})
	return s
}

// 是否获取 其它_score,routing,sort 数据
func (s *SearchSevice) IsGetOtherSearchInfo(v bool) *SearchSevice {
	s.getOtherInfo = v
	return s
}
func (s *SearchSevice) setError(e error) {
	if e != nil {
		s.IsError = true
		s.ErrorDetails = e
	}
}
func (s *SearchSevice) clearError() {
	s.IsError = false
	s.ErrorDetails = nil
	s.searchDo.clearError()
}
