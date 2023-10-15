package query

type EsFields struct {
	DataType        string
	StringData      string
	IntData         int
	GeoData         FilterGeo
	StringArrayData []string
	IntArrayData    []int
}

func (e *EsFields) setType(v string) {
	e.DataType = v
	return
}
func (e *EsFields) SetIntValue(v int) {
	e.IntData = v
	return
}
func (e *EsFields) SetGeo(v FilterGeo) {
	e.GeoData = v
}
func (e *EsFields) SetStringValue(v string) {

	e.StringData = v

}
func (e *EsFields) SetStringValueForArray(v []string) {
	e.StringArrayData = v
}
func (e *EsFields) SetIntValueForArray(v []int) {
	e.IntArrayData = v
}
func (e *EsFields) AppendStringValue(v string) {
	e.StringArrayData = append(e.StringArrayData, v)
}
func (e *EsFields) AppendIntValue(v int) {
	e.IntArrayData = append(e.IntArrayData, v)
}
func (e *EsFields) getData() interface{} {
	switch e.DataType {
	case "int":
		return e.IntData
	case "string":
		return e.StringData
	}
	return nil
}
func (e *EsFields) getGeoData() interface{} {
	g, _ := e.GeoData.GetSourceMap()
	return g
}
func (e *EsFields) getArrayData() []interface{} {
	switch e.DataType {
	case "intArray":
		r := []interface{}{}
		for _, t := range e.IntArrayData {
			r = append(r, t)
		}
		return r
	case "stringArray":
		r := []interface{}{}
		for _, t := range e.StringArrayData {
			r = append(r, t)
		}
		return r
	}
	return nil
}
