package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	apigateway "github.com/xiajason/zervi-basic/basic/backend/internal/api-gateway"
)

// ProxyHandler 代理处理器
type ProxyHandler struct {
	serviceRegistry *apigateway.ServiceRegistry
}

// NewProxyHandler 创建代理处理器
func NewProxyHandler(serviceRegistry *apigateway.ServiceRegistry) *ProxyHandler {
	return &ProxyHandler{
		serviceRegistry: serviceRegistry,
	}
}

// ServiceProxy 服务代理
func (ph *ProxyHandler) ServiceProxy(c *gin.Context) {
	// 提取服务名称
	serviceName := ph.extractServiceName(c.Request.URL.Path)
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无法从路径中提取服务名称",
			"path":  c.Request.URL.Path,
		})
		return
	}

	// 获取健康服务URL
	serviceURL, err := ph.getHealthyServiceURL(serviceName)
	if err != nil {
		log.Printf("❌ 获取服务 %s 的URL失败: %v", serviceName, err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   fmt.Sprintf("服务 %s 不可用", serviceName),
			"details": err.Error(),
		})
		return
	}

	// 创建反向代理
	proxy := ph.createReverseProxy(serviceURL, serviceName)

	// 执行代理
	proxy.ServeHTTP(c.Writer, c.Request)
}

// extractServiceName 从路径中提取服务名称
func (ph *ProxyHandler) extractServiceName(path string) string {
	// 路径格式: /api/v1/{service-name}/...
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "v1" {
		return parts[2]
	}
	return ""
}

// getHealthyServiceURL 获取健康服务的URL
func (ph *ProxyHandler) getHealthyServiceURL(serviceName string) (string, error) {
	return ph.serviceRegistry.GetHealthyServiceURL(serviceName)
}

// createReverseProxy 创建反向代理
func (ph *ProxyHandler) createReverseProxy(targetURL, serviceName string) *httputil.ReverseProxy {
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Printf("❌ 解析目标URL失败: %v", err)
		return nil
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// 自定义代理逻辑
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = ph.rewritePath(req.URL.Path, serviceName)

		// 添加请求头
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		req.Header.Set("X-Forwarded-Proto", "http")
		req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		req.Header.Set("X-Service-Name", serviceName)

		log.Printf("🔄 代理请求: %s %s -> %s", req.Method, req.URL.Path, targetURL)
	}

	// 错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("❌ 代理错误: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(fmt.Sprintf(`{"error": "代理错误", "details": "%s"}`, err.Error())))
	}

	return proxy
}

// rewritePath 重写路径
func (ph *ProxyHandler) rewritePath(path, serviceName string) string {
	// 移除 /api/v1/{service-name} 前缀
	prefix := fmt.Sprintf("/api/v1/%s", serviceName)
	if strings.HasPrefix(path, prefix) {
		return strings.TrimPrefix(path, prefix)
	}
	return path
}

// GetServiceInfo 获取服务信息
func (ph *ProxyHandler) GetServiceInfo(c *gin.Context) {
	serviceName := c.Param("serviceName")

	services, err := ph.serviceRegistry.DiscoverService(serviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("发现服务失败: %v", err),
		})
		return
	}

	if len(services) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("服务 %s 不存在", serviceName),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service_name": serviceName,
		"instances":    services,
		"count":        len(services),
	})
}

// ListServices 列出所有服务
func (ph *ProxyHandler) ListServices(c *gin.Context) {
	services, err := ph.serviceRegistry.ListServices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取服务列表失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"services": services,
		"count":    len(services),
	})
}
