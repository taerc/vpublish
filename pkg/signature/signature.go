package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
)

const (
	HeaderAppKey    = "X-App-Key"
	HeaderTimestamp = "X-Timestamp"
	HeaderSignature = "X-Signature"
	SignatureExpire = 300 // 签名有效期（秒）
)

// GenerateSignature 生成签名
// params: 请求参数
// appSecret: 应用密钥
// timestamp: 时间戳
func GenerateSignature(params map[string]string, appSecret string, timestamp int64) string {
	// 1. 按key排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 2. 拼接
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString("&")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// 3. 添加 timestamp
	if len(keys) > 0 {
		builder.WriteString("&")
	}
	builder.WriteString("timestamp=")
	builder.WriteString(strconv.FormatInt(timestamp, 10))

	// 4. HMAC-SHA256
	h := hmac.New(sha256.New, []byte(appSecret))
	h.Write([]byte(builder.String()))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySignature 验证签名
func VerifySignature(params map[string]string, appSecret string, timestamp int64, signature string) bool {
	expected := GenerateSignature(params, appSecret, timestamp)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// GenerateDownloadToken 生成下载令牌
// versionID: 版本ID
// appSecret: 应用密钥
// expires: 过期时间戳
func GenerateDownloadToken(versionID uint, appSecret string, expires int64) string {
	message := strconv.FormatUint(uint64(versionID), 10) + "|" + strconv.FormatInt(expires, 10)
	h := hmac.New(sha256.New, []byte(appSecret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyDownloadToken 验证下载令牌
func VerifyDownloadToken(versionID uint, appSecret string, expires int64, token string) bool {
	expected := GenerateDownloadToken(versionID, appSecret, expires)
	return hmac.Equal([]byte(expected), []byte(token))
}
