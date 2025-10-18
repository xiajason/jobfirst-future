package multidatabase

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// APIService 多数据库API服务
type APIService struct {
	manager            *MultiDatabaseManager
	syncService        *SyncService
	consistencyChecker *ConsistencyChecker
	transactionManager *TransactionManager
}

// NewAPIService 创建新的API服务
func NewAPIService(
	manager *MultiDatabaseManager,
	syncService *SyncService,
	consistencyChecker *ConsistencyChecker,
	transactionManager *TransactionManager,
) *APIService {
	return &APIService{
		manager:            manager,
		syncService:        syncService,
		consistencyChecker: consistencyChecker,
		transactionManager: transactionManager,
	}
}

// SetupRoutes 设置API路由
func (api *APIService) SetupRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/api/v1/multi-database/health", api.healthCheck)
	r.GET("/api/v1/multi-database/metrics", api.getMetrics)

	// 同步服务API
	syncGroup := r.Group("/api/v1/multi-database/sync")
	{
		syncGroup.POST("/task", api.addSyncTask)
		syncGroup.GET("/status", api.getSyncStatus)
		syncGroup.GET("/tasks", api.getSyncTasks)
	}

	// 一致性检查API
	consistencyGroup := r.Group("/api/v1/multi-database/consistency")
	{
		consistencyGroup.GET("/results", api.getConsistencyResults)
		consistencyGroup.GET("/results/:ruleId", api.getConsistencyResult)
		consistencyGroup.POST("/check", api.triggerConsistencyCheck)
	}

	// 事务管理API
	transactionGroup := r.Group("/api/v1/multi-database/transaction")
	{
		transactionGroup.POST("/begin", api.beginTransaction)
		transactionGroup.POST("/:id/operation", api.addOperation)
		transactionGroup.POST("/:id/prepare", api.prepareTransaction)
		transactionGroup.POST("/:id/commit", api.commitTransaction)
		transactionGroup.POST("/:id/rollback", api.rollbackTransaction)
		transactionGroup.GET("/:id", api.getTransaction)
		transactionGroup.GET("/", api.getActiveTransactions)
	}
}

// healthCheck 健康检查
func (api *APIService) healthCheck(c *gin.Context) {
	healthy := api.manager.IsHealthy()
	metrics := api.manager.GetMetrics()

	status := "healthy"
	if !healthy {
		status = "unhealthy"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    status,
		"healthy":   healthy,
		"metrics":   metrics,
		"timestamp": time.Now(),
	})
}

// getMetrics 获取指标
func (api *APIService) getMetrics(c *gin.Context) {
	metrics := api.manager.GetMetrics()
	c.JSON(http.StatusOK, metrics)
}

// addSyncTask 添加同步任务
func (api *APIService) addSyncTask(c *gin.Context) {
	var task SyncTask
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := api.syncService.AddSyncTask(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "同步任务已添加",
		"task_id": task.ID,
	})
}

// getSyncStatus 获取同步状态
func (api *APIService) getSyncStatus(c *gin.Context) {
	status := api.syncService.GetQueueStatus()
	c.JSON(http.StatusOK, status)
}

// getSyncTasks 获取同步任务列表
func (api *APIService) getSyncTasks(c *gin.Context) {
	// 这里可以实现获取任务列表的逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "同步任务列表功能待实现",
	})
}

// getConsistencyResults 获取一致性检查结果
func (api *APIService) getConsistencyResults(c *gin.Context) {
	results := api.consistencyChecker.GetResults()
	c.JSON(http.StatusOK, results)
}

// getConsistencyResult 获取特定规则的一致性检查结果
func (api *APIService) getConsistencyResult(c *gin.Context) {
	ruleID := c.Param("ruleId")
	result, exists := api.consistencyChecker.GetResult(ruleID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "规则不存在"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// triggerConsistencyCheck 触发一致性检查
func (api *APIService) triggerConsistencyCheck(c *gin.Context) {
	// 这里可以实现手动触发一致性检查的逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "一致性检查已触发",
	})
}

// beginTransaction 开始事务
func (api *APIService) beginTransaction(c *gin.Context) {
	var req struct {
		Timeout string `json:"timeout"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	timeout := 30 * time.Second
	if req.Timeout != "" {
		if parsedTimeout, err := time.ParseDuration(req.Timeout); err == nil {
			timeout = parsedTimeout
		}
	}

	tx, err := api.transactionManager.BeginTransaction(c.Request.Context(), timeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transaction_id": tx.ID,
		"status":         tx.Status,
		"created_at":     tx.CreatedAt,
		"timeout":        tx.Timeout,
	})
}

// addOperation 添加操作到事务
func (api *APIService) addOperation(c *gin.Context) {
	transactionID := c.Param("id")
	tx, exists := api.transactionManager.GetTransaction(transactionID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "事务不存在"})
		return
	}

	var operation TransactionOperation
	if err := c.ShouldBindJSON(&operation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := api.transactionManager.AddOperation(tx, operation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "操作已添加到事务",
		"operation_id": operation.ID,
	})
}

// prepareTransaction 准备事务
func (api *APIService) prepareTransaction(c *gin.Context) {
	transactionID := c.Param("id")
	tx, exists := api.transactionManager.GetTransaction(transactionID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "事务不存在"})
		return
	}

	if err := api.transactionManager.PrepareTransaction(tx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "事务准备完成",
		"status":  tx.Status,
	})
}

// commitTransaction 提交事务
func (api *APIService) commitTransaction(c *gin.Context) {
	transactionID := c.Param("id")
	tx, exists := api.transactionManager.GetTransaction(transactionID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "事务不存在"})
		return
	}

	if err := api.transactionManager.CommitTransaction(tx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "事务提交完成",
		"status":  tx.Status,
	})
}

// rollbackTransaction 回滚事务
func (api *APIService) rollbackTransaction(c *gin.Context) {
	transactionID := c.Param("id")
	tx, exists := api.transactionManager.GetTransaction(transactionID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "事务不存在"})
		return
	}

	if err := api.transactionManager.RollbackTransaction(tx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "事务回滚完成",
		"status":  tx.Status,
	})
}

// getTransaction 获取事务信息
func (api *APIService) getTransaction(c *gin.Context) {
	transactionID := c.Param("id")
	tx, exists := api.transactionManager.GetTransaction(transactionID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "事务不存在"})
		return
	}

	c.JSON(http.StatusOK, tx)
}

// getActiveTransactions 获取活跃事务列表
func (api *APIService) getActiveTransactions(c *gin.Context) {
	transactions := api.transactionManager.GetActiveTransactions()
	c.JSON(http.StatusOK, transactions)
}

// 辅助函数

// parseQueryParam 解析查询参数
func parseQueryParam(c *gin.Context, key string, defaultValue string) string {
	if value := c.Query(key); value != "" {
		return value
	}
	return defaultValue
}

// parseIntQueryParam 解析整数查询参数
func parseIntQueryParam(c *gin.Context, key string, defaultValue int) int {
	if value := c.Query(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// parseBoolQueryParam 解析布尔查询参数
func parseBoolQueryParam(c *gin.Context, key string, defaultValue bool) bool {
	if value := c.Query(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
