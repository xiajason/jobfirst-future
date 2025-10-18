package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"superadmin/errors"
)

// Manager 配置管理器
type Manager struct {
	config *ConfigManagerConfig
}

// ConfigManagerConfig 配置管理器配置
type ConfigManagerConfig struct {
	ConfigPath     string `json:"config_path"`
	BackupPath     string `json:"backup_path"`
	ValidationPath string `json:"validation_path"`
}

// NewManager 创建配置管理器
func NewManager(config *ConfigManagerConfig) *Manager {
	return &Manager{
		config: config,
	}
}

// ConfigBackup 配置备份
type ConfigBackup struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	Size        int64     `json:"size"`
	Files       []string  `json:"files"`
	Checksum    string    `json:"checksum"`
}

// ConfigValidation 配置验证
type ConfigValidation struct {
	Valid      bool                `json:"valid"`
	Errors     []ValidationError   `json:"errors"`
	Warnings   []ValidationWarning `json:"warnings"`
	CheckedAt  time.Time           `json:"checked_at"`
	FilesCount int                 `json:"files_count"`
}

// ValidationError 验证错误
type ValidationError struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// ValidationWarning 验证警告
type ValidationWarning struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// Environment 环境配置
type Environment struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Config      map[string]string `json:"config"`
	IsActive    bool              `json:"is_active"`
	LastUpdated time.Time         `json:"last_updated"`
}

// CollectAllConfigs 收集所有配置
func (m *Manager) CollectAllConfigs() error {
	// 确保配置目录存在
	if err := os.MkdirAll(m.config.ConfigPath, 0755); err != nil {
		return errors.WrapError(errors.ErrCodeFile, "创建配置目录失败", err)
	}

	// 收集各种配置文件
	configs := map[string]string{
		"database":    "/etc/mysql/mysql.conf.d/mysqld.cnf",
		"redis":       "/etc/redis/redis.conf",
		"nginx":       "/etc/nginx/nginx.conf",
		"systemd":     "/etc/systemd/system/",
		"environment": "/etc/environment",
		"crontab":     "/etc/crontab",
	}

	for name, path := range configs {
		if err := m.collectConfig(name, path); err != nil {
			// 记录错误但继续收集其他配置
			fmt.Printf("收集配置 %s 失败: %v\n", name, err)
		}
	}

	return nil
}

// collectConfig 收集单个配置
func (m *Manager) collectConfig(name, path string) error {
	// 检查文件或目录是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.NewError(errors.ErrCodeFileNotFound, fmt.Sprintf("配置文件不存在: %s", path))
	}

	// 创建目标目录
	targetDir := filepath.Join(m.config.ConfigPath, name)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return errors.WrapError(errors.ErrCodeFile, "创建目标目录失败", err)
	}

	// 复制文件或目录
	if err := m.copyPath(path, targetDir); err != nil {
		return errors.WrapError(errors.ErrCodeFile, "复制配置文件失败", err)
	}

	return nil
}

// copyPath 复制路径
func (m *Manager) copyPath(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return m.copyDirectory(src, dst)
	} else {
		return m.copyFile(src, dst)
	}
}

// copyDirectory 复制目录
func (m *Manager) copyDirectory(src, dst string) error {
	// 创建目标目录
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// 读取源目录
	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	// 复制每个条目
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := m.copyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := m.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile 复制文件
func (m *Manager) copyFile(src, dst string) error {
	// 读取源文件
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	// 写入目标文件
	return ioutil.WriteFile(dst, data, 0644)
}

// BackupConfigs 备份配置
func (m *Manager) BackupConfigs() (*ConfigBackup, error) {
	// 生成备份ID
	backupID := fmt.Sprintf("backup_%d", time.Now().Unix())
	backupName := fmt.Sprintf("config_backup_%s", time.Now().Format("20060102_150405"))

	// 创建备份目录
	backupDir := filepath.Join(m.config.BackupPath, backupID)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, errors.WrapError(errors.ErrCodeFile, "创建备份目录失败", err)
	}

	// 复制配置文件到备份目录
	if err := m.copyDirectory(m.config.ConfigPath, backupDir); err != nil {
		return nil, errors.WrapError(errors.ErrCodeFile, "复制配置文件失败", err)
	}

	// 计算备份信息
	backup := &ConfigBackup{
		ID:          backupID,
		Name:        backupName,
		Description: "自动配置备份",
		CreatedAt:   time.Now(),
		Files:       []string{},
	}

	// 统计文件信息
	if err := m.calculateBackupInfo(backupDir, backup); err != nil {
		return nil, errors.WrapError(errors.ErrCodeFile, "计算备份信息失败", err)
	}

	// 生成备份元数据文件
	metadataFile := filepath.Join(backupDir, "backup_metadata.json")
	metadata, _ := json.MarshalIndent(backup, "", "  ")
	if err := ioutil.WriteFile(metadataFile, metadata, 0644); err != nil {
		return nil, errors.WrapError(errors.ErrCodeFile, "写入备份元数据失败", err)
	}

	return backup, nil
}

