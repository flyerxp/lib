package stt

import "encoding/json"

// SearchHit is a single hit.
type SearchHit struct {
	Routing string          `json:"_routing,omitempty"` // routing meta field
	Id      string          `json:"_id,omitempty"`
	Score   float64         `json:"_score,omitempty"`
	Source  json.RawMessage `json:"_source,omitempty"` // stored document source
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
