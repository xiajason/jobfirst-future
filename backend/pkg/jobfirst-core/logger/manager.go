package logger

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

// Level 日志级别
type Level string

const (
	LevelTrace Level = "trace"
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
	LevelFatal Level = "fatal"
	LevelPanic Level = "panic"
)

// Format 日志格式
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// Config 日志配置
type Config struct {
	Level  Level  `json:"level"`
	Format Format `json:"format"`
	Output string `json:"output"` // stdout, stderr, file
	File   string `json:"file"`
}

// Manager 日志管理器
type Manager struct {
	logger *logrus.Logger
	config Config
}

// NewManager 创建日志管理器
func NewManager(config Config) (*Manager, error) {
	logger := logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(string(config.Level))
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// 设置日志格式
	switch config.Format {
	case FormatJSON:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	case FormatText:
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.RFC3339,
			FullTimestamp:   true,
		})
	default:
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.RFC3339,
			FullTimestamp:   true,
		})
	}

	// 设置输出
	switch config.Output {
	case "stdout":
		logger.SetOutput(os.Stdout)
	case "stderr":
		logger.SetOutput(os.Stderr)
	case "file":
		if config.File != "" {
			// 确保日志目录存在
			dir := filepath.Dir(config.File)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, err
			}

			file, err := os.OpenFile(config.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				return nil, err
			}
			logger.SetOutput(file)
		} else {
			logger.SetOutput(os.Stdout)
		}
	default:
		logger.SetOutput(os.Stdout)
	}

	manager := &Manager{
		logger: logger,
		config: config,
	}

	return manager, nil
}

// GetLogger 获取日志实例
func (lm *Manager) GetLogger() *logrus.Logger {
	return lm.logger
}

// Trace 记录Trace级别日志
func (lm *Manager) Trace(args ...interface{}) {
	lm.logger.Trace(args...)
}

// Tracef 记录Trace级别格式化日志
func (lm *Manager) Tracef(format string, args ...interface{}) {
	lm.logger.Tracef(format, args...)
}

// Debug 记录Debug级别日志
func (lm *Manager) Debug(args ...interface{}) {
	lm.logger.Debug(args...)
}

// Debugf 记录Debug级别格式化日志
func (lm *Manager) Debugf(format string, args ...interface{}) {
	lm.logger.Debugf(format, args...)
}

// Info 记录Info级别日志
func (lm *Manager) Info(args ...interface{}) {
	lm.logger.Info(args...)
}

// Infof 记录Info级别格式化日志
func (lm *Manager) Infof(format string, args ...interface{}) {
	lm.logger.Infof(format, args...)
}

// Warn 记录Warn级别日志
func (lm *Manager) Warn(args ...interface{}) {
	lm.logger.Warn(args...)
}

// Warnf 记录Warn级别格式化日志
func (lm *Manager) Warnf(format string, args ...interface{}) {
	lm.logger.Warnf(format, args...)
}

// Error 记录Error级别日志
func (lm *Manager) Error(args ...interface{}) {
	lm.logger.Error(args...)
}

// Errorf 记录Error级别格式化日志
func (lm *Manager) Errorf(format string, args ...interface{}) {
	lm.logger.Errorf(format, args...)
}

// Fatal 记录Fatal级别日志并退出程序
func (lm *Manager) Fatal(args ...interface{}) {
	lm.logger.Fatal(args...)
}

// Fatalf 记录Fatal级别格式化日志并退出程序
func (lm *Manager) Fatalf(format string, args ...interface{}) {
	lm.logger.Fatalf(format, args...)
}

// Panic 记录Panic级别日志并panic
func (lm *Manager) Panic(args ...interface{}) {
	lm.logger.Panic(args...)
}

// Panicf 记录Panic级别格式化日志并panic
func (lm *Manager) Panicf(format string, args ...interface{}) {
	lm.logger.Panicf(format, args...)
}

// WithField 添加字段到日志
func (lm *Manager) WithField(key string, value interface{}) *logrus.Entry {
	return lm.logger.WithField(key, value)
}

// WithFields 添加多个字段到日志
func (lm *Manager) WithFields(fields logrus.Fields) *logrus.Entry {
	return lm.logger.WithFields(fields)
}

// WithError 添加错误到日志
func (lm *Manager) WithError(err error) *logrus.Entry {
	return lm.logger.WithError(err)
}

// SetLevel 设置日志级别
func (lm *Manager) SetLevel(level Level) error {
	logLevel, err := logrus.ParseLevel(string(level))
	if err != nil {
		return err
	}
	lm.logger.SetLevel(logLevel)
	lm.config.Level = level
	return nil
}

// SetOutput 设置日志输出
func (lm *Manager) SetOutput(output io.Writer) {
	lm.logger.SetOutput(output)
}

// SetFormatter 设置日志格式
func (lm *Manager) SetFormatter(formatter logrus.Formatter) {
	lm.logger.SetFormatter(formatter)
}

// 全局日志实例
var globalLogger *Manager

// InitGlobal 初始化全局日志实例
func InitGlobal(config Config) error {
	manager, err := NewManager(config)
	if err != nil {
		return err
	}
	globalLogger = manager
	return nil
}

// GetGlobal 获取全局日志实例
func GetGlobal() *Manager {
	if globalLogger == nil {
		// 使用默认配置
		config := Config{
			Level:  LevelInfo,
			Format: FormatText,
			Output: "stdout",
		}
		globalLogger, _ = NewManager(config)
	}
	return globalLogger
}

// 全局日志函数
func Trace(args ...interface{}) {
	GetGlobal().Trace(args...)
}

func Tracef(format string, args ...interface{}) {
	GetGlobal().Tracef(format, args...)
}

func Debug(args ...interface{}) {
	GetGlobal().Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	GetGlobal().Debugf(format, args...)
}

func Info(args ...interface{}) {
	GetGlobal().Info(args...)
}

func Infof(format string, args ...interface{}) {
	GetGlobal().Infof(format, args...)
}

func Warn(args ...interface{}) {
	GetGlobal().Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	GetGlobal().Warnf(format, args...)
}

func Error(args ...interface{}) {
	GetGlobal().Error(args...)
}

func Errorf(format string, args ...interface{}) {
	GetGlobal().Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	GetGlobal().Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	GetGlobal().Fatalf(format, args...)
}

func Panic(args ...interface{}) {
	GetGlobal().Panic(args...)
}

func Panicf(format string, args ...interface{}) {
	GetGlobal().Panicf(format, args...)
}

func WithField(key string, value interface{}) *logrus.Entry {
	return GetGlobal().WithField(key, value)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return GetGlobal().WithFields(fields)
}

func WithError(err error) *logrus.Entry {
	return GetGlobal().WithError(err)
}
