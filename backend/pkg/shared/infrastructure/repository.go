package infrastructure

import (
	"context"
)

// Repository 仓储接口
type Repository[T any] interface {
	Save(ctx context.Context, entity T) error
	FindByID(ctx context.Context, id string) (T, error)
	FindAll(ctx context.Context) ([]T, error)
	Update(ctx context.Context, entity T) error
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)
}

// BaseRepository 基础仓储实现
type BaseRepository[T any] struct {
	db Database
}

// NewBaseRepository 创建基础仓储
func NewBaseRepository[T any](db Database) *BaseRepository[T] {
	return &BaseRepository[T]{
		db: db,
	}
}

// Save 保存实体
func (r *BaseRepository[T]) Save(ctx context.Context, entity T) error {
	return r.db.Create(ctx, entity)
}

// FindByID 根据ID查找实体
func (r *BaseRepository[T]) FindByID(ctx context.Context, id string) (T, error) {
	var entity T
	err := r.db.First(ctx, &entity, "id = ?", id)
	return entity, err
}

// FindAll 查找所有实体
func (r *BaseRepository[T]) FindAll(ctx context.Context) ([]T, error) {
	var entities []T
	err := r.db.Find(ctx, &entities)
	return entities, err
}

// Update 更新实体
func (r *BaseRepository[T]) Update(ctx context.Context, entity T) error {
	return r.db.Save(ctx, entity)
}

// Delete 删除实体
func (r *BaseRepository[T]) Delete(ctx context.Context, id string) error {
	return r.db.Delete(ctx, "id = ?", id)
}

// Exists 检查实体是否存在
func (r *BaseRepository[T]) Exists(ctx context.Context, id string) (bool, error) {
	var count int64
	err := r.db.Count(ctx, &count, "id = ?", id)
	return count > 0, err
}

// Database 数据库接口
type Database interface {
	Create(ctx context.Context, entity interface{}) error
	First(ctx context.Context, dest interface{}, query interface{}, args ...interface{}) error
	Find(ctx context.Context, dest interface{}) error
	Save(ctx context.Context, entity interface{}) error
	Delete(ctx context.Context, query interface{}, args ...interface{}) error
	Count(ctx context.Context, count *int64, query interface{}, args ...interface{}) error
	Transaction(ctx context.Context, fn func(tx Database) error) error
}
