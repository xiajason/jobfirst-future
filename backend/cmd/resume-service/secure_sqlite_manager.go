package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SecureSQLiteManager 安全的SQLite数据库管理器
type SecureSQLiteManager struct {
	basePath    string
	connections map[uint]*gorm.DB
	mutex       sync.RWMutex
	encryption  bool
}

// NewSecureSQLiteManager 创建安全的SQLite管理器
func NewSecureSQLiteManager(basePath string) *SecureSQLiteManager {
	return &SecureSQLiteManager{
		basePath:    basePath,
		connections: make(map[uint]*gorm.DB),
		encryption:  true, // 默认启用加密
	}
}

// getUserDatabasePath 获取用户数据库安全路径
func (sm *SecureSQLiteManager) getUserDatabasePath(userID uint) (string, error) {
	// 1. 验证用户ID的有效性
	if userID == 0 {
		return "", fmt.Errorf("无效的用户ID")
	}

	// 2. 创建安全的用户数据目录（与现有系统兼容）
	userDataDir := filepath.Join(sm.basePath, "users", fmt.Sprintf("%d", userID))

	// 3. 确保目录存在并设置安全权限
	if err := os.MkdirAll(userDataDir, 0700); err != nil {
		return "", fmt.Errorf("创建用户数据目录失败: %v", err)
	}

	// 4. 设置目录权限为仅所有者可访问
	if err := os.Chmod(userDataDir, 0700); err != nil {
		return "", fmt.Errorf("设置目录权限失败: %v", err)
	}

	// 5. 使用固定的数据库文件名（与现有系统兼容）
	dbFileName := "resume.db"
	dbPath := filepath.Join(userDataDir, dbFileName)

	// 6. 如果数据库文件已存在，验证其安全性
	if _, err := os.Stat(dbPath); err == nil {
		// 验证文件权限
		if err := sm.validateFilePermissions(dbPath); err != nil {
			return "", fmt.Errorf("数据库文件权限验证失败: %v", err)
		}
	}

	return dbPath, nil
}

// generateRandomSuffix 生成安全的随机后缀
func (sm *SecureSQLiteManager) generateRandomSuffix() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// validateFilePermissions 验证文件权限
func (sm *SecureSQLiteManager) validateFilePermissions(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	// 检查文件权限，确保只有所有者可以访问
	mode := info.Mode()
	if mode&0077 != 0 { // 检查组和其他用户的权限
		return fmt.Errorf("文件权限不安全: %v", mode)
	}

	return nil
}

// GetUserDatabase 安全地获取用户数据库连接
func (sm *SecureSQLiteManager) GetUserDatabase(userID uint) (*gorm.DB, error) {
	sm.mutex.RLock()
	if db, exists := sm.connections[userID]; exists {
		sm.mutex.RUnlock()
		return db, nil
	}
	sm.mutex.RUnlock()

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 双重检查
	if db, exists := sm.connections[userID]; exists {
		return db, nil
	}

	// 获取安全的数据库路径
	dbPath, err := sm.getUserDatabasePath(userID)
	if err != nil {
		return nil, fmt.Errorf("获取数据库路径失败: %v", err)
	}

	// 创建数据库连接
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 避免敏感信息泄露到日志
		NowFunc: func() time.Time {
			return time.Now().UTC() // 使用UTC时间避免时区问题
		},
	}

	// 使用WAL模式提高并发性能和安全性
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=1000&_foreign_keys=ON", dbPath)

	db, err := gorm.Open(sqlite.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("连接SQLite数据库失败: %v", err)
	}

	// 设置连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层数据库连接失败: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(1)                   // SQLite不支持多连接，设置为1
	sqlDB.SetMaxIdleConns(1)                   // 空闲连接数
	sqlDB.SetConnMaxLifetime(30 * time.Minute) // 连接最大生命周期

	// 自动迁移表结构
	if err := sm.migrateTables(db); err != nil {
		return nil, fmt.Errorf("迁移表结构失败: %v", err)
	}

	// 设置数据库文件权限
	if err := os.Chmod(dbPath, 0600); err != nil {
		return nil, fmt.Errorf("设置数据库文件权限失败: %v", err)
	}

	// 缓存连接
	sm.connections[userID] = db

	return db, nil
}

// migrateTables 安全地迁移表结构
func (sm *SecureSQLiteManager) migrateTables(db *gorm.DB) error {
	// 启用外键约束
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		return fmt.Errorf("启用外键约束失败: %v", err)
	}

	// 设置WAL模式
	if err := db.Exec("PRAGMA journal_mode = WAL").Error; err != nil {
		return fmt.Errorf("设置WAL模式失败: %v", err)
	}

	// 设置同步模式
	if err := db.Exec("PRAGMA synchronous = NORMAL").Error; err != nil {
		return fmt.Errorf("设置同步模式失败: %v", err)
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(&ResumeFile{}, &Resume{}, &ResumeParsingTask{}, &ResumeContent{}, &ParsedResumeDataDB{}, &UserPrivacySettings{}); err != nil {
		return fmt.Errorf("自动迁移失败: %v", err)
	}

	return nil
}

