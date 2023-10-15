package result

import "encoding/json"

type ErrReturn struct {
	Took         int           `json:"took"`
	TimeOut      int           `json:"time_out"`
	Hits         SearchHits    `json:"hits"`
	Aggregations []interface{} `json:"aggregations"`
}

type SearchHits struct {
	TotalHits *TotalHits   `json:"total,omitempty"`     // total number of hits found
	MaxScore  *float64     `json:"max_score,omitempty"` // maximum score of all hits
	Hits      []*SearchHit `json:"hits,omitempty"`      // the actual hits returned
}

// SearchHit is a single hit.
type SearchHit struct {
	Routing string          `json:"_routing,omitempty"` // routing meta field
	Id      string          `json:"_id,omitempty"`
	Score   float64         `json:"_score,omitempty"`
	Source  json.RawMessage `json:"_source,omitempty"` // stored document source
}

// TotalHits specifies total number of hits and its relation
type TotalHits struct {
	Value    int64  `json:"value"`    // value of the total hit count
	Relation string `json:"relation"` // how the value should be interpreted: accurate ("eq") or a lower bound ("gte")
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
