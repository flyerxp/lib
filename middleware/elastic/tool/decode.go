package tool

import json2 "encoding/json"

type Decoder interface {
	Encode(v interface{}) ([]byte, error)
	Decode(data []byte, v interface{}) error
}

type DefaultDecoder struct{}

// Decode json
func (u *DefaultDecoder) Decode(data []byte, v interface{}) error {
	return json2.Unmarshal(data, v)
}

// EnCode json
func (u *DefaultDecoder) EnCode(v interface{}) ([]byte, error) {
	return json2.Marshal(v)
}
