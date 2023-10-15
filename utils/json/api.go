package json

import (
	jsoniter "github.com/json-iterator/go"
)

type jsonTool struct {
	configDefault jsoniter.API
	isInitEd      bool
}

var (
	tool jsonTool
)

func init() {
	initTool()
}
func initTool() {
	if !tool.isInitEd {
		tool.configDefault = jsoniter.ConfigCompatibleWithStandardLibrary
		tool.isInitEd = true
	}
}

// 转为json
func Encode(val interface{}) ([]byte, error) {
	return tool.configDefault.Marshal(val)
}

// 转为byte
func Decode(data []byte, v interface{}) error {
	return tool.configDefault.Unmarshal(data, v)
}
