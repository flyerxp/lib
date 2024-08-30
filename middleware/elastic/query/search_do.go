package query

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/flyerxp/lib/v2/logger"
	"github.com/flyerxp/lib/v2/middleware/elastic/tool"
	"github.com/valyala/fasthttp"
	"strconv"
	"strings"
	"time"
)

type searchDo struct {
	context      context.Context
	isError      bool
	ErrorDetails error
}

var decodeTool = tool.DefaultDecoder{}

func (sd *searchDo) GetRequestBody(s *SearchSevice) (string, error) {
	if s.where.isError {
		sd.setError(s, s.where.ErrorDetails)
		return "", sd.ErrorDetails
	} else {
		if s.size > 10000 {
			s.size = 9999
		}
		maxFrom := s.table.MaxWindowResult
		if (s.from + s.size) > maxFrom {
			e := errors.New("es max windows scrool size" + strconv.Itoa(maxFrom) + " now " + s.table.Name + " offset " + strconv.Itoa(s.from) + " size " + strconv.Itoa(s.size))
			sd.setError(s, e)
			return "", e
		}
		body := map[string]interface{}{
			"from": s.from,
			"size": s.size,
		}

		if s.scroll != "" {
			delete(body, "from")
		}
		if len(s.where.And) > 0 || len(s.match.And) > 0 {
			body["query"] = sd.getQuery(s)
		}
		if s.source != nil {
			body["_source"] = s.source
		}
		if len(s.sort.Sort) > 0 {
			body["sort"] = s.sort.GetSourceMap()
		}
		if s.trackTotalHits > 0 {
			body["track_total_hits"] = s.trackTotalHits
		}
		if len(s.agge.Agge) > 0 {
			for _, v := range s.agge.Agge {
				v.size = s.size
				v.From = s.from
			}
			body["aggregations"] = s.agge.GetSourceMap()
			body["size"] = 0
			body["from"] = 0
		}
		t, e := decodeTool.EnCode(body)
		if e != nil {
			sd.setError(s, e)
		}
		return string(t), nil
	}
}
func (sd *searchDo) getQuery(s *SearchSevice) map[string]interface{} {
	tmpM := s.where.GetSourceMap()
	if len(s.match.And) > 0 {
		RMatch := map[string]interface{}{}
		tmpMatch := s.match.GetSourceMap()
		for _, v := range tmpMatch {
			for strI, vv := range v["bool"] {
				RMatch[strI] = vv
			}

		}
		if !(s.where.And == nil || len(s.where.And) == 0) {
			tFilter := map[string]interface{}{
				"bool": map[string][]map[string]map[string][]interface{}{
					"must": tmpM,
				},
			}
			RMatch["filter"] = tFilter
		}
		return map[string]interface{}{
			"bool": RMatch,
		}
	} else if s.constantScore {
		return map[string]interface{}{
			"constant_score": map[string]interface{}{
				"filter": map[string]interface{}{
					"bool": map[string][]map[string]map[string][]interface{}{
						"must": tmpM,
					},
				},
			},
		}
	} else {
		return map[string]interface{}{
			"bool": map[string][]map[string]map[string][]interface{}{
				"must": tmpM,
			},
		}
	}
}

// 返回 错误 是否找到了
func (sd *searchDo) Find(s *SearchSevice, bean interface{}) (error, bool) {
	sd.clearError()
	s.size = 1
	r, e := sd.Do(s)

	if e == nil {
		resp := new(SearchResult)
		e = decodeTool.Decode(r, resp)
		if e == nil {
			if resp.TotalHits() < 1 {
				return nil, false
			}
			e = decodeTool.Decode(resp.Hits.Hits[0].Source, bean)
			if e != nil {
				sd.setError(s, e)
				return e, true
			}
			return nil, true
		} else {
			//解析出错
			sd.setError(s, e)
			return e, false
		}
	}
	sd.setError(s, e)
	return e, false
}

