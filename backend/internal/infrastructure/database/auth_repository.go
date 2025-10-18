package database

import (
	"context"
	"errors"

	"github.com/xiajason/zervi-basic/basic/backend/internal/domain/auth"

	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateUser(ctx context.Context, user *auth.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *AuthRepository) GetUserByID(ctx context.Context, id uint) (*auth.User, error) {
	var user auth.User
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) GetUserByUsername(ctx context.Context, username string) (*auth.User, error) {
	var user auth.User
	err := r.db.WithContext(ctx).Where("username = ? AND deleted_at IS NULL", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	var user auth.User
	err := r.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) UpdateUser(ctx context.Context, user *auth.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *AuthRepository) GetSuperAdmin(ctx context.Context) (*auth.User, error) {
	var user auth.User
	err := r.db.WithContext(ctx).Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("roles.name = ? AND users.deleted_at IS NULL", auth.RoleSuperAdmin).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 超级管理员不存在
		}
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) CreateRole(ctx context.Context, role *auth.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *AuthRepository) GetRoleByName(ctx context.Context, name auth.RoleName) (*auth.Role, error) {
	var role auth.Role
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}
	return &role, nil
}

func (r *AuthRepository) CreateUserRole(ctx context.Context, userRole *auth.UserRole) error {
	return r.db.WithContext(ctx).Create(userRole).Error
}

func (r *AuthRepository) CreatePermission(ctx context.Context, permission *auth.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

func (r *AuthRepository) CreateRolePermission(ctx context.Context, rolePermission *auth.RolePermission) error {
	return r.db.WithContext(ctx).Create(rolePermission).Error
}
