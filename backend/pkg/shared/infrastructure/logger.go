package infrastructure

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// LogLevel 日志级别
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// Logger 日志接口
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)

	WithContext(ctx context.Context) Logger
	WithFields(fields ...Field) Logger

	SetLevel(level LogLevel)
	SetOutput(output io.Writer)
}

// LogrusLogger Logrus日志实现
type LogrusLogger struct {
	logger *logrus.Logger
	fields logrus.Fields
}

// NewLogrusLogger 创建Logrus日志器
func NewLogrusLogger() *LogrusLogger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	return &LogrusLogger{
		logger: logger,
		fields: make(logrus.Fields),
	}
}

// Debug 调试日志
func (l *LogrusLogger) Debug(msg string, fields ...Field) {
	l.logger.WithFields(l.mergeFields(fields)).Debug(msg)
}

// Info 信息日志
func (l *LogrusLogger) Info(msg string, fields ...Field) {
	l.logger.WithFields(l.mergeFields(fields)).Info(msg)
}

// Warn 警告日志
func (l *LogrusLogger) Warn(msg string, fields ...Field) {
	l.logger.WithFields(l.mergeFields(fields)).Warn(msg)
}

// Error 错误日志
func (l *LogrusLogger) Error(msg string, fields ...Field) {
	l.logger.WithFields(l.mergeFields(fields)).Error(msg)
}

// Fatal 致命错误日志
func (l *LogrusLogger) Fatal(msg string, fields ...Field) {
	l.logger.WithFields(l.mergeFields(fields)).Fatal(msg)
}

// WithContext 添加上下文
func (l *LogrusLogger) WithContext(ctx context.Context) Logger {
	// 从上下文提取追踪ID
	if traceID := getTraceIDFromContext(ctx); traceID != "" {
		newLogger := &LogrusLogger{
			logger: l.logger,
			fields: l.mergeFields([]Field{{Key: "trace_id", Value: traceID}}),
		}
		return newLogger
	}
	return l
}

// WithFields 添加字段
func (l *LogrusLogger) WithFields(fields ...Field) Logger {
	newLogger := &LogrusLogger{
		logger: l.logger,
		fields: l.mergeFields(fields),
	}
	return newLogger
}

// SetLevel 设置日志级别
func (l *LogrusLogger) SetLevel(level LogLevel) {
	switch level {
	case DebugLevel:
		l.logger.SetLevel(logrus.DebugLevel)
	case InfoLevel:
		l.logger.SetLevel(logrus.InfoLevel)
	case WarnLevel:
		l.logger.SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		l.logger.SetLevel(logrus.ErrorLevel)
	case FatalLevel:
		l.logger.SetLevel(logrus.FatalLevel)
	}
}

// SetOutput 设置输出
func (l *LogrusLogger) SetOutput(output io.Writer) {
	l.logger.SetOutput(output)
}

// mergeFields 合并字段
func (l *LogrusLogger) mergeFields(fields []Field) logrus.Fields {
	result := make(logrus.Fields)

	// 复制现有字段
	for k, v := range l.fields {
		result[k] = v
	}

	// 添加新字段
	for _, field := range fields {
		result[field.Key] = field.Value
	}

	return result
}

// getTraceIDFromContext 从上下文获取追踪ID
func getTraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// 这里可以集成Jaeger或其他追踪系统
	// 暂时返回空字符串
	return ""
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      LogLevel `yaml:"level" json:"level"`
	Output     string   `yaml:"output" json:"output"`
	Format     string   `yaml:"format" json:"format"`
	MaxSize    int      `yaml:"max_size" json:"max_size"`
	MaxBackups int      `yaml:"max_backups" json:"max_backups"`
	MaxAge     int      `yaml:"max_age" json:"max_age"`
	Compress   bool     `yaml:"compress" json:"compress"`
}

// NewLogger 创建日志器
func NewLogger(config *LoggerConfig) Logger {
	logger := NewLogrusLogger()

	if config != nil {
		logger.SetLevel(config.Level)

		// 设置输出文件
		if config.Output != "" && config.Output != "stdout" {
			file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err == nil {
				logger.SetOutput(file)
			}
		}
	}

	return logger
}

// 全局日志器实例
var globalLogger Logger

// InitGlobalLogger 初始化全局日志器
func InitGlobalLogger(config *LoggerConfig) {
	globalLogger = NewLogger(config)
}

// GetLogger 获取全局日志器
func GetLogger() Logger {
	if globalLogger == nil {
		globalLogger = NewLogrusLogger()
	}
	return globalLogger
}

// 便捷函数
func Debug(msg string, fields ...Field) {
	GetLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	GetLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	GetLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	GetLogger().Error(msg, fields...)
}

func Fatal(msg string, fields ...Field) {
	GetLogger().Fatal(msg, fields...)
}
