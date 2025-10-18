package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// MD5Hash 计算MD5哈希
func MD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) (string, error) {
	if length <= 0 || length > 100 {
		return "", fmt.Errorf("length must be between 1 and 100")
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes)[:length], nil
}

// FormatJSON 格式化JSON
func FormatJSON(jsonStr string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", err
	}

	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return string(formatted), nil
}

// GenerateToken 生成简单的访问令牌
func GenerateToken(prefix string) string {
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	return prefix + "-" + hex.EncodeToString(randomBytes)
}

// ValidateToken 验证令牌（简化版本）
func ValidateToken(token string) bool {
	if token == "" {
		return false
	}

	// 开发环境的简化验证
	validTokens := []string{"test-token", "wx-token-123", "admin-token-123"}
	for _, validToken := range validTokens {
		if token == validToken {
			return true
		}
	}

	// 检查令牌格式
	if strings.Contains(token, "-") && len(token) > 10 {
		return true
	}

	return false
}

// ExtractUserID 从令牌中提取用户ID（简化版本）
func ExtractUserID(token string) string {
	// 开发环境的简化实现
	switch token {
	case "test-token":
		return "1"
	case "wx-token-123":
		return "2"
	case "admin-token-123":
		return "999"
	default:
		// 从令牌中提取用户ID的逻辑
		if strings.Contains(token, "-") {
			parts := strings.Split(token, "-")
			if len(parts) > 1 {
				return parts[len(parts)-1][:8] // 取最后一部分的前8位作为用户ID
			}
		}
		return "0"
	}
}

// IsWhitelistPath 检查是否为白名单路径
func IsWhitelistPath(path string, whitelist []string) bool {
	for _, whitePath := range whitelist {
		if strings.HasPrefix(path, whitePath) {
			return true
		}
	}
	return false
}

// GetDefaultWhitelist 获取默认白名单路径
func GetDefaultWhitelist() []string {
	return []string{
		"/health",
		"/version",
		"/v2/api-docs",
		"/swagger/",
		"/metrics",
		"/utils/",
		"/monitor/",
		"/config/",
	}
}

// FormatTimestamp 格式化时间戳
func FormatTimestamp(timestamp time.Time) string {
	return timestamp.Format("2006-01-02 15:04:05")
}

// ParseDuration 解析持续时间字符串
func ParseDuration(durationStr string) (time.Duration, error) {
	return time.ParseDuration(durationStr)
}

// TruncateString 截断字符串
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// IsValidEmail 验证邮箱格式（简单验证）
func IsValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// IsValidPhone 验证手机号格式（简单验证）
func IsValidPhone(phone string) bool {
	// 移除所有非数字字符
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")

	return len(phone) >= 10 && len(phone) <= 15
}
