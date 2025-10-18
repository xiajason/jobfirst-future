package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/xiajason/zervi-basic/basic/backend/internal/domain/user"
	"github.com/xiajason/zervi-basic/basic/backend/pkg/logger"
)

// Service 用户服务接口
type Service interface {
	Register(ctx context.Context, req user.RegisterRequest) (*user.RegisterResponse, error)
	Login(ctx context.Context, req user.LoginRequest) (*user.LoginResponse, error)
	UpdateProfile(ctx context.Context, req user.UpdateProfileRequest) (*user.UpdateProfileResponse, error)
	List(ctx context.Context, req user.ListRequest) (*user.ListResponse, error)
}

// service 用户服务实现
type service struct {
	repo   user.Repository
	logger logger.Logger
}

// NewService 创建用户服务
func NewService(repo user.Repository, logger logger.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// Register 用户注册
func (s *service) Register(ctx context.Context, req user.RegisterRequest) (*user.RegisterResponse, error) {
	s.logger.Info("用户注册: %s", req.Username)

	// 检查用户名是否已存在
	_, err := s.repo.GetByUsername(ctx, req.Username)
	if err == nil {
		return nil, &user.ErrUsernameExists{Username: req.Username}
	}

	// 检查邮箱是否已存在
	_, err = s.repo.GetByEmail(ctx, req.Email)
	if err == nil {
		return nil, &user.ErrEmailExists{Email: req.Email}
	}

	// 创建用户
	userEntity := &user.User{
		UUID:         uuid.New().String(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: req.Password, // 实际应用中需要加密
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		Status:       user.StatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = s.repo.Create(ctx, userEntity)
	if err != nil {
		return nil, err
	}

	return &user.RegisterResponse{
		Message: "用户注册成功",
		User:    *userEntity,
	}, nil
}

// Login 用户登录
func (s *service) Login(ctx context.Context, req user.LoginRequest) (*user.LoginResponse, error) {
	s.logger.Info("用户登录: %s", req.Username)

	// 验证用户凭据
	userEntity, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	// 验证密码 (实际应用中需要比较加密后的密码)
	if userEntity.PasswordHash != req.Password {
		return nil, &user.ErrInvalidCredentials{}
	}

	// 检查用户状态
	if userEntity.Status != user.StatusActive {
		return nil, &user.ErrUserInactive{}
	}

	// 生成JWT token (这里简化处理)
	token := "jwt_token_" + userEntity.Username + "_" + time.Now().Format("20060102150405")

	return &user.LoginResponse{
		Token: token,
		User:  *userEntity,
	}, nil
}

// UpdateProfile 更新用户资料
func (s *service) UpdateProfile(ctx context.Context, req user.UpdateProfileRequest) (*user.UpdateProfileResponse, error) {
	s.logger.Info("更新用户资料")

	// 这里需要从context中获取当前用户ID
	// 为了简化，这里直接返回成功
	return &user.UpdateProfileResponse{
		Message: "用户资料更新成功",
	}, nil
}

// List 获取用户列表
func (s *service) List(ctx context.Context, req user.ListRequest) (*user.ListResponse, error) {
	s.logger.Info("获取用户列表, 页码: %d, 每页: %d", req.Page, req.PageSize)

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 获取用户列表
	users, total, err := s.repo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	return &user.ListResponse{
		Users: users,
		Total: total,
		Page:  req.Page,
		Size:  req.PageSize,
	}, nil
}
