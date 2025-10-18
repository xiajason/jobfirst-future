package usersync

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// IntegrationExample 集成示例
type IntegrationExample struct {
	syncService *UserSyncService
	publisher   *UserEventPublisher
	subscriber  *UserEventSubscriber
}

// NewIntegrationExample 创建集成示例
func NewIntegrationExample(redisClient *redis.Client) *IntegrationExample {
	// 创建同步服务
	config := DefaultSyncConfig()
	syncService := NewUserSyncService(config, redisClient)

	// 创建事件发布器
	publisher := NewUserEventPublisher(redisClient, "basic-server")

	// 创建事件订阅器
	subscriber := NewUserEventSubscriber(redisClient, "user-sync-group", "user-sync-consumer")

	return &IntegrationExample{
		syncService: syncService,
		publisher:   publisher,
		subscriber:  subscriber,
	}
}

// Start 启动集成示例
func (e *IntegrationExample) Start(ctx context.Context) error {
	// 启动同步服务
	if err := e.syncService.Start(); err != nil {
		return err
	}

	// 注册事件处理器
	e.subscriber.RegisterHandler(EventTypeUserCreated, e.handleUserCreated)
	e.subscriber.RegisterHandler(EventTypeUserUpdated, e.handleUserUpdated)
	e.subscriber.RegisterHandler(EventTypeUserDeleted, e.handleUserDeleted)
	e.subscriber.RegisterHandler(EventTypeUserStatusChanged, e.handleUserStatusChanged)

	// 启动事件订阅
	go func() {
		if err := e.subscriber.Start(ctx); err != nil {
			log.Printf("事件订阅器启动失败: %v", err)
		}
	}()

	log.Println("用户数据同步集成示例启动成功")
	return nil
}

// Stop 停止集成示例
func (e *IntegrationExample) Stop() {
	e.syncService.Stop()
	log.Println("用户数据同步集成示例已停止")
}

// SimulateUserCreation 模拟用户创建
func (e *IntegrationExample) SimulateUserCreation() error {
	user := &User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		Role:      "guest",
		Status:    "active",
		Phone:     "1234567890",
		CreatedAt: timePtr(time.Now()),
		UpdatedAt: timePtr(time.Now()),
	}

	// 发布用户创建事件
	return e.publisher.PublishUserCreated(user)
}

// SimulateUserUpdate 模拟用户更新
func (e *IntegrationExample) SimulateUserUpdate() error {
	user := &User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		Role:      "system_admin", // 角色变更
		Status:    "active",
		Phone:     "1234567890",
		CreatedAt: timePtr(time.Now()),
		UpdatedAt: timePtr(time.Now()),
	}

	changes := map[string]interface{}{
		"role": "system_admin",
	}

	// 发布用户更新事件
	return e.publisher.PublishUserUpdated(user, changes)
}

// SimulateUserStatusChange 模拟用户状态变更
func (e *IntegrationExample) SimulateUserStatusChange() error {
	user := &User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		Role:      "system_admin",
		Status:    "inactive", // 状态变更
		Phone:     "1234567890",
		CreatedAt: timePtr(time.Now()),
		UpdatedAt: timePtr(time.Now()),
	}

	// 发布用户状态变更事件
	return e.publisher.PublishUserStatusChanged(user, "active", "inactive")
}

// 事件处理器
func (e *IntegrationExample) handleUserCreated(event UserEvent) error {
	log.Printf("处理用户创建事件: %s, 用户: %s", event.ID, event.Username)

	// 这里可以添加具体的业务逻辑
	// 例如：发送欢迎邮件、初始化用户配置等

	return nil
}

func (e *IntegrationExample) handleUserUpdated(event UserEvent) error {
	log.Printf("处理用户更新事件: %s, 用户: %s", event.ID, event.Username)

	// 这里可以添加具体的业务逻辑
	// 例如：更新缓存、通知其他服务等

	return nil
}

func (e *IntegrationExample) handleUserDeleted(event UserEvent) error {
	log.Printf("处理用户删除事件: %s, 用户: %s", event.ID, event.Username)

	// 这里可以添加具体的业务逻辑
	// 例如：清理用户数据、发送通知等

	return nil
}

func (e *IntegrationExample) handleUserStatusChanged(event UserEvent) error {
	log.Printf("处理用户状态变更事件: %s, 用户: %s", event.ID, event.Username)

	// 这里可以添加具体的业务逻辑
	// 例如：更新用户权限、发送状态变更通知等

	return nil
}

// GetStats 获取统计信息
func (e *IntegrationExample) GetStats() map[string]interface{} {
	stats := e.syncService.GetStats()
	queueStatus := e.syncService.GetQueueStatus()

	return map[string]interface{}{
		"sync_stats":   stats,
		"queue_status": queueStatus,
		"service_name": "user-sync-integration",
		"uptime":       time.Now().Format(time.RFC3339),
	}
}

// TestSync 测试同步功能
func (e *IntegrationExample) TestSync() error {
	log.Println("开始测试用户数据同步功能...")

	// 测试用户创建同步
	if err := e.SimulateUserCreation(); err != nil {
		return err
	}

	// 等待一段时间
	time.Sleep(2 * time.Second)

	// 测试用户更新同步
	if err := e.SimulateUserUpdate(); err != nil {
		return err
	}

	// 等待一段时间
	time.Sleep(2 * time.Second)

	// 测试用户状态变更同步
	if err := e.SimulateUserStatusChange(); err != nil {
		return err
	}

	log.Println("用户数据同步功能测试完成")
	return nil
}

// 辅助函数
func timePtr(t time.Time) *time.Time {
	return &t
}

// BasicServerIntegration Basic Server集成示例
type BasicServerIntegration struct {
	syncService *UserSyncService
	publisher   *UserEventPublisher
}

// NewBasicServerIntegration 创建Basic Server集成
func NewBasicServerIntegration(redisClient *redis.Client) *BasicServerIntegration {
	config := DefaultSyncConfig()
	syncService := NewUserSyncService(config, redisClient)
	publisher := NewUserEventPublisher(redisClient, "basic-server")

	return &BasicServerIntegration{
		syncService: syncService,
		publisher:   publisher,
	}
}

// Start 启动Basic Server集成
func (b *BasicServerIntegration) Start() error {
	return b.syncService.Start()
}

// Stop 停止Basic Server集成
func (b *BasicServerIntegration) Stop() {
	b.syncService.Stop()
}

// OnUserCreated 用户创建时的处理
func (b *BasicServerIntegration) OnUserCreated(user *User) error {
	// 发布用户创建事件
	return b.publisher.PublishUserCreated(user)
}

// OnUserUpdated 用户更新时的处理
func (b *BasicServerIntegration) OnUserUpdated(user *User, changes map[string]interface{}) error {
	// 发布用户更新事件
	return b.publisher.PublishUserUpdated(user, changes)
}

// OnUserDeleted 用户删除时的处理
func (b *BasicServerIntegration) OnUserDeleted(userID uint, username string) error {
	// 发布用户删除事件
	return b.publisher.PublishUserDeleted(userID, username)
}

// OnUserStatusChanged 用户状态变更时的处理
func (b *BasicServerIntegration) OnUserStatusChanged(user *User, oldStatus, newStatus string) error {
	// 发布用户状态变更事件
	return b.publisher.PublishUserStatusChanged(user, oldStatus, newStatus)
}
