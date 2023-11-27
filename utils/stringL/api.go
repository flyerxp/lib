package stringL

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/russross/blackfriday"
	"regexp"
	"strconv"
	"strings"
)

func GetMd5(str string) string {
	md5New := md5.New()
	data := []byte(str)
	md5New.Write(data)
	//fmt.Println(str)
	v := hex.EncodeToString(md5New.Sum(nil))
	//fmt.Println(v)
	return v
}
func MarkdownToHtml(b []byte) string {
	htmlFlags := 0
	htmlFlags |= 4096
	renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")
	return string(blackfriday.Markdown(b, renderer, 0))
}
func TrimHtml(src string) string {
	//将HTML标签全转换成小写
	re, _ := regexp.Compile("<[\\S\\s]+?>")
	src = re.ReplaceAllStringFunc(src, strings.ToLower)
	//去除注释
	re, _ = regexp.Compile("<!--[\\S\\s` ]+?-->")
	src = re.ReplaceAllString(src, "")
	//去除STYLE
	re, _ = regexp.Compile("<style[\\S\\s]+?</style>")
	src = re.ReplaceAllString(src, "")
	//去除SCRIPT
	re, _ = regexp.Compile("<script[\\S\\s]+?</script>")
	src = re.ReplaceAllString(src, "")
	//去除所有尖括号内的HTML代码，并换成换行符
	re, _ = regexp.Compile("<[\\S\\s-]+?>")
	src = re.ReplaceAllString(src, "\n")
	//去除连续的换行符
	re, _ = regexp.Compile("\\s{2,}")
	src = re.ReplaceAllString(src, "\n")
	return strings.TrimSpace(src)
}
func Substr(s string, start int, len int) string {
	nameRune := []rune(s)
	return string(nameRune[start:len])
}

// 提取中文
func GetCn(str string) (cnStr string) {
	r := []rune(str)
	strSlice := []string{}
	for i := 0; i < len(r); i++ {
		if r[i] <= 40869 && r[i] >= 19968 {
			cnStr = cnStr + string(r[i])
			strSlice = append(strSlice, cnStr)
		}
	}
	return
}

// 从float 32 转到float
func Float32toFloat64(v float32) float64 {
	str := strconv.FormatFloat(float64(v), 'f', -1, 32)
	money64, _ := strconv.ParseFloat(str, 64)
	return money64
}
