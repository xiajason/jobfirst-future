package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jobfirst/jobfirst-core"
)

// AI优化功能API路由设置
func setupAIOptimizationRoutes(r *gin.Engine, core *jobfirst.Core) {
	authMiddleware := core.AuthMiddleware.RequireAuth()

	// AI服务分层API路由组
	aiLayering := r.Group("/api/v1/ai/layering")
	aiLayering.Use(authMiddleware)
	{
		// 获取用户可用服务
		aiLayering.GET("/services", func(c *gin.Context) { getUserAvailableServices(c, core) })
		// 获取服务分层信息
		aiLayering.GET("/info", func(c *gin.Context) { getServiceLayeringInfo(c, core) })
		// 升级用户服务层级
		aiLayering.PUT("/upgrade", func(c *gin.Context) { upgradeUserServiceLevel(c, core) })
		// 获取用户使用统计
		aiLayering.GET("/usage", func(c *gin.Context) { getUserUsageStats(c, core) })
		// 检查服务限制
		aiLayering.GET("/limits", func(c *gin.Context) { checkServiceLimits(c, core) })
		// 记录服务使用
		aiLayering.POST("/usage", func(c *gin.Context) { recordServiceUsage(c, core) })
	}

	// 个性化分析API路由组
	personalizedAnalysis := r.Group("/api/v1/ai/analysis")
	personalizedAnalysis.Use(authMiddleware)
	{
		// 更新用户画像
		personalizedAnalysis.PUT("/profile", func(c *gin.Context) { updateUserProfile(c, core) })
		// 获取用户画像
		personalizedAnalysis.GET("/profile", func(c *gin.Context) { getUserProfile(c, core) })
		// 执行个性化分析
		personalizedAnalysis.POST("/analyze", func(c *gin.Context) { performPersonalizedAnalysis(c, core) })
		// 获取个性化推荐
		personalizedAnalysis.GET("/recommendations", func(c *gin.Context) { getPersonalizedRecommendations(c, core) })
		// 获取分析历史
		personalizedAnalysis.GET("/history", func(c *gin.Context) { getAnalysisHistory(c, core) })
	}

	// 用户数据主权保护API路由组
	dataSovereignty := r.Group("/api/v1/ai/data-sovereignty")
	dataSovereignty.Use(authMiddleware)
	{
		// 获取用户数据主权状态
		dataSovereignty.GET("/status", func(c *gin.Context) { getDataSovereigntyStatus(c, core) })
		// 设置数据授权级别
		dataSovereignty.PUT("/authorization", func(c *gin.Context) { setDataAuthorizationLevel(c, core) })
		// 获取数据使用记录
		dataSovereignty.GET("/usage", func(c *gin.Context) { getDataUsageRecords(c, core) })
		// 撤回数据授权
		dataSovereignty.DELETE("/authorization", func(c *gin.Context) { revokeDataAuthorization(c, core) })
	}

	// 隐私保护API路由组
	privacyProtection := r.Group("/api/v1/ai/privacy")
	privacyProtection.Use(authMiddleware)
	{
		// 获取隐私设置
		privacyProtection.GET("/settings", func(c *gin.Context) { getPrivacySettings(c, core) })
		// 更新隐私设置
		privacyProtection.PUT("/settings", func(c *gin.Context) { updatePrivacySettings(c, core) })
		// 数据匿名化处理
		privacyProtection.POST("/anonymize", func(c *gin.Context) { anonymizeData(c, core) })
		// 获取隐私保护状态
		privacyProtection.GET("/status", func(c *gin.Context) { getPrivacyProtectionStatus(c, core) })
	}
}

// AI服务分层相关函数

// 获取用户可用服务
func getUserAvailableServices(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	// 获取用户授权级别
	authorizationLevel := c.Query("authorization_level")
	if authorizationLevel == "" {
		authorizationLevel = "partial_consent" // 默认授权级别
	}

	// 创建AI服务分层管理器
	manager := NewAIServiceLayeringManager()
	services, err := manager.GetUserAvailableServices(userID, authorizationLevel)
	if err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to get available services", err.Error())
		return
	}

	standardSuccessResponse(c, services, "Available services retrieved successfully")
}

// 获取服务分层信息
func getServiceLayeringInfo(c *gin.Context, core *jobfirst.Core) {
	manager := NewAIServiceLayeringManager()
	info := manager.GetServiceLayeringInfo()
	standardSuccessResponse(c, info, "Service layering info retrieved successfully")
}