// 是否存在 返回 错误 是否存在标记
func (sd *searchDo) isExist(s *SearchSevice) (error, bool) {
	sd.clearError()
	s.size = 0
	r, e := sd.Do(s)
	if e == nil {
		resp := new(SearchResult)
		e = decodeTool.Decode(r, resp)
		if e == nil {
			if resp.TotalHits() >= 1 {
				return nil, true
			}
		} else {
			//解析出错
			sd.setError(s, e)
			return e, false
		}
	}
	//查询出错
	sd.setError(s, e)
	return e, false
}

// 是否存在 返回 成功标记 数量
func (sd *searchDo) Count(s *SearchSevice) (error, int64) {
	sd.clearError()
	s.size = 0
	r, e := sd.Do(s)
	if e == nil {
		resp := new(SearchResult)
		e = decodeTool.Decode(r, resp)
		if e == nil {
			return nil, resp.TotalHits()
		} else {
			//解析出错
			sd.setError(s, e)
			return e, 0
		}
	}
	//查询出错
	sd.setError(s, e)
	return e, 0
}

// 获取列表
func (sd *searchDo) Rows(s *SearchSevice, bean interface{}) (error, *SearchRows) {
	sd.clearError()
	if s.size >= 10000 {
		s.size = 10000
	}
	r, e := sd.Do(s)
	if e == nil {
		resp := new(SearchResult)
		e = decodeTool.Decode(r, &resp)
		if e == nil {
			total := resp.TotalHits()
			hasmore := false
			if total < 1 {
				return nil, &SearchRows{0, false, []SearchOtherInfo{}}
			}
			if total > int64(s.from+s.size) {
				hasmore = true
			}
			tmpRows := make([]json.RawMessage, len(resp.Hits.Hits))
			tmpOtherSearchInfo := make([]SearchOtherInfo, len(resp.Hits.Hits))
			for i := range resp.Hits.Hits {
				tmpRows[i] = resp.Hits.Hits[i].Source
				if s.getOtherInfo {
					tmpOtherSearchInfo[i] = SearchOtherInfo{
						Score:   &resp.Hits.Hits[i].Score,
						Id:      resp.Hits.Hits[i].Id,
						Routing: resp.Hits.Hits[i].Routing,
					}
				}
			}
			jsonStr, _ := json.Marshal(tmpRows)
			e = json.Unmarshal(jsonStr, &bean)
			if e != nil {
				sd.setError(s, e)
				return e, &SearchRows{resp.TotalHits(), hasmore, []SearchOtherInfo{}}
			}
			return nil, &SearchRows{resp.TotalHits(), hasmore, tmpOtherSearchInfo}
		} else {
			//解析出错
			sd.setError(s, e)
			return e, &SearchRows{0, false, []SearchOtherInfo{}}
		}
	}
	sd.setError(s, e)
	return e, &SearchRows{0, false, []SearchOtherInfo{}}
}

// 获取统计信息
func (sd *searchDo) GroupRows(s *SearchSevice) (error, *SearchResultGroupBy) {
	sd.clearError()
	r, e := sd.Do(s)
	if e == nil {
		resp := new(SearchResultGroupBy)
		e = decodeTool.Decode(r, &resp)
		if e == nil {
			return nil, resp
		} else {
			//解析出错
			sd.setError(s, e)
			return e, new(SearchResultGroupBy)
		}
	}
	sd.setError(s, e)
	return e, new(SearchResultGroupBy)
}

// 获取统计信息
func (sd *searchDo) Stats(s *SearchSevice, f string) (error, *SearchAggeStats) {
	sd.clearError()
	s.size = 0
	a := &AggeSearch{Field: f, Alias: "stats", CalChar: "stats"}
	s.agge.AddAgge(a)
	r, e := sd.Do(s)
	if e == nil {
		resp := new(SearchResult)
		e = decodeTool.Decode(r, &resp)
		if e == nil {
			t := new(SearchAggeStats)
			e = decodeTool.Decode(resp.Aggregations["stats"], t)
			if e == nil {
				return nil, t
			} else {
				sd.setError(s, e)
			}
		} else {
			//解析出错
			sd.setError(s, e)
			return e, new(SearchAggeStats)
		}
	}
	sd.setError(s, e)
	return e, new(SearchAggeStats)
}

