package query

type Between struct {
	field string
	min   int
	max   int
}
type Gt struct {
	field string
	value int
}
type Gte struct {
	field string
	value int
}
type Lt struct {
	field string
	value int
}
type Lte struct {
	field string
	value int
}

func (t *Between) GetSourceMap() (interface{}, error) {
	m := map[string]interface{}{
		"range": map[string]interface{}{
			t.field: map[string]interface{}{
				"gte": t.min,
				"lte": t.max,
			},
		},
	}
	return m, nil
}
func (t *Lt) GetSourceMap() (interface{}, error) {
	m := map[string]interface{}{
		"range": map[string]interface{}{
			t.field: map[string]interface{}{
				"lt": t.value,
			},
		},
	}
	return m, nil
}
func (t *Lte) GetSourceMap() (interface{}, error) {
	m := map[string]interface{}{
		"range": map[string]interface{}{
			t.field: map[string]interface{}{
				"lte": t.value,
			},
		},
	}
	return m, nil
}
func (t *Gt) GetSourceMap() (interface{}, error) {
	m := map[string]interface{}{
		"range": map[string]interface{}{
			t.field: map[string]interface{}{
				"gt": t.value,
			},
		},
	}
	return m, nil
}
func (t *Gte) GetSourceMap() (interface{}, error) {
	m := map[string]interface{}{
		"range": map[string]interface{}{
			t.field: map[string]interface{}{
				"gte": t.value,
			},
		},
	}
	return m, nil
}
