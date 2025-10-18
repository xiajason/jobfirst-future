package usersync

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// MockRedisClient 模拟Redis客户端
type MockRedisClient struct {
	data map[string]string
}

func NewMockRedisClient() *redis.Client {
	// 返回一个真实的Redis客户端，但在测试环境中会失败
	// 在实际测试中，我们需要一个真实的Redis实例或者更好的模拟
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
}

// TestUserSyncService 测试用户同步服务
func TestUserSyncService(t *testing.T) {
	// 创建模拟Redis客户端
	mockRedis := NewMockRedisClient()

	// 创建同步服务
	config := DefaultSyncConfig()
	config.Enabled = true
	config.Workers = 1
	config.QueueSize = 10
	config.Timeout = 5 * time.Second

	syncService := NewUserSyncService(config, mockRedis)

	// 启动同步服务
	if err := syncService.Start(); err != nil {
		t.Fatalf("启动同步服务失败: %v", err)
	}
	defer syncService.Stop()

	// 创建测试用户
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

	// 创建同步任务
	task := UserSyncTask{
		UserID:    user.ID,
		Username:  user.Username,
		EventType: EventTypeUserCreated,
		Targets: []SyncTarget{
			{
				Service: "redis-cache",
				URL:     "redis://localhost:6379",
				Method:  "SET",
				Enabled: true,
			},
		},
		Data: map[string]interface{}{
			"user": user,
		},
		Priority:   1,
		MaxRetries: 3,
	}

	// 添加同步任务
	if err := syncService.AddSyncTask(task); err != nil {
		t.Fatalf("添加同步任务失败: %v", err)
	}

	// 等待任务处理
	time.Sleep(2 * time.Second)

	// 检查统计信息
	stats := syncService.GetStats()
	if stats.TotalTasks == 0 {
		t.Error("期望有同步任务，但统计显示为0")
	}

	// 检查队列状态
	queueStatus := syncService.GetQueueStatus()
	if queueStatus["is_running"] != true {
		t.Error("期望同步服务正在运行")
	}
}

// TestUserEventPublisher 测试用户事件发布器
func TestUserEventPublisher(t *testing.T) {
	// 创建模拟Redis客户端
	mockRedis := NewMockRedisClient()

	// 创建事件发布器
	publisher := NewUserEventPublisher(mockRedis, "test-service")

	// 创建测试用户
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

	// 测试发布用户创建事件
	if err := publisher.PublishUserCreated(user); err != nil {
		t.Fatalf("发布用户创建事件失败: %v", err)
	}

	// 测试发布用户更新事件
	changes := map[string]interface{}{
		"role": "system_admin",
	}
	if err := publisher.PublishUserUpdated(user, changes); err != nil {
		t.Fatalf("发布用户更新事件失败: %v", err)
	}

	// 测试发布用户状态变更事件
	if err := publisher.PublishUserStatusChanged(user, "active", "inactive"); err != nil {
		t.Fatalf("发布用户状态变更事件失败: %v", err)
	}

	// 测试发布用户删除事件
	if err := publisher.PublishUserDeleted(user.ID, user.Username); err != nil {
		t.Fatalf("发布用户删除事件失败: %v", err)
	}
}

// TestHTTPSyncExecutor 测试HTTP同步执行器
func TestHTTPSyncExecutor(t *testing.T) {
	// 创建HTTP执行器
	executor := NewHTTPSyncExecutor(5 * time.Second)

	// 创建测试任务
	task := UserSyncTask{
		ID:        "test-task-1",
		UserID:    1,
		Username:  "testuser",
		EventType: EventTypeUserCreated,
		Data: map[string]interface{}{
			"user": &User{
				ID:       1,
				Username: "testuser",
				Email:    "test@example.com",
				Role:     "guest",
				Status:   "active",
			},
		},
		CreatedAt: time.Now(),
	}

	// 测试同步到统一认证服务（会失败，因为没有实际服务）
	ctx := context.Background()
	err := executor.SyncToUnifiedAuth(ctx, task)
	if err == nil {
		t.Error("期望同步到统一认证服务失败，但成功了")
	}

	// 测试同步到用户服务（会失败，因为没有实际服务）
	err = executor.SyncToUserService(ctx, task)
	if err == nil {
		t.Error("期望同步到用户服务失败，但成功了")
	}

	// 测试同步到基础服务（会失败，因为没有实际服务）
	err = executor.SyncToBasicServer(ctx, task)
	if err == nil {
		t.Error("期望同步到基础服务失败，但成功了")
	}
}

// TestIntegrationExample 测试集成示例
func TestIntegrationExample(t *testing.T) {
	// 创建模拟Redis客户端
	mockRedis := NewMockRedisClient()

	// 创建集成示例
	integration := NewIntegrationExample(mockRedis)

	// 启动集成示例
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := integration.Start(ctx); err != nil {
		t.Fatalf("启动集成示例失败: %v", err)
	}
	defer integration.Stop()

	// 测试同步功能
	if err := integration.TestSync(); err != nil {
		t.Fatalf("测试同步功能失败: %v", err)
	}

	// 获取统计信息
	stats := integration.GetStats()
	if stats["service_name"] != "user-sync-integration" {
		t.Error("期望服务名称为 user-sync-integration")
	}
}

// TestBasicServerIntegration 测试Basic Server集成
func TestBasicServerIntegration(t *testing.T) {
	// 创建模拟Redis客户端
	mockRedis := NewMockRedisClient()

	// 创建Basic Server集成
	integration := NewBasicServerIntegration(mockRedis)

	// 启动集成
	if err := integration.Start(); err != nil {
		t.Fatalf("启动Basic Server集成失败: %v", err)
	}
	defer integration.Stop()

	// 创建测试用户
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

	// 测试用户创建
	if err := integration.OnUserCreated(user); err != nil {
		t.Fatalf("用户创建处理失败: %v", err)
	}

	// 测试用户更新
	changes := map[string]interface{}{
		"role": "system_admin",
	}
	if err := integration.OnUserUpdated(user, changes); err != nil {
		t.Fatalf("用户更新处理失败: %v", err)
	}

	// 测试用户状态变更
	if err := integration.OnUserStatusChanged(user, "active", "inactive"); err != nil {
		t.Fatalf("用户状态变更处理失败: %v", err)
	}

	// 测试用户删除
	if err := integration.OnUserDeleted(user.ID, user.Username); err != nil {
		t.Fatalf("用户删除处理失败: %v", err)
	}
}
