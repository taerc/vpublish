package pinyin

import (
	"strings"

	"github.com/mozillazg/go-pinyin"
)

// GenerateCode 根据中文名称生成枚举代码
// 输入: "无人机" -> 输出: "TYPE_WU_REN_JI"
// 输入: "地面站" -> 输出: "TYPE_DI_MIAN_ZHAN"
func GenerateCode(chineseName string) string {
	// 使用 Normal 风格，不带声调
	a := pinyin.NewArgs()
	a.Style = pinyin.Normal
	a.Heteronym = false // 不考虑多音字

	// 转换为拼音数组
	py := pinyin.LazyPinyin(chineseName, a)

	// 转大写并用下划线连接
	result := strings.ToUpper(strings.Join(py, "_"))

	// 添加前缀
	return "TYPE_" + result
}

// IsValidCode 检查代码是否有效
func IsValidCode(code string) bool {
	if !strings.HasPrefix(code, "TYPE_") {
		return false
	}
	// 检查是否全是大写字母、数字和下划线
	suffix := code[5:] // 去掉 "TYPE_" 前缀
	for _, c := range suffix {
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return len(suffix) > 0
}
