package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// 配置信息
const (
	BaseURL   = "http://172.16.50.41:8080" // 修改为你的服务器地址
	AppKey    = "JCasbrUdESmxGHuNpL7tUNXhvqFSLt1y"
	AppSecret = "ecf93cc892d18fe7edafe2355a836bd89a56516655e9487580e59d4e638bed21"
)

// 报错上报请求结构
type ErrorReportRequest struct {
	RequestID    string                 `json:"request_id"`
	Timestamp    int64                  `json:"timestamp"`
	Code         string                 `json:"code"`
	ErrorMessage string                 `json:"error_message"`
	DeviceInfo   map[string]string      `json:"device_info"`
	RequestParams map[string]interface{} `json:"request_params,omitempty"`
}

// 响应结构
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  VPublish 报错上报接口测试")
	fmt.Println("========================================")
	fmt.Println()

	// 显示配置信息
	fmt.Println("配置信息:")
	fmt.Printf("  Base URL: %s\n", BaseURL)
	fmt.Printf("  App Key: %s\n", AppKey)
	fmt.Printf("  App Secret: %s\n", AppSecret)
	fmt.Println()

	// 测试单条上报
	fmt.Println("----------------------------------------")
	fmt.Println("测试 1: 单条报错上报")
	fmt.Println("----------------------------------------")
	testSingleReport()

	fmt.Println()

	// 测试批量上报
	fmt.Println("----------------------------------------")
	fmt.Println("测试 2: 批量报错上报")
	fmt.Println("----------------------------------------")
	testBatchReport()
}

// 测试单条上报
func testSingleReport() {
	// 准备请求数据
	request := ErrorReportRequest{
		RequestID:    fmt.Sprintf("req-%d", time.Now().Unix()),
		Timestamp:    time.Now().UnixMilli(),
		Code:         "500",
		ErrorMessage: "Internal Server Error",
		DeviceInfo: map[string]string{
			"app_version":  "1.0.0",
			"device_model": "Test Device",
			"os_version":   "Windows 11",
		},
	}

	// 发送请求
	response, err := sendRequest("/api/v1/app/error/report", request, nil)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 单条上报成功\n")
	fmt.Printf("   响应: %s\n", response.Message)
	if response.Data != nil {
		fmt.Printf("   数据: %v\n", response.Data)
	}
}

// 测试批量上报
func testBatchReport() {
	// 准备多条错误
	requests := []map[string]interface{}{
		{
			"request_id": fmt.Sprintf("req-batch1-%d", time.Now().Unix()),
			"timestamp":   time.Now().UnixMilli(),
			"code":        "500",
			"error_message": "Error 1: Internal Server Error",
			"device_info": map[string]string{
				"app_version":  "1.0.0",
				"device_model": "Test Device",
				"os_version":   "Windows 11",
			},
		},
		{
			"request_id": fmt.Sprintf("req-batch2-%d", time.Now().Unix()),
			"timestamp":   time.Now().UnixMilli(),
			"code":        "400",
			"error_message": "Error 2: Bad Request",
			"device_info": map[string]string{
				"app_version":  "1.0.0",
				"device_model": "Test Device",
				"os_version":   "Windows 11",
			},
		},
		{
			"request_id": fmt.Sprintf("req-batch3-%d", time.Now().Unix()),
			"timestamp":   time.Now().UnixMilli(),
			"code":        "404",
			"error_message": "Error 3: Not Found",
			"device_info": map[string]string{
				"app_version":  "1.0.0",
				"device_model": "Test Device",
				"os_version":   "Windows 11",
			},
		},
	}

	// 发送请求
	body := map[string]interface{}{
		"records": requests,
	}

	response, err := sendRequest("/api/v1/app/error/report/batch", body, nil)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 批量上报成功\n")
	fmt.Printf("   响应: %s\n", response.Message)
	if response.Data != nil {
		// 解析返回的数据
		if dataMap, ok := response.Data.(map[string]interface{}); ok {
			if successCount, ok := dataMap["success_count"].(float64); ok {
				fmt.Printf("   成功: %.0f 条\n", successCount)
			}
			if failedCount, ok := dataMap["failed_count"].(float64); ok {
				fmt.Printf("   失败: %.0f 条\n", failedCount)
			}
		}
		fmt.Printf("   数据: %v\n", response.Data)
	}
}

// 发送 HTTP 请求
func sendRequest(path string, body interface{}, queryParams map[string]string) (*Response, error) {
	// 构建完整 URL
	fullURL := BaseURL + path
	if queryParams != nil {
		urlObj, _ := url.Parse(fullURL)
		q := urlObj.Query()
		for k, v := range queryParams {
			q.Set(k, v)
		}
		urlObj.RawQuery = q.Encode()
		fullURL = urlObj.String()
	}

	// 准备时间戳
	timestamp := time.Now().Unix()
	timestampRFC3339 := time.Unix(timestamp, 0).Format(time.RFC3339)

	// 生成签名
	signature := generateSignature(queryParams, AppSecret, timestamp)

	// 打印请求信息
	fmt.Printf("\n请求信息:\n")
	fmt.Printf("  URL: %s\n", fullURL)
	fmt.Printf("  X-App-Key: %s\n", AppKey)
	fmt.Printf("  X-Timestamp: %s\n", timestampRFC3339)
	fmt.Printf("  X-Signature: %s\n", signature)

	// 序列化请求体
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}
	fmt.Printf("  Body: %s\n", string(jsonBody))

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Key", AppKey)
	req.Header.Set("X-Timestamp", timestampRFC3339)
	req.Header.Set("X-Signature", signature)

	// 发送请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	fmt.Printf("\n响应信息:\n")
	fmt.Printf("  Status Code: %d\n", resp.StatusCode)
	fmt.Printf("  Body: %s\n", string(respBody))

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP 错误: %d, %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var response Response
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &response, nil
}

// 生成签名
func generateSignature(params map[string]string, appSecret string, timestamp int64) string {
	// 1. 按字母顺序排序参数的 key
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 2. 拼接参数: key1=value1&key2=value2&...
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString("&")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// 3. 添加时间戳: ...&timestamp=1234567890
	if len(keys) > 0 {
		builder.WriteString("&")
	}
	builder.WriteString("timestamp=")
	builder.WriteString(strconv.FormatInt(timestamp, 10))

	// 4. HMAC-SHA256
	dataToSign := builder.String()
	h := hmac.New(sha256.New, []byte(appSecret))
	h.Write([]byte(dataToSign))
	signature := hex.EncodeToString(h.Sum(nil))

	// 打印调试信息
	fmt.Printf("\n签名生成详情:\n")
	fmt.Printf("  待签名字符串: %s\n", dataToSign)
	fmt.Printf("  签名结果: %s\n", signature)

	return signature
}
