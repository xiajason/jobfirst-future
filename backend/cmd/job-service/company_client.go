package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CompanyClient Company服务客户端
type CompanyClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewCompanyClient 创建Company服务客户端
func NewCompanyClient(baseURL string) *CompanyClient {
	return &CompanyClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CompanyResponse Company服务响应结构
type CompanyResponse struct {
	Status string `json:"status"`
	Data   struct {
		ID                uint   `json:"id"`
		Name              string `json:"name"`
		ShortName         string `json:"short_name"`
		LogoURL           string `json:"logo_url"`
		Industry          string `json:"industry"`
		CompanySize       string `json:"company_size"`
		Location          string `json:"location"`
		Website           string `json:"website"`
		Description       string `json:"description"`
		FoundedYear       int    `json:"founded_year"`
		Status            string `json:"status"`
		VerificationLevel string `json:"verification_level"`
		JobCount          int    `json:"job_count"`
		ViewCount         int    `json:"view_count"`
		CreatedBy         uint   `json:"created_by"`
	} `json:"data"`
}

// GetCompany 获取公司信息
func (cc *CompanyClient) GetCompany(companyID uint) (*CompanyInfo, error) {
	url := fmt.Sprintf("%s/api/v1/company/public/companies/%d", cc.baseURL, companyID)

	resp, err := cc.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求Company服务失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Company服务返回错误状态: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var companyResp CompanyResponse
	if err := json.Unmarshal(body, &companyResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if companyResp.Status != "success" {
		return nil, fmt.Errorf("Company服务返回错误: %s", companyResp.Status)
	}

	// 转换为Job服务需要的CompanyInfo结构
	companyInfo := &CompanyInfo{
		ID:        companyResp.Data.ID,
		Name:      companyResp.Data.Name,
		ShortName: companyResp.Data.ShortName,
		LogoURL:   companyResp.Data.LogoURL,
		Industry:  companyResp.Data.Industry,
		Location:  companyResp.Data.Location,
	}

	return companyInfo, nil
}

// GetCompanyList 获取公司列表（用于搜索）
func (cc *CompanyClient) GetCompanyList(page, pageSize int, industry, location string) ([]CompanyInfo, error) {
	url := fmt.Sprintf("%s/api/v1/company/public/companies?page=%d&page_size=%d", cc.baseURL, page, pageSize)
	if industry != "" {
		url += "&industry=" + industry
	}
	if location != "" {
		url += "&location=" + location
	}

	resp, err := cc.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求Company服务失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Company服务返回错误状态: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var listResp struct {
		Status string `json:"status"`
		Data   struct {
			Companies []struct {
				ID        uint   `json:"id"`
				Name      string `json:"name"`
				ShortName string `json:"short_name"`
				LogoURL   string `json:"logo_url"`
				Industry  string `json:"industry"`
				Location  string `json:"location"`
			} `json:"companies"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if listResp.Status != "success" {
		return nil, fmt.Errorf("Company服务返回错误: %s", listResp.Status)
	}

	// 转换为CompanyInfo列表
	var companies []CompanyInfo
	for _, c := range listResp.Data.Companies {
		companies = append(companies, CompanyInfo{
			ID:        c.ID,
			Name:      c.Name,
			ShortName: c.ShortName,
			LogoURL:   c.LogoURL,
			Industry:  c.Industry,
			Location:  c.Location,
		})
	}

	return companies, nil
}
