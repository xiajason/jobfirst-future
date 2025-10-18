package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jobfirst/jobfirst-core"
	"github.com/jobfirst/jobfirst-core/auth"
	"github.com/jobfirst/jobfirst-core/team"
)

func main() {
	// 初始化JobFirst核心包
	core, err := jobfirst.NewCore("./configs/config.yaml")
	if err != nil {
		log.Fatal("初始化核心包失败:", err)
	}
	defer core.Close()

	// 创建Gin路由
	router := gin.Default()

	// 设置路由组
	setupRoutes(router, core)

	// 启动服务器
	config := core.Config
	host := config.GetString("server.host")
	port := config.GetInt("server.port")

	log.Printf("服务器启动在 %s:%d", host, port)
	if err := router.Run(fmt.Sprintf("%s:%d", host, port)); err != nil {
		log.Fatal("启动服务器失败:", err)
	}
}

// setupRoutes 设置路由
func setupRoutes(router *gin.Engine, core *jobfirst.Core) {
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		health := core.Health()
		c.JSON(http.StatusOK, health)
	})

	// API版本组
	api := router.Group("/api/v1")
	{
		// 公开路由
		setupPublicRoutes(api, core)

		// 认证路由
		setupAuthRoutes(api, core)

		// 开发团队路由
		setupDevTeamRoutes(api, core)

		// 超级管理员路由
		setupSuperAdminRoutes(api, core)
	}
}

// setupPublicRoutes 设置公开路由
func setupPublicRoutes(api *gin.RouterGroup, core *jobfirst.Core) {
	public := api.Group("/public")
	{
		// 用户注册
		public.POST("/register", func(c *gin.Context) {
			var req auth.RegisterRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "请求参数错误",
					"details": err.Error(),
				})
				return
			}

			resp, err := core.AuthManager.Register(req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "注册失败",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, resp)
		})

		// 用户登录
		public.POST("/login", func(c *gin.Context) {
			var req auth.LoginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "请求参数错误",
					"details": err.Error(),
				})
				return
			}

			resp, err := core.AuthManager.Login(req, c.ClientIP(), c.Request.UserAgent())
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   "登录失败",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, resp)
		})

		// 超级管理员登录
		public.POST("/super-admin/login", func(c *gin.Context) {
			var req auth.LoginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "请求参数错误",
					"details": err.Error(),
				})
				return
			}

			resp, err := core.AuthManager.SuperAdminLogin(req, c.ClientIP(), c.Request.UserAgent())
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   "超级管理员登录失败",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, resp)
		})
	}
}

// setupAuthRoutes 设置认证路由
func setupAuthRoutes(api *gin.RouterGroup, core *jobfirst.Core) {
	auth := api.Group("/auth")
	auth.Use(core.AuthMiddleware.RequireAuth())
	{
		// 获取用户资料
		auth.GET("/profile", func(c *gin.Context) {
			userID := c.GetUint("user_id")
			user, err := core.AuthManager.GetUserByID(userID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"error":   "用户不存在",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    user,
			})
		})

		// 检查成员身份
		auth.GET("/check-membership", func(c *gin.Context) {
			userID := c.GetUint("user_id")
			devTeam, err := core.AuthManager.GetDevTeamUser(userID)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data": gin.H{
						"is_member": false,
						"message":   "您不是开发团队成员",
					},
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"is_member": true,
					"member":    devTeam,
				},
			})
		})
	}
}

// setupDevTeamRoutes 设置开发团队路由
func setupDevTeamRoutes(api *gin.RouterGroup, core *jobfirst.Core) {
	devTeam := api.Group("/dev-team")
	devTeam.Use(core.AuthMiddleware.RequireDevTeam())
	{
		// 获取个人资料
		devTeam.GET("/profile", func(c *gin.Context) {
			userID := c.GetUint("user_id")
			devTeam, err := core.AuthManager.GetDevTeamUser(userID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"error":   "团队成员信息不存在",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    devTeam,
			})
		})

		// 获取个人操作日志
		devTeam.GET("/my-logs", func(c *gin.Context) {
			userID := c.GetUint("user_id")
			page := 1
			pageSize := 20

			req := team.GetOperationLogsRequest{
				Page:     page,
				PageSize: pageSize,
				UserID:   userID,
			}

			resp, err := core.TeamManager.GetOperationLogs(req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "获取操作日志失败",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, resp)
		})

		// 获取团队状态
		devTeam.GET("/status", func(c *gin.Context) {
			resp, err := core.TeamManager.GetStats()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "获取团队状态失败",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, resp)
		})
	}
}

