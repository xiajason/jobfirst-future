package usersync

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// UserSyncService 用户数据同步服务
type UserSyncService struct {
	config       *SyncConfig
	redisClient  *redis.Client
	httpExecutor *HTTPSyncExecutor
	queue        chan UserSyncTask
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	mu           sync.RWMutex
	stats        *SyncStats
}

// SyncStats 同步统计
type SyncStats struct {
	TotalTasks     int64 `json:"total_tasks"`
	CompletedTasks int64 `json:"completed_tasks"`
	FailedTasks    int64 `json:"failed_tasks"`
	RetryTasks     int64 `json:"retry_tasks"`
	mu             sync.RWMutex
}

// NewUserSyncService 创建新的用户同步服务
func NewUserSyncService(config *SyncConfig, redisClient *redis.Client) *UserSyncService {
	if config == nil {
		config = DefaultSyncConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &UserSyncService{
		config:       config,
		redisClient:  redisClient,
		httpExecutor: NewHTTPSyncExecutor(config.Timeout),
		queue:        make(chan UserSyncTask, config.QueueSize),
		ctx:          ctx,
		cancel:       cancel,
		stats:        &SyncStats{},
	}
}

// Start 启动同步服务
func (s *UserSyncService) Start() error {
	if !s.config.Enabled {
		log.Println("用户数据同步服务已禁用")
		return nil
	}

	log.Printf("启动用户数据同步服务，工作协程数: %d", s.config.Workers)

	// 启动工作协程
	for i := 0; i < s.config.Workers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}

	// 启动一致性检查
	if s.config.ConsistencyCheck {
		s.wg.Add(1)
		go s.consistencyChecker()
	}

	log.Println("用户数据同步服务启动成功")
	return nil
}

// Stop 停止同步服务
func (s *UserSyncService) Stop() {
	log.Println("停止用户数据同步服务")
	s.cancel()
	close(s.queue)
	s.wg.Wait()
	log.Println("用户数据同步服务已停止")
}

// worker 工作协程
func (s *UserSyncService) worker(id int) {
	defer s.wg.Done()

	log.Printf("用户同步工作协程 %d 启动", id)

	for {
		select {
		case <-s.ctx.Done():
			log.Printf("用户同步工作协程 %d 停止", id)
			return
		case task, ok := <-s.queue:
			if !ok {
				log.Printf("用户同步工作协程 %d 队列关闭", id)
				return
			}

			s.processTask(task)
		}
	}
}

// processTask 处理同步任务
func (s *UserSyncService) processTask(task UserSyncTask) {
	startTime := time.Now()
	task.Status = SyncTaskStatusProcessing
	task.UpdatedAt = time.Now()

	log.Printf("处理用户同步任务: %s, 用户: %s, 类型: %s",
		task.ID, task.Username, task.EventType)

	// 更新统计
	s.stats.mu.Lock()
	s.stats.TotalTasks++
	s.stats.mu.Unlock()

	// 执行同步到各个目标
	successCount := 0
	totalTargets := len(task.Targets)

	for _, target := range task.Targets {
		if !target.Enabled {
			continue
		}

		result := s.executeSync(task, target)
		if result.Success {
			successCount++
		}

		// 记录同步结果
		s.recordSyncResult(result)
	}

	duration := time.Since(startTime)

	// 更新任务状态
	if successCount == totalTargets {
		task.Status = SyncTaskStatusCompleted
		s.stats.mu.Lock()
		s.stats.CompletedTasks++
		s.stats.mu.Unlock()
		log.Printf("用户同步任务完成: %s, 耗时: %v", task.ID, duration)
	} else if successCount > 0 {
		// 部分成功，标记为失败但记录部分成功
		task.Status = SyncTaskStatusFailed
		task.Error = fmt.Sprintf("部分同步失败: %d/%d", successCount, totalTargets)
		s.stats.mu.Lock()
		s.stats.FailedTasks++
		s.stats.mu.Unlock()
		log.Printf("用户同步任务部分失败: %s, 成功: %d/%d", task.ID, successCount, totalTargets)
	} else {
		// 完全失败，检查是否需要重试
		task.Status = SyncTaskStatusFailed
		task.Error = "所有同步目标都失败"
		task.RetryCount++

		s.stats.mu.Lock()
		s.stats.FailedTasks++
		s.stats.mu.Unlock()

		log.Printf("用户同步任务失败: %s, 重试次数: %d/%d",
			task.ID, task.RetryCount, s.config.MaxRetries)

		// 如果还有重试次数，重新加入队列
		if task.RetryCount < s.config.MaxRetries {
			task.Status = SyncTaskStatusRetrying
			s.stats.mu.Lock()
			s.stats.RetryTasks++
			s.stats.mu.Unlock()

			go func() {
				// 指数退避重试
				retryDelay := time.Duration(task.RetryCount) * s.config.RetryInterval
				time.Sleep(retryDelay)

				select {
				case s.queue <- task:
					log.Printf("用户同步任务重新加入队列: %s", task.ID)
				case <-s.ctx.Done():
					log.Printf("用户同步任务重试取消: %s", task.ID)
				}
			}()
		}
	}
}

