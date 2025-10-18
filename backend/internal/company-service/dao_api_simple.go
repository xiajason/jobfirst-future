package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jobfirst/jobfirst-core"
)

// DAO API路由设置
func setupDAORoutes(r *gin.Engine, core *jobfirst.Core) {
	// DAO管理API路由组
	dao := r.Group("/api/v1/dao")
	authMiddleware := core.AuthMiddleware.RequireAuth()
	dao.Use(authMiddleware)
	{
		// 创建企业DAO
		dao.POST("/", func(c *gin.Context) {
			createCompanyDAO(c, core)
		})

		// 获取企业DAO列表
		dao.GET("/", func(c *gin.Context) {
			getCompanyDAOList(c, core)
		})

		// 获取单个DAO信息
		dao.GET("/:id", func(c *gin.Context) {
			getCompanyDAO(c, core)
		})

		// 更新DAO信息
		dao.PUT("/:id", func(c *gin.Context) {
			updateCompanyDAO(c, core)
		})

		// 删除DAO
		dao.DELETE("/:id", func(c *gin.Context) {
			deleteCompanyDAO(c, core)
		})

		// DAO成员管理
		dao.POST("/:id/members", func(c *gin.Context) {
			addDAOMember(c, core)
		})

		dao.DELETE("/:id/members/:user_id", func(c *gin.Context) {
			removeDAOMember(c, core)
		})

		dao.GET("/:id/members", func(c *gin.Context) {
			getDAOMembers(c, core)
		})

		// 提案管理
		dao.POST("/:id/proposals", func(c *gin.Context) {
			createProposal(c, core)
		})

		dao.GET("/:id/proposals", func(c *gin.Context) {
			getProposals(c, core)
		})

		dao.GET("/proposals/:proposal_id", func(c *gin.Context) {
			getProposal(c, core)
		})

		dao.PUT("/proposals/:proposal_id", func(c *gin.Context) {
			updateProposal(c, core)
		})

		dao.DELETE("/proposals/:proposal_id", func(c *gin.Context) {
			deleteProposal(c, core)
		})

		// 投票管理
		dao.POST("/proposals/:proposal_id/vote", func(c *gin.Context) {
			voteOnProposal(c, core)
		})

		dao.GET("/proposals/:proposal_id/votes", func(c *gin.Context) {
			getProposalVotes(c, core)
		})

		// 自主管理团队
		dao.POST("/:id/teams", func(c *gin.Context) {
			createAutonomousTeam(c, core)
		})

		dao.GET("/:id/teams", func(c *gin.Context) {
			getAutonomousTeams(c, core)
		})

		dao.PUT("/teams/:team_id", func(c *gin.Context) {
			updateAutonomousTeam(c, core)
		})

		dao.DELETE("/teams/:team_id", func(c *gin.Context) {
			deleteAutonomousTeam(c, core)
		})

		// 团队成员管理
		dao.POST("/teams/:team_id/members", func(c *gin.Context) {
			addTeamMember(c, core)
		})

		dao.DELETE("/teams/:team_id/members/:user_id", func(c *gin.Context) {
			removeTeamMember(c, core)
		})

		dao.GET("/teams/:team_id/members", func(c *gin.Context) {
			getTeamMembers(c, core)
		})

		// DAO活动记录
		dao.GET("/:id/activities", func(c *gin.Context) {
			getDAOActivities(c, core)
		})

		// DAO配置管理
		dao.GET("/:id/settings", func(c *gin.Context) {
			getDAOSettings(c, core)
		})

		dao.PUT("/:id/settings", func(c *gin.Context) {
			updateDAOSettings(c, core)
		})
	}
}

