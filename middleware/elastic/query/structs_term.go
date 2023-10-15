package query

import (
	"errors"
	"reflect"
)

type Term struct {
	field string
	value interface{}
}
type Terms struct {
	field  string
	values []interface{}
}
type FilterGeo struct {
	Field string
	Lat   float32
	Lon   float32
	Unit  string
}

func (t *Term) GetSourceMap() (interface{}, error) {
	if t == nil {
		return nil, errors.New("term value error now is nil")
	}
	switch reflect.TypeOf(t.value).Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.String:
		return map[string]interface{}{
			"term": map[string]interface{}{
				t.field: t.value,
			},
		}, nil
	default:
		return nil, errors.New("term value error" + t.field)
	}
}
func (t *Terms) GetSourceMap() (interface{}, error) {
	if t == nil {
		return nil, errors.New("term value error now is nil")
	}
	switch reflect.TypeOf(t.values).Kind() {
	case reflect.Array, reflect.Slice:
		return map[string]interface{}{
			"terms": map[string]interface{}{
				t.field: t.values,
			},
		}, nil
	default:
		return nil, errors.New("terms value error" + t.field)
	}
}
func (t *FilterGeo) GetSourceMap() (interface{}, error) {
	if t == nil {
		return nil, errors.New("term value error now is nil")
	}
	return map[string]interface{}{
		"geo_distance": map[string]interface{}{
			"distance": t.Unit,
			t.Field: map[string]interface{}{
				"lat": t.Lat,
				"lon": t.Lon,
			},
		},
	}, nil
}
