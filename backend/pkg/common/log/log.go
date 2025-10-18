package log

import (
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

// LogManager 日志管理器
type LogManager struct {
	logger *logrus.Logger
	config *LogConfig
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `json:"level"`       // 日志级别
	Format     string `json:"format"`      // 日志格式 (json/text)
	Output     string `json:"output"`      // 输出目标 (file/stdout)
	FilePath   string `json:"file_path"`   // 文件路径
	MaxSize    int    `json:"max_size"`    // 最大文件大小(MB)
	MaxBackups int    `json:"max_backups"` // 最大备份数量
	MaxAge     int    `json:"max_age"`     // 最大保存天数
	Compress   bool   `json:"compress"`    // 是否压缩
}

// DefaultLogConfig 默认日志配置
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:      "info",
		Format:     "json",
		Output:     "stdout",
		FilePath:   "logs/app.log",
		MaxSize:    100,
		MaxBackups: 10,
		MaxAge:     30,
		Compress:   true,
	}
}

// NewLogManager 创建日志管理器
func NewLogManager(config *LogConfig) (*LogManager, error) {
	if config == nil {
		config = DefaultLogConfig()
	}

	lm := &LogManager{
		logger: logrus.New(),
		config: config,
	}

	// 设置日志级别
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	lm.logger.SetLevel(level)

	// 设置日志格式
	if config.Format == "json" {
		lm.logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		lm.logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	// 设置输出
	if config.Output == "file" {
		if err := lm.setupFileOutput(); err != nil {
			return nil, err
		}
	} else {
		lm.logger.SetOutput(os.Stdout)
	}

	return lm, nil
}

// setupFileOutput 设置文件输出
func (l *LogManager) setupFileOutput() error {
	// 确保目录存在
	dir := filepath.Dir(l.config.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 打开文件
	file, err := os.OpenFile(l.config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	l.logger.SetOutput(file)
	return nil
}

// GetLogger 获取日志实例
func (l *LogManager) GetLogger() *logrus.Logger {
	return l.logger
}

// Info 信息日志
func (l *LogManager) Info(args ...interface{}) {
	l.logger.Info(args...)
}

// Warn 警告日志
func (l *LogManager) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

// Error 错误日志
func (l *LogManager) Error(args ...interface{}) {
	l.logger.Error(args...)
}

// Debug 调试日志
func (l *LogManager) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

// Fatal 致命错误日志
func (l *LogManager) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

// WithField 添加字段
func (l *LogManager) WithField(key string, value interface{}) *logrus.Entry {
	return l.logger.WithField(key, value)
}

// WithFields 添加多个字段
func (l *LogManager) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.logger.WithFields(fields)
}
