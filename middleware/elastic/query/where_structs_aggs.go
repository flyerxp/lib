package query

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
)

// Es 返回的Result
type SearchResultGroupBy struct {
	Header       fasthttp.ResponseHeader `json:"-"`
	TookInMillis int64                   `json:"took,omitempty"`
	Aggregations json.RawMessage         `json:"aggregations,omitempty"` // results from aggregations
	TimedOut     bool                    `json:"timed_out,omitempty"`
	LastDsl      string                  `json:"last_send"`
}

// Agge 搜索条件
type AggeSearch struct {
	Field        string
	Alias        string
	size         int
	From         int
	Hint         string
	CalChar      string
	Order        []map[string]string //_key:asc
	Aggregations []*AggeSearchOrder
}

// Agge 搜索条件排序
type AggeSearchOrder struct {
	Field      string
	size       int
	Alias      string
	Cal        string
	Source     []string
	SortMethod string //desc || asc
	Sort       []map[string]map[string]string
}

func (a *AggeSearch) addSort(v *AggeSearchOrder) {
	a.Aggregations = append(a.Aggregations, v)
}
func (a *AggeSearchOrder) GetSourceMap() map[string]interface{} {
	t := map[string]interface{}{}
	if a.Cal == " top_hits" {
		t["size"] = a.size
	}
	if a.Field != "" {
		t["field"] = a.Field
	}
	if len(a.Source) > 0 {
		t["_source"] = a.Source
	}
	if len(a.Sort) > 0 {
		t["sort"] = a.Sort
	}
	return map[string]interface{}{
		a.Cal: t,
	}
}
func (a *AggeSearch) GetSourceMap() interface{} {
	if a.Alias == "" {
		a.Alias = a.Field
	}
	if a.CalChar != "" {
		return map[string]map[string]string{
			a.CalChar: {
				"field": a.Field,
			},
		}
	}

	aSort := make(map[string]interface{}, 0)
	if len(a.Aggregations) > 0 {
		for _, v := range a.Aggregations {
			if a.From == 0 {
				v.size = a.size
			} else {
				v.size = a.size + a.From + 10 //防止各分片合并后，最终结果不一致
			}

			if v.SortMethod != "desc" {
				v.SortMethod = "asc"
			}
			aSort[v.Alias] = v.GetSourceMap()
		}
	}
	r := map[string]map[string]interface{}{
		"terms": {
			"field":          a.Field,
			"size":           a.size,
			"execution_hint": a.Hint,
		},
		"aggregations": aSort,
	}
	if a.From > 0 && a.CalChar != "top_hits" {
		r["terms"]["size"] = a.From + a.size + 10
		aSort["bucket_truncate"] = map[string]interface{}{
			"bucket_sort": map[string]interface{}{
				"from": a.From,
				"size": a.size,
			},
		}
	}

	if a.CalChar == "top_hits" {
		tOrder := make([]map[string]map[string]string, 0)
		if a.Order != nil {
			for _, v := range a.Order {
				for k, value := range v {
					tOrder = append(tOrder, map[string]map[string]string{
						k: {
							"order": value,
						},
					})
				}
			}
			r["terms"]["sort"] = tOrder
		}
	} else {
		if a.Order != nil {
			r["terms"]["order"] = a.Order
		}
	}
	return r
}
func (a *AggeSearch) AddAggregations(o *AggeSearchOrder) *AggeSearch {
	a.Aggregations = append(a.Aggregations, o)
	return a
}
func (a *AggeSearch) AddOrder(o map[string]string) *AggeSearch {
	a.Order = append(a.Order, o)
	return a
}
