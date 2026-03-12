package pinyin

import (
	"strings"
	"unicode"

	"github.com/mozillazg/go-pinyin"
)

// GenerateCode 根据名称生成枚举代码
// 支持中文、英文、数字的混合输入
// 规则：
//   - 中文字符转换为拼音（如：无人机 -> WU_REN_JI）
//   - 英文字母和数字作为整体保留并转大写（如：V2 -> V2，Pro -> PRO）
//   - 其他字符（空格、标点、小数点等）作为分隔符
//
// 示例:
//   - "无人机" -> "TYPE_WU_REN_JI"
//   - "无人机V2" -> "TYPE_WU_REN_JI_V2"
//   - "地面站Pro" -> "TYPE_DI_MIAN_ZHAN_PRO"
//   - "APP 1.0版本" -> "TYPE_APP_1_0_BAN_BEN"
//   - "无人机-专业版V2.1" -> "TYPE_WU_REN_JI_ZHUAN_YE_BAN_V2_1"
func GenerateCode(name string) string {
	var segments []string
	var chineseBuf, alphanumericBuf strings.Builder

	flushChinese := func() {
		if chineseBuf.Len() > 0 {
			py := convertToPinyin(chineseBuf.String())
			if py != "" {
				segments = append(segments, py)
			}
			chineseBuf.Reset()
		}
	}

	flushAlphanumeric := func() {
		if alphanumericBuf.Len() > 0 {
			segments = append(segments, strings.ToUpper(alphanumericBuf.String()))
			alphanumericBuf.Reset()
		}
	}

	for _, r := range name {
		if unicode.Is(unicode.Han, r) {
			flushAlphanumeric()
			chineseBuf.WriteRune(r)
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			flushChinese()
			alphanumericBuf.WriteRune(r)
		} else {
			flushChinese()
			flushAlphanumeric()
		}
	}

	flushChinese()
	flushAlphanumeric()

	if len(segments) == 0 {
		return "TYPE_UNNAMED"
	}

	result := strings.Join(segments, "_")
	return "TYPE_" + result
}

func convertToPinyin(chinese string) string {
	a := pinyin.NewArgs()
	a.Style = pinyin.Normal
	a.Heteronym = false

	py := pinyin.LazyPinyin(chinese, a)
	return strings.ToUpper(strings.Join(py, "_"))
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
