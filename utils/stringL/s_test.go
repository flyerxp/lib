package stringL

import (
	"fmt"
	"testing"
)

func TestEncode(t *testing.T) {
	str := "  <title>内容管理 | 编辑内容</title>\n<meta charset=\"utf-8\">\n"
	fmt.Println(TrimHtml(str))
	fmt.Println(Substr("as中国星", 0, 4))
}