// calculateBackupInfo 计算备份信息
func (m *Manager) calculateBackupInfo(backupDir string, backup *ConfigBackup) error {
	var totalSize int64
	var fileCount int

	err := filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			totalSize += info.Size()
			fileCount++

			// 记录文件路径
			relPath, _ := filepath.Rel(backupDir, path)
			backup.Files = append(backup.Files, relPath)
		}

		return nil
	})

	if err != nil {
		return err
	}

	backup.Size = totalSize
	backup.Checksum = fmt.Sprintf("%x", totalSize) // 简化的校验和

	return nil
}

// ValidateConfigs 验证配置
func (m *Manager) ValidateConfigs() (*ConfigValidation, error) {
	validation := &ConfigValidation{
		Valid:      true,
		Errors:     []ValidationError{},
		Warnings:   []ValidationWarning{},
		CheckedAt:  time.Now(),
		FilesCount: 0,
	}

	// 验证配置文件
	if err := m.validateConfigFiles(validation); err != nil {
		return nil, errors.WrapError(errors.ErrCodeValidation, "验证配置文件失败", err)
	}

	// 验证环境变量
	if err := m.validateEnvironmentVariables(validation); err != nil {
		return nil, errors.WrapError(errors.ErrCodeValidation, "验证环境变量失败", err)
	}

	// 验证服务配置
	if err := m.validateServiceConfigs(validation); err != nil {
		return nil, errors.WrapError(errors.ErrCodeValidation, "验证服务配置失败", err)
	}

	// 确定整体验证状态
	validation.Valid = len(validation.Errors) == 0

	return validation, nil
}

// validateConfigFiles 验证配置文件
func (m *Manager) validateConfigFiles(validation *ConfigValidation) error {
	// 验证JSON配置文件
	jsonFiles := []string{
		"config.json",
		"database.json",
		"redis.json",
		"ai.json",
	}

	for _, filename := range jsonFiles {
		filepath := filepath.Join(m.config.ConfigPath, filename)
		if _, err := os.Stat(filepath); err == nil {
			validation.FilesCount++
			if err := m.validateJSONFile(filepath, validation); err != nil {
				validation.Errors = append(validation.Errors, ValidationError{
					File:    filename,
					Message: err.Error(),
					Type:    "json_syntax",
				})
			}
		}
	}

	// 验证YAML配置文件
	yamlFiles := []string{
		"docker-compose.yml",
		"k8s-config.yaml",
		"nginx.conf",
	}

	for _, filename := range yamlFiles {
		filepath := filepath.Join(m.config.ConfigPath, filename)
		if _, err := os.Stat(filepath); err == nil {
			validation.FilesCount++
			if err := m.validateYAMLFile(filepath, validation); err != nil {
				validation.Errors = append(validation.Errors, ValidationError{
					File:    filename,
					Message: err.Error(),
					Type:    "yaml_syntax",
				})
			}
		}
	}

	return nil
}

// validateJSONFile 验证JSON文件
func (m *Manager) validateJSONFile(filepath string, validation *ConfigValidation) error {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return err
	}

	return nil
}

// validateYAMLFile 验证YAML文件
func (m *Manager) validateYAMLFile(filepath string, validation *ConfigValidation) error {
	// 简化的YAML验证
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	// 检查基本的YAML语法
	content := string(data)
	if strings.Contains(content, "\t") {
		validation.Warnings = append(validation.Warnings, ValidationWarning{
			File:    filepath,
			Message: "YAML文件包含制表符，建议使用空格",
			Type:    "yaml_format",
		})
	}

	return nil
}

