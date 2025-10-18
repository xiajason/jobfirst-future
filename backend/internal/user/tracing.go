package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

var tracer opentracing.Tracer

// InitTracing 初始化链路追踪
func InitTracing() error {
	cfg := config.Configuration{
		ServiceName: "zervigo-user-service",
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  "host.docker.internal:14268",
		},
	}

	var err error
	tracer, _, err = cfg.NewTracer()
	if err != nil {
		return fmt.Errorf("failed to initialize jaeger tracer: %w", err)
	}

	opentracing.SetGlobalTracer(tracer)
	return nil
}

// TracingMiddleware 创建链路追踪中间件
func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从HTTP头中提取span context
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))

		// 创建新的span
		span := tracer.StartSpan(
			fmt.Sprintf("%s %s", c.Request.Method, c.FullPath()),
			ext.RPCServerOption(spanCtx),
			ext.SpanKindRPCServer,
		)
		defer span.Finish()

		// 设置span标签
		span.SetTag("http.method", c.Request.Method)
		span.SetTag("http.url", c.Request.URL.String())
		span.SetTag("http.user_agent", c.Request.UserAgent())
		span.SetTag("service.name", "zervigo-user-service")

		// 将span context存储到gin context中
		ctx := opentracing.ContextWithSpan(c.Request.Context(), span)
		c.Request = c.Request.WithContext(ctx)

		// 处理请求
		c.Next()

		// 设置响应标签
		span.SetTag("http.status_code", c.Writer.Status())
		if c.Writer.Status() >= 400 {
			span.SetTag("error", true)
		}
	}
}

// StartSpan 开始一个新的span
func StartSpan(ctx context.Context, operationName string) opentracing.Span {
	return opentracing.StartSpan(operationName, opentracing.ChildOf(opentracing.SpanFromContext(ctx).Context()))
}

// InjectSpanContext 将span context注入到HTTP请求头中
func InjectSpanContext(span opentracing.Span, req *http.Request) error {
	return tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
}

// CloseTracing 关闭链路追踪
func CloseTracing() {
	if tracer != nil {
		if closer, ok := tracer.(interface{ Close() }); ok {
			closer.Close()
		}
	}
}
