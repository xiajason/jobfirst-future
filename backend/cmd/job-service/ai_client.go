package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AIClient AI服务客户端
type AIClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewAIClient 创建AI服务客户端
func NewAIClient(baseURL string) *AIClient {
	return &AIClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// AIJobMatchingRequest AI服务职位匹配请求
type AIJobMatchingRequest struct {
	ResumeID uint                   `json:"resume_id"`
	Limit    int                    `json:"limit"`
	Filters  map[string]interface{} `json:"filters"`
}

// AIJobMatchingResponse AI服务职位匹配响应
type AIJobMatchingResponse struct {
	Success   bool                   `json:"success"`
	Data      []AIMatchResult        `json:"data"`
	Message   string                 `json:"message"`
	Timestamp string                 `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// AIMatchResult AI匹配结果
type AIMatchResult struct {
	JobID       uint               `json:"job_id"`
	MatchScore  float64            `json:"match_score"`
	Breakdown   map[string]float64 `json:"breakdown"`
	Confidence  float64            `json:"confidence"`
	JobInfo     Job                `json:"job_info"`
	CompanyInfo CompanyInfo        `json:"company_info"`
	Reason      string             `json:"reason"`
}

// MatchJob 调用AI服务进行职位匹配
func (c *AIClient) MatchJob(req AIJobMatchingRequest, authToken string) (*AIJobMatchingResponse, error) {
	// 准备请求数据
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", c.BaseURL+"/api/v1/ai/job-matching", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+authToken)

	// 发送请求
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应
	var aiResponse AIJobMatchingResponse
	if err := json.Unmarshal(body, &aiResponse); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI服务返回错误: %s (状态码: %d)", aiResponse.Message, resp.StatusCode)
	}

	return &aiResponse, nil
}

// EnhancedMatchJob 调用AI服务进行增强版职位匹配
func (c *AIClient) EnhancedMatchJob(req AIJobMatchingRequest, authToken string) (*AIJobMatchingResponse, error) {
	// 准备请求数据
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	// 创建HTTP请求 - 使用增强版匹配API
	httpReq, err := http.NewRequest("POST", c.BaseURL+"/api/v1/ai/enhanced-job-matching", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+authToken)

	// 发送请求
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应
	var aiResponse AIJobMatchingResponse
	if err := json.Unmarshal(body, &aiResponse); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI服务返回错误: %s (状态码: %d)", aiResponse.Message, resp.StatusCode)
	}

	return &aiResponse, nil
}

// GetMatchingRecommendations 获取匹配推荐建议
func (c *AIClient) GetMatchingRecommendations(resumeID uint, authToken string) (map[string]interface{}, error) {
	// 创建HTTP请求
	httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/ai/matching/recommendations/%d", c.BaseURL, resumeID), nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	httpReq.Header.Set("Authorization", "Bearer "+authToken)

	// 发送请求
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应
	var recommendations map[string]interface{}
	if err := json.Unmarshal(body, &recommendations); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI服务返回错误: (状态码: %d)", resp.StatusCode)
	}

	return recommendations, nil
}

// GetMatchingAnalysis 获取匹配分析报告
func (c *AIClient) GetMatchingAnalysis(resumeID uint, authToken string) (map[string]interface{}, error) {
	// 创建HTTP请求
	httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/ai/matching/analysis/%d", c.BaseURL, resumeID), nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	httpReq.Header.Set("Authorization", "Bearer "+authToken)

	// 发送请求
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应
	var analysis map[string]interface{}
	if err := json.Unmarshal(body, &analysis); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI服务返回错误: (状态码: %d)", resp.StatusCode)
	}

	return analysis, nil
}

// GetAIHealth 检查AI服务健康状态
func (c *AIClient) GetAIHealth() (map[string]interface{}, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/health")
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var health map[string]interface{}
	if err := json.Unmarshal(body, &health); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return health, nil
}
