package swagger

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"resume-centre/common/core"
)

// SwaggerManager Swagger管理器
type SwaggerManager struct {
	config   *SwaggerConfig
	document *SwaggerDocument
}

// NewSwaggerManager 创建Swagger管理器
func NewSwaggerManager(config *SwaggerConfig) *SwaggerManager {
	if config == nil {
		config = DefaultSwaggerConfig()
	}

	// 创建默认文档结构
	document := &SwaggerDocument{
		Swagger:             "2.0",
		Info:                config.Info,
		Host:                config.Info.Host,
		BasePath:            config.Info.BasePath,
		Schemes:             config.Info.Schemes,
		Consumes:            []string{"application/json"},
		Produces:            []string{"application/json"},
		SecurityDefinitions: make(map[string]SecurityScheme),
		Security: []map[string][]string{
			{
				"Authorization": {},
				"accessToken":   {},
			},
		},
		Paths:       make(map[string]interface{}),
		Definitions: make(map[string]interface{}),
		Tags:        []TagInfo{},
	}

	// 添加安全定义
	for _, scheme := range config.Security {
		document.SecurityDefinitions[scheme.Name] = scheme
	}

	return &SwaggerManager{
		config:   config,
		document: document,
	}
}

// GenerateSwaggerJSON 生成Swagger JSON文档
func (s *SwaggerManager) GenerateSwaggerJSON() ([]byte, error) {
	return json.MarshalIndent(s.document, "", "  ")
}

// AddPath 添加API路径
func (s *SwaggerManager) AddPath(path string, method string, operation interface{}) {
	if s.document.Paths[path] == nil {
		s.document.Paths[path] = make(map[string]interface{})
	}

	pathItem := s.document.Paths[path].(map[string]interface{})
	pathItem[method] = operation
}

// AddDefinition 添加数据模型定义
func (s *SwaggerManager) AddDefinition(name string, definition interface{}) {
	s.document.Definitions[name] = definition
}

// AddTag 添加标签
func (s *SwaggerManager) AddTag(tag TagInfo) {
	s.document.Tags = append(s.document.Tags, tag)
}

// CreateOperation 创建API操作定义
func (s *SwaggerManager) CreateOperation(tags []string, summary, description string, parameters []interface{}, responses map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"tags":        tags,
		"summary":     summary,
		"description": description,
		"parameters":  parameters,
		"responses":   responses,
		"security": []map[string][]string{
			{
				"Authorization": {},
				"accessToken":   {},
			},
		},
	}
}

// CreateParameter 创建参数定义
func (s *SwaggerManager) CreateParameter(name, in, description, paramType string, required bool) map[string]interface{} {
	param := map[string]interface{}{
		"name":        name,
		"in":          in,
		"description": description,
		"required":    required,
	}

	if paramType != "" {
		param["type"] = paramType
	}

	return param
}

// CreateResponse 创建响应定义
func (s *SwaggerManager) CreateResponse(description string, schema interface{}) map[string]interface{} {
	return map[string]interface{}{
		"description": description,
		"schema":      schema,
	}
}

// CreateSchema 创建数据模型
func (s *SwaggerManager) CreateSchema(schemaType string, properties map[string]interface{}) map[string]interface{} {
	schema := map[string]interface{}{
		"type":       schemaType,
		"properties": properties,
	}

	if schemaType == "object" && properties != nil {
		schema["properties"] = properties
	}

	return schema
}

// CreateProperty 创建属性定义
func (s *SwaggerManager) CreateProperty(propType, description string, example interface{}) map[string]interface{} {
	prop := map[string]interface{}{
		"type":        propType,
		"description": description,
	}

	if example != nil {
		prop["example"] = example
	}

	return prop
}

