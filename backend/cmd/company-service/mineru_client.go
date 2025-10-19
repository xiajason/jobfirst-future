package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// MinerUClient MinerU服务客户端
type MinerUClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewMinerUClient 创建MinerU客户端
func NewMinerUClient(baseURL string) *MinerUClient {
	return &MinerUClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ParseDocumentRequest 解析文档请求
type ParseDocumentRequest struct {
	FilePath     string `json:"file_path"`
	UserID       int    `json:"user_id"`
	BusinessType string `json:"business_type"`
}

// MinerUParseDocumentResponse 解析文档响应
type MinerUParseDocumentResponse struct {
	Status string      `json:"status"`
	Result interface{} `json:"result"`
	Error  string      `json:"error,omitempty"`
}

// UploadAndParseRequest 上传并解析请求
type UploadAndParseRequest struct {
	File   *os.File
	UserID int
}

// CompanyDocumentInfo 企业文档信息
type CompanyDocumentInfo struct {
	Type         string                 `json:"type"`
	BusinessType string                 `json:"business_type"`
	Pages        int                    `json:"pages,omitempty"`
	Content      string                 `json:"content"`
	Structure    map[string]interface{} `json:"structure"`
	Metadata     map[string]interface{} `json:"metadata"`
	FileInfo     map[string]interface{} `json:"file_info"`
	ParsedAt     string                 `json:"parsed_at"`
	Status       string                 `json:"status"`
	Confidence   float64                `json:"confidence"`
	// 企业画像特有字段
	CompanyName   string `json:"company_name,omitempty"`
	Industry      string `json:"industry,omitempty"`
	Location      string `json:"location,omitempty"`
	EmployeeCount int    `json:"employee_count,omitempty"`
	FoundedYear   int    `json:"founded_year,omitempty"`
	Revenue       string `json:"revenue,omitempty"`
}

// ParseDocument 解析文档
func (c *MinerUClient) ParseDocument(filePath string, userID int) (*CompanyDocumentInfo, error) {
	request := ParseDocumentRequest{
		FilePath:     filePath,
		UserID:       userID,
		BusinessType: "company", // 企业文档解析
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	resp, err := c.HTTPClient.Post(
		c.BaseURL+"/api/v1/parse/document",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("请求MinerU服务失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var response MinerUParseDocumentResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if response.Status != "success" {
		return nil, fmt.Errorf("MinerU解析失败: %s", response.Error)
	}

	// 将结果转换为CompanyDocumentInfo
	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		return nil, fmt.Errorf("序列化结果失败: %v", err)
	}

	var documentInfo CompanyDocumentInfo
	if err := json.Unmarshal(resultBytes, &documentInfo); err != nil {
		return nil, fmt.Errorf("解析文档信息失败: %v", err)
	}

	return &documentInfo, nil
}

// UploadAndParse 上传并解析文档
func (c *MinerUClient) UploadAndParse(filePath string, userID int) (*CompanyDocumentInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 创建multipart表单
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加文件字段
	fileWriter, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("创建文件字段失败: %v", err)
	}

	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return nil, fmt.Errorf("复制文件内容失败: %v", err)
	}

	// 添加用户ID字段
	err = writer.WriteField("user_id", fmt.Sprintf("%d", userID))
	if err != nil {
		return nil, fmt.Errorf("添加用户ID字段失败: %v", err)
	}

	// 添加业务类型字段
	err = writer.WriteField("business_type", "company")
	if err != nil {
		return nil, fmt.Errorf("添加业务类型字段失败: %v", err)
	}

	writer.Close()

	// 发送请求
	resp, err := c.HTTPClient.Post(
		c.BaseURL+"/api/v1/parse/upload",
		writer.FormDataContentType(),
		&buf,
	)
	if err != nil {
		return nil, fmt.Errorf("请求MinerU服务失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var response MinerUParseDocumentResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if response.Status != "success" {
		return nil, fmt.Errorf("MinerU解析失败: %s", response.Error)
	}

	// 将结果转换为CompanyDocumentInfo
	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		return nil, fmt.Errorf("序列化结果失败: %v", err)
	}

	var documentInfo CompanyDocumentInfo
	if err := json.Unmarshal(resultBytes, &documentInfo); err != nil {
		return nil, fmt.Errorf("解析文档信息失败: %v", err)
	}

	return &documentInfo, nil
}

// GetParseStatus 获取解析状态
func (c *MinerUClient) GetParseStatus() (map[string]interface{}, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/v1/parse/status")
	if err != nil {
		return nil, fmt.Errorf("请求MinerU服务失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return response, nil
}

// HealthCheck 健康检查
func (c *MinerUClient) HealthCheck() (map[string]interface{}, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/health")
	if err != nil {
		return nil, fmt.Errorf("请求MinerU服务失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return response, nil
}
