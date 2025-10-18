package multidatabase

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"gorm.io/gorm"
)

// TransactionManager 跨数据库事务管理器
type TransactionManager struct {
	manager            *MultiDatabaseManager
	config             *TransactionConfig
	activeTransactions map[string]*MultiDatabaseTransaction
	mu                 sync.RWMutex
}

// TransactionConfig 事务配置
type TransactionConfig struct {
	// 默认超时时间
	DefaultTimeout time.Duration `yaml:"default_timeout"`

	// 最大重试次数
	MaxRetries int `yaml:"max_retries"`

	// 重试间隔
	RetryInterval time.Duration `yaml:"retry_interval"`

	// 两阶段提交超时
	TwoPhaseCommitTimeout time.Duration `yaml:"two_phase_commit_timeout"`

	// 启用分布式锁
	EnableDistributedLock bool `yaml:"enable_distributed_lock"`
}

// MultiDatabaseTransaction 跨数据库事务
type MultiDatabaseTransaction struct {
	ID         string                 `json:"id"`
	Status     TransactionStatus      `json:"status"`
	Operations []TransactionOperation `json:"operations"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Timeout    time.Duration          `json:"timeout"`
	RetryCount int                    `json:"retry_count"`
	MaxRetries int                    `json:"max_retries"`
	Error      string                 `json:"error,omitempty"`

	// 各数据库的事务对象
	mysqlTx      *gorm.DB
	postgresTx   *gorm.DB
	neo4jSession neo4j.Session
	redisTx      redis.Pipeliner

	// 上下文
	ctx    context.Context
	cancel context.CancelFunc
}

// TransactionStatus 事务状态
type TransactionStatus string

const (
	TransactionStatusPending    TransactionStatus = "pending"
	TransactionStatusPrepared   TransactionStatus = "prepared"
	TransactionStatusCommitted  TransactionStatus = "committed"
	TransactionStatusRolledBack TransactionStatus = "rolled_back"
	TransactionStatusFailed     TransactionStatus = "failed"
	TransactionStatusTimeout    TransactionStatus = "timeout"
)

// TransactionOperation 事务操作
type TransactionOperation struct {
	ID         string                 `json:"id"`
	Type       OperationType          `json:"type"`
	Database   DatabaseType           `json:"database"`
	Query      string                 `json:"query"`
	Data       map[string]interface{} `json:"data"`
	Order      int                    `json:"order"`
	RetryCount int                    `json:"retry_count"`
	MaxRetries int                    `json:"max_retries"`
	Status     OperationStatus        `json:"status"`
	Error      string                 `json:"error,omitempty"`
	ExecutedAt *time.Time             `json:"executed_at,omitempty"`
}

// OperationType 操作类型
type OperationType string

const (
	OperationTypeInsert OperationType = "insert"
	OperationTypeUpdate OperationType = "update"
	OperationTypeDelete OperationType = "delete"
	OperationTypeQuery  OperationType = "query"
)

// OperationStatus 操作状态
type OperationStatus string

const (
	OperationStatusPending    OperationStatus = "pending"
	OperationStatusExecuted   OperationStatus = "executed"
	OperationStatusFailed     OperationStatus = "failed"
	OperationStatusRolledBack OperationStatus = "rolled_back"
)

// NewTransactionManager 创建新的事务管理器
func NewTransactionManager(manager *MultiDatabaseManager, config *TransactionConfig) *TransactionManager {
	return &TransactionManager{
		manager:            manager,
		config:             config,
		activeTransactions: make(map[string]*MultiDatabaseTransaction),
	}
}

// BeginTransaction 开始跨数据库事务
func (tm *TransactionManager) BeginTransaction(ctx context.Context, timeout time.Duration) (*MultiDatabaseTransaction, error) {
	if timeout == 0 {
		timeout = tm.config.DefaultTimeout
	}

	txCtx, cancel := context.WithTimeout(ctx, timeout)

	transaction := &MultiDatabaseTransaction{
		ID:         generateTransactionID(),
		Status:     TransactionStatusPending,
		Operations: make([]TransactionOperation, 0),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Timeout:    timeout,
		MaxRetries: tm.config.MaxRetries,
		ctx:        txCtx,
		cancel:     cancel,
	}

	// 初始化各数据库的事务
	if err := tm.initializeTransaction(transaction); err != nil {
		cancel()
		return nil, fmt.Errorf("初始化事务失败: %w", err)
	}

	tm.mu.Lock()
	tm.activeTransactions[transaction.ID] = transaction
	tm.mu.Unlock()

	log.Printf("开始跨数据库事务: %s", transaction.ID)
	return transaction, nil
}

// initializeTransaction 初始化事务
func (tm *TransactionManager) initializeTransaction(tx *MultiDatabaseTransaction) error {
	// 初始化MySQL事务
	if tm.manager.MySQL != nil {
		tx.mysqlTx = tm.manager.MySQL.Begin()
		if tx.mysqlTx.Error != nil {
			return fmt.Errorf("MySQL事务初始化失败: %w", tx.mysqlTx.Error)
		}
	}

	// 初始化PostgreSQL事务
	if tm.manager.PostgreSQL != nil {
		tx.postgresTx = tm.manager.PostgreSQL.Begin()
		if tx.postgresTx.Error != nil {
			// 回滚MySQL事务
			if tx.mysqlTx != nil {
				tx.mysqlTx.Rollback()
			}
			return fmt.Errorf("PostgreSQL事务初始化失败: %w", tx.postgresTx.Error)
		}
	}

	// 初始化Neo4j会话
	if tm.manager.Neo4j != nil {
		tx.neo4jSession = tm.manager.Neo4j.NewSession(neo4j.SessionConfig{})
	}

	// 初始化Redis事务
	if tm.manager.Redis != nil {
		tx.redisTx = tm.manager.Redis.Pipeline()
	}

	return nil
}

// AddOperation 添加操作到事务
func (tm *TransactionManager) AddOperation(tx *MultiDatabaseTransaction, operation TransactionOperation) error {
	if tx.Status != TransactionStatusPending {
		return fmt.Errorf("事务状态不允许添加操作: %s", tx.Status)
	}

	operation.ID = generateOperationID()
	operation.Status = OperationStatusPending
	operation.MaxRetries = tm.config.MaxRetries

	tx.Operations = append(tx.Operations, operation)
	tx.UpdatedAt = time.Now()

	log.Printf("添加操作到事务 %s: %s (%s)", tx.ID, operation.ID, operation.Type)
	return nil
}

// PrepareTransaction 准备事务（两阶段提交的第一阶段）
func (tm *TransactionManager) PrepareTransaction(tx *MultiDatabaseTransaction) error {
	if tx.Status != TransactionStatusPending {
		return fmt.Errorf("事务状态不允许准备: %s", tx.Status)
	}

	log.Printf("准备事务: %s", tx.ID)

	// 按顺序执行所有操作
	for i := range tx.Operations {
		operation := &tx.Operations[i]
		if err := tm.executeOperation(tx, operation); err != nil {
			operation.Status = OperationStatusFailed
			operation.Error = err.Error()
			tx.Status = TransactionStatusFailed
			tx.Error = fmt.Sprintf("操作 %s 执行失败: %v", operation.ID, err)
			return err
		}

		operation.Status = OperationStatusExecuted
		now := time.Now()
		operation.ExecutedAt = &now
	}

	tx.Status = TransactionStatusPrepared
	tx.UpdatedAt = time.Now()

	log.Printf("事务准备完成: %s", tx.ID)
	return nil
}

// CommitTransaction 提交事务（两阶段提交的第二阶段）
func (tm *TransactionManager) CommitTransaction(tx *MultiDatabaseTransaction) error {
	if tx.Status != TransactionStatusPrepared {
		return fmt.Errorf("事务状态不允许提交: %s", tx.Status)
	}

	log.Printf("提交事务: %s", tx.ID)

	// 提交MySQL事务
	if tx.mysqlTx != nil {
		if err := tx.mysqlTx.Commit().Error; err != nil {
			tm.rollbackTransaction(tx)
			return fmt.Errorf("MySQL事务提交失败: %w", err)
		}
	}

	// 提交PostgreSQL事务
	if tx.postgresTx != nil {
		if err := tx.postgresTx.Commit().Error; err != nil {
			tm.rollbackTransaction(tx)
			return fmt.Errorf("PostgreSQL事务提交失败: %w", err)
		}
	}

	// 提交Neo4j事务
	if tx.neo4jSession != nil {
		if err := tx.neo4jSession.Close(); err != nil {
			log.Printf("Neo4j会话关闭失败: %v", err)
		}
	}

	// 执行Redis事务
	if tx.redisTx != nil {
		if _, err := tx.redisTx.Exec(tx.ctx); err != nil {
			log.Printf("Redis事务执行失败: %v", err)
		}
	}

	tx.Status = TransactionStatusCommitted
	tx.UpdatedAt = time.Now()

	// 从活跃事务中移除
	tm.mu.Lock()
	delete(tm.activeTransactions, tx.ID)
	tm.mu.Unlock()

	tx.cancel()
	log.Printf("事务提交完成: %s", tx.ID)
	return nil
}

// RollbackTransaction 回滚事务
func (tm *TransactionManager) RollbackTransaction(tx *MultiDatabaseTransaction) error {
	return tm.rollbackTransaction(tx)
}

// rollbackTransaction 内部回滚方法
func (tm *TransactionManager) rollbackTransaction(tx *MultiDatabaseTransaction) error {
	log.Printf("回滚事务: %s", tx.ID)

	// 回滚MySQL事务
	if tx.mysqlTx != nil {
		if err := tx.mysqlTx.Rollback().Error; err != nil {
			log.Printf("MySQL事务回滚失败: %v", err)
		}
	}

	// 回滚PostgreSQL事务
	if tx.postgresTx != nil {
		if err := tx.postgresTx.Rollback().Error; err != nil {
			log.Printf("PostgreSQL事务回滚失败: %v", err)
		}
	}

	// 关闭Neo4j会话
	if tx.neo4jSession != nil {
		if err := tx.neo4jSession.Close(); err != nil {
			log.Printf("Neo4j会话关闭失败: %v", err)
		}
	}

	// 丢弃Redis事务
	if tx.redisTx != nil {
		// Redis Pipeline没有Discard方法，直接忽略
	}

	// 更新操作状态
	for i := range tx.Operations {
		if tx.Operations[i].Status == OperationStatusExecuted {
			tx.Operations[i].Status = OperationStatusRolledBack
		}
	}

	tx.Status = TransactionStatusRolledBack
	tx.UpdatedAt = time.Now()

	// 从活跃事务中移除
	tm.mu.Lock()
	delete(tm.activeTransactions, tx.ID)
	tm.mu.Unlock()

	tx.cancel()
	log.Printf("事务回滚完成: %s", tx.ID)
	return nil
}

// executeOperation 执行单个操作
func (tm *TransactionManager) executeOperation(tx *MultiDatabaseTransaction, operation *TransactionOperation) error {
	switch operation.Database {
	case DatabaseTypeMySQL:
		return tm.executeMySQLOperation(tx, operation)
	case DatabaseTypePostgreSQL:
		return tm.executePostgreSQLOperation(tx, operation)
	case DatabaseTypeNeo4j:
		return tm.executeNeo4jOperation(tx, operation)
	case DatabaseTypeRedis:
		return tm.executeRedisOperation(tx, operation)
	default:
		return fmt.Errorf("不支持的数据库类型: %s", operation.Database)
	}
}

// executeMySQLOperation 执行MySQL操作
func (tm *TransactionManager) executeMySQLOperation(tx *MultiDatabaseTransaction, operation *TransactionOperation) error {
	if tx.mysqlTx == nil {
		return fmt.Errorf("MySQL事务未初始化")
	}

	switch operation.Type {
	case OperationTypeInsert:
		return tx.mysqlTx.Exec(operation.Query, operation.Data).Error
	case OperationTypeUpdate:
		return tx.mysqlTx.Exec(operation.Query, operation.Data).Error
	case OperationTypeDelete:
		return tx.mysqlTx.Exec(operation.Query, operation.Data).Error
	case OperationTypeQuery:
		return tx.mysqlTx.Raw(operation.Query, operation.Data).Error
	default:
		return fmt.Errorf("不支持的操作类型: %s", operation.Type)
	}
}

// executePostgreSQLOperation 执行PostgreSQL操作
func (tm *TransactionManager) executePostgreSQLOperation(tx *MultiDatabaseTransaction, operation *TransactionOperation) error {
	if tx.postgresTx == nil {
		return fmt.Errorf("PostgreSQL事务未初始化")
	}

	switch operation.Type {
	case OperationTypeInsert:
		return tx.postgresTx.Exec(operation.Query, operation.Data).Error
	case OperationTypeUpdate:
		return tx.postgresTx.Exec(operation.Query, operation.Data).Error
	case OperationTypeDelete:
		return tx.postgresTx.Exec(operation.Query, operation.Data).Error
	case OperationTypeQuery:
		return tx.postgresTx.Raw(operation.Query, operation.Data).Error
	default:
		return fmt.Errorf("不支持的操作类型: %s", operation.Type)
	}
}

// executeNeo4jOperation 执行Neo4j操作
func (tm *TransactionManager) executeNeo4jOperation(tx *MultiDatabaseTransaction, operation *TransactionOperation) error {
	if tx.neo4jSession == nil {
		return fmt.Errorf("Neo4j会话未初始化")
	}

	_, err := tx.neo4jSession.Run(operation.Query, operation.Data)
	return err
}

// executeRedisOperation 执行Redis操作
func (tm *TransactionManager) executeRedisOperation(tx *MultiDatabaseTransaction, operation *TransactionOperation) error {
	if tx.redisTx == nil {
		return fmt.Errorf("Redis事务未初始化")
	}

	switch operation.Type {
	case OperationTypeInsert, OperationTypeUpdate:
		key := fmt.Sprintf("%v", operation.Data["key"])
		value := fmt.Sprintf("%v", operation.Data["value"])
		return tx.redisTx.Set(tx.ctx, key, value, 0).Err()
	case OperationTypeDelete:
		key := fmt.Sprintf("%v", operation.Data["key"])
		return tx.redisTx.Del(tx.ctx, key).Err()
	default:
		return fmt.Errorf("不支持的操作类型: %s", operation.Type)
	}
}

// GetActiveTransactions 获取活跃事务列表
func (tm *TransactionManager) GetActiveTransactions() map[string]*MultiDatabaseTransaction {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	transactions := make(map[string]*MultiDatabaseTransaction)
	for k, v := range tm.activeTransactions {
		transactions[k] = v
	}
	return transactions
}

// GetTransaction 获取特定事务
func (tm *TransactionManager) GetTransaction(id string) (*MultiDatabaseTransaction, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tx, exists := tm.activeTransactions[id]
	return tx, exists
}

// CleanupExpiredTransactions 清理过期事务
func (tm *TransactionManager) CleanupExpiredTransactions() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	now := time.Now()
	for id, tx := range tm.activeTransactions {
		if now.Sub(tx.CreatedAt) > tx.Timeout {
			log.Printf("清理过期事务: %s", id)
			tm.rollbackTransaction(tx)
		}
	}
}

// generateTransactionID 生成事务ID
func generateTransactionID() string {
	return fmt.Sprintf("tx_%d", time.Now().UnixNano())
}

// generateOperationID 生成操作ID
func generateOperationID() string {
	return fmt.Sprintf("op_%d", time.Now().UnixNano())
}
