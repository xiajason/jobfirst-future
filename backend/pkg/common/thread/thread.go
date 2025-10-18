package thread

import (
	"context"
	"sync"
	"time"
)

// ThreadPool 线程池
type ThreadPool struct {
	workers    int
	tasks      chan Task
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	stats      *PoolStats
	statsMutex sync.RWMutex
}

// Task 任务接口
type Task interface {
	Execute() error
	GetID() string
}

// PoolStats 线程池统计
type PoolStats struct {
	TotalTasks     int64 `json:"total_tasks"`     // 总任务数
	CompletedTasks int64 `json:"completed_tasks"` // 完成任务数
	FailedTasks    int64 `json:"failed_tasks"`    // 失败任务数
	ActiveWorkers  int64 `json:"active_workers"`  // 活跃工作线程数
	QueueSize      int   `json:"queue_size"`      // 队列大小
}

// NewThreadPool 创建线程池
func NewThreadPool(workers int, queueSize int) *ThreadPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &ThreadPool{
		workers: workers,
		tasks:   make(chan Task, queueSize),
		ctx:     ctx,
		cancel:  cancel,
		stats:   &PoolStats{},
	}

	// 启动工作线程
	for i := 0; i < workers; i++ {
		pool.wg.Add(1)
		go pool.worker(i)
	}

	return pool
}

// worker 工作线程
func (pool *ThreadPool) worker(id int) {
	defer pool.wg.Done()

	for {
		select {
		case task, ok := <-pool.tasks:
			if !ok {
				return
			}

			pool.statsMutex.Lock()
			pool.stats.ActiveWorkers++
			pool.stats.TotalTasks++
			pool.statsMutex.Unlock()

			// 执行任务
			err := task.Execute()

			pool.statsMutex.Lock()
			pool.stats.ActiveWorkers--
			if err != nil {
				pool.stats.FailedTasks++
			} else {
				pool.stats.CompletedTasks++
			}
			pool.statsMutex.Unlock()

		case <-pool.ctx.Done():
			return
		}
	}
}

// Submit 提交任务
func (pool *ThreadPool) Submit(task Task) error {
	select {
	case pool.tasks <- task:
		pool.statsMutex.Lock()
		pool.stats.QueueSize = len(pool.tasks)
		pool.statsMutex.Unlock()
		return nil
	case <-pool.ctx.Done():
		return context.Canceled
	default:
		return context.DeadlineExceeded
	}
}

// SubmitWithTimeout 带超时的任务提交
func (pool *ThreadPool) SubmitWithTimeout(task Task, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(pool.ctx, timeout)
	defer cancel()

	select {
	case pool.tasks <- task:
		pool.statsMutex.Lock()
		pool.stats.QueueSize = len(pool.tasks)
		pool.statsMutex.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Shutdown 关闭线程池
func (pool *ThreadPool) Shutdown() {
	pool.cancel()
	close(pool.tasks)
	pool.wg.Wait()
}

// GetStats 获取统计信息
func (pool *ThreadPool) GetStats() *PoolStats {
	pool.statsMutex.RLock()
	defer pool.statsMutex.RUnlock()

	stats := *pool.stats
	stats.QueueSize = len(pool.tasks)
	return &stats
}

// GetQueueSize 获取队列大小
func (pool *ThreadPool) GetQueueSize() int {
	return len(pool.tasks)
}

// GetActiveWorkers 获取活跃工作线程数
func (pool *ThreadPool) GetActiveWorkers() int64 {
	pool.statsMutex.RLock()
	defer pool.statsMutex.RUnlock()
	return pool.stats.ActiveWorkers
}
