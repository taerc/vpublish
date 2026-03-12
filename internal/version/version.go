package version

import (
	"fmt"
	"runtime"
)

var (
	// 以下变量通过 -ldflags 在编译时注入
	Version   = "dev"     // 语义化版本，如 v2.1.0
	GitCommit = "unknown" // Git 提交哈希（短，7位）
	BuildTime = "unknown" // 构建时间
)

// Info 版本信息结构
type Info struct {
	// 完整版本号
	Version string `json:"version"`
	// Git 提交哈希
	GitCommit string `json:"git_commit"`
	// 构建时间
	BuildTime string `json:"build_time"`
	// Go 版本
	GoVersion string `json:"go_version"`
	// 运行平台
	Platform string `json:"platform"`
}

// Get 获取版本信息
func Get() Info {
	return Info{
		Version:   String(),
		GitCommit: GitCommit,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String 返回完整版本号字符串
// 格式: {语义化版本}-{Git短提交哈希}-{发布时间}
// 示例: v2.1.0-a1b2c3d-20260312-1530
func String() string {
	if GitCommit == "unknown" {
		return Version
	}
	return fmt.Sprintf("%s-%s-%s", Version, GitCommit, BuildTime)
}
