package query

type Sort struct {
	Sort []map[string]map[string]interface{} `json:"sort,omitempty"`
}

func (s *Sort) GetSourceMap() interface{} {
	return s.Sort
}
func (s *Sort) AddScriptSort(ss SortScript) {
	s.Sort = append(s.Sort, ss.GetSourceMap())
}
func (s *Sort) AddFieldSort(ss SortField) {
	s.Sort = append(s.Sort, ss.GetSourceMap())
}
func (s *Sort) AddGeoSort(ss SortGeo) {
	s.Sort = append(s.Sort, ss.GetSourceMap())
}
