package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ServiceIntegration 服务间集成处理器
type ServiceIntegration struct {
	notificationBusiness *NotificationBusiness
}

// NewServiceIntegration 创建服务间集成处理器
func NewServiceIntegration(nb *NotificationBusiness) *ServiceIntegration {
	return &ServiceIntegration{
		notificationBusiness: nb,
	}
}

// UserQuotaInfo 用户配额信息
type UserQuotaInfo struct {
	UserID           uint    `json:"user_id"`
	SubscriptionType string  `json:"subscription_type"`
	ServiceType      string  `json:"service_type"`
	DailyLimit       int     `json:"daily_limit"`
	MonthlyLimit     int     `json:"monthly_limit"`
	DailyUsed        int     `json:"daily_used"`
	MonthlyUsed      int     `json:"monthly_used"`
	DailyCostLimit   float64 `json:"daily_cost_limit"`
	MonthlyCostLimit float64 `json:"monthly_cost_limit"`
	DailyCostUsed    float64 `json:"daily_cost_used"`
	MonthlyCostUsed  float64 `json:"monthly_cost_used"`
	IsActive         bool    `json:"is_active"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID                 uint   `json:"id"`
	Username           string `json:"username"`
	Email              string `json:"email"`
	SubscriptionStatus string `json:"subscription_status"`
	SubscriptionType   string `json:"subscription_type"`
	IsActive           bool   `json:"is_active"`
}

// CheckUserQuotaAndSendNotification 检查用户配额并发送通知
func (si *ServiceIntegration) CheckUserQuotaAndSendNotification(userID uint) error {
	// 1. 获取用户配额信息
	quotaInfo, err := si.getUserQuotaFromCompanyService(userID)
	if err != nil {
		return fmt.Errorf("获取用户配额信息失败: %v", err)
	}

	// 2. 检查每日配额使用情况
	if quotaInfo.DailyLimit > 0 {
		usagePercentage := float64(quotaInfo.DailyUsed) / float64(quotaInfo.DailyLimit) * 100

		// 如果使用量超过80%，发送警告通知
		if usagePercentage >= 80 && usagePercentage < 100 {
			err := si.notificationBusiness.SendAIServiceNotification(
				userID,
				"ai_service_limit_warning",
				"AI服务使用量警告",
				fmt.Sprintf("您的AI服务每日使用量已达到%.1f%%，当前使用量：%d/%d", usagePercentage, quotaInfo.DailyUsed, quotaInfo.DailyLimit),
				map[string]interface{}{
					"service_type": quotaInfo.ServiceType,
					"usage":        quotaInfo.DailyUsed,
					"limit":        quotaInfo.DailyLimit,
					"percentage":   usagePercentage,
				},
			)
			if err != nil {
				return fmt.Errorf("发送使用量警告通知失败: %v", err)
			}
		}

		// 如果使用量达到100%，发送超出限制通知
		if usagePercentage >= 100 {
			err := si.notificationBusiness.SendAIServiceNotification(
				userID,
				"ai_service_limit_exceeded",
				"AI服务使用限制超出",
				fmt.Sprintf("您的AI服务每日使用量已超出限制，当前使用量：%d，限制：%d。请升级订阅或等待配额重置。", quotaInfo.DailyUsed, quotaInfo.DailyLimit),
				map[string]interface{}{
					"service_type": quotaInfo.ServiceType,
					"usage":        quotaInfo.DailyUsed,
					"limit":        quotaInfo.DailyLimit,
				},
			)
			if err != nil {
				return fmt.Errorf("发送使用限制超出通知失败: %v", err)
			}
		}
	}

	// 3. 检查成本使用情况
	if quotaInfo.DailyCostLimit > 0 {
		costPercentage := quotaInfo.DailyCostUsed / quotaInfo.DailyCostLimit * 100

		// 如果成本使用超过90%，发送成本警告通知
		if costPercentage >= 90 {
			err := si.notificationBusiness.SendCostControlNotification(
				userID,
				"cost_limit_warning",
				"成本使用警告",
				fmt.Sprintf("您的每日成本使用已达到%.1f%%，当前成本：$%.2f，限制：$%.2f", costPercentage, quotaInfo.DailyCostUsed, quotaInfo.DailyCostLimit),
				map[string]interface{}{
					"current_cost": quotaInfo.DailyCostUsed,
					"limit":        quotaInfo.DailyCostLimit,
					"percentage":   costPercentage,
				},
			)
			if err != nil {
				return fmt.Errorf("发送成本警告通知失败: %v", err)
			}
		}
	}

	return nil
}

// HandleSubscriptionStatusChange 处理订阅状态变更
func (si *ServiceIntegration) HandleSubscriptionStatusChange(userID uint, oldStatus, newStatus string) error {
	// 获取用户信息
	userInfo, err := si.getUserInfoFromUserService(userID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %v", err)
	}

	// 根据状态变更发送相应通知
	switch {
	case newStatus == "premium" && oldStatus == "trial":
		// 从试用升级到付费
		err := si.notificationBusiness.SendSubscriptionNotification(
			userID,
			"subscription_upgraded",
			"订阅升级成功",
			fmt.Sprintf("恭喜！您已成功从试用版升级到%s，现在可以享受更多功能和服务。", newStatus),
		)
		if err != nil {
			return fmt.Errorf("发送订阅升级通知失败: %v", err)
		}

	case newStatus == "expired":
		// 订阅过期
		err := si.notificationBusiness.SendSubscriptionNotification(
			userID,
			"subscription_expired",
			"订阅已过期",
			fmt.Sprintf("您的%s订阅已过期，部分功能将受到限制。请及时续费以恢复完整功能。", oldStatus),
		)
		if err != nil {
			return fmt.Errorf("发送订阅过期通知失败: %v", err)
		}

	case newStatus == "active" && oldStatus == "expired":
		// 订阅续费
		err := si.notificationBusiness.SendSubscriptionNotification(
			userID,
			"subscription_renewed",
			"订阅续费成功",
			fmt.Sprintf("您的%s订阅已成功续费，感谢您的支持！", userInfo.SubscriptionType),
		)
		if err != nil {
			return fmt.Errorf("发送订阅续费通知失败: %v", err)
		}
	}

	return nil
}

// SendWelcomeNotification 发送欢迎通知
func (si *ServiceIntegration) SendWelcomeNotification(userID uint) error {
	// 获取用户信息
	userInfo, err := si.getUserInfoFromUserService(userID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %v", err)
	}

	// 发送欢迎通知
	title := "欢迎使用JobFirst平台"
	content := fmt.Sprintf("欢迎%s！您已成功注册JobFirst平台，现在可以开始使用我们的AI服务。", userInfo.Username)

	err = si.notificationBusiness.CreateNotification(
		userID,
		"welcome",
		title,
		content,
		"system",
		"normal",
		`{"type":"welcome","timestamp":"`+fmt.Sprintf("%d", time.Now().Unix())+`"}`,
	)
	if err != nil {
		return fmt.Errorf("发送欢迎通知失败: %v", err)
	}

	return nil
}

// getUserQuotaFromCompanyService 从Company服务获取用户配额信息
func (si *ServiceIntegration) getUserQuotaFromCompanyService(userID uint) (*UserQuotaInfo, error) {
	url := fmt.Sprintf("http://localhost:8083/api/v1/quota/user/%d", userID)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求Company服务失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Company服务返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var response struct {
		QuotaInfo   map[string]interface{} `json:"quota_info"`
		ServiceType string                 `json:"service_type"`
		UserID      uint                   `json:"user_id"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 将quota_info转换为UserQuotaInfo结构
	quotaInfo := &UserQuotaInfo{
		UserID:      response.UserID,
		ServiceType: response.ServiceType,
	}

	// 从quota_info中提取字段
	if dailyUsed, ok := response.QuotaInfo["daily_used"].(float64); ok {
		quotaInfo.DailyUsed = int(dailyUsed)
	}
	if dailyLimit, ok := response.QuotaInfo["daily_limit"].(float64); ok {
		quotaInfo.DailyLimit = int(dailyLimit)
	}
	if monthlyUsed, ok := response.QuotaInfo["monthly_used"].(float64); ok {
		quotaInfo.MonthlyUsed = int(monthlyUsed)
	}
	if monthlyLimit, ok := response.QuotaInfo["monthly_limit"].(float64); ok {
		quotaInfo.MonthlyLimit = int(monthlyLimit)
	}
	if dailyCostUsed, ok := response.QuotaInfo["cost_used"].(float64); ok {
		quotaInfo.DailyCostUsed = dailyCostUsed
	}
	if dailyCostLimit, ok := response.QuotaInfo["cost_limit"].(float64); ok {
		quotaInfo.DailyCostLimit = dailyCostLimit
	}

	return quotaInfo, nil
}

// getUserInfoFromUserService 从User服务获取用户信息
func (si *ServiceIntegration) getUserInfoFromUserService(userID uint) (*UserInfo, error) {
	url := fmt.Sprintf("http://localhost:8081/api/v1/users/%d", userID)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求User服务失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("User服务返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var response struct {
		Status string   `json:"status"`
		Data   UserInfo `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if response.Status != "success" {
		return nil, fmt.Errorf("User服务返回错误: %s", response.Status)
	}

	return &response.Data, nil
}

// StartQuotaMonitoring 启动配额监控
func (si *ServiceIntegration) StartQuotaMonitoring() {
	// 这里可以实现定时任务，定期检查所有活跃用户的配额使用情况
	// 并自动发送相应的通知
	go func() {
		ticker := time.NewTicker(30 * time.Minute) // 每30分钟检查一次
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 这里需要获取所有活跃用户列表，然后检查他们的配额
				// 由于这是一个示例，我们暂时跳过具体的实现
				// 在实际应用中，这里会调用User服务获取活跃用户列表
				// 然后对每个用户调用CheckUserQuotaAndSendNotification
			}
		}
	}()
}
