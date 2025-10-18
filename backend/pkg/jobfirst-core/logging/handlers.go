package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ConsoleHandler 控制台处理器
type ConsoleHandler struct {
	format LogFormat
	writer io.Writer
}

// NewConsoleHandler 创建控制台处理器
func NewConsoleHandler(format LogFormat, writer io.Writer) (*ConsoleHandler, error) {
	return &ConsoleHandler{
		format: format,
		writer: writer,
	}, nil
}

// Handle 处理日志条目
func (h *ConsoleHandler) Handle(entry *LogEntry) error {
	var output string
	var err error

	switch h.format {
	case LogFormatJSON:
		output, err = h.formatJSON(entry)
	case LogFormatText:
		output, err = h.formatText(entry)
	default:
		return fmt.Errorf("unsupported format: %s", h.format)
	}

	if err != nil {
		return fmt.Errorf("failed to format log entry: %v", err)
	}

	_, err = fmt.Fprintln(h.writer, output)
	return err
}

// formatJSON 格式化为JSON
func (h *ConsoleHandler) formatJSON(entry *LogEntry) (string, error) {
	data, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// formatText 格式化为文本
func (h *ConsoleHandler) formatText(entry *LogEntry) (string, error) {
	var parts []string

	// 时间戳
	parts = append(parts, entry.Timestamp.Format("2006-01-02 15:04:05.000"))

	// 级别
	parts = append(parts, strings.ToUpper(string(entry.Level)))

	// 服务
	if entry.Service != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.Service))
	}

	// 模块
	if entry.Module != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.Module))
	}

	// 文件位置
	if entry.File != "" && entry.Line > 0 {
		parts = append(parts, fmt.Sprintf("[%s:%d]", entry.File, entry.Line))
	}

	// 跟踪ID
	if entry.TraceID != "" {
		parts = append(parts, fmt.Sprintf("[trace:%s]", entry.TraceID))
	}

	// 用户ID
	if entry.UserID != "" {
		parts = append(parts, fmt.Sprintf("[user:%s]", entry.UserID))
	}

	// 请求ID
	if entry.RequestID != "" {
		parts = append(parts, fmt.Sprintf("[req:%s]", entry.RequestID))
	}

	// 消息
	parts = append(parts, entry.Message)

	// 错误
	if entry.Error != nil {
		parts = append(parts, fmt.Sprintf("error=%v", entry.Error))
	}

	// 字段
	if len(entry.Fields) > 0 {
		var fieldParts []string
		for k, v := range entry.Fields {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", k, v))
		}
		parts = append(parts, strings.Join(fieldParts, " "))
	}

	return strings.Join(parts, " "), nil
}

// Close 关闭处理器
func (h *ConsoleHandler) Close() error {
	return nil
}

// FileHandler 文件处理器
type FileHandler struct {
	config   *LoggerConfig
	file     *os.File
	filePath string
}

