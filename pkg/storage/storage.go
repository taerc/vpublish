package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) (*LocalStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("create storage directory: %w", err)
	}
	return &LocalStorage{basePath: basePath}, nil
}

// Save 保存上传的文件
// 返回: 相对路径, 文件大小, 文件哈希, 错误
func (s *LocalStorage) Save(file *multipart.FileHeader, category string) (string, int64, string, error) {
	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return "", 0, "", fmt.Errorf("open uploaded file: %w", err)
	}
	defer src.Close()

	// 生成存储路径
	dateDir := time.Now().Format("2006/01/02")
	storageDir := filepath.Join(s.basePath, category, dateDir)
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return "", 0, "", fmt.Errorf("create storage directory: %w", err)
	}

	// 生成文件名：时间戳_原文件名
	ext := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(),
		strings.TrimSuffix(file.Filename, ext), ext)

	// 清理文件名中的特殊字符
	fileName = sanitizeFileName(fileName)

	relativePath := filepath.Join(category, dateDir, fileName)
	absPath := filepath.Join(s.basePath, relativePath)

	// 创建目标文件
	dst, err := os.Create(absPath)
	if err != nil {
		return "", 0, "", fmt.Errorf("create destination file: %w", err)
	}
	defer dst.Close()

	// 计算哈希并写入文件
	hasher := sha256.New()
	writer := io.MultiWriter(dst, hasher)

	size, err := io.Copy(writer, src)
	if err != nil {
		os.Remove(absPath)
		return "", 0, "", fmt.Errorf("write file: %w", err)
	}

	hash := hex.EncodeToString(hasher.Sum(nil))
	return relativePath, size, hash, nil
}

// Delete 删除文件
func (s *LocalStorage) Delete(relativePath string) error {
	absPath := filepath.Join(s.basePath, relativePath)
	return os.Remove(absPath)
}

// GetFilePath 获取文件的绝对路径
func (s *LocalStorage) GetFilePath(relativePath string) string {
	return filepath.Join(s.basePath, relativePath)
}

// Exists 检查文件是否存在
func (s *LocalStorage) Exists(relativePath string) bool {
	absPath := filepath.Join(s.basePath, relativePath)
	_, err := os.Stat(absPath)
	return !os.IsNotExist(err)
}

// sanitizeFileName 清理文件名
func sanitizeFileName(name string) string {
	// 替换不安全字符
	replacer := strings.NewReplacer(
		" ", "_",
		"\\", "_",
		"/", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(name)
}
