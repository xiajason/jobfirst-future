package usersync

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// UserEventPublisher 用户事件发布器
type UserEventPublisher struct {
	redisClient *redis.Client
	topic       string
	serviceName string
}

// NewUserEventPublisher 创建新的用户事件发布器
func NewUserEventPublisher(redisClient *redis.Client, serviceName string) *UserEventPublisher {
	return &UserEventPublisher{
		redisClient: redisClient,
		topic:       "user_events",
		serviceName: serviceName,
	}
}

// PublishUserCreated 发布用户创建事件
func (p *UserEventPublisher) PublishUserCreated(user *User) error {
	event := UserEvent{
		ID:        p.generateEventID(),
		Type:      EventTypeUserCreated,
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Data:      user,
		Timestamp: time.Now(),
		Source:    p.serviceName,
	}

	return p.publishEvent(event)
}

// PublishUserUpdated 发布用户更新事件
func (p *UserEventPublisher) PublishUserUpdated(user *User, changes map[string]interface{}) error {
	event := UserEvent{
		ID:       p.generateEventID(),
		Type:     EventTypeUserUpdated,
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Data: map[string]interface{}{
			"user":    user,
			"changes": changes,
		},
		Timestamp: time.Now(),
		Source:    p.serviceName,
	}

	return p.publishEvent(event)
}

// PublishUserDeleted 发布用户删除事件
func (p *UserEventPublisher) PublishUserDeleted(userID uint, username string) error {
	event := UserEvent{
		ID:       p.generateEventID(),
		Type:     EventTypeUserDeleted,
		UserID:   userID,
		Username: username,
		Email:    "",
		Data: map[string]interface{}{
			"user_id":  userID,
			"username": username,
		},
		Timestamp: time.Now(),
		Source:    p.serviceName,
	}

	return p.publishEvent(event)
}

// PublishUserStatusChanged 发布用户状态变更事件
func (p *UserEventPublisher) PublishUserStatusChanged(user *User, oldStatus, newStatus string) error {
	event := UserEvent{
		ID:       p.generateEventID(),
		Type:     EventTypeUserStatusChanged,
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Data: map[string]interface{}{
			"user":       user,
			"old_status": oldStatus,
			"new_status": newStatus,
		},
		Timestamp: time.Now(),
		Source:    p.serviceName,
	}

	return p.publishEvent(event)
}

// publishEvent 发布事件
func (p *UserEventPublisher) publishEvent(event UserEvent) error {
	if p.redisClient == nil {
		log.Printf("Redis客户端未初始化，跳过事件发布: %s", event.Type)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 序列化事件
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("序列化用户事件失败: %w", err)
	}

	// 发布到Redis Stream
	_, err = p.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: p.topic,
		Values: map[string]interface{}{
			"event": string(data),
		},
	}).Result()

	if err != nil {
		return fmt.Errorf("发布用户事件失败: %w", err)
	}

	log.Printf("用户事件已发布: %s, 用户: %s, ID: %s", event.Type, event.Username, event.ID)
	return nil
}

// generateEventID 生成事件ID
func (p *UserEventPublisher) generateEventID() string {
	return fmt.Sprintf("%s_%d_%d", p.serviceName, time.Now().Unix(), time.Now().UnixNano()%1000000)
}

// UserEventSubscriber 用户事件订阅器
type UserEventSubscriber struct {
	redisClient  *redis.Client
	topic        string
	groupName    string
	consumerName string
	handlers     map[EventType]func(UserEvent) error
}

// NewUserEventSubscriber 创建新的用户事件订阅器
func NewUserEventSubscriber(redisClient *redis.Client, groupName, consumerName string) *UserEventSubscriber {
	return &UserEventSubscriber{
		redisClient:  redisClient,
		topic:        "user_events",
		groupName:    groupName,
		consumerName: consumerName,
		handlers:     make(map[EventType]func(UserEvent) error),
	}
}

// RegisterHandler 注册事件处理器
func (s *UserEventSubscriber) RegisterHandler(eventType EventType, handler func(UserEvent) error) {
	s.handlers[eventType] = handler
}

// Start 启动事件订阅
func (s *UserEventSubscriber) Start(ctx context.Context) error {
	if s.redisClient == nil {
		return fmt.Errorf("Redis客户端未初始化")
	}

	// 创建消费者组
	err := s.redisClient.XGroupCreateMkStream(ctx, s.topic, s.groupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("创建消费者组失败: %w", err)
	}

	log.Printf("用户事件订阅器启动，消费者组: %s, 消费者: %s", s.groupName, s.consumerName)

	// 开始消费事件
	for {
		select {
		case <-ctx.Done():
			log.Println("用户事件订阅器停止")
			return nil
		default:
			// 读取事件
			streams, err := s.redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    s.groupName,
				Consumer: s.consumerName,
				Streams:  []string{s.topic, ">"},
				Count:    10,
				Block:    time.Second,
			}).Result()

			if err != nil {
				if err == redis.Nil {
					continue // 没有新事件，继续等待
				}
				log.Printf("读取用户事件失败: %v", err)
				continue
			}

			// 处理事件
			for _, stream := range streams {
				for _, message := range stream.Messages {
					if err := s.processMessage(ctx, message); err != nil {
						log.Printf("处理用户事件失败: %v", err)
					}
				}
			}
		}
	}
}

// processMessage 处理消息
func (s *UserEventSubscriber) processMessage(ctx context.Context, message redis.XMessage) error {
	// 解析事件
	eventData, exists := message.Values["event"]
	if !exists {
		return fmt.Errorf("消息中缺少事件数据")
	}

	eventStr, ok := eventData.(string)
	if !ok {
		return fmt.Errorf("事件数据格式错误")
	}

	var event UserEvent
	if err := json.Unmarshal([]byte(eventStr), &event); err != nil {
		return fmt.Errorf("解析用户事件失败: %w", err)
	}

	// 查找处理器
	handler, exists := s.handlers[event.Type]
	if !exists {
		log.Printf("未找到事件处理器: %s", event.Type)
		return nil
	}

	// 执行处理器
	if err := handler(event); err != nil {
		return fmt.Errorf("处理用户事件失败: %w", err)
	}

	// 确认消息
	return s.redisClient.XAck(ctx, s.topic, s.groupName, message.ID).Err()
}
