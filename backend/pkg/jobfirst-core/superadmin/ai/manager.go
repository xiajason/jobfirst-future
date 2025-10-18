package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"superadmin/errors"
)

// Manager AI服务管理器
type Manager struct {
	config *AIConfig
}

// AIConfig AI配置
type AIConfig struct {
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
	BaseURL  string `json:"base_url"`
	Model    string `json:"model"`
	Port     int    `json:"port"`
}

// NewManager 创建AI服务管理器
func NewManager(config *AIConfig) *Manager {
	return &Manager{
		config: config,
	}
}

// AIServiceStatus AI服务状态
type AIServiceStatus struct {
	Service   *ServiceInfo  `json:"service"`
	Provider  *ProviderInfo `json:"provider"`
	Model     *ModelInfo    `json:"model"`
	Health    *HealthInfo   `json:"health"`
	LastCheck time.Time     `json:"last_check"`
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Port      int       `json:"port"`
	PID       int       `json:"pid,omitempty"`
	Uptime    string    `json:"uptime,omitempty"`
	Version   string    `json:"version,omitempty"`
	LastCheck time.Time `json:"last_check"`
	Error     string    `json:"error,omitempty"`
}

// ProviderInfo 提供商信息
type ProviderInfo struct {
	Name    string `json:"name"`
	APIKey  string `json:"api_key_masked"`
	BaseURL string `json:"base_url"`
	Status  string `json:"status"`
}

// ModelInfo 模型信息
type ModelInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	ContextSize int    `json:"context_size"`
	MaxTokens   int    `json:"max_tokens"`
	Status      string `json:"status"`
}

// HealthInfo 健康信息
type HealthInfo struct {
	Status       string    `json:"status"`
	ResponseTime float64   `json:"response_time_ms"`
	ErrorRate    float64   `json:"error_rate"`
	LastTest     time.Time `json:"last_test"`
}

