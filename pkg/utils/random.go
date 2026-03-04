package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
)

const (
	// 字母表用于生成随机字符串
	letters      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits       = "0123456789"
	alphanumeric = letters + digits
)

// GenerateRandomString 生成指定长度的随机字符串（字母+数字）
func GenerateRandomString(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphanumeric))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		result[i] = alphanumeric[num.Int64()]
	}
	return string(result), nil
}

// GenerateRandomHex 生成指定长度的十六进制随机字符串
func GenerateRandomHex(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}

	bytes := make([]byte, (length+1)/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	hexStr := hex.EncodeToString(bytes)
	if len(hexStr) > length {
		hexStr = hexStr[:length]
	}
	return hexStr, nil
}

// GenerateAppKey 生成 AppKey（32位字母数字组合）
func GenerateAppKey() (string, error) {
	return GenerateRandomString(32)
}

// GenerateAppSecret 生成 AppSecret（64位十六进制）
func GenerateAppSecret() (string, error) {
	return GenerateRandomHex(64)
}
