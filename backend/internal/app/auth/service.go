package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/xiajason/zervi-basic/basic/backend/internal/domain/auth"
	"github.com/xiajason/zervi-basic/basic/backend/internal/infrastructure/database"
	"github.com/xiajason/zervi-basic/basic/backend/pkg/logger"
)

// Service 认证服务
type Service struct {
	repo   *database.AuthRepository
	logger logger.Logger
}

// NewService 创建认证服务
func NewService(repo *database.AuthRepository, logger logger.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// InitializeSuperAdmin 初始化超级管理员
func (s *Service) InitializeSuperAdmin(ctx context.Context, req auth.InitializeSuperAdminRequest) (*auth.InitializeSuperAdminResponse, error) {
	s.logger.Info("初始化超级管理员: %s", req.Username)

	// 检查是否已存在超级管理员
	superAdmin, err := s.repo.GetSuperAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if superAdmin != nil {
		return &auth.InitializeSuperAdminResponse{
			Message: "超级管理员已存在",
		}, nil
	}

	// 创建超级管理员用户
	user := &auth.User{
		UUID:         uuid.New().String(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: req.Password, // 实际应用中需要加密
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Status:       auth.StatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	// 创建超级管理员角色
	role := &auth.Role{
		Name:        auth.RoleSuperAdmin,
		DisplayName: "超级管理员",
		Description: "系统超级管理员",
		Status:      auth.StatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = s.repo.CreateRole(ctx, role)
	if err != nil {
		return nil, err
	}

	// 关联用户和角色
	userRole := &auth.UserRole{
		UserID: user.ID,
		RoleID: role.ID,
	}

	err = s.repo.CreateUserRole(ctx, userRole)
	if err != nil {
		return nil, err
	}

	return &auth.InitializeSuperAdminResponse{
		Message: "超级管理员初始化成功",
		UserID:  user.ID,
	}, nil
}

// CheckSuperAdminStatus 检查超级管理员状态
func (s *Service) CheckSuperAdminStatus(ctx context.Context, req auth.CheckSuperAdminStatusRequest) (*auth.CheckSuperAdminStatusResponse, error) {
	superAdmin, err := s.repo.GetSuperAdmin(ctx)
	if err != nil {
		return nil, err
	}

	status := "not_exists"
	var userID *uint
	if superAdmin != nil {
		status = "exists"
		userID = &superAdmin.ID
	}

	return &auth.CheckSuperAdminStatusResponse{
		Exists: superAdmin != nil,
		UserID: userID,
		Status: status,
	}, nil
}

// ResetSuperAdminPassword 重置超级管理员密码
func (s *Service) ResetSuperAdminPassword(ctx context.Context, req auth.ResetSuperAdminPasswordRequest) error {
	s.logger.Info("重置超级管理员密码")

	// 获取超级管理员
	superAdmin, err := s.repo.GetSuperAdmin(ctx)
	if err != nil {
		return err
	}

	if superAdmin == nil {
		return &auth.ErrSuperAdminNotFound{}
	}

	// 更新密码
	superAdmin.PasswordHash = req.NewPassword // 实际应用中需要加密
	superAdmin.UpdatedAt = time.Now()

	return s.repo.UpdateUser(ctx, superAdmin)
}
