package jobfirst

import (
	"fmt"
	"time"

	"github.com/jobfirst/jobfirst-core/auth"
	"github.com/jobfirst/jobfirst-core/config"
	"github.com/jobfirst/jobfirst-core/database"
	"github.com/jobfirst/jobfirst-core/logger"
	"github.com/jobfirst/jobfirst-core/middleware"
	"github.com/jobfirst/jobfirst-core/service/errors"
	"github.com/jobfirst/jobfirst-core/service/health"
	"github.com/jobfirst/jobfirst-core/service/registry"

	// 注释掉superadmin导入，因为它是独立模块
	// "github.com/jobfirst/jobfirst-core/superadmin"
	"github.com/jobfirst/jobfirst-core/team"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Core JobFirst核心包
type Core struct {
	Config         *config.Manager
	Database       *database.Manager
	Logger         *logger.Manager
	AuthManager    *auth.AuthManager
	TeamManager    *team.Manager
	AuthMiddleware *middleware.AuthMiddleware
	ErrorHandler   *errors.ErrorHandler
	ServiceHealth  *health.ServiceHealth
	ConsulRegistry *registry.ConsulRegistry
	// SuperAdmin     *superadmin.Manager  // 注释掉，因为superadmin是独立模块
}

// NewCore 创建JobFirst核心实例
func NewCore(configPath string) (*Core, error) {
	// 1. 初始化配置管理器
	configManager, err := config.NewManager(configPath)
	if err != nil {
		return nil, fmt.Errorf("初始化配置管理器失败: %w", err)
	}

	// 2. 加载应用配置
	appConfig, err := configManager.LoadAppConfig()
	if err != nil {
		return nil, fmt.Errorf("加载应用配置失败: %w", err)
	}

	// 3. 初始化日志管理器
	logConfig := logger.Config{
		Level:  logger.Level(appConfig.Log.Level),
		Format: logger.Format(appConfig.Log.Format),
		Output: appConfig.Log.Output,
		File:   appConfig.Log.File,
	}

	logManager, err := logger.NewManager(logConfig)
	if err != nil {
		return nil, fmt.Errorf("初始化日志管理器失败: %w", err)
	}

	// 设置全局日志
	if err := logger.InitGlobal(logConfig); err != nil {
		return nil, fmt.Errorf("初始化全局日志失败: %w", err)
	}

	// 4. 初始化数据库管理器
	dbConfig := database.Config{
		MySQL: database.MySQLConfig{
			Host:        appConfig.Database.Host,
			Port:        appConfig.Database.Port,
			Username:    appConfig.Database.Username,
			Password:    appConfig.Database.Password,
			Database:    appConfig.Database.Database,
			Charset:     appConfig.Database.Charset,
			MaxIdle:     appConfig.Database.MaxIdle,
			MaxOpen:     appConfig.Database.MaxOpen,
			MaxLifetime: parseDuration(appConfig.Database.MaxLifetime),
			LogLevel:    parseGORMLogLevel(appConfig.Database.LogLevel),
		},
		Redis: database.RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			Database: 0,
			PoolSize: 10,
			MinIdle:  5,
		},
		PostgreSQL: database.PostgreSQLConfig{
			Host:        "", // 设置为空以禁用PostgreSQL
			Port:        5432,
			Username:    "szjason72",
			Password:    "",
			Database:    "jobfirst_vector",
			SSLMode:     "disable",
			MaxIdle:     10,
			MaxOpen:     100,
			MaxLifetime: parseDuration("1h"),
			LogLevel:    parseGORMLogLevel("warn"),
		},
		Neo4j: database.Neo4jConfig{
			URI:      "", // 设置为空以禁用Neo4j
			Username: "neo4j",
			Password: "password",
			Database: "neo4j",
		},
	}

	dbManager, err := database.NewManager(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("初始化数据库管理器失败: %w", err)
	}

	// 5. 执行数据库迁移（迁移失败时继续启动服务）
	if err := dbManager.Migrate(&auth.User{}, &auth.DevTeamUser{}, &auth.DevOperationLog{}); err != nil {
		// 记录迁移错误但不中断服务启动
		fmt.Printf("警告: 数据库迁移失败，但服务将继续启动: %v\n", err)
	}

	// 6. 初始化认证管理器
	authConfig := auth.AuthConfig{
		JWTSecret:        appConfig.Auth.JWTSecret,
		TokenExpiry:      parseDuration(appConfig.Auth.TokenExpiry),
		RefreshExpiry:    parseDuration(appConfig.Auth.RefreshExpiry),
		PasswordMin:      appConfig.Auth.PasswordMin,
		MaxLoginAttempts: appConfig.Auth.MaxLoginAttempts,
		LockoutDuration:  parseDuration(appConfig.Auth.LockoutDuration),
	}

	authManager := auth.NewAuthManager(dbManager.GetDB(), authConfig)

	// 7. 初始化团队管理器
	teamManager := team.NewManager(dbManager.GetDB())

	// 8. 初始化认证中间件
	authMiddleware := middleware.NewAuthMiddleware(authManager)

	// 9. 初始化错误处理器
	errorHandler := errors.NewErrorHandler()

	// 10. 初始化服务健康检查器
	serviceHealth := health.NewServiceHealth("jobfirst-core", "1.0.0")

	// 添加数据库健康检查
	if dbManager.GetDB() != nil {
		dbChecker := health.NewComponentHealth("database", func() error {
			sqlDB, err := dbManager.GetDB().DB()
			if err != nil {
				return err
			}
			return sqlDB.Ping()
		})
		serviceHealth.AddChecker(dbChecker)
	}

	// 11. 初始化Consul注册器
	consulRegistry, err := registry.NewConsulRegistry("localhost:8500")
	if err != nil {
		logManager.Warn("创建Consul注册器失败: %v", err)
	}

	// 12. 注释掉超级管理员管理器初始化，因为superadmin是独立模块
	// superAdminConfig := &superadmin.Config{
	// 	System: system.MonitorConfig{
	// 		// 系统监控配置
	// 	},
	// 	User: user.UserConfig{
	// 		// 用户管理配置
	// 	},
	// 	Database: superadmindatabase.DatabaseConfig{
	// 		// 数据库管理配置
	// 	},
	// 	AI: ai.AIConfig{
	// 		// AI管理配置
	// 	},
	// 	Config: superadminconfig.ConfigManagerConfig{
	// 		// 配置管理配置
	// 	},
	// 	CICD: cicd.CICDConfig{
	// 		// CI/CD管理配置
	// 	},
	// }
	// superAdminManager, err := superadmin.NewManager(superAdminConfig)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create super admin manager: %v", err)
	// }

	core := &Core{
		Config:         configManager,
		Database:       dbManager,
		Logger:         logManager,
		AuthManager:    authManager,
		TeamManager:    teamManager,
		AuthMiddleware: authMiddleware,
		ErrorHandler:   errorHandler,
		ServiceHealth:  serviceHealth,
		ConsulRegistry: consulRegistry,
		// SuperAdmin:     superAdminManager,  // 注释掉
	}

	logManager.Info("JobFirst核心包初始化成功")
	return core, nil
}