// 批量查询，返回游标
func (sd *searchDo) Batch(s *SearchSevice) *ResultScroll {
	if s.size >= 10000 {
		s.size = 10000
	}
	resp := new(ResultScroll)
	resp.searchService = s
	return resp
}

// 批量查询，返回游标
func (sd *searchDo) BatchSearch(s *SearchSevice) (error, *ResultScroll) {
	sd.clearError()
	if s.size >= 10000 {
		s.size = 10000
	}
	r, e := sd.Do(s)
	resp := new(ResultScroll)
	resp.searchService = s
	if e == nil {
		e = decodeTool.Decode(r, &resp)
		if e == nil {
			return nil, resp
		} else {
			//解析出错
			sd.setError(s, e)
			return e, resp
		}
	}
	sd.setError(s, e)
	return e, resp
}

// 批量查询，返回游标
func (sd *searchDo) DeleteScroll(s *SearchSevice, scrollId string) error {
	sd.clearError()
	body := "{\"scroll_id\":\"" + scrollId + "\"}"
	_, e := s.httpClient.SendRequest(s.context, fasthttp.MethodDelete, "/_search/scroll", body, s.timeOut, 0, true)
	return e
}
func (sd *searchDo) BatchByScroll(s *SearchSevice, scrollId string) (error, *ResultScroll) {
	start := time.Now()
	r, e := s.httpClient.SendRequest(s.context, fasthttp.MethodGet, sd.getScrollUrl(s, scrollId), "", s.timeOut, 0, true)
	tTime := time.Since(start).Microseconds()
	logger.AddEsTime(s.context, int(tTime))
	resp := new(ResultScroll)
	resp.searchService = s
	if e == nil {
		e = decodeTool.Decode(r, &resp)
		if e == nil {
			return nil, resp
		} else {
			//解析出错
			sd.setError(s, e)
			return e, resp
		}
	}
	sd.setError(s, e)
	return e, resp
}
func (sd *searchDo) Do(s *SearchSevice) ([]byte, error) {
	body, _ := sd.GetRequestBody(s)

	if sd.isError {
		s.setError(sd.ErrorDetails)
		return []byte{}, sd.ErrorDetails
	}

	s.Dsl = body
	start := time.Now()
	r, e := s.httpClient.SendRequest(s.context, fasthttp.MethodPost, sd.getUrl(s), body, s.timeOut, 0, true)
	tTime := time.Since(start).Microseconds()
	logger.AddEsTime(s.context, int(tTime))
	return r, e
}
func (sd *searchDo) RequestApi(s *SearchSevice, method string, url string, body string) ([]byte, error) {
	s.Dsl = body
	start := time.Now()
	r, e := s.httpClient.SendRequest(s.context, method, url, body, s.timeOut, 0, true)
	tTime := time.Since(start).Microseconds()
	logger.AddEsTime(s.context, int(tTime))
	return r, e
}
func (sd *searchDo) ContextF(ctx context.Context) {
	sd.context = ctx
}
func (sd *searchDo) getScrollUrl(s *SearchSevice, scrollId string) string {
	if s.table.Name == "" {
		sd.setError(s, errors.New("unkow table name"))
		return "/"
	}
	url := "/_search/scroll"
	qs := make([]string, 0)
	qs = append(qs, "scroll_id="+scrollId)
	qs = append(qs, "scroll="+s.scroll)
	if len(qs) > 0 {
		url += "?" + strings.Join(qs, "&")
	}
	return url
}
func (sd *searchDo) getUrl(s *SearchSevice) string {
	if s.table.Name == "" {
		sd.setError(s, errors.New("unkow table name"))
		return "/"
	}
	url := "/" + s.table.Name + "/_search"
	qs := make([]string, 0)
	if s.routing != "" {
		qs = append(qs, "routing="+s.routing)
	}
	if s.scroll != "" {
		qs = append(qs, "scroll="+s.scroll)
	}

	if len(qs) > 0 {
		url += "?" + strings.Join(qs, "&")
	}
	return url
}
func (sd *searchDo) setError(s *SearchSevice, e error) {
	sd.isError = true
	sd.ErrorDetails = e
	s.setError(e)
}
func (sd *searchDo) clearError() {
	sd.isError = false
	sd.ErrorDetails = nil
}