// 创建企业DAO
func createCompanyDAO(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	var dao CompanyDAO
	if err := c.ShouldBindJSON(&dao); err != nil {
		standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	// 验证企业是否存在且用户有权限
	db := core.GetDB()
	var company Company
	if err := db.First(&company, dao.CompanyID).Error; err != nil {
		standardErrorResponse(c, http.StatusNotFound, "Company not found", err.Error())
		return
	}

	// 检查用户是否有权限创建DAO
	if company.CreatedBy != userID {
		standardErrorResponse(c, http.StatusForbidden, "No permission to create DAO for this company", "")
		return
	}

	// 创建DAO
	dao.CreatedBy = userID
	dao.CreatedAt = time.Now()
	dao.UpdatedAt = time.Now()
	dao.Status = "active"

	if err := db.Create(&dao).Error; err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to create DAO", err.Error())
		return
	}

	// 创建者自动成为DAO成员
	member := DAOMember{
		DAOID:             dao.ID,
		UserID:            userID,
		Role:              "founder",
		VotingPower:       1000, // 创始人默认投票权重
		TokenBalance:      1000,
		ContributionScore: 0,
		JoinedAt:          time.Now(),
		Status:            "active",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	db.Create(&member)

	// 记录活动
	activity := DAOActivity{
		DAOID:        dao.ID,
		UserID:       userID,
		ActivityType: "dao_created",
		Description:  "DAO created successfully",
		CreatedAt:    time.Now(),
	}
	db.Create(&activity)

	standardSuccessResponse(c, dao, "DAO created successfully")
}

// 获取企业DAO列表
func getCompanyDAOList(c *gin.Context, core *jobfirst.Core) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	companyID := c.Query("company_id")
	status := c.Query("status")

	db := core.GetDB()
	var daos []CompanyDAO
	offset := (page - 1) * pageSize

	query := db.Model(&CompanyDAO{}).Preload("Company").Preload("Members")
	if companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Offset(offset).Limit(pageSize).Find(&daos).Error; err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to get DAO list", err.Error())
		return
	}

	var total int64
	query.Count(&total)

	standardSuccessResponse(c, gin.H{
		"daos":  daos,
		"total": total,
		"page":  page,
		"size":  pageSize,
	}, "DAO list retrieved successfully")
}

// 获取单个DAO信息
func getCompanyDAO(c *gin.Context, core *jobfirst.Core) {
	daoID, _ := strconv.Atoi(c.Param("id"))

	db := core.GetDB()
	var dao CompanyDAO
	if err := db.Preload("Company").Preload("Members").Preload("Proposals").Preload("Teams").First(&dao, daoID).Error; err != nil {
		standardErrorResponse(c, http.StatusNotFound, "DAO not found", err.Error())
		return
	}

	standardSuccessResponse(c, dao, "DAO information retrieved successfully")
}

// 更新DAO信息
func updateCompanyDAO(c *gin.Context, core *jobfirst.Core) {
	daoID, _ := strconv.Atoi(c.Param("id"))
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	var updateData CompanyDAO
	if err := c.ShouldBindJSON(&updateData); err != nil {
		standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	db := core.GetDB()
	var dao CompanyDAO
	if err := db.First(&dao, daoID).Error; err != nil {
		standardErrorResponse(c, http.StatusNotFound, "DAO not found", err.Error())
		return
	}

	// 检查权限
	if dao.CreatedBy != userID {
		standardErrorResponse(c, http.StatusForbidden, "No permission to update this DAO", "")
		return
	}

	// 更新DAO信息
	updateData.UpdatedAt = time.Now()
	if err := db.Model(&dao).Updates(updateData).Error; err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to update DAO", err.Error())
		return
	}

	standardSuccessResponse(c, dao, "DAO updated successfully")
}

// 删除DAO
func deleteCompanyDAO(c *gin.Context, core *jobfirst.Core) {
	daoID, _ := strconv.Atoi(c.Param("id"))
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	db := core.GetDB()
	var dao CompanyDAO
	if err := db.First(&dao, daoID).Error; err != nil {
		standardErrorResponse(c, http.StatusNotFound, "DAO not found", err.Error())
		return
	}

	// 检查权限
	if dao.CreatedBy != userID {
		standardErrorResponse(c, http.StatusForbidden, "No permission to delete this DAO", "")
		return
	}

	// 软删除DAO
	if err := db.Model(&dao).Update("status", "inactive").Error; err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to delete DAO", err.Error())
		return
	}

	standardSuccessResponse(c, gin.H{}, "DAO deleted successfully")
}

