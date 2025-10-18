// Package common 提供JobFirst项目的通用功能模块
//
// 该包包含以下功能：
// - 核心功能 (core) - 控制器基类、响应模型、常量定义
// - 安全框架 (security) - 认证和授权功能
// - JWT令牌处理 (jwt) - 令牌创建、解析、验证功能
// - Swagger文档配置 (swagger) - API文档配置和生成
// - 缓存处理 (cache) - 缓存管理功能
// - 日志处理 (log) - 日志管理功能
// - 线程池管理 (thread) - 线程池管理功能
// - 存储服务 (storage) - 文件存储服务
// - ElasticSearch集成 (es) - ES搜索功能
// - 消息队列 (mq) - 消息队列功能
// - 工具函数 (utils)
// - 配置管理 (config)
// - 中间件 (middleware)
// - 通用处理器 (handlers)
package common

import (
	"resume-centre/common/config"
	"resume-centre/common/handlers"
	"resume-centre/common/jwt"
	"resume-centre/common/middleware"
	"resume-centre/common/security"
	"resume-centre/common/swagger"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
)

// Common 通用模块结构
type Common struct {
	Config   *config.Config
	Logger   *logrus.Logger
	Consul   *api.Client
	Handlers *handlers.CommonHandlers
	Security *security.SecurityFilter
	JWT      *jwt.JWTManager
	Swagger  *swagger.SwaggerManager
}

// New 创建新的Common实例
func New() *Common {
	return &Common{
		Config: config.NewConfig(),
		Logger: logrus.New(),
	}
}

// Init 初始化Common模块
func (c *Common) Init() error {
	// 初始化配置
	if err := c.Config.Load(); err != nil {
		return err
	}

	// 初始化日志
	c.Logger.SetFormatter(&logrus.JSONFormatter{})
	c.Logger.SetLevel(c.Config.GetLogLevel())

	// 初始化Consul客户端
	if err := c.initConsul(); err != nil {
		return err
	}

	// 初始化处理器
	c.Handlers = handlers.NewCommonHandlers(c.Consul)

	// 初始化安全过滤器
	c.Security = security.NewSecurityFilter(security.DefaultWhitelist())

	// 初始化JWT管理器
	c.JWT = jwt.NewJWTManager(nil) // 使用默认配置

	// 初始化Swagger管理器
	c.Swagger = swagger.NewSwaggerManager(nil) // 使用默认配置
	c.Swagger.GenerateDefaultSwaggerDoc()      // 生成默认文档

	return nil
}

// initConsul 初始化Consul客户端
func (c *Common) initConsul() error {
	config := api.DefaultConfig()
	config.Address = c.Config.GetConsulAddress()
	config.Datacenter = c.Config.GetConsulDatacenter()

	var err error
	c.Consul, err = api.NewClient(config)
	if err != nil {
		return err
	}

	// 测试连接
	_, err = c.Consul.Agent().Self()
	return err
}

// SetupRoutes 设置通用路由
func (c *Common) SetupRoutes(router *gin.Engine, serviceName string) {
	// 健康检查路由
	router.GET("/health", c.Handlers.HealthHandler())

	// 版本信息路由
	router.GET("/version", c.Handlers.VersionHandler(serviceName))

	// 工具函数路由组
	utils := router.Group("/utils")
	{
		utils.POST("/md5", c.Handlers.MD5Handler())
		utils.POST("/random", c.Handlers.RandomHandler())
		utils.POST("/format/json", c.Handlers.JSONFormatHandler())
	}

	// 监控路由组
	monitor := router.Group("/monitor")
	{
		monitor.GET("/status", c.Handlers.StatusHandler())
		monitor.GET("/services", c.Handlers.ServicesHandler())
	}

	// 配置管理路由组
	config := router.Group("/config")
	{
		config.GET("/:key", c.Handlers.GetConfigHandler())
		config.PUT("/:key", c.Handlers.SetConfigHandler())
		config.DELETE("/:key", c.Handlers.DeleteConfigHandler())
	}

	// Swagger文档路由
	c.Swagger.SetupSwaggerRoutes(router, "")
}

// SetupMiddleware 设置通用中间件
func (c *Common) SetupMiddleware(router *gin.Engine, whitelist []string) {
	// 添加通用中间件
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.MetricsMiddleware())

	// 添加安全过滤器
	if len(whitelist) > 0 {
		c.Security = security.NewSecurityFilter(whitelist)
	}
	router.Use(c.Security.Filter())
}

// GetConfig 获取配置实例
func (c *Common) GetConfig() *config.Config {
	return c.Config
}

// GetLogger 获取日志实例
func (c *Common) GetLogger() *logrus.Logger {
	return c.Logger
}

// GetConsul 获取Consul客户端
func (c *Common) GetConsul() *api.Client {
	return c.Consul
}

// GetHandlers 获取处理器实例
func (c *Common) GetHandlers() *handlers.CommonHandlers {
	return c.Handlers
}

// GetSecurity 获取安全过滤器实例
func (c *Common) GetSecurity() *security.SecurityFilter {
	return c.Security
}

// GetJWT 获取JWT管理器实例
func (c *Common) GetJWT() *jwt.JWTManager {
	return c.JWT
}

// GetSwagger 获取Swagger管理器实例
func (c *Common) GetSwagger() *swagger.SwaggerManager {
	return c.Swagger
}
