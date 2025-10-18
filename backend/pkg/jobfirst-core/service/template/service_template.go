package template

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jobfirst/jobfirst-core"
	"github.com/jobfirst/jobfirst-core/service/health"
	"github.com/jobfirst/jobfirst-core/service/registry"
)

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name        string
	Version     string
	Port        int
	Address     string
	ConfigPath  string
	HealthCheck bool
}

// ServiceTemplate 服务模板
type ServiceTemplate struct {
	config      *ServiceConfig
	core        *jobfirst.Core
	router      *gin.Engine
	health      *health.ServiceHealth
	consulReg   *registry.ConsulRegistry
	serviceInfo *registry.ServiceInfo
}

// NewServiceTemplate 创建服务模板
func NewServiceTemplate(config *ServiceConfig) (*ServiceTemplate, error) {
	// 初始化JobFirst核心包
	core, err := jobfirst.NewCore(config.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("初始化JobFirst核心包失败: %v", err)
	}

	// 创建健康检查器
	serviceHealth := health.NewServiceHealth(config.Name, config.Version)

	// 添加数据库健康检查
	if core.GetDB() != nil {
		dbChecker := health.NewComponentHealth("database", func() error {
			sqlDB, err := core.GetDB().DB()
			if err != nil {
				return err
			}
			return sqlDB.Ping()
		})
		serviceHealth.AddChecker(dbChecker)
	}

	// 创建Gin引擎
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// 创建Consul注册器
	consulReg, err := registry.NewConsulRegistry("localhost:8500")
	if err != nil {
		log.Printf("警告: 创建Consul注册器失败: %v", err)
	}

	// 创建服务信息
	serviceInfo := &registry.ServiceInfo{
		ID:      fmt.Sprintf("%s-%d", config.Name, config.Port),
		Name:    config.Name,
		Version: config.Version,
		Address: config.Address,
		Port:    config.Port,
		Health:  &registry.HealthStatus{Status: "healthy"},
	}

	return &ServiceTemplate{
		config:      config,
		core:        core,
		router:      router,
		health:      serviceHealth,
		consulReg:   consulReg,
		serviceInfo: serviceInfo,
	}, nil
}

// SetupHealthCheck 设置健康检查端点
func (st *ServiceTemplate) SetupHealthCheck() {
	st.router.GET("/health", health.HealthHandler(st.health))
}

// SetupAuthMiddleware 设置认证中间件
func (st *ServiceTemplate) SetupAuthMiddleware() gin.HandlerFunc {
	return st.core.AuthMiddleware.RequireAuth()
}

// GetRouter 获取路由引擎
func (st *ServiceTemplate) GetRouter() *gin.Engine {
	return st.router
}

// GetCore 获取核心实例
func (st *ServiceTemplate) GetCore() *jobfirst.Core {
	return st.core
}

// RegisterToConsul 注册到Consul
func (st *ServiceTemplate) RegisterToConsul() error {
	if st.consulReg == nil {
		return fmt.Errorf("Consul注册器未初始化")
	}

	return st.consulReg.Register(st.serviceInfo)
}

// DeregisterFromConsul 从Consul注销
func (st *ServiceTemplate) DeregisterFromConsul() error {
	if st.consulReg == nil {
		return nil
	}

	return st.consulReg.Deregister(st.serviceInfo.ID)
}

// StartHealthMonitoring 启动健康监控
func (st *ServiceTemplate) StartHealthMonitoring() {
	if st.consulReg == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			health := st.health.Check()
			st.consulReg.UpdateHealth(st.serviceInfo.ID, &registry.HealthStatus{
				Status:  health.Status,
				Message: health.Message,
			})
		}
	}()
}

// Start 启动服务
func (st *ServiceTemplate) Start() error {
	// 设置健康检查
	st.SetupHealthCheck()

	// 注册到Consul
	if err := st.RegisterToConsul(); err != nil {
		log.Printf("警告: 注册到Consul失败: %v", err)
	}

	// 启动健康监控
	st.StartHealthMonitoring()

	// 启动服务器
	addr := fmt.Sprintf(":%d", st.config.Port)
	log.Printf("Starting %s v%s on %s", st.config.Name, st.config.Version, addr)

	return st.router.Run(addr)
}

// Stop 停止服务
func (st *ServiceTemplate) Stop() error {
	// 从Consul注销
	if err := st.DeregisterFromConsul(); err != nil {
		log.Printf("警告: 从Consul注销失败: %v", err)
	}

	// 关闭核心包
	if st.core != nil {
		st.core.Close()
	}

	return nil
}

// SetupAPIRoutes 设置API路由的通用方法
func (st *ServiceTemplate) SetupAPIRoutes(apiPath string, setupFunc func(*gin.RouterGroup, *jobfirst.Core)) {
	authMiddleware := st.SetupAuthMiddleware()
	api := st.router.Group(apiPath)
	api.Use(authMiddleware)

	setupFunc(api, st.core)
}
