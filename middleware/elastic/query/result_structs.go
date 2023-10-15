package query

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
)

// Es 返回的Result
type SearchResult struct {
	Header          fasthttp.ResponseHeader `json:"-"`
	TookInMillis    int64                   `json:"took,omitempty"`
	TerminatedEarly bool                    `json:"terminated_early,omitempty"`
	NumReducePhases int                     `json:"num_reduce_phases,omitempty"`
	ScrollId        string                  `json:"_scroll_id,omitempty"`
	Hits            *SearchHits             `json:"hits,omitempty"`
	Aggregations    Aggregations            `json:"aggregations,omitempty"` // results from aggregations
	TimedOut        bool                    `json:"timed_out,omitempty"`
	LastDsl         string                  `json:"last_send"`
}

type SearchResultStats struct {
	stats SearchAggeStats
}
type SearchAggeStats struct {
	Count int64   `json:"count"`
	Min   float32 `json:"min"`
	Max   float32 `json:"max"`
	Avg   float32 `json:"avg"`
	Sum   float32 `json:"sum"`
}

// 返回给业务方的Result
type ListResult struct {
	List    interface{} `json:"list,omitempty"`
	HasMore byte        `json:"has_more,omitempty"`
	Total   int         `json:"total,omitempty"`
}

// 返回的请求信息,默认不带
type RequestInfo struct {
	TookInMillis int64  `json:"took,omitempty"` // search time in milliseconds
	LastDsl      string `json:"last_dsl"`
}
type SearchHits struct {
	TotalHits *TotalHits   `json:"total,omitempty"`     // total number of hits found
	MaxScore  *float64     `json:"max_score,omitempty"` // maximum score of all hits
	Hits      []*SearchHit `json:"hits,omitempty"`      // the actual hits returned
}

// TotalHits specifies total number of hits and its relation
type TotalHits struct {
	Value    int64  `json:"value"`    // value of the total hit count
	Relation string `json:"relation"` // how the value should be interpreted: accurate ("eq") or a lower bound ("gte")
}

// TotalHits 总量
func (r *SearchResult) TotalHits() int64 {
	if r != nil && r.Hits != nil && r.Hits.TotalHits != nil {
		return r.Hits.TotalHits.Value
	}
	return 0
}

type Aggregations map[string]json.RawMessage

// SearchHit is a single hit.
type SearchHit struct {
	Routing string          `json:"_routing,omitempty"` // routing meta field
	Id      string          `json:"_id,omitempty"`
	Score   float64         `json:"_score,omitempty"`
	Source  json.RawMessage `json:"_source,omitempty"` // stored document source
}

// SearchHitFields helps to simplify resolving slices of specific types.
type SearchHitFields map[string]interface{}

// SearchHitInnerHits is used for inner hits.
type SearchHitInnerHits struct {
	Hits *SearchHits `json:"hits,omitempty"`
}
type SearchRows struct {
	Total     int64             `json:"total"`
	HasMore   bool              `json:"has_more"`
	OtherInfo []SearchOtherInfo `json:"other_info"`
}
type SearchOtherInfo struct {
	Score   *float64      `json:"_score,omitempty"`   // computed score
	Id      string        `json:"_id,omitempty"`      // external or internal
	Routing string        `json:"_routing,omitempty"` // routing meta field
	Sort    []interface{} `json:"sort,omitempty"`     // sort information
}
type SearchNodes struct {
	NodesT struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Failed     int `json:"failed"`
	} `json:"_nodes"`
	ClusterName string `json:"cluster_name"`
	Nodes       map[string]struct {
		Name string `json:"name"`
		Host string `json:"host"`
		Ip   string `json:"ip"`
		Http struct {
			BoundAddress            []string `json:"bound_address"`
			PublishAddress          string   `json:"publish_address"`
			MaxContentLengthInBytes int      `json:"max_content_length_in_bytes"`
		} `json:"http"`
	} `json:"nodes"`
}
