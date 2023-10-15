package query

import (
	"encoding/json"
	"github.com/flyerxp/lib/logger"
	"go.uber.org/zap"
	"io"
)

type ResultScroll struct {
	searchService *SearchSevice
	Scroll        string `json:"scroll"`
	ScrollId      string `json:"_scroll_id"`
	Took          int    `json:"took"`
	Hits          struct {
		Hits []struct {
			Index   string          `json:"_index"`
			Type    string          `json:"_type"`
			Id      string          `json:"_id"`
			Score   float32         `json:"_score"`
			Routing string          `json:"_routing"`
			Source  json.RawMessage `json:"_source"`
		} `json:"hits"`
	}
	IsEnd bool
}
type SearchScroll struct {
	Scroll   string `json:"scroll"`
	ScrollId string `json:"_scroll_id"`
}

func (s *ResultScroll) Next() (error, *ResultScroll) {
	ss := new(SearchScroll)
	ss.Scroll = s.Scroll
	if s.IsEnd {
		return io.EOF, new(ResultScroll)
	}
	if s.ScrollId == "" {
		s.searchService.trackTotalHits = 0
		e, resp := s.searchService.searchDo.BatchSearch(s.searchService)
		if e == nil {
			s.ScrollId = resp.ScrollId
			return nil, resp
		} else {
			return e, nil
		}
	} else {
		e, resp := s.searchService.searchDo.BatchByScroll(s.searchService, s.ScrollId)
		if len(resp.Hits.Hits) == 0 {
			resp.IsEnd = true
			e = s.searchService.searchDo.DeleteScroll(s.searchService, s.ScrollId)
			if e != nil {
				logger.AddWarn(s.searchService.context, zap.Error(e))
			}
			return io.EOF, resp
		}
		if e == nil {
			s.ScrollId = resp.ScrollId
			return nil, resp
		} else {
			s.IsEnd = true
			return e, nil
		}

	}
}
