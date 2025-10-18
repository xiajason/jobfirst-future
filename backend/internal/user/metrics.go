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
	userOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_operations_total",
			Help: "Total number of user operations",
		},
		[]string{"operation", "status"},
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
}

// RecordUserOperation 记录用户操作指标
func RecordUserOperation(operation, status string) {
	userOperationsTotal.WithLabelValues(operation, status).Inc()
}

// UpdateDBMetrics 更新数据库连接池指标
func UpdateDBMetrics(database string, active, idle int) {
	dbConnectionsActive.WithLabelValues(database).Set(float64(active))
	dbConnectionsIdle.WithLabelValues(database).Set(float64(idle))
}
