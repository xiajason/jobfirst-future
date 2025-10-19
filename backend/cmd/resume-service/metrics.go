package main

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP请求计数器
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	// HTTP请求持续时间
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// 活跃连接数
	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active connections",
		},
	)

	// 业务指标
	resumeOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "resume_operations_total",
			Help: "Total number of resume operations",
		},
		[]string{"operation", "status"},
	)

	// 简历处理指标
	resumeProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "resume_processing_duration_seconds",
			Help:    "Resume processing duration in seconds",
			Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0, 60.0},
		},
		[]string{"operation"},
	)

	// 数据库连接池指标
	dbConnectionsActive = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
		[]string{"database"},
	)

	dbConnectionsIdle = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections",
		},
		[]string{"database"},
	)
)

// PrometheusMetricsMiddleware 创建Prometheus metrics中间件
func PrometheusMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 增加活跃连接数
		activeConnections.Inc()
		defer activeConnections.Dec()

		// 处理请求
		c.Next()

		// 记录指标
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		
		httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			status,
		).Inc()

		httpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)
	}
}

// SetupMetricsRoutes 设置metrics路由
func SetupMetricsRoutes(r *gin.Engine) {
	// Prometheus metrics端点
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	
	// 健康检查端点（用于Prometheus健康检查）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"service":   "resume-service",
			"timestamp": time.Now().Unix(),
		})
	})
}

// RecordResumeOperation 记录简历操作指标
func RecordResumeOperation(operation, status string) {
	resumeOperationsTotal.WithLabelValues(operation, status).Inc()
}

// RecordResumeProcessingDuration 记录简历处理时间
func RecordResumeProcessingDuration(operation string, duration time.Duration) {
	resumeProcessingDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// UpdateDBMetrics 更新数据库连接池指标
func UpdateDBMetrics(database string, active, idle int) {
	dbConnectionsActive.WithLabelValues(database).Set(float64(active))
	dbConnectionsIdle.WithLabelValues(database).Set(float64(idle))
}
