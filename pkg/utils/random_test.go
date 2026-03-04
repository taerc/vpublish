package utils

import (
	"regexp"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"length 16", 16},
		{"length 32", 32},
		{"length 64", 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateRandomString(tt.length)
			if err != nil {
				t.Fatalf("GenerateRandomString() error = %v", err)
			}
			if len(result) != tt.length {
				t.Errorf("GenerateRandomString() length = %v, want %v", len(result), tt.length)
			}
			// 检查是否只包含字母数字
			matched, _ := regexp.MatchString("^[a-zA-Z0-9]+$", result)
			if !matched {
				t.Errorf("GenerateRandomString() contains invalid characters: %v", result)
			}
		})
	}

	// 测试每次生成的结果不同
	str1, _ := GenerateRandomString(32)
	str2, _ := GenerateRandomString(32)
	if str1 == str2 {
		t.Error("GenerateRandomString() should generate different values")
	}
}

func TestGenerateRandomHex(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"length 16", 16},
		{"length 32", 32},
		{"length 64", 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateRandomHex(tt.length)
			if err != nil {
				t.Fatalf("GenerateRandomHex() error = %v", err)
			}
			if len(result) != tt.length {
				t.Errorf("GenerateRandomHex() length = %v, want %v", len(result), tt.length)
			}
			// 检查是否只包含十六进制字符
			matched, _ := regexp.MatchString("^[a-f0-9]+$", result)
			if !matched {
				t.Errorf("GenerateRandomHex() contains invalid characters: %v", result)
			}
		})
	}
}

func TestGenerateAppKey(t *testing.T) {
	key, err := GenerateAppKey()
	if err != nil {
		t.Fatalf("GenerateAppKey() error = %v", err)
	}
	if len(key) != 32 {
		t.Errorf("GenerateAppKey() length = %v, want 32", len(key))
	}
	matched, _ := regexp.MatchString("^[a-zA-Z0-9]+$", key)
	if !matched {
		t.Errorf("GenerateAppKey() contains invalid characters: %v", key)
	}
}

func TestGenerateAppSecret(t *testing.T) {
	secret, err := GenerateAppSecret()
	if err != nil {
		t.Fatalf("GenerateAppSecret() error = %v", err)
	}
	if len(secret) != 64 {
		t.Errorf("GenerateAppSecret() length = %v, want 64", len(secret))
	}
	matched, _ := regexp.MatchString("^[a-f0-9]+$", secret)
	if !matched {
		t.Errorf("GenerateAppSecret() contains invalid characters: %v", secret)
	}
}
