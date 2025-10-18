# 用户数据同步机制

⚠️ **注意：此包已废弃，将被集群级数据同步机制替代**

## 概述

用户数据同步机制是一个基于事件驱动的分布式数据同步系统，用于确保用户数据在多个微服务之间保持一致。该系统支持实时同步、异步处理、错误重试和一致性检查。

## 特性

- **事件驱动架构** - 基于用户操作事件触发同步
- **异步处理** - 不阻塞用户操作主流程
- **多目标同步** - 支持同步到多个服务和数据库
- **错误重试** - 智能重试机制和指数退避
- **一致性检查** - 定期检查和自动修复数据不一致
- **监控统计** - 完整的同步状态监控和统计
- **可配置** - 灵活的配置选项

## 架构组件

### 1. 用户同步服务 (UserSyncService)
核心同步服务，负责任务队列管理和同步执行。

### 2. 事件发布器 (UserEventPublisher)
负责发布用户操作事件到Redis Stream。

### 3. 事件订阅器 (UserEventSubscriber)
负责订阅和处理用户操作事件。

### 4. HTTP同步执行器 (HTTPSyncExecutor)
负责通过HTTP API同步数据到其他服务。

## 快速开始

### 1. 基本使用

```go
package main

import (
    "log"
    "time"
    
    "github.com/redis/go-redis/v9"
    "github.com/xiajason/zervi-basic/basic/backend/pkg/user-sync"
)

func main() {
    // 创建Redis客户端
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        DB:   0,
    })

    // 创建同步服务
    config := usersync.DefaultSyncConfig()
    syncService := usersync.NewUserSyncService(config, redisClient)

    // 启动同步服务
    if err := syncService.Start(); err != nil {
        log.Fatal(err)
    }
    defer syncService.Stop()

    // 创建事件发布器
    publisher := usersync.NewUserEventPublisher(redisClient, "my-service")

    // 发布用户创建事件
    user := &usersync.User{
        ID:        1,
        Username:  "testuser",
        Email:     "test@example.com",
        Role:      "guest",
        Status:    "active",
        CreatedAt: timePtr(time.Now()),
        UpdatedAt: timePtr(time.Now()),
    }

    if err := publisher.PublishUserCreated(user); err != nil {
        log.Printf("发布事件失败: %v", err)
    }
}

func timePtr(t time.Time) *time.Time {
    return &t
}
```

### 2. 集成到现有服务

```go
// 在Basic Server中集成
type BasicServer struct {
    userSyncIntegration *UserSyncIntegration
}

func (s *BasicServer) OnUserCreated(userID uint, username, email, role, status, phone string) error {
    return s.userSyncIntegration.OnUserCreated(userID, username, email, role, status, phone)
}

func (s *BasicServer) OnUserUpdated(userID uint, username, email, role, status, phone string, changes map[string]interface{}) error {
    return s.userSyncIntegration.OnUserUpdated(userID, username, email, role, status, phone, changes)
}
```

## 配置选项

### 同步配置 (SyncConfig)

```go
config := &usersync.SyncConfig{
    Enabled:           true,              // 启用同步服务
    Workers:           3,                 // 工作协程数
    QueueSize:         1000,              // 队列大小
    RetryInterval:     5 * time.Second,   // 重试间隔
    MaxRetries:        3,                 // 最大重试次数
    Timeout:           30 * time.Second,  // 超时时间
    ConsistencyCheck:  true,              // 启用一致性检查
    CheckInterval:     5 * time.Minute,   // 检查间隔
    AutoRepair:        true,              // 自动修复
}
```

### 同步目标配置

```go
targets := []usersync.SyncTarget{
    {
        Service: "unified-auth",
        URL:     "http://localhost:8207/api/v1/auth/sync/user",
        Method:  "POST",
        Enabled: true,
    },
    {
        Service: "user-service",
        URL:     "http://localhost:8081/api/v1/users/sync",
        Method:  "POST",
        Enabled: true,
    },
    {
        Service: "redis-cache",
        URL:     "redis://localhost:6379",
        Method:  "SET",
        Enabled: true,
    },
}
```

## 事件类型

### 支持的事件类型

- `user.created` - 用户创建
- `user.updated` - 用户更新
- `user.deleted` - 用户删除
- `user.status_changed` - 用户状态变更

### 事件结构

