package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// TracingService 分布式追踪服务接口
type TracingService interface {
	StartSpan(name string) Span
	StartSpanWithContext(ctx context.Context, name string) (Span, context.Context)
	InjectContext(ctx context.Context, carrier interface{}) error
	ExtractContext(carrier interface{}) (context.Context, error)
	Shutdown(ctx context.Context) error
}

// Span 追踪跨度接口
type Span interface {
	End()
	SetAttributes(attrs ...Field)
	AddEvent(name string, attrs ...Field)
	SetStatus(code int, description string)
	GetTraceID() string
}

// TracingConfig 追踪配置
type TracingConfig struct {
	ServiceName    string `yaml:"service_name" json:"service_name"`
	ServiceVersion string `yaml:"service_version" json:"service_version"`
	Environment    string `yaml:"environment" json:"environment"`

	JaegerEndpoint string `yaml:"jaeger_endpoint" json:"jaeger_endpoint"`
	JaegerUsername string `yaml:"jaeger_username" json:"jaeger_username"`
	JaegerPassword string `yaml:"jaeger_password" json:"jaeger_password"`

	SampleRate float64 `yaml:"sample_rate" json:"sample_rate"`
	Enabled    bool    `yaml:"enabled" json:"enabled"`
}

// SimpleTracing 简单追踪实现
type SimpleTracing struct {
	config *TracingConfig
	spans  map[string]*SimpleSpan
}

// NewSimpleTracing 创建简单追踪服务
func NewSimpleTracing(config *TracingConfig) *SimpleTracing {
	return &SimpleTracing{
		config: config,
		spans:  make(map[string]*SimpleSpan),
	}
}

// StartSpan 开始追踪跨度
func (t *SimpleTracing) StartSpan(name string) Span {
	span := &SimpleSpan{
		name:       name,
		startTime:  time.Now(),
		attributes: make(map[string]interface{}),
		events:     make([]Event, 0),
		traceID:    generateTraceID(),
		spanID:     generateSpanID(),
	}

	t.spans[span.spanID] = span

	Info("Span started",
		Field{Key: "span_name", Value: name},
		Field{Key: "trace_id", Value: span.traceID},
		Field{Key: "span_id", Value: span.spanID},
	)

	return span
}

// StartSpanWithContext 使用上下文开始追踪跨度
func (t *SimpleTracing) StartSpanWithContext(ctx context.Context, name string) (Span, context.Context) {
	span := t.StartSpan(name)

	// 将追踪信息注入上下文
	newCtx := context.WithValue(ctx, "trace_id", span.GetTraceID())
	newCtx = context.WithValue(newCtx, "span_id", span.(*SimpleSpan).spanID)

	return span, newCtx
}

// InjectContext 注入追踪上下文
func (t *SimpleTracing) InjectContext(ctx context.Context, carrier interface{}) error {
	// 简单实现，将追踪信息记录到日志
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		Info("Tracing context injected",
			Field{Key: "trace_id", Value: traceID},
		)
	}
	return nil
}

// ExtractContext 提取追踪上下文
func (t *SimpleTracing) ExtractContext(carrier interface{}) (context.Context, error) {
	// 简单实现，返回新的上下文
	return context.Background(), nil
}

// Shutdown 关闭追踪服务
func (t *SimpleTracing) Shutdown(ctx context.Context) error {
	Info("Tracing service shutdown")
	return nil
}

// SimpleSpan 简单跨度实现
type SimpleSpan struct {
	name       string
	startTime  time.Time
	endTime    time.Time
	attributes map[string]interface{}
	events     []Event
	traceID    string
	spanID     string
	status     int
	statusDesc string
}

// Event 事件结构
type Event struct {
	Name       string                 `json:"name"`
	Timestamp  time.Time              `json:"timestamp"`
	Attributes map[string]interface{} `json:"attributes"`
}

// End 结束跨度
func (s *SimpleSpan) End() {
	s.endTime = time.Now()
	duration := s.endTime.Sub(s.startTime)

	Info("Span ended",
		Field{Key: "span_name", Value: s.name},
		Field{Key: "trace_id", Value: s.traceID},
		Field{Key: "span_id", Value: s.spanID},
		Field{Key: "duration", Value: duration.String()},
		Field{Key: "status", Value: s.status},
		Field{Key: "status_desc", Value: s.statusDesc},
	)
}

// SetAttributes 设置属性
func (s *SimpleSpan) SetAttributes(attrs ...Field) {
	for _, attr := range attrs {
		s.attributes[attr.Key] = attr.Value
	}
}

