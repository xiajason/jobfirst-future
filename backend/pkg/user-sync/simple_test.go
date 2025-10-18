package usersync

import (
	"testing"
	"time"
)

// TestUserSyncServiceWithoutRedis 测试用户同步服务（不依赖Redis）
func TestUserSyncServiceWithoutRedis(t *testing.T) {
	// 创建同步服务（不使用Redis）
	config := DefaultSyncConfig()
	config.Enabled = true
	config.Workers = 1
	config.QueueSize = 10
	config.Timeout = 5 * time.Second

	syncService := NewUserSyncService(config, nil)

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
				Service: "unified-auth",
				URL:     "http://localhost:8207/api/v1/auth/sync/user",
				Method:  "POST",
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

// TestUserEventPublisherWithoutRedis 测试用户事件发布器（不依赖Redis）
func TestUserEventPublisherWithoutRedis(t *testing.T) {
	// 创建事件发布器（不使用Redis）
	publisher := NewUserEventPublisher(nil, "test-service")

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

	// 测试发布用户创建事件（应该成功，因为Redis为nil时会跳过）
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

// TestBasicServerIntegrationWithoutRedis 测试Basic Server集成（不依赖Redis）
func TestBasicServerIntegrationWithoutRedis(t *testing.T) {
	// 创建Basic Server集成（不使用Redis）
	integration := NewBasicServerIntegration(nil)

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

// TestSyncConfig 测试同步配置
func TestSyncConfig(t *testing.T) {
	config := DefaultSyncConfig()

	if !config.Enabled {
		t.Error("期望默认配置启用同步服务")
	}

	if config.Workers != 3 {
		t.Errorf("期望默认工作协程数为3，实际为%d", config.Workers)
	}

	if config.QueueSize != 1000 {
		t.Errorf("期望默认队列大小为1000，实际为%d", config.QueueSize)
	}

	if config.MaxRetries != 3 {
		t.Errorf("期望默认最大重试次数为3，实际为%d", config.MaxRetries)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("期望默认超时时间为30秒，实际为%v", config.Timeout)
	}
}

// TestUserEventTypes 测试用户事件类型
func TestUserEventTypes(t *testing.T) {
	// 测试事件类型常量
	if EventTypeUserCreated != "user.created" {
		t.Errorf("期望用户创建事件类型为'user.created'，实际为'%s'", EventTypeUserCreated)
	}

	if EventTypeUserUpdated != "user.updated" {
		t.Errorf("期望用户更新事件类型为'user.updated'，实际为'%s'", EventTypeUserUpdated)
	}

	if EventTypeUserDeleted != "user.deleted" {
		t.Errorf("期望用户删除事件类型为'user.deleted'，实际为'%s'", EventTypeUserDeleted)
	}

	if EventTypeUserStatusChanged != "user.status_changed" {
		t.Errorf("期望用户状态变更事件类型为'user.status_changed'，实际为'%s'", EventTypeUserStatusChanged)
	}
}

// TestSyncTaskStatus 测试同步任务状态
func TestSyncTaskStatus(t *testing.T) {
	// 测试任务状态常量
	if SyncTaskStatusPending != "pending" {
		t.Errorf("期望待处理状态为'pending'，实际为'%s'", SyncTaskStatusPending)
	}

	if SyncTaskStatusProcessing != "processing" {
		t.Errorf("期望处理中状态为'processing'，实际为'%s'", SyncTaskStatusProcessing)
	}

	if SyncTaskStatusCompleted != "completed" {
		t.Errorf("期望已完成状态为'completed'，实际为'%s'", SyncTaskStatusCompleted)
	}

	if SyncTaskStatusFailed != "failed" {
		t.Errorf("期望失败状态为'failed'，实际为'%s'", SyncTaskStatusFailed)
	}

	if SyncTaskStatusRetrying != "retrying" {
		t.Errorf("期望重试状态为'retrying'，实际为'%s'", SyncTaskStatusRetrying)
	}
}
