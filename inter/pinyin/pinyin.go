package pinyin

import (
	"bytes"
	"github.com/Lofanmi/pinyin-golang/pinyin"
	"strings"
)

var dict = pinyin.NewDict()

// 转换拼音字符串
func Convert(s string) string {
	return dict.Convert(s, " ").None()
}

// 转换拼音字符串切片
func ConvertSlice(s string) []string {
	str := dict.Convert(s, " ").None()
	return pinyin.ToSlice(str)
}

// 转换拼音字符串 - 首字母
func ConvertFirstLetter(s string) string {
	return dict.Abbr(s, "")
}

// 转换拼音字符串切片 - 首字母
func ConvertFirstLetterSlice(s string) []string {
	str := dict.Abbr(s, " ")
	return pinyin.ToSlice(str)
}

// 转换拼音字符串切片 - 姓名
func ConvertName(s string) string {
	return dict.Name(s, " ").None()
}

// 转换拼音字符串切片 - 姓名
func ConvertNameSlice(s string) []string {
	str := dict.Name(s, " ").None()
	return pinyin.ToSlice(str)
}

// 获取首字母
func FormatSliceFirstLetter(s []string) string {
	total := len(s)
	buf := new(bytes.Buffer)
	for i := 0; i < total; i++ {
		buf.WriteString(s[i][:1])
	}
	return buf.String()
}

// 字符串切片转小写
func FormatSliceToLower(s []string) []string {
	total := len(s)
	for i := 0; i < total; i++ {
		s[i] = strings.ToLower(s[i])
	}
	return s
}