// SetupSwaggerRoutes 设置Swagger路由
func (s *SwaggerManager) SetupSwaggerRoutes(router *gin.Engine, basePath string) {
	// Swagger JSON文档路由
	router.GET(basePath+"/v2/api-docs", s.SwaggerJSONHandler())

	// Swagger UI路由
	router.GET(basePath+"/swagger/*any", s.SwaggerUIHandler())
}

// SwaggerJSONHandler Swagger JSON处理器
func (s *SwaggerManager) SwaggerJSONHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成Swagger JSON
		swaggerJSON, err := s.GenerateSwaggerJSON()
		if err != nil {
			c.JSON(http.StatusInternalServerError, core.NewErrorResponse(500, "Failed to generate Swagger JSON"))
			return
		}

		c.Header("Content-Type", "application/json")
		c.Data(http.StatusOK, "application/json", swaggerJSON)
	}
}

// SwaggerUIHandler Swagger UI处理器
func (s *SwaggerManager) SwaggerUIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求路径
		path := c.Param("any")

		// 如果是根路径，返回Swagger UI HTML
		if path == "/" || path == "" {
			s.serveSwaggerUI(c)
			return
		}

		// 处理静态资源
		s.serveSwaggerStatic(c, path)
	}
}

// serveSwaggerUI 提供Swagger UI HTML页面
func (s *SwaggerManager) serveSwaggerUI(c *gin.Context) {
	swaggerUIHTML := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui.css" />
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin:0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '%s',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                validatorUrl: null,
                onComplete: function() {
                    console.log('Swagger UI loaded successfully');
                }
            });
        };
    </script>
