package query

import "strings"

type SortField struct {
	Field string `json:"field"`
	Order string `json:"order"`
}

type SortGeo struct {
	Field        string  `json:"field"`
	Lat          float32 `json:"lat"`
	Lon          float32 `json:"lon"`
	Order        string  `json:"order"`
	Unit         string  `json:"unit"`
	DistanceType string  `json:"distance_type"`
}

type SortScript struct {
	Script SortScriptDetails `json:"script"`
	Type   string            `json:"type"`
	Order  string            `json:"order"`
}
type SortScriptDetails struct {
	Source string                 `json:"source"`
	Lang   string                 `json:"lang"`
	Params map[string]interface{} `json:"params"`
}

func (t *SortField) GetSourceMap() map[string]map[string]interface{} {
	if t.Order != "desc" && t.Order != "asc" {
		t.Order = "desc"
	}
	return map[string]map[string]interface{}{
		t.Field: {
			"order": t.Order,
		},
	}
}
func (t *SortGeo) GetSourceMap() map[string]map[string]interface{} {
	if t.Lat < -90 || t.Lat > 90 {
		t.Lat = 0
	}
	if t.Lon < -180 || t.Lon > 180 {
		t.Lon = 0
	}
	return map[string]map[string]interface{}{
		"_geo_distance": {
			t.Field: map[string]float32{
				"lat": t.Lat,
				"lon": t.Lon,
			},
			"order":         t.Order,
			"unit":          t.Unit,
			"distance_type": t.DistanceType,
		},
	}
}
func (t *SortScript) GetSourceMap() map[string]map[string]interface{} {
	if t.Order != "desc" && t.Order != "asc" {
		t.Order = "desc"
	}
	if t.Type == "" {
		t.Type = "number"
	}
	t.Script.Source = strings.Replace(t.Script.Source, "\t", " ", -1)
	t.Script.Source = strings.Replace(t.Script.Source, "\n", " ", -1)
	t.Script.Source = strings.Replace(t.Script.Source, "\r", " ", -1)
	return map[string]map[string]interface{}{
		"_script": {
			"script": t.Script,
			"type":   t.Type,
			"order":  t.Order,
		},
	}
}
