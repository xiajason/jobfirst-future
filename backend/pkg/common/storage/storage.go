package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// StorageType 存储类型
type StorageType string

const (
	StorageTypeLocal StorageType = "local" // 本地存储
	StorageTypeS3    StorageType = "s3"    // AWS S3
	StorageTypeOSS   StorageType = "oss"   // 阿里云OSS
	StorageTypeCOS   StorageType = "cos"   // 腾讯云COS
)

// StorageConfig 存储配置
type StorageConfig struct {
	Type        StorageType `json:"type"`          // 存储类型
	BasePath    string      `json:"base_path"`     // 基础路径
	MaxFileSize int64       `json:"max_file_size"` // 最大文件大小
	AllowedExts []string    `json:"allowed_exts"`  // 允许的文件扩展名
	URLPrefix   string      `json:"url_prefix"`    // URL前缀
}

// FileInfo 文件信息
type FileInfo struct {
	Name        string    `json:"name"`         // 文件名
	Path        string    `json:"path"`         // 文件路径
	Size        int64     `json:"size"`         // 文件大小
	ContentType string    `json:"content_type"` // 内容类型
	URL         string    `json:"url"`          // 访问URL
	CreatedAt   time.Time `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`   // 更新时间
}

// StorageManager 存储管理器
type StorageManager struct {
	config *StorageConfig
}

// NewStorageManager 创建存储管理器
func NewStorageManager(config *StorageConfig) *StorageManager {
	if config == nil {
		config = &StorageConfig{
			Type:        StorageTypeLocal,
			BasePath:    "./uploads",
			MaxFileSize: 10 * 1024 * 1024, // 10MB
			AllowedExts: []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx"},
			URLPrefix:   "/uploads",
		}
	}

	// 确保基础目录存在
	if config.Type == StorageTypeLocal {
		os.MkdirAll(config.BasePath, 0755)
	}

	return &StorageManager{
		config: config,
	}
}

// UploadFile 上传文件
func (s *StorageManager) UploadFile(ctx context.Context, file *multipart.FileHeader, subPath string) (*FileInfo, error) {
	// 检查文件大小
	if file.Size > s.config.MaxFileSize {
		return nil, fmt.Errorf("file size exceeds limit: %d > %d", file.Size, s.config.MaxFileSize)
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !s.isAllowedExtension(ext) {
		return nil, fmt.Errorf("file extension not allowed: %s", ext)
	}

	// 生成文件路径
	fileName := s.generateFileName(file.Filename)
	filePath := filepath.Join(subPath, fileName)
	fullPath := filepath.Join(s.config.BasePath, filePath)

	// 确保目录存在
	os.MkdirAll(filepath.Dir(fullPath), 0755)

	// 保存文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close()

	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %v", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %v", err)
	}

	// 返回文件信息
	fileInfo := &FileInfo{
		Name:        fileName,
		Path:        filePath,
		Size:        file.Size,
		ContentType: file.Header.Get("Content-Type"),
		URL:         s.config.URLPrefix + "/" + filePath,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return fileInfo, nil
}

// DeleteFile 删除文件
func (s *StorageManager) DeleteFile(ctx context.Context, filePath string) error {
	fullPath := filepath.Join(s.config.BasePath, filePath)
	return os.Remove(fullPath)
}

// GetFileInfo 获取文件信息
func (s *StorageManager) GetFileInfo(ctx context.Context, filePath string) (*FileInfo, error) {
	fullPath := filepath.Join(s.config.BasePath, filePath)

	stat, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		Name:      filepath.Base(filePath),
		Path:      filePath,
		Size:      stat.Size(),
		URL:       s.config.URLPrefix + "/" + filePath,
		CreatedAt: stat.ModTime(),
		UpdatedAt: stat.ModTime(),
	}, nil
}

// FileExists 检查文件是否存在
func (s *StorageManager) FileExists(ctx context.Context, filePath string) (bool, error) {
	fullPath := filepath.Join(s.config.BasePath, filePath)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// isAllowedExtension 检查文件扩展名是否允许
func (s *StorageManager) isAllowedExtension(ext string) bool {
	for _, allowedExt := range s.config.AllowedExts {
		if strings.ToLower(allowedExt) == ext {
			return true
		}
	}
	return false
}

// generateFileName 生成文件名
func (s *StorageManager) generateFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%d%s", timestamp, ext)
}

// GetConfig 获取配置
func (s *StorageManager) GetConfig() *StorageConfig {
	return s.config
}