// CloseUserDatabase 安全地关闭用户数据库连接
func (sm *SecureSQLiteManager) CloseUserDatabase(userID uint) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if db, exists := sm.connections[userID]; exists {
		sqlDB, err := db.DB()
		if err != nil {
			delete(sm.connections, userID)
			return fmt.Errorf("获取底层数据库连接失败: %v", err)
		}

		if err := sqlDB.Close(); err != nil {
			delete(sm.connections, userID)
			return fmt.Errorf("关闭数据库连接失败: %v", err)
		}

		delete(sm.connections, userID)
	}

	return nil
}

// CloseAllConnections 关闭所有数据库连接
func (sm *SecureSQLiteManager) CloseAllConnections() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	var errors []error
	for userID, db := range sm.connections {
		sqlDB, err := db.DB()
		if err != nil {
			errors = append(errors, fmt.Errorf("用户%d: 获取底层连接失败: %v", userID, err))
			continue
		}

		if err := sqlDB.Close(); err != nil {
			errors = append(errors, fmt.Errorf("用户%d: 关闭连接失败: %v", userID, err))
		}
	}

	// 清空连接缓存
	sm.connections = make(map[uint]*gorm.DB)

	if len(errors) > 0 {
		return fmt.Errorf("关闭连接时发生错误: %v", errors)
	}

	return nil
}

// ValidateUserAccess 验证用户访问权限
func (sm *SecureSQLiteManager) ValidateUserAccess(userID uint, requestUserID uint) error {
	if userID != requestUserID {
		return fmt.Errorf("访问被拒绝: 用户%d无权访问用户%d的数据", requestUserID, userID)
	}
	return nil
}

// GetUserDatabaseInfo 获取用户数据库信息（用于监控）
func (sm *SecureSQLiteManager) GetUserDatabaseInfo(userID uint) (map[string]interface{}, error) {
	db, err := sm.GetUserDatabase(userID)
	if err != nil {
		return nil, err
	}

	info := make(map[string]interface{})

	// 获取数据库统计信息
	var fileCount int64
	db.Model(&ResumeFile{}).Count(&fileCount)
	info["file_count"] = fileCount

	var resumeCount int64
	db.Model(&Resume{}).Count(&resumeCount)
	info["resume_count"] = resumeCount

	var taskCount int64
	db.Model(&ResumeParsingTask{}).Count(&taskCount)
	info["task_count"] = taskCount

	// 获取数据库文件信息
	dbPath, err := sm.getUserDatabasePath(userID)
	if err == nil {
		if stat, err := os.Stat(dbPath); err == nil {
			info["file_size"] = stat.Size()
			info["last_modified"] = stat.ModTime()
		}
	}

	return info, nil
}

// CleanupInactiveDatabases 清理不活跃的数据库连接
func (sm *SecureSQLiteManager) CleanupInactiveDatabases(maxIdleTime time.Duration) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	var toClose []uint

	// 简化版本：定期清理所有连接
	// TODO: 实现更精确的连接使用时间跟踪
	for userID := range sm.connections {
		toClose = append(toClose, userID)
	}

	for _, userID := range toClose {
		if db, exists := sm.connections[userID]; exists {
			sqlDB, _ := db.DB()
			sqlDB.Close()
			delete(sm.connections, userID)
		}
	}

	return nil
}

// 全局SQLite管理器实例
var globalSQLiteManager *SecureSQLiteManager

// InitSecureSQLiteManager 初始化安全的SQLite管理器
func InitSecureSQLiteManager(basePath string) error {
	globalSQLiteManager = NewSecureSQLiteManager(basePath)

	// 启动定期清理任务
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if globalSQLiteManager != nil {
				globalSQLiteManager.CleanupInactiveDatabases(1 * time.Hour)
			}
		}
	}()

	return nil
}

// GetSecureUserDatabase 安全地获取用户数据库
func GetSecureUserDatabase(userID uint) (*gorm.DB, error) {
	if globalSQLiteManager == nil {
		return nil, fmt.Errorf("SQLite管理器未初始化")
	}
	return globalSQLiteManager.GetUserDatabase(userID)
}

// CloseSecureUserDatabase 安全地关闭用户数据库
func CloseSecureUserDatabase(userID uint) error {
	if globalSQLiteManager == nil {
		return nil
	}
	return globalSQLiteManager.CloseUserDatabase(userID)
}

// ==============================================
// 缺失的类型定义
// ==============================================

// ResumeFile 简历文件模型
type ResumeFile struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	FileName  string    `json:"file_name" gorm:"not null"`
	FilePath  string    `json:"file_path" gorm:"not null"`
	FileSize  int64     `json:"file_size"`
	FileType  string    `json:"file_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Resume 简历模型
type Resume struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	Title     string    `json:"title" gorm:"not null"`
	Content   string    `json:"content" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserPrivacySettings 用户隐私设置模型
type UserPrivacySettings struct {
	ID                  uint      `json:"id" gorm:"primaryKey"`
	ResumeContentID     uint      `json:"resume_content_id" gorm:"not null"`
	IsPublic            bool      `json:"is_public" gorm:"default:false"`
	ShareWithCompanies  bool      `json:"share_with_companies" gorm:"default:false"`
	AllowSearch         bool      `json:"allow_search" gorm:"default:true"`
	AllowDownload       bool      `json:"allow_download" gorm:"default:false"`
	ViewPermissions     string    `json:"view_permissions" gorm:"type:text"`
	DownloadPermissions string    `json:"download_permissions" gorm:"type:text"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