// GetDB 获取数据库实例
func (c *Core) GetDB() *gorm.DB {
	return c.Database.GetDB()
}

// Close 关闭核心包
func (c *Core) Close() error {
	if err := c.Database.Close(); err != nil {
		return fmt.Errorf("关闭数据库连接失败: %w", err)
	}
	c.Logger.Info("JobFirst核心包已关闭")
	return nil
}

// Health 健康检查
func (c *Core) Health() map[string]interface{} {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// 检查数据库健康状态
	dbHealth := c.Database.Health()
	health["database"] = dbHealth

	// 检查配置
	health["config"] = map[string]interface{}{
		"loaded": c.Config != nil,
	}

	// 检查日志
	health["logger"] = map[string]interface{}{
		"initialized": c.Logger != nil,
	}

	return health
}

// CreateServiceTemplate 创建服务模板
// func (c *Core) CreateServiceTemplate(config *template.ServiceConfig) (*template.ServiceTemplate, error) {
// 	return template.NewServiceTemplate(config)
// }

// GetErrorHandler 获取错误处理器
func (c *Core) GetErrorHandler() *errors.ErrorHandler {
	return c.ErrorHandler
}

// GetServiceHealth 获取服务健康检查器
func (c *Core) GetServiceHealth() *health.ServiceHealth {
	return c.ServiceHealth
}

// GetConsulRegistry 获取Consul注册器
func (c *Core) GetConsulRegistry() *registry.ConsulRegistry {
	return c.ConsulRegistry
}

// 辅助函数

// parseDuration 解析时间字符串
func parseDuration(s string) time.Duration {
	if s == "" {
		return time.Hour * 24 // 默认24小时
	}

	duration, err := time.ParseDuration(s)
	if err != nil {
		return time.Hour * 24 // 默认24小时
	}

	return duration
}

// parseGORMLogLevel 解析GORM日志级别
func parseGORMLogLevel(level string) gormlogger.LogLevel {
	switch level {
	case "trace":
		return gormlogger.Silent
	case "debug":
		return gormlogger.Info
	case "info":
		return gormlogger.Info
	case "warn":
		return gormlogger.Warn
	case "error":
		return gormlogger.Error
	case "fatal":
		return gormlogger.Error
	case "panic":
		return gormlogger.Error
	default:
		return gormlogger.Info
	}
}
