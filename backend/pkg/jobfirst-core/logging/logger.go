package logging

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// StandardLogger 标准日志记录器实现
type StandardLogger struct {
	config    *LoggerConfig
	handlers  []LogHandler
	fields    map[string]interface{}
	service   string
	module    string
	traceID   string
	userID    string
	requestID string
	mutex     sync.RWMutex
	metrics   *LogMetrics
}

// NewStandardLogger 创建标准日志记录器
func NewStandardLogger(config *LoggerConfig) (*StandardLogger, error) {
	if config == nil {
		config = &LoggerConfig{
			Level:      LogLevelInfo,
			Format:     LogFormatJSON,
			Service:    "unknown",
			Module:     "unknown",
			Output:     []string{"stdout"},
			Caller:     true,
			Stacktrace: false,
		}
	}

	logger := &StandardLogger{
		config:   config,
		handlers: make([]LogHandler, 0),
		fields:   make(map[string]interface{}),
		service:  config.Service,
		module:   config.Module,
		metrics: &LogMetrics{
			LogsByLevel:   make(map[LogLevel]int64),
			LogsByService: make(map[string]int64),
		},
	}

	// 初始化处理器
	if err := logger.initHandlers(); err != nil {
		return nil, fmt.Errorf("failed to initialize handlers: %v", err)
	}

	return logger, nil
}

// initHandlers 初始化处理器
func (l *StandardLogger) initHandlers() error {
	for _, output := range l.config.Output {
		var handler LogHandler
		var err error

		switch output {
		case "stdout":
			handler, err = NewConsoleHandler(l.config.Format, os.Stdout)
		case "stderr":
			handler, err = NewConsoleHandler(l.config.Format, os.Stderr)
		case "file":
			if l.config.FilePath == "" {
				l.config.FilePath = "./logs/app.log"
			}
			handler, err = NewFileHandler(l.config)
		default:
			return fmt.Errorf("unsupported output: %s", output)
		}

		if err != nil {
			return fmt.Errorf("failed to create handler for %s: %v", output, err)
		}

		l.handlers = append(l.handlers, handler)
	}

	return nil
}

// Debug 记录调试日志
func (l *StandardLogger) Debug(msg string, fields ...map[string]interface{}) {
	l.log(LogLevelDebug, msg, nil, fields...)
}

// Info 记录信息日志
func (l *StandardLogger) Info(msg string, fields ...map[string]interface{}) {
	l.log(LogLevelInfo, msg, nil, fields...)
}

// Warn 记录警告日志
func (l *StandardLogger) Warn(msg string, fields ...map[string]interface{}) {
	l.log(LogLevelWarn, msg, nil, fields...)
}

// Error 记录错误日志
func (l *StandardLogger) Error(msg string, err error, fields ...map[string]interface{}) {
	l.log(LogLevelError, msg, err, fields...)
}

// Fatal 记录致命错误日志
func (l *StandardLogger) Fatal(msg string, err error, fields ...map[string]interface{}) {
	l.log(LogLevelFatal, msg, err, fields...)
	os.Exit(1)
}

// Panic 记录恐慌日志
func (l *StandardLogger) Panic(msg string, err error, fields ...map[string]interface{}) {
	l.log(LogLevelPanic, msg, err, fields...)
	panic(fmt.Sprintf("%s: %v", msg, err))
}

// log 记录日志的核心方法
func (l *StandardLogger) log(level LogLevel, msg string, err error, fields ...map[string]interface{}) {
	// 检查日志级别
	if !l.shouldLog(level) {
		return
	}

	// 创建日志条目
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		Service:   l.service,
		Module:    l.module,
		TraceID:   l.traceID,
		UserID:    l.userID,
		RequestID: l.requestID,
		Error:     err,
		Fields:    make(map[string]interface{}),
	}

	// 添加字段
	l.mutex.RLock()
	for k, v := range l.fields {
		entry.Fields[k] = v
	}
	l.mutex.RUnlock()

	// 添加额外字段
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			entry.Fields[k] = v
		}
	}

	// 添加调用者信息
	if l.config.Caller {
		if file, line, ok := l.getCaller(); ok {
			entry.File = file
			entry.Line = line
		}
	}

	// 添加堆栈跟踪
	if l.config.Stacktrace && (level == LogLevelError || level == LogLevelFatal || level == LogLevelPanic) {
		entry.Stack = l.getStackTrace()
	}

	// 发送到处理器
	for _, handler := range l.handlers {
		if err := handler.Handle(entry); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to handle log entry: %v\n", err)
		}
	}

	// 更新指标
	l.updateMetrics(entry)
}