// 升级用户服务层级
func upgradeUserServiceLevel(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	var request struct {
		ServiceLevel       string `json:"service_level"`
		AuthorizationLevel string `json:"authorization_level"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	manager := NewAIServiceLayeringManager()
	err := manager.UpgradeUserServiceLevel(userID, request.ServiceLevel, request.AuthorizationLevel)
	if err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to upgrade service level", err.Error())
		return
	}

	standardSuccessResponse(c, nil, "Service level upgraded successfully")
}

// 获取用户使用统计
func getUserUsageStats(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	manager := NewAIServiceLayeringManager()
	stats, err := manager.GetUserUsageStats(userID)
	if err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to get usage stats", err.Error())
		return
	}

	standardSuccessResponse(c, stats, "Usage stats retrieved successfully")
}

// 检查服务限制
func checkServiceLimits(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	serviceID := c.Query("service_id")
	if serviceID == "" {
		standardErrorResponse(c, http.StatusBadRequest, "Service ID is required", "")
		return
	}

	manager := NewAIServiceLayeringManager()
	allowed, err := manager.CheckServiceLimits(userID, serviceID)
	if err != nil {
		standardErrorResponse(c, http.StatusForbidden, "Service limit check failed", err.Error())
		return
	}

	response := map[string]interface{}{
		"allowed": allowed,
		"message": "Service limits checked successfully",
	}

	standardSuccessResponse(c, response, "Service limits checked successfully")
}

// 记录服务使用
func recordServiceUsage(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	var request struct {
		ServiceID          string `json:"service_id"`
		FeatureID          string `json:"feature_id"`
		RequestType        string `json:"request_type"`
		DataSize           int64  `json:"data_size"`
		ProcessingTime     int64  `json:"processing_time"`
		AuthorizationLevel string `json:"authorization_level"`
		Anonymized         bool   `json:"anonymized"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	manager := NewAIServiceLayeringManager()
	err := manager.RecordServiceUsage(
		userID,
		request.ServiceID,
		request.FeatureID,
		request.RequestType,
		request.DataSize,
		request.ProcessingTime,
		request.AuthorizationLevel,
		request.Anonymized,
	)
	if err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to record service usage", err.Error())
		return
	}

	standardSuccessResponse(c, nil, "Service usage recorded successfully")
}

// 个性化分析相关函数

// 更新用户画像
func updateUserProfile(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	var profile UserProfile
	if err := c.ShouldBindJSON(&profile); err != nil {
		standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	manager := NewPersonalizedAnalysisManager()
	err := manager.UpdateUserProfile(userID, profile)
	if err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to update user profile", err.Error())
		return
	}

	standardSuccessResponse(c, profile, "User profile updated successfully")
}

// 获取用户画像
func getUserProfile(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	manager := NewPersonalizedAnalysisManager()
	profile, exists := manager.userProfiles[userID]
	if !exists {
		standardErrorResponse(c, http.StatusNotFound, "User profile not found", "")
		return
	}

	standardSuccessResponse(c, profile, "User profile retrieved successfully")
}

