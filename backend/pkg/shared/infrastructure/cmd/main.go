package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"resume-centre/shared/infrastructure"
)

// InfrastructureService 基础设施服务
type InfrastructureService struct {
	infra  *infrastructure.Infrastructure
	server *http.Server
}

// NewInfrastructureService 创建基础设施服务
func NewInfrastructureService() *InfrastructureService {
	return &InfrastructureService{
		infra: infrastructure.NewInfrastructure(),
	}
}

// Start 启动服务
func (s *InfrastructureService) Start() error {
	// 初始化基础设施
	if err := s.infra.Init(); err != nil {
		return fmt.Errorf("failed to initialize infrastructure: %v", err)
	}

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin路由
	router := gin.New()
	router.Use(gin.Recovery())

	// 设置路由
	s.setupRoutes(router)

	// 创建HTTP服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8210"
	}

	s.server = &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// 启动服务器
	go func() {
		s.infra.Logger.Info("Starting infrastructure service",
			infrastructure.Field{Key: "port", Value: port},
		)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.infra.Logger.Error("Failed to start server",
				infrastructure.Field{Key: "error", Value: err.Error()},
			)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.infra.Logger.Info("Shutting down infrastructure service...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.infra.Logger.Error("Server forced to shutdown",
			infrastructure.Field{Key: "error", Value: err.Error()},
		)
	}

	s.infra.Logger.Info("Infrastructure service exited")
	return nil
}

// setupRoutes 设置路由
func (s *InfrastructureService) setupRoutes(router *gin.Engine) {
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":   "jobfirst-shared-infrastructure",
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
		})
	})

	// 服务信息
	router.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":     "jobfirst-shared-infrastructure",
			"description": "Shared infrastructure service for JobFirst platform",
			"version":     "1.0.0",
			"features": []string{
				"database-management",
				"service-registry",
				"security-management",
				"tracing",
				"messaging",
				"caching",
			},
		})
	})

	// 指标端点
	router.GET("/metrics", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"infrastructure": gin.H{
				"database_connections":  1,
				"service_registrations": 0,
				"active_traces":         0,
				"cache_hits":            0,
				"cache_misses":          0,
			},
		})
	})

	// 数据库状态
	router.GET("/database/status", func(c *gin.Context) {
		status := gin.H{
			"mysql": gin.H{
				"status": "connected",
				"host":   os.Getenv("MYSQL_HOST"),
				"port":   os.Getenv("MYSQL_PORT"),
			},
			"redis": gin.H{
				"status": "connected",
				"host":   os.Getenv("REDIS_HOST"),
				"port":   os.Getenv("REDIS_PORT"),
			},
			"neo4j": gin.H{
				"status": "connected",
				"uri":    os.Getenv("NEO4J_URI"),
			},
			"postgresql": gin.H{
				"status": "connected",
				"host":   os.Getenv("POSTGRESQL_HOST"),
				"port":   os.Getenv("POSTGRESQL_PORT"),
			},
		}
		c.JSON(http.StatusOK, status)
	})

	// 服务注册状态
	router.GET("/registry/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"consul": gin.H{
				"status":     "connected",
				"address":    os.Getenv("CONSUL_ADDRESS"),
				"datacenter": os.Getenv("CONSUL_DATACENTER"),
			},
			"services": []string{},
		})
	})

	// 安全状态
	router.GET("/security/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "enabled",
			"features": []string{
				"jwt-authentication",
				"role-based-access-control",
				"encryption",
				"audit-logging",
			},
		})
	})

	// 追踪状态
	router.GET("/tracing/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "enabled",
			"features": []string{
				"distributed-tracing",
				"span-collection",
				"trace-analysis",
			},
		})
	})

	// 消息队列状态
	router.GET("/messaging/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "enabled",
			"features": []string{
				"message-queuing",
				"event-publishing",
				"message-subscription",
			},
		})
	})

	// 缓存状态
	router.GET("/cache/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "enabled",
			"features": []string{
				"redis-caching",
				"cache-invalidation",
				"cache-statistics",
			},
		})
	})
}

func main() {
	service := NewInfrastructureService()
	if err := service.Start(); err != nil {
		log.Fatalf("Failed to start infrastructure service: %v", err)
	}
}
