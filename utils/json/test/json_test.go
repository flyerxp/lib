package json

import (
	json2 "github.com/flyerxp/lib/utils/json"
	"testing"
)

// 普通Encode测试
func TestEncode(t *testing.T) {
	data := map[string]string{"a": "<>asdf中国asdf%$", "b": "dddddddd<a href=\"\">asfdasdfasfd</a>"}
	jsons, _ := json2.Encode(&data)
	var data1 map[string]string
	e := json2.Decode(jsons, &data1)
	t.Logf("decode %#v %#v", data1, e)
	//json2.Encode(v, EscapeHTML)

}
