package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Storage  StorageConfig  `yaml:"storage"`
	Log      LogConfig      `yaml:"log"`
	CORS     CORSConfig     `yaml:"cors"`
	MCP      MCPConfig      `yaml:"mcp"`
}

type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	Mode         string        `yaml:"mode"` // debug, release
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	DBName       string `yaml:"dbname"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}

type JWTConfig struct {
	Secret        string        `yaml:"secret"`
	Expire        time.Duration `yaml:"expire"`
	RefreshExpire time.Duration `yaml:"refresh_expire"`
}

type StorageConfig struct {
	Type        string `yaml:"type"` // local
	Path        string `yaml:"path"`
	MaxFileSize int64  `yaml:"max_file_size"` // bytes
}

type LogConfig struct {
	Level string `yaml:"level"` // debug, info, warn, error
}

type CORSConfig struct {
	Enabled      bool     `yaml:"enabled"`
	AllowOrigins []string `yaml:"allow_origins"`
	AllowMethods []string `yaml:"allow_methods"`
	AllowHeaders []string `yaml:"allow_headers"`
}

// MCPConfig MCP 服务配置
type MCPConfig struct {
	HTTP MCPHTTPConfig `yaml:"http"`
	Auth MCPAuthConfig `yaml:"auth"`
}

// MCPHTTPConfig MCP HTTP 传输配置
type MCPHTTPConfig struct {
	Enabled      bool   `yaml:"enabled"`       // 是否启用 HTTP 传输
	Host         string `yaml:"host"`          // 监听地址
	Port         int    `yaml:"port"`          // 监听端口
	EndpointPath string `yaml:"endpoint_path"` // MCP 端点路径
}

// MCPAuthConfig MCP 认证配置
type MCPAuthConfig struct {
	AppKey    string `yaml:"app_key"`    // 应用 Key
	AppSecret string `yaml:"app_secret"` // 应用 Secret
}

// ResolveConfigPath 解析配置文件路径
// 优先级：
// 1. 环境变量 MCP_CONFIG_PATH
// 2. 可执行文件所在目录的 configs/config.yaml
// 3. 当前工作目录的 ./configs/config.yaml
func ResolveConfigPath(defaultPath string) string {
	// 1. 检查环境变量
	if envPath := os.Getenv("MCP_CONFIG_PATH"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
	}

	// 2. 尝试可执行文件所在目录
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		execConfigPath := filepath.Join(execDir, "configs", "config.yaml")
		if _, err := os.Stat(execConfigPath); err == nil {
			return execConfigPath
		}
	}

	// 3. 返回默认路径（当前工作目录）
	return defaultPath
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// set defaults
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 30 * time.Second
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 30 * time.Second
	}
	if cfg.JWT.Expire == 0 {
		cfg.JWT.Expire = 24 * time.Hour
	}
	if cfg.JWT.RefreshExpire == 0 {
		cfg.JWT.RefreshExpire = 168 * time.Hour
	}
	if cfg.Storage.Path == "" {
		cfg.Storage.Path = "./uploads"
	}
	if cfg.Storage.MaxFileSize == 0 {
		cfg.Storage.MaxFileSize = 100 * 1024 * 1024 // 100MB
	}
	// CORS defaults
	if len(cfg.CORS.AllowOrigins) == 0 {
		cfg.CORS.AllowOrigins = []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	}
	if len(cfg.CORS.AllowMethods) == 0 {
		cfg.CORS.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	if len(cfg.CORS.AllowHeaders) == 0 {
		cfg.CORS.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "X-App-Key", "X-Timestamp", "X-Signature"}
	}

	// MCP HTTP defaults
	if cfg.MCP.HTTP.Port == 0 {
		cfg.MCP.HTTP.Port = 8080 // 默认与主服务共用端口
	}
	if cfg.MCP.HTTP.Host == "" {
		cfg.MCP.HTTP.Host = "localhost"
	}
	if cfg.MCP.HTTP.EndpointPath == "" {
		cfg.MCP.HTTP.EndpointPath = "/mcp"
	}

	return &cfg, nil
}
