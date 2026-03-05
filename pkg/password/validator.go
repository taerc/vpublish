package password

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

var (
	// ErrPasswordTooShort 密码太短
	ErrPasswordTooShort = errors.New("密码长度至少需要8位")
	// ErrPasswordTooLong 密码太长
	ErrPasswordTooLong = errors.New("密码长度不能超过72位")
	// ErrPasswordNoUpper 密码缺少大写字母
	ErrPasswordNoUpper = errors.New("密码必须包含至少一个大写字母")
	// ErrPasswordNoLower 密码缺少小写字母
	ErrPasswordNoLower = errors.New("密码必须包含至少一个小写字母")
	// ErrPasswordNoDigit 密码缺少数字
	ErrPasswordNoDigit = errors.New("密码必须包含至少一个数字")
	// ErrPasswordNoSpecial 密码缺少特殊字符
	ErrPasswordNoSpecial = errors.New("密码必须包含至少一个特殊字符")
	// ErrPasswordTooWeak 密码过于简单
	ErrPasswordTooWeak = errors.New("密码过于简单，请使用更复杂的密码")
	// ErrPasswordCommon 密码是常见弱密码
	ErrPasswordCommon = errors.New("密码过于常见，请使用更复杂的密码")
)

// ValidationConfig 密码验证配置
type ValidationConfig struct {
	MinLength       int  // 最小长度，默认8
	MaxLength       int  // 最大长度，默认72（bcrypt限制）
	RequireUpper    bool // 是否需要大写字母
	RequireLower    bool // 是否需要小写字母
	RequireDigit    bool // 是否需要数字
	RequireSpecial  bool // 是否需要特殊字符
	CheckCommonList bool // 是否检查常见弱密码列表
}

// DefaultConfig 默认验证配置
var DefaultConfig = ValidationConfig{
	MinLength:       8,
	MaxLength:       72,
	RequireUpper:    true,
	RequireLower:    true,
	RequireDigit:    true,
	RequireSpecial:  true,
	CheckCommonList: true,
}

// 常见弱密码列表（部分）
var commonPasswords = map[string]bool{
	"password":  true,
	"123456":    true,
	"12345678":  true,
	"qwerty":    true,
	"abc123":    true,
	"monkey":    true,
	"1234567":   true,
	"letmein":   true,
	"trustno1":  true,
	"dragon":    true,
	"baseball":  true,
	"iloveyou":  true,
	"master":    true,
	"sunshine":  true,
	"ashley":    true,
	"bailey":    true,
	"shadow":    true,
	"123123":    true,
	"654321":    true,
	"superman":  true,
	"qazwsx":    true,
	"michael":   true,
	"football":  true,
	"password1": true,
	"password2": true,
	"admin":     true,
	"admin123":  true,
	"root":      true,
	"toor":      true,
	"test":      true,
	"test123":   true,
	"user":      true,
	"user123":   true,
	"guest":     true,
	"welcome":   true,
	"welcome1":  true,
	"hello":     true,
	"hello123":  true,
	"passw0rd":  true,
	"p@ssw0rd":  true,
	"pass123":   true,
	"changeme":  true,
}

// Validator 密码验证器
type Validator struct {
	config ValidationConfig
}

// NewValidator 创建密码验证器
func NewValidator(config ...ValidationConfig) *Validator {
	cfg := DefaultConfig
	if len(config) > 0 {
		cfg = config[0]
	}
	return &Validator{config: cfg}
}

// Validate 验证密码是否符合复杂度要求
func (v *Validator) Validate(password string) error {
	// 检查长度
	if len(password) < v.config.MinLength {
		return ErrPasswordTooShort
	}
	if len(password) > v.config.MaxLength {
		return ErrPasswordTooLong
	}

	// 检查常见弱密码
	if v.config.CheckCommonList {
		if commonPasswords[strings.ToLower(password)] {
			return ErrPasswordCommon
		}
	}

	// 检查是否包含用户名或常见单词模式
	if v.config.CheckCommonList {
		if err := v.checkPatterns(password); err != nil {
			return err
		}
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if v.config.RequireUpper && !hasUpper {
		return ErrPasswordNoUpper
	}
	if v.config.RequireLower && !hasLower {
		return ErrPasswordNoLower
	}
	if v.config.RequireDigit && !hasDigit {
		return ErrPasswordNoDigit
	}
	if v.config.RequireSpecial && !hasSpecial {
		return ErrPasswordNoSpecial
	}

	return nil
}

// checkPatterns 检查密码中的常见模式
func (v *Validator) checkPatterns(password string) error {
	lowerPass := strings.ToLower(password)

	// 检查连续字符模式（如123456, abcdef）
	if hasSequentialChars(lowerPass, 4) {
		return ErrPasswordTooWeak
	}

	// 检查重复字符模式（如aaaa, 1111）
	if hasRepeatedChars(password, 3) {
		return ErrPasswordTooWeak
	}

	// 检查键盘模式（如qwerty, asdfgh）
	keyboardPatterns := []string{"qwerty", "asdfgh", "zxcvbn", "qazwsx", "edcrfv"}
	for _, pattern := range keyboardPatterns {
		if strings.Contains(lowerPass, pattern) {
			return ErrPasswordTooWeak
		}
	}

	return nil
}

// hasSequentialChars 检查是否有连续字符
func hasSequentialChars(s string, length int) bool {
	if len(s) < length {
		return false
	}

	sequential := 1
	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1]+1 || s[i] == s[i-1]-1 {
			sequential++
			if sequential >= length {
				return true
			}
		} else {
			sequential = 1
		}
	}
	return false
}

