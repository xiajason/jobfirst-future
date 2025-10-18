package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"resume-centre/common/utils"
)

// CommonHandlers 通用处理器结构
type CommonHandlers struct {
	consulClient *api.Client
}

// NewCommonHandlers 创建新的通用处理器
func NewCommonHandlers(consulClient *api.Client) *CommonHandlers {
	return &CommonHandlers{
		consulClient: consulClient,
	}
}

// HealthHandler 健康检查处理器
func (h *CommonHandlers) HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	}
}

// VersionHandler 版本信息处理器
func (h *CommonHandlers) VersionHandler(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"version": "1.0.0",
				"build":   time.Now().Format("2006-01-02"),
				"service": serviceName,
			},
			"msg": "success",
		})
	}
}

// MD5Handler MD5加密处理器
func (h *CommonHandlers) MD5Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Text string `json:"text" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "参数错误",
			})
			return
		}

		md5Hash := utils.MD5Hash(req.Text)

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"original": req.Text,
				"md5":      md5Hash,
			},
			"msg": "MD5加密成功",
		})
	}
}

// RandomHandler 随机字符串生成处理器
func (h *CommonHandlers) RandomHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Length int `json:"length" binding:"required,min=1,max=100"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "参数错误",
			})
			return
		}

		randomStr, err := utils.GenerateRandomString(req.Length)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "生成随机字符串失败",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"length": req.Length,
				"random": randomStr,
			},
			"msg": "随机字符串生成成功",
		})
	}
}

// JSONFormatHandler JSON格式化处理器
func (h *CommonHandlers) JSONFormatHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			JSON string `json:"json" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "参数错误",
			})
			return
		}

		formatted, err := utils.FormatJSON(req.JSON)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "JSON格式错误",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"original":  req.JSON,
				"formatted": formatted,
			},
			"msg": "JSON格式化成功",
		})
	}
}

// StatusHandler 系统状态处理器
func (h *CommonHandlers) StatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查Consul连接
		consulStatus := "healthy"
		if h.consulClient != nil {
			_, err := h.consulClient.Agent().Self()
			if err != nil {
				consulStatus = "unhealthy"
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"timestamp": time.Now().Format(time.RFC3339),
				"services": gin.H{
					"discovery": gin.H{
						"status": consulStatus,
						"type":   "consul",
					},
				},
			},
			"msg": "系统状态检查完成",
		})
	}
}

// ServicesHandler 服务列表处理器
func (h *CommonHandlers) ServicesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.consulClient == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "Consul客户端未初始化",
			})
			return
		}

		services, err := h.consulClient.Agent().Services()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "获取服务列表失败",
			})
			return
		}

		serviceList := make([]gin.H, 0)
		for _, service := range services {
			serviceList = append(serviceList, gin.H{
				"id":      service.ID,
				"name":    service.Service,
				"address": service.Address,
				"port":    service.Port,
				"tags":    service.Tags,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"services": serviceList,
				"total":    len(serviceList),
			},
			"msg": "获取服务列表成功",
		})
	}
}

// GetConfigHandler 获取配置处理器
func (h *CommonHandlers) GetConfigHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Param("key")
		if key == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "配置键不能为空",
			})
			return
		}

		if h.consulClient == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "Consul客户端未初始化",
			})
			return
		}

		pair, _, err := h.consulClient.KV().Get(key, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "获取配置失败",
			})
			return
		}

		if pair == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"msg":  "配置不存在",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"key":   key,
				"value": string(pair.Value),
			},
			"msg": "获取配置成功",
		})
	}
}

// SetConfigHandler 设置配置处理器
func (h *CommonHandlers) SetConfigHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Param("key")
		if key == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "配置键不能为空",
			})
			return
		}

		var req struct {
			Value string `json:"value" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "参数错误",
			})
			return
		}

		if h.consulClient == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "Consul客户端未初始化",
			})
			return
		}

		pair := &api.KVPair{
			Key:   key,
			Value: []byte(req.Value),
		}

		_, err := h.consulClient.KV().Put(pair, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "设置配置失败",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"key":   key,
				"value": req.Value,
			},
			"msg": "设置配置成功",
		})
	}
}

// DeleteConfigHandler 删除配置处理器
func (h *CommonHandlers) DeleteConfigHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Param("key")
		if key == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "配置键不能为空",
			})
			return
		}

		if h.consulClient == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "Consul客户端未初始化",
			})
			return
		}

		_, err := h.consulClient.KV().Delete(key, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "删除配置失败",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"key": key,
			},
			"msg": "删除配置成功",
		})
	}
}
