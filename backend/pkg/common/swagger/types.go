package swagger

// SwaggerInfo Swagger信息配置
type SwaggerInfo struct {
	Title       string   `json:"title"`       // API标题
	Description string   `json:"description"` // API描述
	Version     string   `json:"version"`     // API版本
	Host        string   `json:"host"`        // 主机地址
	BasePath    string   `json:"basePath"`    // 基础路径
	Schemes     []string `json:"schemes"`     // 协议方案
}

// ContactInfo 联系信息
type ContactInfo struct {
	Name  string `json:"name"`  // 联系人姓名
	URL   string `json:"url"`   // 联系URL
	Email string `json:"email"` // 联系邮箱
}

// LicenseInfo 许可证信息
type LicenseInfo struct {
	Name string `json:"name"` // 许可证名称
	URL  string `json:"url"`  // 许可证URL
}

// SwaggerConfig Swagger配置
type SwaggerConfig struct {
	Info     SwaggerInfo      `json:"info"`     // API信息
	Contact  ContactInfo      `json:"contact"`  // 联系信息
	License  LicenseInfo      `json:"license"`  // 许可证信息
	Security []SecurityScheme `json:"security"` // 安全认证方案
}

// SecurityScheme 安全认证方案
type SecurityScheme struct {
	Type         string            `json:"type"`         // 认证类型
	Name         string            `json:"name"`         // 认证名称
	In           string            `json:"in"`           // 认证位置
	Description  string            `json:"description"`  // 认证描述
	Scheme       string            `json:"scheme"`       // 认证方案
	BearerFormat string            `json:"bearerFormat"` // Bearer格式
	Scopes       map[string]string `json:"scopes"`       // 权限范围
}

// SwaggerDocument Swagger文档结构
type SwaggerDocument struct {
	Swagger             string                    `json:"swagger"`
	Info                SwaggerInfo               `json:"info"`
	Host                string                    `json:"host"`
	BasePath            string                    `json:"basePath"`
	Schemes             []string                  `json:"schemes"`
	Consumes            []string                  `json:"consumes"`
	Produces            []string                  `json:"produces"`
	SecurityDefinitions map[string]SecurityScheme `json:"securityDefinitions"`
	Security            []map[string][]string     `json:"security"`
	Paths               map[string]interface{}    `json:"paths"`
	Definitions         map[string]interface{}    `json:"definitions"`
	Tags                []TagInfo                 `json:"tags"`
}

// TagInfo 标签信息
type TagInfo struct {
	Name        string `json:"name"`        // 标签名称
	Description string `json:"description"` // 标签描述
}

// DefaultSwaggerConfig 默认Swagger配置
func DefaultSwaggerConfig() *SwaggerConfig {
	return &SwaggerConfig{
		Info: SwaggerInfo{
			Title:       "JobFirst API",
			Description: "JobFirst微服务API文档",
			Version:     "1.0.0",
			Host:        "localhost:8000",
			BasePath:    "/",
			Schemes:     []string{"http", "https"},
		},
		Contact: ContactInfo{
			Name:  "JobFirst Team",
			URL:   "https://jobfirst.com",
			Email: "support@jobfirst.com",
		},
		License: LicenseInfo{
			Name: "MIT",
			URL:  "https://opensource.org/licenses/MIT",
		},
		Security: []SecurityScheme{
			{
				Type:        "apiKey",
				Name:        "Authorization",
				In:          "header",
				Description: "Bearer token for API authentication",
			},
			{
				Type:        "apiKey",
				Name:        "accessToken",
				In:          "header",
				Description: "Access token for API authentication",
			},
		},
	}
}

// ServiceSwaggerConfig 服务特定Swagger配置
func ServiceSwaggerConfig(serviceName, serviceDescription string) *SwaggerConfig {
	config := DefaultSwaggerConfig()
	config.Info.Title = serviceName + " API"
	config.Info.Description = serviceDescription
	return config
}