</body>
</html>`, s.config.Info.Title, "/v2/api-docs")

	c.Header("Content-Type", "text/html")
	c.Data(http.StatusOK, "text/html", []byte(swaggerUIHTML))
}

// serveSwaggerStatic 提供Swagger静态资源
func (s *SwaggerManager) serveSwaggerStatic(c *gin.Context, path string) {
	// 这里可以添加静态资源处理逻辑
	// 目前简单返回404，实际项目中可能需要提供静态文件服务
	c.JSON(http.StatusNotFound, core.NewErrorResponse(404, "Static resource not found"))
}

// GenerateDefaultSwaggerDoc 生成默认Swagger文档
func (s *SwaggerManager) GenerateDefaultSwaggerDoc() {
	// 添加默认标签
	s.AddTag(TagInfo{Name: "认证", Description: "用户认证相关API"})
	s.AddTag(TagInfo{Name: "用户", Description: "用户管理相关API"})
	s.AddTag(TagInfo{Name: "简历", Description: "简历管理相关API"})
	s.AddTag(TagInfo{Name: "职位", Description: "职位管理相关API"})
	s.AddTag(TagInfo{Name: "企业", Description: "企业管理相关API"})
	s.AddTag(TagInfo{Name: "统计", Description: "统计分析相关API"})
	s.AddTag(TagInfo{Name: "积分", Description: "积分管理相关API"})
	s.AddTag(TagInfo{Name: "存储", Description: "文件存储相关API"})

	// 添加默认数据模型
	s.addDefaultDefinitions()

	// 添加默认API路径
	s.addDefaultPaths()
}

// addDefaultDefinitions 添加默认数据模型定义
func (s *SwaggerManager) addDefaultDefinitions() {
	// 基础响应模型
	baseResponse := s.CreateSchema("object", map[string]interface{}{
		"code":    s.CreateProperty("integer", "状态码", 0),
		"message": s.CreateProperty("string", "消息", "success"),
		"data":    s.CreateProperty("object", "数据", nil),
		"time":    s.CreateProperty("string", "时间戳", time.Now().Format(time.RFC3339)),
	})
	s.AddDefinition("BaseResponse", baseResponse)

	// 用户信息模型
	userInfo := s.CreateSchema("object", map[string]interface{}{
		"user_id":    s.CreateProperty("integer", "用户ID", 123),
		"username":   s.CreateProperty("string", "用户名", "john_doe"),
		"email":      s.CreateProperty("string", "邮箱", "john@example.com"),
		"phone":      s.CreateProperty("string", "手机号", "13800138000"),
		"role":       s.CreateProperty("string", "角色", "user"),
		"status":     s.CreateProperty("integer", "状态", 1),
		"created_at": s.CreateProperty("string", "创建时间", time.Now().Format(time.RFC3339)),
		"updated_at": s.CreateProperty("string", "更新时间", time.Now().Format(time.RFC3339)),
	})
	s.AddDefinition("UserInfo", userInfo)

	// 登录请求模型
	loginRequest := s.CreateSchema("object", map[string]interface{}{
		"username": s.CreateProperty("string", "用户名", "john_doe"),
		"password": s.CreateProperty("string", "密码", "password123"),
		"captcha":  s.CreateProperty("string", "验证码", "1234"),
		"remember": s.CreateProperty("boolean", "记住我", true),
	})
	s.AddDefinition("LoginRequest", loginRequest)

	// 登录响应模型
	loginResponse := s.CreateSchema("object", map[string]interface{}{
		"access_token":  s.CreateProperty("string", "访问令牌", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."),
		"refresh_token": s.CreateProperty("string", "刷新令牌", "refresh_token_123"),
		"token_type":    s.CreateProperty("string", "令牌类型", "Bearer"),
		"expires_in":    s.CreateProperty("integer", "过期时间（秒）", 86400),
		"user":          s.CreateProperty("object", "用户信息", userInfo),
	})
	s.AddDefinition("LoginResponse", loginResponse)
}

// addDefaultPaths 添加默认API路径
func (s *SwaggerManager) addDefaultPaths() {
	// 健康检查
	healthOperation := s.CreateOperation(
		[]string{"系统"},
		"健康检查",
		"检查服务健康状态",
		[]interface{}{},
		map[string]interface{}{
			"200": s.CreateResponse("成功", map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"status": s.CreateProperty("string", "状态", "healthy"),
					"time":   s.CreateProperty("string", "时间", time.Now().Format(time.RFC3339)),
				},
			}),
		},
	)
	s.AddPath("/health", "get", healthOperation)

	// 版本信息
	versionOperation := s.CreateOperation(
		[]string{"系统"},
		"版本信息",
		"获取服务版本信息",
		[]interface{}{},
		map[string]interface{}{
			"200": s.CreateResponse("成功", map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code": s.CreateProperty("integer", "状态码", 0),
					"data": s.CreateProperty("object", "数据", map[string]interface{}{
						"version": s.CreateProperty("string", "版本", "1.0.0"),
						"build":   s.CreateProperty("string", "构建时间", "2024-01-01"),
						"service": s.CreateProperty("string", "服务名", "user-service"),
					}),
					"msg": s.CreateProperty("string", "消息", "success"),
				},
			}),
		},
	)
	s.AddPath("/version", "get", versionOperation)

	// 用户登录
	loginOperation := s.CreateOperation(
		[]string{"认证"},
		"用户登录",
		"用户登录接口",
		[]interface{}{
			s.CreateParameter("body", "body", "登录请求参数", "object", true),
		},
		map[string]interface{}{
			"200": s.CreateResponse("登录成功", map[string]interface{}{
				"$ref": "#/definitions/LoginResponse",
			}),
			"400": s.CreateResponse("请求参数错误", map[string]interface{}{
				"$ref": "#/definitions/BaseResponse",
			}),
			"401": s.CreateResponse("认证失败", map[string]interface{}{
				"$ref": "#/definitions/BaseResponse",
			}),
		},
	)
	s.AddPath("/api/v1/user/auth/login", "post", loginOperation)
}

// GetConfig 获取Swagger配置
func (s *SwaggerManager) GetConfig() *SwaggerConfig {
	return s.config
}

// GetDocument 获取Swagger文档
func (s *SwaggerManager) GetDocument() *SwaggerDocument {
	return s.document
}