```go
type UserEvent struct {
    ID        string      `json:"id"`
    Type      EventType   `json:"type"`
    UserID    uint        `json:"user_id"`
    Username  string      `json:"username"`
    Email     string      `json:"email"`
    Data      interface{} `json:"data"`
    Timestamp time.Time   `json:"timestamp"`
    Source    string      `json:"source"`
}
```

## API接口

### 同步接收端点

其他服务需要实现以下端点来接收同步数据：

#### 统一认证服务
```
POST /api/v1/auth/sync/user
Content-Type: application/json

{
    "task_id": "user_sync_1_1234567890",
    "user_id": 1,
    "username": "testuser",
    "event_type": "user.created",
    "data": {
        "user": {
            "id": 1,
            "username": "testuser",
            "email": "test@example.com",
            "role": "guest",
            "status": "active"
        }
    },
    "timestamp": "2025-09-19T21:51:28Z"
}
```

#### 用户服务
```
POST /api/v1/users/sync
Content-Type: application/json

{
    "task_id": "user_sync_1_1234567890",
    "user_id": 1,
    "username": "testuser",
    "event_type": "user.created",
    "data": {
        "user": {
            "id": 1,
            "username": "testuser",
            "email": "test@example.com",
            "role": "guest",
            "status": "active"
        }
    },
    "timestamp": "2025-09-19T21:51:28Z"
}
```

## 监控和统计

### 获取同步统计

```go
stats := syncService.GetStats()
// 返回:
// {
//     "total_tasks": 100,
//     "completed_tasks": 95,
//     "failed_tasks": 3,
//     "retry_tasks": 2
// }
```

### 获取队列状态

```go
queueStatus := syncService.GetQueueStatus()
// 返回:
// {
//     "queue_length": 5,
//     "queue_cap": 1000,
//     "workers": 3,
//     "is_running": true,
//     "enabled": true
// }
```

## 测试

### 运行测试

```bash
cd basic/backend/pkg/user-sync
go test -v
```

### 测试覆盖

- 用户同步服务测试
- 事件发布器测试
- HTTP同步执行器测试
- 集成示例测试
- 配置测试

## 部署和运维

### 环境要求

- Go 1.21+
- Redis 6.0+
- 网络连接到目标服务

### 部署建议

1. **Redis配置**
   - 启用持久化
   - 配置内存限制
   - 设置适当的超时

2. **服务配置**
   - 根据负载调整工作协程数
   - 设置合适的队列大小
   - 配置合理的超时时间

3. **监控告警**
   - 监控同步成功率
   - 监控队列长度
   - 设置失败率告警

### 故障排除

#### 常见问题

1. **同步失败**
   - 检查目标服务是否可用
   - 验证网络连接
   - 检查API端点是否正确

2. **队列积压**
   - 增加工作协程数
   - 检查目标服务性能
   - 考虑批量同步

3. **Redis连接问题**
   - 检查Redis服务状态
   - 验证连接配置
   - 检查网络连接

## 最佳实践

1. **错误处理**
   - 实现适当的重试机制
   - 记录详细的错误日志
   - 设置告警阈值

2. **性能优化**
   - 使用批量同步
   - 合理配置队列大小
   - 监控资源使用

3. **数据一致性**
   - 定期运行一致性检查
   - 实现数据修复机制
   - 监控数据质量

4. **安全性**
   - 使用HTTPS进行同步
   - 实现API认证
   - 验证数据完整性

## 扩展和定制

### 添加新的同步目标

```go
// 在getDefaultSyncTargets中添加新的目标
{
    Service: "new-service",
    URL:     "http://localhost:8080/api/v1/sync/user",
    Method:  "POST",
    Enabled: true,
}
```

### 自定义事件处理器

```go
subscriber.RegisterHandler(usersync.EventTypeUserCreated, func(event usersync.UserEvent) error {
    // 自定义处理逻辑
    log.Printf("处理用户创建事件: %s", event.Username)
    return nil
})
```

### 自定义同步执行器

```go
type CustomSyncExecutor struct {
    // 自定义实现
}

func (e *CustomSyncExecutor) ExecuteSync(ctx context.Context, task usersync.UserSyncTask, target usersync.SyncTarget) error {
    // 自定义同步逻辑
    return nil
}
```

## 许可证

本项目采用MIT许可证。
