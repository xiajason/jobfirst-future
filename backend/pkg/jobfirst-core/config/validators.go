package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// StringValidator 字符串验证器
type StringValidator struct {
	MinLength int
	MaxLength int
	Pattern   string
	Required  bool
}

func (v *StringValidator) Validate(value interface{}) error {
	if value == nil {
		if v.Required {
			return fmt.Errorf("value is required")
		}
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}

	if v.MinLength > 0 && len(str) < v.MinLength {
		return fmt.Errorf("string length must be at least %d", v.MinLength)
	}

	if v.MaxLength > 0 && len(str) > v.MaxLength {
		return fmt.Errorf("string length must be at most %d", v.MaxLength)
	}

	if v.Pattern != "" {
		matched, err := regexp.MatchString(v.Pattern, str)
		if err != nil {
			return fmt.Errorf("invalid pattern: %v", err)
		}
		if !matched {
			return fmt.Errorf("string does not match pattern: %s", v.Pattern)
		}
	}

	return nil
}

func (v *StringValidator) GetName() string {
	return "string"
}

// NumberValidator 数字验证器
type NumberValidator struct {
	Min     float64
	Max     float64
	Integer bool
}

func (v *NumberValidator) Validate(value interface{}) error {
	if value == nil {
		return nil
	}

	var num float64
	switch val := value.(type) {
	case int:
		num = float64(val)
	case int32:
		num = float64(val)
	case int64:
		num = float64(val)
	case float32:
		num = float64(val)
	case float64:
		num = val
	default:
		return fmt.Errorf("value must be a number")
	}

	if v.Min != 0 && num < v.Min {
		return fmt.Errorf("number must be at least %f", v.Min)
	}

	if v.Max != 0 && num > v.Max {
		return fmt.Errorf("number must be at most %f", v.Max)
	}

	if v.Integer && num != float64(int64(num)) {
		return fmt.Errorf("number must be an integer")
	}

	return nil
}

func (v *NumberValidator) GetName() string {
	return "number"
}

// BooleanValidator 布尔验证器
type BooleanValidator struct{}

func (v *BooleanValidator) Validate(value interface{}) error {
	if value == nil {
		return nil
	}

	_, ok := value.(bool)
	if !ok {
		return fmt.Errorf("value must be a boolean")
	}

	return nil
}

func (v *BooleanValidator) GetName() string {
	return "boolean"
}

// PortValidator 端口验证器
type PortValidator struct{}

func (v *PortValidator) Validate(value interface{}) error {
	if value == nil {
		return nil
	}

	var port int
	switch val := value.(type) {
	case int:
		port = val
	case int32:
		port = int(val)
	case int64:
		port = int(val)
	case float32:
		port = int(val)
	case float64:
		port = int(val)
	case string:
		p, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid port number: %v", err)
		}
		port = p
	default:
		return fmt.Errorf("port must be a number")
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	return nil
}

func (v *PortValidator) GetName() string {
	return "port"
}

// HostValidator 主机验证器
type HostValidator struct{}

func (v *HostValidator) Validate(value interface{}) error {
	if value == nil {
		return nil
	}

	host, ok := value.(string)
	if !ok {
		return fmt.Errorf("host must be a string")
	}

	if host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	// 简单的IP地址或域名验证
	if host == "localhost" {
		return nil
	}

	// 检查是否是有效的IP地址格式
	ipPattern := `^(\d{1,3}\.){3}\d{1,3}$`
	matched, err := regexp.MatchString(ipPattern, host)
	if err != nil {
		return fmt.Errorf("invalid host format: %v", err)
	}

	if matched {
		// 验证IP地址的每个部分
		parts := strings.Split(host, ".")
		for _, part := range parts {
			num, err := strconv.Atoi(part)
			if err != nil || num < 0 || num > 255 {
				return fmt.Errorf("invalid IP address: %s", host)
			}
		}
		return nil
	}

	// 检查是否是有效的域名格式
	domainPattern := `^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`
	matched, err = regexp.MatchString(domainPattern, host)
	if err != nil {
		return fmt.Errorf("invalid host format: %v", err)
	}

	if !matched {
		return fmt.Errorf("invalid host format: %s", host)
	}

	return nil
}

func (v *HostValidator) GetName() string {
	return "host"
}

// LogLevelValidator 日志级别验证器
type LogLevelValidator struct{}

func (v *LogLevelValidator) Validate(value interface{}) error {
	if value == nil {
		return nil
	}

	level, ok := value.(string)
	if !ok {
		return fmt.Errorf("log level must be a string")
	}

	validLevels := []string{"debug", "info", "warn", "warning", "error", "fatal", "panic"}
	level = strings.ToLower(level)

	for _, validLevel := range validLevels {
		if level == validLevel {
			return nil
		}
	}

	return fmt.Errorf("invalid log level: %s. Valid levels are: %s", level, strings.Join(validLevels, ", "))
}

func (v *LogLevelValidator) GetName() string {
	return "log_level"
}

// URLValidator URL验证器
type URLValidator struct{}

func (v *URLValidator) Validate(value interface{}) error {
	if value == nil {
		return nil
	}

	url, ok := value.(string)
	if !ok {
		return fmt.Errorf("URL must be a string")
	}

	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// 简单的URL格式验证
	urlPattern := `^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(/.*)?$`
	matched, err := regexp.MatchString(urlPattern, url)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}

	if !matched {
		return fmt.Errorf("invalid URL format: %s", url)
	}

	return nil
}

func (v *URLValidator) GetName() string {
	return "url"
}

// NewStringValidator 创建字符串验证器
func NewStringValidator(minLength, maxLength int, pattern string, required bool) *StringValidator {
	return &StringValidator{
		MinLength: minLength,
		MaxLength: maxLength,
		Pattern:   pattern,
		Required:  required,
	}
}

// NewNumberValidator 创建数字验证器
func NewNumberValidator(min, max float64, integer bool) *NumberValidator {
	return &NumberValidator{
		Min:     min,
		Max:     max,
		Integer: integer,
	}
}

// NewPortValidator 创建端口验证器
func NewPortValidator() *PortValidator {
	return &PortValidator{}
}

// NewHostValidator 创建主机验证器
func NewHostValidator() *HostValidator {
	return &HostValidator{}
}

// NewLogLevelValidator 创建日志级别验证器
func NewLogLevelValidator() *LogLevelValidator {
	return &LogLevelValidator{}
}

// NewURLValidator 创建URL验证器
func NewURLValidator() *URLValidator {
	return &URLValidator{}
}
