package usersync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPSyncExecutor HTTP同步执行器
type HTTPSyncExecutor struct {
	client  *http.Client
	timeout time.Duration
}

// NewHTTPSyncExecutor 创建新的HTTP同步执行器
func NewHTTPSyncExecutor(timeout time.Duration) *HTTPSyncExecutor {
	return &HTTPSyncExecutor{
		client: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// ExecuteSync 执行HTTP同步
func (e *HTTPSyncExecutor) ExecuteSync(ctx context.Context, task UserSyncTask, target SyncTarget) error {
	// 准备请求数据
	requestData := map[string]interface{}{
		"task_id":    task.ID,
		"user_id":    task.UserID,
		"username":   task.Username,
		"event_type": task.EventType,
		"data":       task.Data,
		"timestamp":  task.CreatedAt,
	}

	// 序列化请求数据
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("序列化请求数据失败: %w", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, target.Method, target.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Sync-Source", "user-sync-service")
	req.Header.Set("X-Sync-Task-ID", task.ID)
	req.Header.Set("X-Sync-Event-Type", string(task.EventType))
	req.Header.Set("User-Agent", "UserSyncService/1.0")

	// 执行请求
	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("执行HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查响应状态
	if resp.StatusCode >= 400 {
		return fmt.Errorf("同步失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应（可选）
	var response map[string]interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &response); err != nil {
			// 响应解析失败不是致命错误，记录日志即可
			fmt.Printf("解析响应失败: %v, 响应内容: %s\n", err, string(body))
		}
	}

	return nil
}

// SyncToUnifiedAuth 同步到统一认证服务
func (e *HTTPSyncExecutor) SyncToUnifiedAuth(ctx context.Context, task UserSyncTask) error {
	target := SyncTarget{
		Service: "unified-auth",
		URL:     "http://localhost:8207/api/v1/auth/sync/user",
		Method:  "POST",
		Enabled: true,
	}

	return e.ExecuteSync(ctx, task, target)
}

// SyncToUserService 同步到用户服务
func (e *HTTPSyncExecutor) SyncToUserService(ctx context.Context, task UserSyncTask) error {
	target := SyncTarget{
		Service: "user-service",
		URL:     "http://localhost:8081/api/v1/users/sync",
		Method:  "POST",
		Enabled: true,
	}

	return e.ExecuteSync(ctx, task, target)
}

// SyncToBasicServer 同步到基础服务
func (e *HTTPSyncExecutor) SyncToBasicServer(ctx context.Context, task UserSyncTask) error {
	target := SyncTarget{
		Service: "basic-server",
		URL:     "http://localhost:8080/api/v1/auth/sync/user",
		Method:  "POST",
		Enabled: true,
	}

	return e.ExecuteSync(ctx, task, target)
}

// TestConnection 测试连接
func (e *HTTPSyncExecutor) TestConnection(ctx context.Context, url string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建测试请求失败: %w", err)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("测试连接失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("连接测试失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// BatchSync 批量同步
func (e *HTTPSyncExecutor) BatchSync(ctx context.Context, tasks []UserSyncTask, target SyncTarget) error {
	if len(tasks) == 0 {
		return nil
	}

	// 准备批量请求数据
	batchData := map[string]interface{}{
		"tasks":     tasks,
		"timestamp": time.Now(),
	}

	// 序列化请求数据
	jsonData, err := json.Marshal(batchData)
	if err != nil {
		return fmt.Errorf("序列化批量请求数据失败: %w", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", target.URL+"/batch", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建批量HTTP请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Sync-Source", "user-sync-service")
	req.Header.Set("X-Sync-Batch-Size", fmt.Sprintf("%d", len(tasks)))
	req.Header.Set("User-Agent", "UserSyncService/1.0")

	// 执行请求
	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("执行批量HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取批量响应失败: %w", err)
	}

	// 检查响应状态
	if resp.StatusCode >= 400 {
		return fmt.Errorf("批量同步失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return nil
}