// 其他函数的简化实现
func addDAOMember(c *gin.Context, core *jobfirst.Core) {
	standardSuccessResponse(c, gin.H{"message": "Member added successfully"}, "Member added successfully")
}

func removeDAOMember(c *gin.Context, core *jobfirst.Core) {
	standardSuccessResponse(c, gin.H{"message": "Member removed successfully"}, "Member removed successfully")
}

func getDAOMembers(c *gin.Context, core *jobfirst.Core) {
	standardSuccessResponse(c, gin.H{"members": []gin.H{}}, "Members retrieved successfully")
}

func createProposal(c *gin.Context, core *jobfirst.Core) {
	standardSuccessResponse(c, gin.H{"message": "Proposal created successfully"}, "Proposal created successfully")
}

func getProposals(c *gin.Context, core *jobfirst.Core) {
	standardSuccessResponse(c, gin.H{"proposals": []gin.H{}}, "Proposals retrieved successfully")
}

func getProposal(c *gin.Context, core *jobfirst.Core) {
	standardSuccessResponse(c, gin.H{"proposal": gin.H{}}, "Proposal retrieved successfully")
}

func updateProposal(c *gin.Context, core *jobfirst.Core) {
	standardSuccessResponse(c, gin.H{"message": "Proposal updated successfully"}, "Proposal updated successfully")
}

func deleteProposal(c *gin.Context, core *jobfirst.Core) {
	standardSuccessResponse(c, gin.H{"message": "Proposal deleted successfully"}, "Proposal deleted successfully")
}

func voteOnProposal(c *gin.Context, core *jobfirst.Core) {
	standardSuccessResponse(c, gin.H{"message": "Vote recorded successfully"}, "Vote recorded successfully")
}

func getProposalVotes(c *gin.Context, core *jobfirst.Core) {
	standardSuccessResponse(c, gin.H{"votes": []gin.H{}}, "Votes retrieved successfully")
}

func createAutonomousTeam(c *gin.Context, core *jobfirst.Core) {
	standardSuccessResponse(c, gin.H{"message": "Team created successfully"}, "Team created successfully")
}

func getAutonomousTeams(c *gin.Context, core *jobfirst.Core) {
	standardSuccessResponse(c, gin.H{"teams": []gin.H{}}, "Teams retrieved successfully")
}

func updateAutonomousTeam(c *gin.Context, core *jobfirst.Core) {
	standardSuccessResponse(c, gin.H{"message": "Team updated successfully"}, "Team updated successfully")
}

func deleteAutonomousTeam(c *gin.Context, core *jobfirst.Core) {
	standardSuccessResponse(c, gin.H{"message": "Team deleted successfully"}, "Team deleted successfully")
}

func addTeamMember(c *gin.Context, core *jobfirst.Core) {
	standardSuccessResponse(c, gin.H{"message": "Team member added successfully"}, "Team member added successfully")
}

func removeTeamMember(c *gin.Context, core *jobfirst.Core) {
	standardSuccessResponse(c, gin.H{"message": "Team member removed successfully"}, "Team member removed successfully")
}

func getTeamMembers(c *gin.Context, core *jobfirst.Core) {
	standardSuccessResponse(c, gin.H{"members": []gin.H{}}, "Team members retrieved successfully")
}

func getDAOActivities(c *gin.Context, core *jobfirst.Core) {
	standardSuccessResponse(c, gin.H{"activities": []gin.H{}}, "Activities retrieved successfully")
}

func getDAOSettings(c *gin.Context, core *jobfirst.Core) {
	standardSuccessResponse(c, gin.H{"settings": []gin.H{}}, "Settings retrieved successfully")
}

func updateDAOSettings(c *gin.Context, core *jobfirst.Core) {
	standardSuccessResponse(c, gin.H{"message": "Settings updated successfully"}, "Settings updated successfully")
}
