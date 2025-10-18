package kernel

import (
	"time"

	"github.com/google/uuid"
)

// BaseEntity 基础实体
type BaseEntity struct {
	ID        string     `json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// NewBaseEntity 创建新的基础实体
func NewBaseEntity() BaseEntity {
	now := time.Now()
	return BaseEntity{
		ID:        uuid.New().String(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Update 更新实体
func (e *BaseEntity) Update() {
	e.UpdatedAt = time.Now()
}

// Delete 删除实体
func (e *BaseEntity) Delete() {
	now := time.Now()
	e.DeletedAt = &now
}

// IsDeleted 检查是否已删除
func (e *BaseEntity) IsDeleted() bool {
	return e.DeletedAt != nil
}