// NewFileHandler 创建文件处理器
func NewFileHandler(config *LoggerConfig) (*FileHandler, error) {
	// 确保目录存在
	dir := filepath.Dir(config.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	// 打开文件
	file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	return &FileHandler{
		config:   config,
		file:     file,
		filePath: config.FilePath,
	}, nil
}

// Handle 处理日志条目
func (h *FileHandler) Handle(entry *LogEntry) error {
	var output string
	var err error

	switch h.config.Format {
	case LogFormatJSON:
		output, err = h.formatJSON(entry)
	case LogFormatText:
		output, err = h.formatText(entry)
	default:
		return fmt.Errorf("unsupported format: %s", h.config.Format)
	}

	if err != nil {
		return fmt.Errorf("failed to format log entry: %v", err)
	}

	_, err = fmt.Fprintln(h.file, output)
	return err
}

// formatJSON 格式化为JSON
func (h *FileHandler) formatJSON(entry *LogEntry) (string, error) {
	data, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// formatText 格式化为文本
func (h *FileHandler) formatText(entry *LogEntry) (string, error) {
	var parts []string

	// 时间戳
	if h.config.LocalTime {
		parts = append(parts, entry.Timestamp.Local().Format("2006-01-02 15:04:05.000"))
	} else {
		parts = append(parts, entry.Timestamp.UTC().Format("2006-01-02 15:04:05.000"))
	}

	// 级别
	parts = append(parts, strings.ToUpper(string(entry.Level)))

	// 服务
	if entry.Service != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.Service))
	}

	// 模块
	if entry.Module != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.Module))
	}

	// 文件位置
	if entry.File != "" && entry.Line > 0 {
		parts = append(parts, fmt.Sprintf("[%s:%d]", entry.File, entry.Line))
	}

	// 跟踪ID
	if entry.TraceID != "" {
		parts = append(parts, fmt.Sprintf("[trace:%s]", entry.TraceID))
	}

	// 用户ID
	if entry.UserID != "" {
		parts = append(parts, fmt.Sprintf("[user:%s]", entry.UserID))
	}

	// 请求ID
	if entry.RequestID != "" {
		parts = append(parts, fmt.Sprintf("[req:%s]", entry.RequestID))
	}

	// 消息
	parts = append(parts, entry.Message)

	// 错误
	if entry.Error != nil {
		parts = append(parts, fmt.Sprintf("error=%v", entry.Error))
	}

	// 字段
	if len(entry.Fields) > 0 {
		var fieldParts []string
		for k, v := range entry.Fields {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", k, v))
		}
		parts = append(parts, strings.Join(fieldParts, " "))
	}

	// 堆栈跟踪
	if entry.Stack != "" {
		parts = append(parts, fmt.Sprintf("\n%s", entry.Stack))
	}

	return strings.Join(parts, " "), nil
}

// Close 关闭处理器
func (h *FileHandler) Close() error {
	if h.file != nil {
		return h.file.Close()
	}
	return nil
}

// RotatingFileHandler 轮转文件处理器
type RotatingFileHandler struct {
	config      *LoggerConfig
	file        *os.File
	filePath    string
	currentSize int64
	maxSize     int64
	maxAge      time.Duration
	maxBackups  int
	backups     []string
}

// NewRotatingFileHandler 创建轮转文件处理器
func NewRotatingFileHandler(config *LoggerConfig) (*RotatingFileHandler, error) {
	// 确保目录存在
	dir := filepath.Dir(config.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	// 打开文件
	file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	// 获取当前文件大小
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stat: %v", err)
	}

	handler := &RotatingFileHandler{
		config:      config,
		file:        file,
		filePath:    config.FilePath,
		currentSize: stat.Size(),
		maxSize:     int64(config.MaxSize) * 1024 * 1024,           // MB to bytes
		maxAge:      time.Duration(config.MaxAge) * 24 * time.Hour, // days to duration
		maxBackups:  config.MaxBackups,
		backups:     make([]string, 0),
	}

	// 加载现有备份文件
	if err := handler.loadBackups(); err != nil {
		return nil, fmt.Errorf("failed to load backups: %v", err)
	}

	return handler, nil
}

// Handle 处理日志条目
func (h *RotatingFileHandler) Handle(entry *LogEntry) error {
	// 检查是否需要轮转
	if h.shouldRotate() {
		if err := h.rotate(); err != nil {
			return fmt.Errorf("failed to rotate log file: %v", err)
		}
	}

	// 格式化日志条目
	var output string
	var err error

	switch h.config.Format {
	case LogFormatJSON:
		output, err = h.formatJSON(entry)
	case LogFormatText:
		output, err = h.formatText(entry)
	default:
		return fmt.Errorf("unsupported format: %s", h.config.Format)
	}

	if err != nil {
		return fmt.Errorf("failed to format log entry: %v", err)
	}

	// 写入文件
	_, err = fmt.Fprintln(h.file, output)
	if err != nil {
		return fmt.Errorf("failed to write to log file: %v", err)
	}

	// 更新文件大小
	h.currentSize += int64(len(output)) + 1 // +1 for newline

	return nil
}

// shouldRotate 检查是否需要轮转
func (h *RotatingFileHandler) shouldRotate() bool {
	return h.currentSize >= h.maxSize
}

