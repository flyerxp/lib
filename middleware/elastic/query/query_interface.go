package query

// 获取所有查询最终代码
type Query interface {
	GetSourceMap() (interface{}, error)
}