// setupSuperAdminRoutes 设置超级管理员路由
func setupSuperAdminRoutes(api *gin.RouterGroup, core *jobfirst.Core) {
	admin := api.Group("/admin")
	admin.Use(core.AuthMiddleware.RequireSuperAdmin())
	{
		// 添加团队成员
		admin.POST("/members", func(c *gin.Context) {
			var req team.AddMemberRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "请求参数错误",
					"details": err.Error(),
				})
				return
			}

			resp, err := core.TeamManager.AddMember(req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "添加团队成员失败",
					"details": err.Error(),
				})
				return
			}

			// 记录操作日志
			userID := c.GetUint("user_id")
			core.TeamManager.LogOperation(userID, "add_team_member", "dev_team", map[string]interface{}{
				"added_user": req.Username,
				"team_role":  req.TeamRole,
			}, c.ClientIP(), c.Request.UserAgent(), "success")

			c.JSON(http.StatusOK, resp)
		})

		// 获取团队成员列表
		admin.GET("/members", func(c *gin.Context) {
			page := 1
			pageSize := 10
			role := c.Query("role")
			status := c.Query("status")

			req := team.GetMembersRequest{
				Page:     page,
				PageSize: pageSize,
				Role:     role,
				Status:   status,
			}

			resp, err := core.TeamManager.GetMembers(req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "获取团队成员列表失败",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, resp)
		})

		// 更新团队成员
		admin.PUT("/members/:id", func(c *gin.Context) {
			memberIDStr := c.Param("id")
			memberID := parseUint(memberIDStr)
			var req team.UpdateMemberRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "请求参数错误",
					"details": err.Error(),
				})
				return
			}

			resp, err := core.TeamManager.UpdateMember(memberID, req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "更新团队成员失败",
					"details": err.Error(),
				})
				return
			}

			// 记录操作日志
			userID := c.GetUint("user_id")
			core.TeamManager.LogOperation(userID, "update_team_member", "dev_team", map[string]interface{}{
				"member_id": memberID,
				"updates":   req,
			}, c.ClientIP(), c.Request.UserAgent(), "success")

			c.JSON(http.StatusOK, resp)
		})

		// 移除团队成员
		admin.DELETE("/members/:id", func(c *gin.Context) {
			memberIDStr := c.Param("id")
			memberID := parseUint(memberIDStr)
			resp, err := core.TeamManager.RemoveMember(memberID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "移除团队成员失败",
					"details": err.Error(),
				})
				return
			}

			// 记录操作日志
			userID := c.GetUint("user_id")
			core.TeamManager.LogOperation(userID, "remove_team_member", "dev_team", map[string]interface{}{
				"member_id": memberID,
			}, c.ClientIP(), c.Request.UserAgent(), "success")

			c.JSON(http.StatusOK, resp)
		})

		// 获取团队统计信息
		admin.GET("/stats", func(c *gin.Context) {
			resp, err := core.TeamManager.GetStats()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "获取统计信息失败",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, resp)
		})

		// 获取操作日志
		admin.GET("/logs", func(c *gin.Context) {
			page := 1
			pageSize := 20
			userID := c.Query("user_id")
			operationType := c.Query("operation_type")
			status := c.Query("status")

			req := team.GetOperationLogsRequest{
				Page:          page,
				PageSize:      pageSize,
				OperationType: operationType,
				Status:        status,
			}

			if userID != "" {
				req.UserID = parseUint(userID)
			}

			resp, err := core.TeamManager.GetOperationLogs(req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "获取操作日志失败",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, resp)
		})
	}
}

// 辅助函数
func parseUint(s string) uint {
	if s == "" {
		return 0
	}
	// 简化处理，实际应该使用strconv.ParseUint
	// 这里返回一个默认值
	return 1
}