// rotate 轮转日志文件
func (h *RotatingFileHandler) rotate() error {
	// 关闭当前文件
	if err := h.file.Close(); err != nil {
		return fmt.Errorf("failed to close current log file: %v", err)
	}

	// 重命名当前文件
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	backupPath := fmt.Sprintf("%s.%s", h.filePath, timestamp)

	if err := os.Rename(h.filePath, backupPath); err != nil {
		return fmt.Errorf("failed to rename log file: %v", err)
	}

	// 添加到备份列表
	h.backups = append(h.backups, backupPath)

	// 清理旧备份
	if err := h.cleanupBackups(); err != nil {
		return fmt.Errorf("failed to cleanup backups: %v", err)
	}

	// 创建新文件
	file, err := os.OpenFile(h.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create new log file: %v", err)
	}

	h.file = file
	h.currentSize = 0

	return nil
}

// cleanupBackups 清理旧备份
func (h *RotatingFileHandler) cleanupBackups() error {
	// 按数量清理
	if len(h.backups) > h.maxBackups {
		toDelete := h.backups[:len(h.backups)-h.maxBackups]
		for _, backup := range toDelete {
			if err := os.Remove(backup); err != nil {
				return fmt.Errorf("failed to remove backup file %s: %v", backup, err)
			}
		}
		h.backups = h.backups[len(h.backups)-h.maxBackups:]
	}

	// 按时间清理
	cutoff := time.Now().Add(-h.maxAge)
	for i := len(h.backups) - 1; i >= 0; i-- {
		backup := h.backups[i]
		stat, err := os.Stat(backup)
		if err != nil {
			continue
		}

		if stat.ModTime().Before(cutoff) {
			if err := os.Remove(backup); err != nil {
				return fmt.Errorf("failed to remove old backup file %s: %v", backup, err)
			}
			h.backups = append(h.backups[:i], h.backups[i+1:]...)
		}
	}

	return nil
}

// loadBackups 加载现有备份文件
func (h *RotatingFileHandler) loadBackups() error {
	dir := filepath.Dir(h.filePath)
	base := filepath.Base(h.filePath)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasPrefix(name, base+".") {
			h.backups = append(h.backups, filepath.Join(dir, name))
		}
	}

	return nil
}

// formatJSON 格式化为JSON
func (h *RotatingFileHandler) formatJSON(entry *LogEntry) (string, error) {
	data, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// formatText 格式化为文本
func (h *RotatingFileHandler) formatText(entry *LogEntry) (string, error) {
	var parts []string

	// 时间戳
	if h.config.LocalTime {
		parts = append(parts, entry.Timestamp.Local().Format("2006-01-02 15:04:05.000"))
	} else {
		parts = append(parts, entry.Timestamp.UTC().Format("2006-01-02 15:04:05.000"))
	}

	// 级别
	parts = append(parts, strings.ToUpper(string(entry.Level)))

	// 服务
	if entry.Service != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.Service))
	}

	// 模块
	if entry.Module != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.Module))
	}

	// 文件位置
	if entry.File != "" && entry.Line > 0 {
		parts = append(parts, fmt.Sprintf("[%s:%d]", entry.File, entry.Line))
	}

	// 跟踪ID
	if entry.TraceID != "" {
		parts = append(parts, fmt.Sprintf("[trace:%s]", entry.TraceID))
	}

	// 用户ID
	if entry.UserID != "" {
		parts = append(parts, fmt.Sprintf("[user:%s]", entry.UserID))
	}

	// 请求ID
	if entry.RequestID != "" {
		parts = append(parts, fmt.Sprintf("[req:%s]", entry.RequestID))
	}

	// 消息
	parts = append(parts, entry.Message)

	// 错误
	if entry.Error != nil {
		parts = append(parts, fmt.Sprintf("error=%v", entry.Error))
	}

	// 字段
	if len(entry.Fields) > 0 {
		var fieldParts []string
		for k, v := range entry.Fields {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", k, v))
		}
		parts = append(parts, strings.Join(fieldParts, " "))
	}

	// 堆栈跟踪
	if entry.Stack != "" {
		parts = append(parts, fmt.Sprintf("\n%s", entry.Stack))
	}

	return strings.Join(parts, " "), nil
}

// Close 关闭处理器
func (h *RotatingFileHandler) Close() error {
	if h.file != nil {
		return h.file.Close()
	}
	return nil
}