// executeSync 执行同步到指定目标
func (s *UserSyncService) executeSync(task UserSyncTask, target SyncTarget) SyncResult {
	startTime := time.Now()
	result := SyncResult{
		TaskID:    task.ID,
		Target:    target.Service,
		Timestamp: time.Now(),
	}

	ctx, cancel := context.WithTimeout(s.ctx, s.config.Timeout)
	defer cancel()

	var err error
	switch target.Service {
	case "unified-auth":
		err = s.httpExecutor.SyncToUnifiedAuth(ctx, task)
	case "user-service":
		err = s.httpExecutor.SyncToUserService(ctx, task)
	case "basic-server":
		err = s.httpExecutor.SyncToBasicServer(ctx, task)
	case "redis-cache":
		err = s.syncToRedis(ctx, task, target)
	default:
		err = fmt.Errorf("不支持的同步目标: %s", target.Service)
	}

	result.Duration = time.Since(startTime)
	result.Success = err == nil
	if err != nil {
		result.Message = err.Error()
	}

	return result
}

// syncToRedis 同步到Redis缓存
func (s *UserSyncService) syncToRedis(ctx context.Context, task UserSyncTask, target SyncTarget) error {
	if s.redisClient == nil {
		return fmt.Errorf("Redis客户端未初始化")
	}

	key := fmt.Sprintf("user:%d", task.UserID)
	data, err := json.Marshal(task.Data)
	if err != nil {
		return fmt.Errorf("序列化用户数据失败: %w", err)
	}

	return s.redisClient.Set(ctx, key, data, 24*time.Hour).Err()
}

// recordSyncResult 记录同步结果
func (s *UserSyncService) recordSyncResult(result SyncResult) {
	// 这里可以将同步结果记录到数据库或日志中
	log.Printf("同步结果: %+v", result)
}

// AddSyncTask 添加同步任务
func (s *UserSyncService) AddSyncTask(task UserSyncTask) error {
	if !s.config.Enabled {
		return fmt.Errorf("用户数据同步服务已禁用")
	}

	if task.ID == "" {
		task.ID = fmt.Sprintf("user_sync_%d_%d", task.UserID, time.Now().UnixNano())
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	if task.MaxRetries == 0 {
		task.MaxRetries = s.config.MaxRetries
	}
	task.Status = SyncTaskStatusPending

	select {
	case s.queue <- task:
		log.Printf("用户同步任务已添加: %s", task.ID)
		return nil
	case <-s.ctx.Done():
		return fmt.Errorf("用户数据同步服务已停止")
	default:
		return fmt.Errorf("用户同步队列已满")
	}
}

// PublishUserEvent 发布用户事件
func (s *UserSyncService) PublishUserEvent(event UserEvent) error {
	if !s.config.Enabled {
		return nil
	}

	// 创建同步任务
	task := UserSyncTask{
		UserID:    event.UserID,
		Username:  event.Username,
		EventType: event.Type,
		Targets:   s.getDefaultSyncTargets(),
		Data:      map[string]interface{}{"event": event},
		Priority:  1,
	}

	return s.AddSyncTask(task)
}

// getDefaultSyncTargets 获取默认同步目标
func (s *UserSyncService) getDefaultSyncTargets() []SyncTarget {
	return []SyncTarget{
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
			Service: "basic-server",
			URL:     "http://localhost:8080/api/v1/auth/sync/user",
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
}

// consistencyChecker 一致性检查器
func (s *UserSyncService) consistencyChecker() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.CheckInterval)
	defer ticker.Stop()

	log.Println("用户数据一致性检查器启动")

	for {
		select {
		case <-s.ctx.Done():
			log.Println("用户数据一致性检查器停止")
			return
		case <-ticker.C:
			if err := s.checkUserConsistency(); err != nil {
				log.Printf("用户数据一致性检查失败: %v", err)
			}
		}
	}
}

// checkUserConsistency 检查用户数据一致性
func (s *UserSyncService) checkUserConsistency() error {
	log.Println("执行用户数据一致性检查")
	// 这里应该实现具体的一致性检查逻辑
	// 暂时返回成功
	return nil
}

// GetStats 获取同步统计
func (s *UserSyncService) GetStats() SyncStats {
	s.stats.mu.RLock()
	defer s.stats.mu.RUnlock()

	return SyncStats{
		TotalTasks:     s.stats.TotalTasks,
		CompletedTasks: s.stats.CompletedTasks,
		FailedTasks:    s.stats.FailedTasks,
		RetryTasks:     s.stats.RetryTasks,
	}
}

// GetQueueStatus 获取队列状态
func (s *UserSyncService) GetQueueStatus() map[string]interface{} {
	return map[string]interface{}{
		"queue_length": len(s.queue),
		"queue_cap":    cap(s.queue),
		"workers":      s.config.Workers,
		"is_running":   s.ctx.Err() == nil,
		"enabled":      s.config.Enabled,
	}
}
