package logging

import (
	"time"
)

// LogLevel 日志级别
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
	LogLevelPanic LogLevel = "panic"
)

// LogFormat 日志格式
type LogFormat string

const (
	LogFormatJSON LogFormat = "json"
	LogFormatText LogFormat = "text"
)

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Service   string                 `json:"service,omitempty"`
	Module    string                 `json:"module,omitempty"`
	Function  string                 `json:"function,omitempty"`
	File      string                 `json:"file,omitempty"`
	Line      int                    `json:"line,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	SpanID    string                 `json:"span_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Error     error                  `json:"error,omitempty"`
	Stack     string                 `json:"stack,omitempty"`
}

// Logger 日志记录器接口
type Logger interface {
	Debug(msg string, fields ...map[string]interface{})
	Info(msg string, fields ...map[string]interface{})
	Warn(msg string, fields ...map[string]interface{})
	Error(msg string, err error, fields ...map[string]interface{})
	Fatal(msg string, err error, fields ...map[string]interface{})
	Panic(msg string, err error, fields ...map[string]interface{})

	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithService(service string) Logger
	WithModule(module string) Logger
	WithTraceID(traceID string) Logger
	WithUserID(userID string) Logger
	WithRequestID(requestID string) Logger

	SetLevel(level LogLevel)
	GetLevel() LogLevel
}

// LogHandler 日志处理器接口
type LogHandler interface {
	Handle(entry *LogEntry) error
	Close() error
}

// LoggerConfig 日志记录器配置
type LoggerConfig struct {
	Level      LogLevel        `json:"level"`
	Format     LogFormat       `json:"format"`
	Service    string          `json:"service"`
	Module     string          `json:"module"`
	Output     []string        `json:"output"` // stdout, stderr, file, syslog
	FilePath   string          `json:"file_path,omitempty"`
	MaxSize    int             `json:"max_size,omitempty"` // MB
	MaxAge     int             `json:"max_age,omitempty"`  // days
	MaxBackups int             `json:"max_backups,omitempty"`
	Compress   bool            `json:"compress"`
	LocalTime  bool            `json:"local_time"`
	Caller     bool            `json:"caller"`
	Stacktrace bool            `json:"stacktrace"`
	Sampling   *SamplingConfig `json:"sampling,omitempty"`
}

// SamplingConfig 采样配置
type SamplingConfig struct {
	Initial    int `json:"initial"`
	Thereafter int `json:"thereafter"`
}

// LogMetrics 日志指标
type LogMetrics struct {
	TotalLogs     int64              `json:"total_logs"`
	LogsByLevel   map[LogLevel]int64 `json:"logs_by_level"`
	LogsByService map[string]int64   `json:"logs_by_service"`
	ErrorRate     float64            `json:"error_rate"`
	LastLogTime   time.Time          `json:"last_log_time"`
}

// LogFilter 日志过滤器
type LogFilter struct {
	Levels    []LogLevel `json:"levels,omitempty"`
	Services  []string   `json:"services,omitempty"`
	Modules   []string   `json:"modules,omitempty"`
	Keywords  []string   `json:"keywords,omitempty"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
}

// LogQuery 日志查询
type LogQuery struct {
	Filter   *LogFilter `json:"filter,omitempty"`
	Limit    int        `json:"limit,omitempty"`
	Offset   int        `json:"offset,omitempty"`
	OrderBy  string     `json:"order_by,omitempty"`
	OrderDir string     `json:"order_dir,omitempty"` // asc, desc
}

// LogSearchResult 日志搜索结果
type LogSearchResult struct {
	Entries []*LogEntry `json:"entries"`
	Total   int64       `json:"total"`
	Limit   int         `json:"limit"`
	Offset  int         `json:"offset"`
}
