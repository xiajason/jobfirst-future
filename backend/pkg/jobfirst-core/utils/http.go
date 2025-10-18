package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient HTTP客户端
type HTTPClient struct {
	client  *http.Client
	baseURL string
	headers map[string]string
}

// NewHTTPClient 创建HTTP客户端
func NewHTTPClient(baseURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
		headers: make(map[string]string),
	}
}

// SetHeader 设置请求头
func (c *HTTPClient) SetHeader(key, value string) {
	c.headers[key] = value
}

// SetHeaders 设置多个请求头
func (c *HTTPClient) SetHeaders(headers map[string]string) {
	for k, v := range headers {
		c.headers[k] = v
	}
}

// Get 发送GET请求
func (c *HTTPClient) Get(path string, params map[string]string) (*HTTPResponse, error) {
	return c.request("GET", path, params, nil)
}

// Post 发送POST请求
func (c *HTTPClient) Post(path string, data interface{}) (*HTTPResponse, error) {
	return c.request("POST", path, nil, data)
}

// Put 发送PUT请求
func (c *HTTPClient) Put(path string, data interface{}) (*HTTPResponse, error) {
	return c.request("PUT", path, nil, data)
}

// Delete 发送DELETE请求
func (c *HTTPClient) Delete(path string) (*HTTPResponse, error) {
	return c.request("DELETE", path, nil, nil)
}

// request 发送HTTP请求
func (c *HTTPClient) request(method, path string, params map[string]string, data interface{}) (*HTTPResponse, error) {
	url := c.baseURL + path

	// 添加查询参数
	if len(params) > 0 {
		url += "?"
		first := true
		for k, v := range params {
			if !first {
				url += "&"
			}
			url += fmt.Sprintf("%s=%s", k, v)
			first = false
		}
	}

	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("序列化请求数据失败: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	// 如果有数据，设置Content-Type
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       responseBody,
	}, nil
}

// HTTPResponse HTTP响应
type HTTPResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// JSON 解析响应为JSON
func (r *HTTPResponse) JSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// String 获取响应字符串
func (r *HTTPResponse) String() string {
	return string(r.Body)
}

// IsSuccess 检查请求是否成功
func (r *HTTPResponse) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}
