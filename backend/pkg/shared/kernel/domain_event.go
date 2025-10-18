package kernel

import (
	"time"
)

// DomainEvent 领域事件接口
type DomainEvent interface {
	EventID() string
	EventType() string
	OccurredOn() time.Time
	AggregateID() string
}

// BaseDomainEvent 基础领域事件
type BaseDomainEvent struct {
	ID          string    `json:"event_id"`
	Type        string    `json:"event_type"`
	OccurredAt  time.Time `json:"occurred_on"`
	AggregateID string    `json:"aggregate_id"`
}

// NewBaseDomainEvent 创建基础领域事件
func NewBaseDomainEvent(eventType, aggregateID string) BaseDomainEvent {
	return BaseDomainEvent{
		ID:          generateID(),
		Type:        eventType,
		OccurredAt:  time.Now(),
		AggregateID: aggregateID,
	}
}

// EventID 获取事件ID
func (e BaseDomainEvent) EventID() string {
	return e.ID
}

// EventType 获取事件类型
func (e BaseDomainEvent) EventType() string {
	return e.Type
}

// OccurredOn 获取事件发生时间
func (e BaseDomainEvent) OccurredOn() time.Time {
	return e.OccurredAt
}

// AggregateID 获取聚合根ID
func (e BaseDomainEvent) AggregateID() string {
	return e.AggregateID
}

// DomainEventHandler 领域事件处理器接口
type DomainEventHandler interface {
	Handle(event DomainEvent) error
}

// DomainEventPublisher 领域事件发布器接口
type DomainEventPublisher interface {
	Publish(event DomainEvent) error
	PublishAll(events []DomainEvent) error
}

// DomainEventStore 领域事件存储接口
type DomainEventStore interface {
	Save(event DomainEvent) error
	GetEvents(aggregateID string) ([]DomainEvent, error)
	GetEventsByType(eventType string) ([]DomainEvent, error)
}

// generateID 生成唯一ID
func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
