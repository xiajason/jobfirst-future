package team

import "github.com/jobfirst/jobfirst-core/auth"

// AddMemberRequest 添加团队成员请求
type AddMemberRequest struct {
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name"`
	TeamRole  string `json:"team_role" binding:"required"`
	Phone     string `json:"phone"`
}

// AddMemberResponse 添加团队成员响应
type AddMemberResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Data    AddMemberData `json:"data"`
}

// AddMemberData 添加团队成员数据
type AddMemberData struct {
	User    auth.User        `json:"user"`
	DevTeam auth.DevTeamUser `json:"dev_team"`
}

// UpdateMemberRequest 更新团队成员请求
type UpdateMemberRequest struct {
	TeamRole                  string `json:"team_role"`
	ServerAccessLevel         string `json:"server_access_level"`
	CodeAccessModules         string `json:"code_access_modules"`
	DatabaseAccess            string `json:"database_access"`
	ServiceRestartPermissions string `json:"service_restart_permissions"`
	Status                    string `json:"status"`
}

// UpdateMemberResponse 更新团队成员响应
type UpdateMemberResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    auth.DevTeamUser `json:"data"`
}

// RemoveMemberResponse 移除团队成员响应
type RemoveMemberResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GetMembersRequest 获取团队成员请求
type GetMembersRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

// GetMembersResponse 获取团队成员响应
type GetMembersResponse struct {
	Success bool           `json:"success"`
	Data    GetMembersData `json:"data"`
}

// GetMembersData 获取团队成员数据
type GetMembersData struct {
	Members    []auth.DevTeamUser `json:"members"`
	Pagination Pagination         `json:"pagination"`
}

// GetStatsResponse 获取统计信息响应
type GetStatsResponse struct {
	Success bool      `json:"success"`
	Data    TeamStats `json:"data"`
}

// TeamStats 团队统计信息
type TeamStats struct {
	TotalMembers    int64            `json:"total_members"`
	ActiveMembers   int64            `json:"active_members"`
	InactiveMembers int64            `json:"inactive_members"`
	RoleStats       map[string]int64 `json:"role_stats"`
}

// GetOperationLogsRequest 获取操作日志请求
type GetOperationLogsRequest struct {
	Page          int    `json:"page"`
	PageSize      int    `json:"page_size"`
	UserID        uint   `json:"user_id"`
	OperationType string `json:"operation_type"`
	Status        string `json:"status"`
}

// GetOperationLogsResponse 获取操作日志响应
type GetOperationLogsResponse struct {
	Success bool                 `json:"success"`
	Data    GetOperationLogsData `json:"data"`
}

// GetOperationLogsData 获取操作日志数据
type GetOperationLogsData struct {
	Logs       []auth.DevOperationLog `json:"logs"`
	Pagination Pagination             `json:"pagination"`
}

// Pagination 分页信息
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}
