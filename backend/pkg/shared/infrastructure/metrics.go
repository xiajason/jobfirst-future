package infrastructure

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP请求计数器
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTP请求持续时间
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// 活跃连接数
	httpRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)

	// 服务健康状态
	serviceHealth = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "service_health",
			Help: "Service health status (1 = healthy, 0 = unhealthy)",
		},
	)
)

func init() {
	// 注册指标
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(httpRequestsInFlight)
	prometheus.MustRegister(serviceHealth)
}

// MetricsMiddleware 提供Prometheus指标收集中间件
func MetricsMiddleware() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

// RequestMetricsMiddleware 提供HTTP请求指标收集中间件
func RequestMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 增加活跃请求计数
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		// 处理请求
		c.Next()

		// 记录请求指标
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		httpRequestsTotal.WithLabelValues(c.Request.Method, endpoint, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, endpoint).Observe(duration)
	}
}

// SetServiceHealth 设置服务健康状态
func SetServiceHealth(healthy bool) {
	if healthy {
		serviceHealth.Set(1)
	} else {
		serviceHealth.Set(0)
	}
}

// isAllowedIP 检查IP是否被允许访问metrics
func isAllowedIP(ipStr string) bool {
	// 允许的IP范围
	allowedRanges := []string{
		"127.0.0.1",      // localhost
		"::1",            // localhost IPv6
		"172.16.0.0/12",  // Docker内部网络
		"192.168.0.0/16", // 私有网络
		"10.0.0.0/8",     // 私有网络
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	for _, allowed := range allowedRanges {
		if strings.Contains(allowed, "/") {
			// CIDR范围
			_, network, err := net.ParseCIDR(allowed)
			if err == nil && network.Contains(ip) {
				return true
			}
		} else {
			// 单个IP
			if ip.String() == allowed {
				return true
			}
		}
	}

	return false
}

// SecureMetricsHandler 返回安全的metrics处理器
func SecureMetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查X-Forwarded-For头，防止外部访问
		forwardedFor := c.GetHeader("X-Forwarded-For")
		if forwardedFor != "" {
			// 检查是否包含外部IP
			ips := strings.Split(forwardedFor, ",")
			for _, ip := range ips {
				ip = strings.TrimSpace(ip)
				if !isAllowedIP(ip) {
					c.JSON(403, gin.H{
						"error":   "Access denied",
						"message": "Metrics endpoint is only accessible from internal networks",
					})
					return
				}
			}
		}

		// 允许访问，返回metrics
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	}
}

// GetMetricsHandler 返回metrics处理器（保持向后兼容）
func GetMetricsHandler() gin.HandlerFunc {
	return SecureMetricsHandler()
}