// 执行个性化分析
func performPersonalizedAnalysis(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	var request struct {
		AnalysisType string `json:"analysis_type"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	manager := NewPersonalizedAnalysisManager()
	result, err := manager.PerformPersonalizedAnalysis(userID, request.AnalysisType)
	if err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to perform analysis", err.Error())
		return
	}

	standardSuccessResponse(c, result, "Analysis completed successfully")
}

// 获取个性化推荐
func getPersonalizedRecommendations(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	limitStr := c.Query("limit")
	limit := 0
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	manager := NewPersonalizedAnalysisManager()
	recommendations, err := manager.GetPersonalizedRecommendations(userID, limit)
	if err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to get recommendations", err.Error())
		return
	}

	standardSuccessResponse(c, recommendations, "Recommendations retrieved successfully")
}

// 获取分析历史
func getAnalysisHistory(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	limitStr := c.Query("limit")
	limit := 0
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	manager := NewPersonalizedAnalysisManager()
	history, err := manager.GetAnalysisHistory(userID, limit)
	if err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to get analysis history", err.Error())
		return
	}

	standardSuccessResponse(c, history, "Analysis history retrieved successfully")
}

// 用户数据主权保护相关函数

// 获取用户数据主权状态
func getDataSovereigntyStatus(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	// 模拟数据主权状态
	status := map[string]interface{}{
		"user_id":              userID,
		"data_sovereignty":     true,
		"local_storage":        true,
		"encryption_enabled":   true,
		"user_control":         true,
		"data_portability":     true,
		"consent_management":   true,
		"data_deletion_right":  true,
		"transparency_level":   "high",
		"last_updated":         time.Now(),
	}

	standardSuccessResponse(c, status, "Data sovereignty status retrieved successfully")
}

// 设置数据授权级别
func setDataAuthorizationLevel(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	var request struct {
		AuthorizationLevel string `json:"authorization_level"`
		DataTypes          []string `json:"data_types"`
		Purpose            string `json:"purpose"`
		Duration           int    `json:"duration"` // days
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	// 模拟设置授权级别
	response := map[string]interface{}{
		"user_id":              userID,
		"authorization_level":  request.AuthorizationLevel,
		"data_types":           request.DataTypes,
		"purpose":              request.Purpose,
		"duration":             request.Duration,
		"expires_at":           time.Now().AddDate(0, 0, request.Duration),
		"status":               "active",
		"created_at":           time.Now(),
	}

	standardSuccessResponse(c, response, "Data authorization level set successfully")
}

// 获取数据使用记录
func getDataUsageRecords(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	// 模拟数据使用记录
	records := []map[string]interface{}{
		{
			"usage_id":      "usage_001",
			"user_id":       userID,
			"data_type":     "resume_analysis",
			"service_type":  "ai_analysis",
			"purpose":       "career_optimization",
			"data_size":     1024,
			"anonymized":    true,
			"consent_level": "full_consent",
			"created_at":    time.Now().Add(-24 * time.Hour),
		},
		{
			"usage_id":      "usage_002",
			"user_id":       userID,
			"data_type":     "skill_assessment",
			"service_type":  "ai_matching",
			"purpose":       "job_recommendation",
			"data_size":     512,
			"anonymized":    true,
			"consent_level": "partial_consent",
			"created_at":    time.Now().Add(-12 * time.Hour),
		},
	}

	standardSuccessResponse(c, records, "Data usage records retrieved successfully")
}

// 撤回数据授权
func revokeDataAuthorization(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	var request struct {
		DataTypes []string `json:"data_types"`
		Purpose   string   `json:"purpose"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	// 模拟撤回授权
	response := map[string]interface{}{
		"user_id":        userID,
		"data_types":     request.DataTypes,
		"purpose":        request.Purpose,
		"status":         "revoked",
		"revoked_at":     time.Now(),
		"message":        "Data authorization revoked successfully",
	}

	standardSuccessResponse(c, response, "Data authorization revoked successfully")
}

// 隐私保护相关函数

// 获取隐私设置
func getPrivacySettings(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	// 模拟隐私设置
	settings := map[string]interface{}{
		"user_id":                userID,
		"profile_visibility":     "connections",
		"data_sharing":           "partial",
		"analytics_opt_in":       true,
		"marketing_opt_in":       false,
		"third_party_sharing":    false,
		"data_retention_period":  365, // days
		"encryption_enabled":     true,
		"anonymization_enabled":  true,
		"consent_management":     true,
		"data_portability":       true,
		"right_to_be_forgotten": true,
		"last_updated":           time.Now(),
	}

	standardSuccessResponse(c, settings, "Privacy settings retrieved successfully")
}

// 更新隐私设置
func updatePrivacySettings(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	var request map[string]interface{}
	if err := c.ShouldBindJSON(&request); err != nil {
		standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	// 模拟更新隐私设置
	request["user_id"] = userID
	request["last_updated"] = time.Now()

	standardSuccessResponse(c, request, "Privacy settings updated successfully")
}

// 数据匿名化处理
func anonymizeData(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	var request struct {
		DataTypes []string               `json:"data_types"`
		Data      map[string]interface{} `json:"data"`
		Level     string                  `json:"level"` // basic, intermediate, advanced
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	// 模拟数据匿名化处理
	anonymizedData := map[string]interface{}{
		"user_id":     userID,
		"data_types":  request.DataTypes,
		"level":       request.Level,
		"anonymized":  true,
		"processed_at": time.Now(),
		"message":     "Data anonymized successfully",
	}

	standardSuccessResponse(c, anonymizedData, "Data anonymized successfully")
}

// 获取隐私保护状态
func getPrivacyProtectionStatus(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	// 模拟隐私保护状态
	status := map[string]interface{}{
		"user_id":                userID,
		"encryption_status":       "enabled",
		"anonymization_status":   "enabled",
		"consent_status":         "active",
		"data_sovereignty":       "enabled",
		"privacy_compliance":     "gdpr_compliant",
		"data_retention":          "compliant",
		"user_control":           "enabled",
		"transparency":            "high",
		"last_audit":             time.Now().Add(-7 * 24 * time.Hour),
		"next_audit":             time.Now().Add(7 * 24 * time.Hour),
	}

	standardSuccessResponse(c, status, "Privacy protection status retrieved successfully")
}
