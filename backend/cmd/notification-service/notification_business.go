package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jobfirst/jobfirst-core"
	"gorm.io/gorm"
)

// Notification 通知模型
type Notification struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	UserID    uint       `json:"user_id" gorm:"not null"`
	Type      string     `json:"type" gorm:"size:50;not null"` // 对应数据库的type字段
	Title     string     `json:"title" gorm:"size:200;not null"`
	Content   string     `json:"content" gorm:"type:text"`
	Category  string     `json:"category" gorm:"size:50;default:system"`
	Priority  string     `json:"priority" gorm:"type:enum('low','normal','high','urgent');default:normal"`
	Status    string     `json:"status" gorm:"size:20;default:unread"` // unread, read
	IsRead    bool       `json:"is_read" gorm:"default:false"`
	ReadAt    *time.Time `json:"read_at" gorm:"column:read_at"`
	ExpiresAt *time.Time `json:"expires_at"`
	Metadata  string     `json:"metadata" gorm:"type:json"` // 存储JSON字符串
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// TableName 指定表名
func (Notification) TableName() string {
	return "notifications"
}

// NotificationBusiness 通知业务逻辑处理器
type NotificationBusiness struct {
	core *jobfirst.Core
	db   *gorm.DB
}

// NewNotificationBusiness 创建通知业务逻辑处理器
func NewNotificationBusiness(core *jobfirst.Core) *NotificationBusiness {
	return &NotificationBusiness{
		core: core,
		db:   core.GetDB(),
	}
}

// CreateNotification 创建通知
func (nb *NotificationBusiness) CreateNotification(userID uint, notificationType, title, content, category, priority, metadata string) error {
	notification := Notification{
		UserID:    userID,
		Type:      notificationType,
		Title:     title,
		Content:   content,
		Category:  category,
		Priority:  priority,
		Status:    "unread",
		IsRead:    false,
		Metadata:  metadata, // 直接存储JSON字符串
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return nb.db.Create(&notification).Error
}

// GetUserNotifications 获取用户通知列表
func (nb *NotificationBusiness) GetUserNotifications(userID uint, limit int) ([]Notification, error) {
	var notifications []Notification
	query := nb.db.Where("user_id = ?", userID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&notifications).Error
	return notifications, err
}

// MarkAsRead 标记通知为已读
func (nb *NotificationBusiness) MarkAsRead(notificationID, userID uint) error {
	now := time.Now()
	return nb.db.Model(&Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"status":     "read",
			"is_read":    true,
			"read_at":    &now,
			"updated_at": now,
		}).Error
}

// GetNotificationStats 获取用户通知统计
func (nb *NotificationBusiness) GetNotificationStats(userID uint) (map[string]interface{}, error) {
	var totalCount, unreadCount int64

	// 总通知数
	if err := nb.db.Model(&Notification{}).Where("user_id = ?", userID).Count(&totalCount).Error; err != nil {
		return nil, err
	}

	// 未读通知数
	if err := nb.db.Model(&Notification{}).Where("user_id = ? AND status = ?", userID, "unread").Count(&unreadCount).Error; err != nil {
		return nil, err
	}

	// 按类型统计
	var typeStats []map[string]interface{}
	rows, err := nb.db.Model(&Notification{}).
		Select("type, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("type").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var notificationType string
		var count int64
		if err := rows.Scan(&notificationType, &count); err != nil {
			continue
		}
		typeStats = append(typeStats, map[string]interface{}{
			"type":  notificationType,
			"count": count,
		})
	}

	return map[string]interface{}{
		"total_count":  totalCount,
		"unread_count": unreadCount,
		"read_count":   totalCount - unreadCount,
		"type_stats":   typeStats,
	}, nil
}

// SendSubscriptionNotification 发送订阅相关通知
func (nb *NotificationBusiness) SendSubscriptionNotification(userID uint, notificationType, title, content string) error {
	metadata := map[string]interface{}{
		"notification_type": notificationType,
		"timestamp":         time.Now().Unix(),
	}

	metadataJSON, _ := json.Marshal(metadata)

	return nb.CreateNotification(
		userID,
		notificationType,
		title,
		content,
		"subscription",
		"high",
		string(metadataJSON),
	)
}

// SendAIServiceNotification 发送AI服务相关通知
func (nb *NotificationBusiness) SendAIServiceNotification(userID uint, notificationType, title, content string, usageData map[string]interface{}) error {
	metadata := map[string]interface{}{
		"notification_type": notificationType,
		"usage_data":        usageData,
		"timestamp":         time.Now().Unix(),
	}

	metadataJSON, _ := json.Marshal(metadata)

	return nb.CreateNotification(
		userID,
		notificationType,
		title,
		content,
		"ai_service",
		"normal",
		string(metadataJSON),
	)
}

// SendCostControlNotification 发送成本控制相关通知
func (nb *NotificationBusiness) SendCostControlNotification(userID uint, notificationType, title, content string, costData map[string]interface{}) error {
	metadata := map[string]interface{}{
		"notification_type": notificationType,
		"cost_data":         costData,
		"timestamp":         time.Now().Unix(),
	}

	metadataJSON, _ := json.Marshal(metadata)

	return nb.CreateNotification(
		userID,
		notificationType,
		title,
		content,
		"cost_control",
		"high",
		string(metadataJSON),
	)
}

// CheckAndSendQuotaWarning 检查并发送配额警告通知
func (nb *NotificationBusiness) CheckAndSendQuotaWarning(userID uint) error {
	// 这里需要调用Company服务的AI配额API来获取用户配额信息
	// 由于服务间通信，我们需要通过HTTP调用

	// 获取用户配额信息
	quotaURL := fmt.Sprintf("http://localhost:8083/api/v1/quota/user/%d", userID)
	resp, err := http.Get(quotaURL)
	if err != nil {
		return fmt.Errorf("获取用户配额信息失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("获取用户配额信息失败，状态码: %d", resp.StatusCode)
	}

	// 解析响应并检查配额使用情况
	// 这里需要根据实际的API响应格式来解析
	// 如果配额使用超过80%，发送警告通知

	return nil
}

// AutoMigrate 自动迁移数据库表
func (nb *NotificationBusiness) AutoMigrate() error {
	return nb.db.AutoMigrate(&Notification{})
}