// AITestResult AI测试结果
type AITestResult struct {
	Success      bool          `json:"success"`
	ResponseTime time.Duration `json:"response_time"`
	Response     string        `json:"response"`
	Error        string        `json:"error,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
}

// GetAIServiceStatus 获取AI服务状态
func (m *Manager) GetAIServiceStatus() (*AIServiceStatus, error) {
	status := &AIServiceStatus{
		LastCheck: time.Now(),
	}

	// 检查服务状态
	serviceInfo, err := m.checkServiceStatus()
	if err != nil {
		status.Service = &ServiceInfo{
			Name:      "ai-service",
			Status:    "error",
			Port:      m.config.Port,
			LastCheck: time.Now(),
			Error:     err.Error(),
		}
	} else {
		status.Service = serviceInfo
	}

	// 检查提供商状态
	providerInfo, err := m.checkProviderStatus()
	if err != nil {
		status.Provider = &ProviderInfo{
			Name:   m.config.Provider,
			Status: "error",
		}
	} else {
		status.Provider = providerInfo
	}

	// 检查模型状态
	modelInfo, err := m.checkModelStatus()
	if err != nil {
		status.Model = &ModelInfo{
			Name:   m.config.Model,
			Status: "error",
		}
	} else {
		status.Model = modelInfo
	}

	// 检查健康状态
	healthInfo, err := m.checkHealthStatus()
	if err != nil {
		status.Health = &HealthInfo{
			Status: "error",
		}
	} else {
		status.Health = healthInfo
	}

	return status, nil
}

// checkServiceStatus 检查服务状态
func (m *Manager) checkServiceStatus() (*ServiceInfo, error) {
	info := &ServiceInfo{
		Name:      "ai-service",
		Port:      m.config.Port,
		LastCheck: time.Now(),
	}

	// 检查端口是否开放
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", m.config.Port))
	output, err := cmd.Output()
	if err != nil {
		info.Status = "down"
		info.Error = "服务未运行"
		return info, err
	}

	// 解析进程信息
	lines := strings.Split(string(output), "\n")
	if len(lines) > 1 {
		fields := strings.Fields(lines[1])
		if len(fields) > 1 {
			// 解析PID
			info.PID = 0 // 简化处理
		}
	}

	// 检查服务健康
	healthURL := fmt.Sprintf("http://localhost:%d/health", m.config.Port)
	resp, err := http.Get(healthURL)
	if err != nil {
		info.Status = "running"
		info.Error = "健康检查失败"
		return info, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		info.Status = "healthy"
	} else {
		info.Status = "warning"
		info.Error = fmt.Sprintf("健康检查返回状态码: %d", resp.StatusCode)
	}

	return info, nil
}

// checkProviderStatus 检查提供商状态
func (m *Manager) checkProviderStatus() (*ProviderInfo, error) {
	info := &ProviderInfo{
		Name:    m.config.Provider,
		BaseURL: m.config.BaseURL,
		Status:  "unknown",
	}

	// 掩码API密钥
	if len(m.config.APIKey) > 8 {
		info.APIKey = m.config.APIKey[:4] + "****" + m.config.APIKey[len(m.config.APIKey)-4:]
	} else {
		info.APIKey = "****"
	}

	// 测试提供商连接
	switch m.config.Provider {
	case "openai":
		info.Status = m.testOpenAI()
	case "anthropic":
		info.Status = m.testAnthropic()
	case "google":
		info.Status = m.testGoogle()
	case "local":
		info.Status = m.testLocal()
	default:
		info.Status = "unsupported"
	}

	return info, nil
}

// testOpenAI 测试OpenAI连接
func (m *Manager) testOpenAI() string {
	client := &http.Client{Timeout: 10 * time.Second}

	reqBody := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
		"max_tokens": 10,
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+m.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "error"
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return "healthy"
	}
	return "error"
}

// testAnthropic 测试Anthropic连接
func (m *Manager) testAnthropic() string {
	client := &http.Client{Timeout: 10 * time.Second}

	reqBody := map[string]interface{}{
		"model":      "claude-3-haiku-20240307",
		"max_tokens": 10,
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	req.Header.Set("x-api-key", m.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(req)
	if err != nil {
		return "error"
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return "healthy"
	}
	return "error"
}

// testGoogle 测试Google连接
func (m *Manager) testGoogle() string {
	client := &http.Client{Timeout: 10 * time.Second}

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": "Hello"},
				},
			},
		},
	}

	jsonData, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", m.config.Model, m.config.APIKey)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "error"
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return "healthy"
	}
	return "error"
}

// testLocal 测试本地模型连接
func (m *Manager) testLocal() string {
	// 测试本地AI服务
	client := &http.Client{Timeout: 10 * time.Second}

	reqBody := map[string]interface{}{
		"model":      m.config.Model,
		"prompt":     "Hello",
		"max_tokens": 10,
	}

	jsonData, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("http://localhost:%d/v1/completions", m.config.Port)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "error"
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return "healthy"
	}
	return "error"
}

// checkModelStatus 检查模型状态
func (m *Manager) checkModelStatus() (*ModelInfo, error) {
	info := &ModelInfo{
		Name:   m.config.Model,
		Status: "unknown",
	}

	// 根据模型设置默认参数
	switch m.config.Model {
	case "gpt-4":
		info.ContextSize = 8192
		info.MaxTokens = 4096
		info.Version = "4.0"
	case "gpt-3.5-turbo":
		info.ContextSize = 4096
		info.MaxTokens = 4096
		info.Version = "3.5"
	case "claude-3-opus-20240229":
		info.ContextSize = 200000
		info.MaxTokens = 4096
		info.Version = "3.0"
	case "claude-3-sonnet-20240229":
		info.ContextSize = 200000
		info.MaxTokens = 4096
		info.Version = "3.0"
	case "claude-3-haiku-20240307":
		info.ContextSize = 200000
		info.MaxTokens = 4096
		info.Version = "3.0"
	case "gemini-pro":
		info.ContextSize = 30720
		info.MaxTokens = 2048
		info.Version = "1.0"
	case "gemma-7b":
		info.ContextSize = 8192
		info.MaxTokens = 2048
		info.Version = "1.0"
	default:
		info.ContextSize = 4096
		info.MaxTokens = 2048
		info.Version = "unknown"
	}

	// 测试模型可用性
	if m.testModelAvailability() {
		info.Status = "available"
	} else {
		info.Status = "unavailable"
	}

	return info, nil
}

// testModelAvailability 测试模型可用性
func (m *Manager) testModelAvailability() bool {
	// 执行简单的测试请求
	testResult, err := m.TestAIService()
	if err != nil {
		return false
	}
	return testResult.Success
}

// checkHealthStatus 检查健康状态
func (m *Manager) checkHealthStatus() (*HealthInfo, error) {
	info := &HealthInfo{
		LastTest: time.Now(),
	}

	// 执行健康测试
	testResult, err := m.TestAIService()
	if err != nil {
		info.Status = "error"
		info.ErrorRate = 100.0
		return info, err
	}

	if testResult.Success {
		info.Status = "healthy"
		info.ResponseTime = float64(testResult.ResponseTime.Milliseconds())
		info.ErrorRate = 0.0
	} else {
		info.Status = "unhealthy"
		info.ErrorRate = 100.0
	}

	return info, nil
}

// ConfigureAIService 配置AI服务
func (m *Manager) ConfigureAIService(provider, apiKey, baseURL, model string) error {
	// 更新配置
	m.config.Provider = provider
	m.config.APIKey = apiKey
	m.config.BaseURL = baseURL
	m.config.Model = model

	// 验证配置
	if err := m.validateConfiguration(); err != nil {
		return errors.WrapError(errors.ErrCodeValidation, "配置验证失败", err)
	}

	// 重启服务以应用新配置
	if err := m.restartAIService(); err != nil {
		return errors.WrapError(errors.ErrCodeService, "重启AI服务失败", err)
	}

	return nil
}

// validateConfiguration 验证配置
func (m *Manager) validateConfiguration() error {
	if m.config.Provider == "" {
		return errors.NewError(errors.ErrCodeValidation, "提供商不能为空")
	}

	if m.config.APIKey == "" && m.config.Provider != "local" {
		return errors.NewError(errors.ErrCodeValidation, "API密钥不能为空")
	}

	if m.config.Model == "" {
		return errors.NewError(errors.ErrCodeValidation, "模型名称不能为空")
	}

	// 验证提供商和模型的兼容性
	if !m.isProviderModelCompatible(m.config.Provider, m.config.Model) {
		return errors.NewError(errors.ErrCodeValidation, "提供商和模型不兼容")
	}

	return nil
}

// isProviderModelCompatible 检查提供商和模型是否兼容
func (m *Manager) isProviderModelCompatible(provider, model string) bool {
	compatibility := map[string][]string{
		"openai":    {"gpt-4", "gpt-3.5-turbo", "gpt-3.5-turbo-16k"},
		"anthropic": {"claude-3-opus-20240229", "claude-3-sonnet-20240229", "claude-3-haiku-20240307"},
		"google":    {"gemini-pro", "gemini-pro-vision"},
		"local":     {"gemma-7b", "gemma-2b", "llama-2-7b", "llama-2-13b"},
	}

	models, exists := compatibility[provider]
	if !exists {
		return false
	}

	for _, compatibleModel := range models {
		if model == compatibleModel {
			return true
		}
	}

	return false
}

// restartAIService 重启AI服务
func (m *Manager) restartAIService() error {
	// 停止现有服务
	cmd := exec.Command("pkill", "-f", "ai-service")
	cmd.Run() // 忽略错误

	// 等待服务停止
	time.Sleep(2 * time.Second)

	// 启动新服务
	cmd = exec.Command("python3", "-m", "ai_service.main")
	cmd.Dir = "/path/to/ai-service" // 需要根据实际路径调整

	// 在后台启动服务
	if err := cmd.Start(); err != nil {
		return errors.WrapError(errors.ErrCodeService, "启动AI服务失败", err)
	}

	// 等待服务启动
	time.Sleep(5 * time.Second)

	// 验证服务是否启动成功
	status, err := m.GetAIServiceStatus()
	if err != nil {
		return errors.WrapError(errors.ErrCodeService, "验证AI服务启动失败", err)
	}

	if status.Service.Status != "healthy" {
		return errors.NewError(errors.ErrCodeService, "AI服务启动后状态异常")
	}

	return nil
}

// TestAIService 测试AI服务
func (m *Manager) TestAIService() (*AITestResult, error) {
	result := &AITestResult{
		Timestamp: time.Now(),
	}

	startTime := time.Now()

	// 构建测试请求
	reqBody := map[string]interface{}{
		"model":      m.config.Model,
		"prompt":     "请回答：1+1等于多少？",
		"max_tokens": 50,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	url := fmt.Sprintf("http://localhost:%d/v1/completions", m.config.Port)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}
	defer resp.Body.Close()

	result.ResponseTime = time.Since(startTime)

	if resp.StatusCode == 200 {
		result.Success = true
		// 解析响应
		var response map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
			if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if text, ok := choice["text"].(string); ok {
						result.Response = text
					}
				}
			}
		}
	} else {
		result.Success = false
		result.Error = fmt.Sprintf("HTTP状态码: %d", resp.StatusCode)
	}

	return result, nil
}
