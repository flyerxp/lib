package query

type Agge struct {
	Agge     []*AggeSearch `json:"sort,omitempty"`
	AggeBody *map[string]interface{}
}

func (a *Agge) GetSourceMap() map[string]interface{} {
	if a.AggeBody != nil {
		return *a.AggeBody
	}
	r := map[string]interface{}{}
	for _, v := range a.Agge {
		if v.Alias == "" {
			r[v.Field] = v.GetSourceMap()
		} else {
			r[v.Alias] = v.GetSourceMap()
		}
	}
	return r
}
func (a *Agge) AddAgge(b *AggeSearch) *Agge {
	if a.Agge == nil {
		a.Agge = make([]*AggeSearch, 0)
	}
	a.Agge = append(a.Agge, b)
	return a
}
func (a *Agge) setBody(b *map[string]interface{}) *Agge {
	a.AggeBody = b
	return a
}