// validateEnvironmentVariables 验证环境变量
func (m *Manager) validateEnvironmentVariables(validation *ConfigValidation) error {
	requiredVars := []string{
		"PATH",
		"HOME",
		"USER",
	}

	optionalVars := []string{
		"DATABASE_URL",
		"REDIS_URL",
		"API_KEY",
		"SECRET_KEY",
	}

	// 检查必需的环境变量
	for _, varName := range requiredVars {
		if os.Getenv(varName) == "" {
			validation.Errors = append(validation.Errors, ValidationError{
				File:    "environment",
				Message: fmt.Sprintf("必需的环境变量 %s 未设置", varName),
				Type:    "missing_env_var",
			})
		}
	}

	// 检查可选的环境变量
	for _, varName := range optionalVars {
		if os.Getenv(varName) == "" {
			validation.Warnings = append(validation.Warnings, ValidationWarning{
				File:    "environment",
				Message: fmt.Sprintf("可选的环境变量 %s 未设置", varName),
				Type:    "missing_optional_env_var",
			})
		}
	}

	return nil
}

// validateServiceConfigs 验证服务配置
func (m *Manager) validateServiceConfigs(validation *ConfigValidation) error {
	// 验证systemd服务配置
	systemdPath := filepath.Join(m.config.ConfigPath, "systemd")
	if _, err := os.Stat(systemdPath); err == nil {
		validation.FilesCount++
		if err := m.validateSystemdConfigs(systemdPath, validation); err != nil {
			validation.Errors = append(validation.Errors, ValidationError{
				File:    "systemd",
				Message: err.Error(),
				Type:    "systemd_config",
			})
		}
	}

	return nil
}

// validateSystemdConfigs 验证systemd配置
func (m *Manager) validateSystemdConfigs(systemdPath string, validation *ConfigValidation) error {
	// 检查systemd服务文件
	serviceFiles := []string{
		"jobfirst-api.service",
		"jobfirst-worker.service",
		"jobfirst-scheduler.service",
	}

	for _, serviceFile := range serviceFiles {
		filepath := filepath.Join(systemdPath, serviceFile)
		if _, err := os.Stat(filepath); err == nil {
			if err := m.validateSystemdService(filepath, validation); err != nil {
				validation.Errors = append(validation.Errors, ValidationError{
					File:    serviceFile,
					Message: err.Error(),
					Type:    "systemd_service",
				})
			}
		}
	}

	return nil
}

// validateSystemdService 验证systemd服务文件
func (m *Manager) validateSystemdService(filepath string, validation *ConfigValidation) error {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	content := string(data)

	// 检查必需的section
	requiredSections := []string{"[Unit]", "[Service]"}
	for _, section := range requiredSections {
		if !strings.Contains(content, section) {
			return fmt.Errorf("缺少必需的section: %s", section)
		}
	}

	// 检查必需的配置项
	requiredKeys := []string{"ExecStart", "User", "Group"}
	for _, key := range requiredKeys {
		if !strings.Contains(content, key+"=") {
			return fmt.Errorf("缺少必需的配置项: %s", key)
		}
	}

	return nil
}

// GetEnvironments 获取环境配置
func (m *Manager) GetEnvironments() ([]Environment, error) {
	environments := []Environment{
		{
			Name:        "development",
			Description: "开发环境",
			Config: map[string]string{
				"DEBUG":        "true",
				"LOG_LEVEL":    "debug",
				"DATABASE_URL": "mysql://localhost:3306/jobfirst_dev",
				"REDIS_URL":    "redis://localhost:6379/0",
				"API_BASE_URL": "http://localhost:8080",
			},
			IsActive:    true,
			LastUpdated: time.Now(),
		},
		{
			Name:        "staging",
			Description: "预发布环境",
			Config: map[string]string{
				"DEBUG":        "false",
				"LOG_LEVEL":    "info",
				"DATABASE_URL": "mysql://staging-db:3306/jobfirst_staging",
				"REDIS_URL":    "redis://staging-redis:6379/0",
				"API_BASE_URL": "https://staging-api.jobfirst.com",
			},
			IsActive:    false,
			LastUpdated: time.Now(),
		},
		{
			Name:        "production",
			Description: "生产环境",
			Config: map[string]string{
				"DEBUG":        "false",
				"LOG_LEVEL":    "warn",
				"DATABASE_URL": "mysql://prod-db:3306/jobfirst_prod",
				"REDIS_URL":    "redis://prod-redis:6379/0",
				"API_BASE_URL": "https://api.jobfirst.com",
			},
			IsActive:    false,
			LastUpdated: time.Now(),
		},
	}

	return environments, nil
}
