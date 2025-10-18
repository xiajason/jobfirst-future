package database

import (
	"context"
	"errors"
	"time"

	"github.com/xiajason/zervi-basic/basic/backend/internal/domain/user"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) user.Repository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *UserRepository) GetByID(ctx context.Context, id uint) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).Where("username = ? AND deleted_at IS NULL", username).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&user.User{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

func (r *UserRepository) List(ctx context.Context, req user.ListRequest) ([]user.User, int64, error) {
	var users []user.User
	var total int64

	query := r.db.WithContext(ctx).Model(&user.User{}).Where("deleted_at IS NULL")

	// 搜索条件
	if req.Search != "" {
		query = query.Where("username LIKE ? OR email LIKE ? OR first_name LIKE ? OR last_name LIKE ?",
			"%"+req.Search+"%", "%"+req.Search+"%", "%"+req.Search+"%", "%"+req.Search+"%")
	}

	// 状态过滤
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	if req.Page > 0 && req.PageSize > 0 {
		offset := (req.Page - 1) * req.PageSize
		query = query.Offset(offset).Limit(req.PageSize)
	}

	// 排序
	query = query.Order("created_at DESC")

	// 查询数据
	if err := query.Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id uint, status user.Status) error {
	return r.db.WithContext(ctx).Model(&user.User{}).Where("id = ?", id).Update("status", status).Error
}