// shouldLog 检查是否应该记录日志
func (l *StandardLogger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
		LogLevelFatal: 4,
		LogLevelPanic: 5,
	}

	return levels[level] >= levels[l.config.Level]
}

// getCaller 获取调用者信息
func (l *StandardLogger) getCaller() (string, int, bool) {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "", 0, false
	}

	// 简化文件路径
	parts := strings.Split(file, "/")
	if len(parts) > 2 {
		file = strings.Join(parts[len(parts)-2:], "/")
	}

	return file, line, true
}

// getStackTrace 获取堆栈跟踪
func (l *StandardLogger) getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// updateMetrics 更新指标
func (l *StandardLogger) updateMetrics(entry *LogEntry) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.metrics.TotalLogs++
	l.metrics.LogsByLevel[entry.Level]++
	l.metrics.LogsByService[entry.Service]++
	l.metrics.LastLogTime = entry.Timestamp

	// 计算错误率
	if entry.Level == LogLevelError || entry.Level == LogLevelFatal || entry.Level == LogLevelPanic {
		l.metrics.ErrorRate = float64(l.metrics.LogsByLevel[LogLevelError]+l.metrics.LogsByLevel[LogLevelFatal]+l.metrics.LogsByLevel[LogLevelPanic]) / float64(l.metrics.TotalLogs) * 100
	}
}

// WithField 添加字段
func (l *StandardLogger) WithField(key string, value interface{}) Logger {
	newLogger := l.copy()
	newLogger.mutex.Lock()
	newLogger.fields[key] = value
	newLogger.mutex.Unlock()
	return newLogger
}

// WithFields 添加多个字段
func (l *StandardLogger) WithFields(fields map[string]interface{}) Logger {
	newLogger := l.copy()
	newLogger.mutex.Lock()
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	newLogger.mutex.Unlock()
	return newLogger
}

// WithService 设置服务名
func (l *StandardLogger) WithService(service string) Logger {
	newLogger := l.copy()
	newLogger.service = service
	return newLogger
}

// WithModule 设置模块名
func (l *StandardLogger) WithModule(module string) Logger {
	newLogger := l.copy()
	newLogger.module = module
	return newLogger
}

// WithTraceID 设置跟踪ID
func (l *StandardLogger) WithTraceID(traceID string) Logger {
	newLogger := l.copy()
	newLogger.traceID = traceID
	return newLogger
}

// WithUserID 设置用户ID
func (l *StandardLogger) WithUserID(userID string) Logger {
	newLogger := l.copy()
	newLogger.userID = userID
	return newLogger
}

// WithRequestID 设置请求ID
func (l *StandardLogger) WithRequestID(requestID string) Logger {
	newLogger := l.copy()
	newLogger.requestID = requestID
	return newLogger
}

// SetLevel 设置日志级别
func (l *StandardLogger) SetLevel(level LogLevel) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.config.Level = level
}

// GetLevel 获取日志级别
func (l *StandardLogger) GetLevel() LogLevel {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.config.Level
}

// GetMetrics 获取日志指标
func (l *StandardLogger) GetMetrics() *LogMetrics {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	// 返回副本
	metrics := &LogMetrics{
		TotalLogs:     l.metrics.TotalLogs,
		LogsByLevel:   make(map[LogLevel]int64),
		LogsByService: make(map[string]int64),
		ErrorRate:     l.metrics.ErrorRate,
		LastLogTime:   l.metrics.LastLogTime,
	}

	for k, v := range l.metrics.LogsByLevel {
		metrics.LogsByLevel[k] = v
	}

	for k, v := range l.metrics.LogsByService {
		metrics.LogsByService[k] = v
	}

	return metrics
}

// Close 关闭日志记录器
func (l *StandardLogger) Close() error {
	for _, handler := range l.handlers {
		if err := handler.Close(); err != nil {
			return fmt.Errorf("failed to close handler: %v", err)
		}
	}
	return nil
}

// copy 复制日志记录器
func (l *StandardLogger) copy() *StandardLogger {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	newLogger := &StandardLogger{
		config:    l.config,
		handlers:  l.handlers, // 共享处理器
		fields:    make(map[string]interface{}),
		service:   l.service,
		module:    l.module,
		traceID:   l.traceID,
		userID:    l.userID,
		requestID: l.requestID,
		metrics:   l.metrics, // 共享指标
	}

	// 复制字段
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	return newLogger
}