// hasRepeatedChars 检查是否有重复字符
func hasRepeatedChars(s string, count int) bool {
	if len(s) < count {
		return false
	}

	repeated := 1
	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1] {
			repeated++
			if repeated >= count {
				return true
			}
		} else {
			repeated = 1
		}
	}
	return false
}

// PasswordStrength 密码强度等级
type PasswordStrength int

const (
	StrengthVeryWeak PasswordStrength = iota
	StrengthWeak
	StrengthMedium
	StrengthStrong
	StrengthVeryStrong
)

func (s PasswordStrength) String() string {
	switch s {
	case StrengthVeryWeak:
		return "非常弱"
	case StrengthWeak:
		return "弱"
	case StrengthMedium:
		return "中等"
	case StrengthStrong:
		return "强"
	case StrengthVeryStrong:
		return "非常强"
	default:
		return "未知"
	}
}

// CalculateStrength 计算密码强度
func (v *Validator) CalculateStrength(password string) PasswordStrength {
	score := 0

	// 长度评分
	if len(password) >= 8 {
		score++
	}
	if len(password) >= 12 {
		score++
	}
	if len(password) >= 16 {
		score++
	}

	// 字符类型评分
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if hasUpper {
		score++
	}
	if hasLower {
		score++
	}
	if hasDigit {
		score++
	}
	if hasSpecial {
		score++
	}

	// 混合类型加分
	typeCount := 0
	if hasUpper {
		typeCount++
	}
	if hasLower {
		typeCount++
	}
	if hasDigit {
		typeCount++
	}
	if hasSpecial {
		typeCount++
	}
	if typeCount >= 3 {
		score++
	}

	// 惩罚常见弱密码
	if commonPasswords[strings.ToLower(password)] {
		score -= 3
	}

	// 惩罚连续或重复字符
	if hasSequentialChars(strings.ToLower(password), 3) {
		score--
	}
	if hasRepeatedChars(password, 3) {
		score--
	}

	// 转换为强度等级
	switch {
	case score <= 2:
		return StrengthVeryWeak
	case score <= 4:
		return StrengthWeak
	case score <= 6:
		return StrengthMedium
	case score <= 8:
		return StrengthStrong
	default:
		return StrengthVeryStrong
	}
}

// 全局默认验证器
var defaultValidator = NewValidator()

// Validate 使用默认配置验证密码
func Validate(password string) error {
	return defaultValidator.Validate(password)
}

// CalculatePasswordStrength 使用默认配置计算密码强度
func CalculatePasswordStrength(password string) PasswordStrength {
	return defaultValidator.CalculateStrength(password)
}

// GetPasswordRequirements 获取密码要求说明（用于前端展示）
func GetPasswordRequirements() []string {
	return []string{
		"密码长度至少8位，最多72位",
		"至少包含一个大写字母（A-Z）",
		"至少包含一个小写字母（a-z）",
		"至少包含一个数字（0-9）",
		"至少包含一个特殊字符（如 !@#$%^&*）",
		"不能使用常见弱密码（如 password、123456）",
		"不能包含连续或重复字符（如 123、aaa）",
	}
}

// IsPasswordValid 检查密码是否有效（简化接口）
func IsPasswordValid(password string) (bool, string) {
	err := Validate(password)
	if err != nil {
		return false, err.Error()
	}
	return true, ""
}

// 匹配密码复杂度的正则表达式（可选，用于前端或快速检查）
var (
	// UpperCaseRegex 大写字母正则
	UpperCaseRegex = regexp.MustCompile(`[A-Z]`)
	// LowerCaseRegex 小写字母正则
	LowerCaseRegex = regexp.MustCompile(`[a-z]`)
	// DigitRegex 数字正则
	DigitRegex = regexp.MustCompile(`[0-9]`)
	// SpecialCharRegex 特殊字符正则
	SpecialCharRegex = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]`)
)