// AddEvent 添加事件
func (s *SimpleSpan) AddEvent(name string, attrs ...Field) {
	event := Event{
		Name:       name,
		Timestamp:  time.Now(),
		Attributes: make(map[string]interface{}),
	}

	for _, attr := range attrs {
		event.Attributes[attr.Key] = attr.Value
	}

	s.events = append(s.events, event)

	Info("Span event added",
		Field{Key: "span_name", Value: s.name},
		Field{Key: "event_name", Value: name},
		Field{Key: "trace_id", Value: s.traceID},
	)
}

// SetStatus 设置状态
func (s *SimpleSpan) SetStatus(code int, description string) {
	s.status = code
	s.statusDesc = description
}

// GetTraceID 获取追踪ID
func (s *SimpleSpan) GetTraceID() string {
	return s.traceID
}

// NoopTracing 空操作追踪实现（用于禁用追踪）
type NoopTracing struct{}

// StartSpan 开始空操作跨度
func (t *NoopTracing) StartSpan(name string) Span {
	return &NoopSpan{}
}

// StartSpanWithContext 使用上下文开始空操作跨度
func (t *NoopTracing) StartSpanWithContext(ctx context.Context, name string) (Span, context.Context) {
	return &NoopSpan{}, ctx
}

// InjectContext 空操作上下文注入
func (t *NoopTracing) InjectContext(ctx context.Context, carrier interface{}) error {
	return nil
}

// ExtractContext 空操作上下文提取
func (t *NoopTracing) ExtractContext(carrier interface{}) (context.Context, error) {
	return context.Background(), nil
}

// Shutdown 空操作关闭
func (t *NoopTracing) Shutdown(ctx context.Context) error {
	return nil
}

// NoopSpan 空操作跨度
type NoopSpan struct{}

// End 空操作结束
func (s *NoopSpan) End() {}

// SetAttributes 空操作设置属性
func (s *NoopSpan) SetAttributes(attrs ...Field) {}

// AddEvent 空操作添加事件
func (s *NoopSpan) AddEvent(name string, attrs ...Field) {}

// SetStatus 空操作设置状态
func (s *NoopSpan) SetStatus(code int, description string) {}

// GetTraceID 空操作获取追踪ID
func (s *NoopSpan) GetTraceID() string {
	return "noop-trace-id"
}

// CreateDefaultTracingConfig 创建默认追踪配置
func CreateDefaultTracingConfig() *TracingConfig {
	return &TracingConfig{
		ServiceName:    "jobfirst",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		JaegerEndpoint: "http://localhost:14268/api/traces",
		SampleRate:     1.0,
		Enabled:        true,
	}
}

// 全局追踪服务实例
var globalTracingService TracingService

// InitGlobalTracingService 初始化全局追踪服务
func InitGlobalTracingService(tracing TracingService) {
	globalTracingService = tracing
}

// GetTracingService 获取全局追踪服务
func GetTracingService() TracingService {
	return globalTracingService
}

// TracingMiddleware Gin追踪中间件
func TracingMiddleware(serviceName string) func(c *gin.Context) {
	return func(c *gin.Context) {
		if globalTracingService == nil {
			c.Next()
			return
		}

		// 开始追踪跨度
		span, ctx := globalTracingService.StartSpanWithContext(c.Request.Context(),
			fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path))
		defer span.End()

		// 设置请求属性
		span.SetAttributes(
			Field{Key: "http.method", Value: c.Request.Method},
			Field{Key: "http.url", Value: c.Request.URL.String()},
			Field{Key: "http.user_agent", Value: c.Request.UserAgent()},
			Field{Key: "http.remote_addr", Value: c.ClientIP()},
		)

		// 将追踪上下文注入到请求中
		c.Request = c.Request.WithContext(ctx)

		// 处理请求
		c.Next()

		// 设置响应属性
		span.SetAttributes(
			Field{Key: "http.status_code", Value: c.Writer.Status()},
			Field{Key: "http.response_size", Value: c.Writer.Size()},
		)

		// 设置状态
		if c.Writer.Status() >= 400 {
			span.SetStatus(1, fmt.Sprintf("HTTP %d", c.Writer.Status()))
		} else {
			span.SetStatus(0, "")
		}
	}
}

// 辅助函数
func generateTraceID() string {
	return fmt.Sprintf("trace_%d", time.Now().UnixNano())
}

func generateSpanID() string {
	return fmt.Sprintf("span_%d", time.Now().UnixNano())
}
