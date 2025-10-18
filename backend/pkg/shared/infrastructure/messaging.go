package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// MessageHandler 消息处理器接口
type MessageHandler func(ctx context.Context, message *Message) error

// Message 消息结构
type Message struct {
	ID         string                 `json:"id"`
	Topic      string                 `json:"topic"`
	Data       interface{}            `json:"data"`
	Headers    map[string]string      `json:"headers"`
	Metadata   map[string]interface{} `json:"metadata"`
	Timestamp  time.Time              `json:"timestamp"`
	RetryCount int                    `json:"retry_count"`
}

// MessageQueue 消息队列接口
type MessageQueue interface {
	Publish(ctx context.Context, topic string, message *Message) error
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error
	Unsubscribe(ctx context.Context, topic string) error
	PublishBatch(ctx context.Context, topic string, messages []*Message) error
	GetDeadLetterQueue(ctx context.Context, topic string) ([]*Message, error)
	RetryMessage(ctx context.Context, topic string, messageID string) error
	Close() error
}

// MessagingConfig 消息队列配置
type MessagingConfig struct {
	RedisAddr     string `yaml:"redis_addr" json:"redis_addr"`
	RedisPassword string `yaml:"redis_password" json:"redis_password"`
	RedisDB       int    `yaml:"redis_db" json:"redis_db"`

	MaxRetries    int           `yaml:"max_retries" json:"max_retries"`
	RetryDelay    time.Duration `yaml:"retry_delay" json:"retry_delay"`
	BatchSize     int           `yaml:"batch_size" json:"batch_size"`
	ConsumerGroup string        `yaml:"consumer_group" json:"consumer_group"`
}

// RedisStreamsQueue Redis Streams消息队列实现
type RedisStreamsQueue struct {
	client *redis.Client
	config *MessagingConfig
	ctx    context.Context
	cancel context.CancelFunc
}

// NewRedisStreamsQueue 创建Redis Streams消息队列
func NewRedisStreamsQueue(config *MessagingConfig) (*RedisStreamsQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &RedisStreamsQueue{
		client: client,
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Publish 发布消息
func (q *RedisStreamsQueue) Publish(ctx context.Context, topic string, message *Message) error {
	if message.ID == "" {
		message.ID = generateMessageID()
	}
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	// 序列化消息
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// 发布到Redis Stream
	streamKey := fmt.Sprintf("stream:%s", topic)
	args := []interface{}{
		"*", // 自动生成ID
		"data", string(data),
		"topic", topic,
		"timestamp", message.Timestamp.Unix(),
	}

	// 添加头部信息
	for key, value := range message.Headers {
		args = append(args, key, value)
	}

	_, err = q.client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: args,
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	Info("Message published successfully",
		Field{Key: "topic", Value: topic},
		Field{Key: "message_id", Value: message.ID},
	)

	return nil
}

// Subscribe 订阅消息
func (q *RedisStreamsQueue) Subscribe(ctx context.Context, topic string, handler MessageHandler) error {
	streamKey := fmt.Sprintf("stream:%s", topic)
	groupKey := fmt.Sprintf("group:%s", q.config.ConsumerGroup)
	consumerName := fmt.Sprintf("consumer:%s", generateConsumerID())

	// 创建消费者组
	err := q.createConsumerGroup(ctx, streamKey, groupKey)
	if err != nil {
		return err
	}

	// 启动消费者协程
	go q.consumeMessages(ctx, streamKey, groupKey, consumerName, handler)

	Info("Message subscription started",
		Field{Key: "topic", Value: topic},
		Field{Key: "consumer_group", Value: q.config.ConsumerGroup},
		Field{Key: "consumer_name", Value: consumerName},
	)

	return nil
}

// Unsubscribe 取消订阅
func (q *RedisStreamsQueue) Unsubscribe(ctx context.Context, topic string) error {
	// 取消上下文，停止所有消费者
	q.cancel()

	Info("Message subscription stopped",
		Field{Key: "topic", Value: topic},
	)

	return nil
}

// PublishBatch 批量发布消息
func (q *RedisStreamsQueue) PublishBatch(ctx context.Context, topic string, messages []*Message) error {
	if len(messages) == 0 {
		return nil
	}

	streamKey := fmt.Sprintf("stream:%s", topic)
	pipe := q.client.Pipeline()

	for _, message := range messages {
		if message.ID == "" {
			message.ID = generateMessageID()
		}
		if message.Timestamp.IsZero() {
			message.Timestamp = time.Now()
		}

		data, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %v", err)
		}

		args := []interface{}{
			"*",
			"data", string(data),
			"topic", topic,
			"timestamp", message.Timestamp.Unix(),
		}

		for key, value := range message.Headers {
			args = append(args, key, value)
		}

		pipe.XAdd(ctx, &redis.XAddArgs{
			Stream: streamKey,
			Values: args,
		})
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to publish batch messages: %v", err)
	}

	Info("Batch messages published successfully",
		Field{Key: "topic", Value: topic},
		Field{Key: "count", Value: len(messages)},
	)

	return nil
}

// GetDeadLetterQueue 获取死信队列
func (q *RedisStreamsQueue) GetDeadLetterQueue(ctx context.Context, topic string) ([]*Message, error) {
	dlqKey := fmt.Sprintf("dlq:%s", topic)

	// 获取死信队列中的所有消息
	result, err := q.client.LRange(ctx, dlqKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get dead letter queue: %v", err)
	}

	var messages []*Message
	for _, data := range result {
		var message Message
		if err := json.Unmarshal([]byte(data), &message); err != nil {
			Error("Failed to unmarshal dead letter message",
				Field{Key: "error", Value: err.Error()},
			)
			continue
		}
		messages = append(messages, &message)
	}

	return messages, nil
}

// RetryMessage 重试消息
func (q *RedisStreamsQueue) RetryMessage(ctx context.Context, topic string, messageID string) error {
	dlqKey := fmt.Sprintf("dlq:%s", topic)

	// 从死信队列中移除消息
	messageData, err := q.client.LPop(ctx, dlqKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get message from dead letter queue: %v", err)
	}

	var message Message
	if err := json.Unmarshal([]byte(messageData), &message); err != nil {
		return fmt.Errorf("failed to unmarshal message: %v", err)
	}

	// 增加重试次数
	message.RetryCount++

	// 重新发布消息
	return q.Publish(ctx, topic, &message)
}

// Close 关闭消息队列
func (q *RedisStreamsQueue) Close() error {
	q.cancel()
	return q.client.Close()
}

// createConsumerGroup 创建消费者组
func (q *RedisStreamsQueue) createConsumerGroup(ctx context.Context, streamKey, _ string) error {
	// 检查消费者组是否存在
	groups, err := q.client.XInfoGroups(ctx, streamKey).Result()
	if err != nil {
		// 如果流不存在，创建流和消费者组
		if err := q.client.XGroupCreate(ctx, streamKey, q.config.ConsumerGroup, "0").Err(); err != nil {
			return fmt.Errorf("failed to create consumer group: %v", err)
		}
		return nil
	}

	// 检查消费者组是否已存在
	for _, group := range groups {
		if group.Name == q.config.ConsumerGroup {
			return nil
		}
	}

	// 创建消费者组
	return q.client.XGroupCreate(ctx, streamKey, q.config.ConsumerGroup, "0").Err()
}

// consumeMessages 消费消息
func (q *RedisStreamsQueue) consumeMessages(ctx context.Context, streamKey, groupKey, consumerName string, handler MessageHandler) {
	for {
		select {
		case <-q.ctx.Done():
			return
		default:
			// 读取消息
			streams, err := q.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    q.config.ConsumerGroup,
				Consumer: consumerName,
				Streams:  []string{streamKey, ">"},
				Count:    int64(q.config.BatchSize),
				Block:    0,
			}).Result()

			if err != nil {
				Error("Failed to read messages",
					Field{Key: "error", Value: err.Error()},
				)
				time.Sleep(time.Second)
				continue
			}

			for _, stream := range streams {
				for _, message := range stream.Messages {
					go q.processMessage(ctx, streamKey, groupKey, message, handler)
				}
			}
		}
	}
}

// processMessage 处理消息
func (q *RedisStreamsQueue) processMessage(ctx context.Context, streamKey, groupKey string, message redis.XMessage, handler MessageHandler) {
	// 解析消息数据
	data, exists := message.Values["data"]
	if !exists {
		Error("Message data not found",
			Field{Key: "message_id", Value: message.ID},
		)
		q.ackMessage(ctx, streamKey, groupKey, message.ID)
		return
	}

	var msg Message
	if err := json.Unmarshal([]byte(data.(string)), &msg); err != nil {
		Error("Failed to unmarshal message",
			Field{Key: "error", Value: err.Error()},
			Field{Key: "message_id", Value: message.ID},
		)
		q.ackMessage(ctx, streamKey, groupKey, message.ID)
		return
	}

	// 处理消息
	if err := handler(ctx, &msg); err != nil {
		Error("Message processing failed",
			Field{Key: "error", Value: err.Error()},
			Field{Key: "message_id", Value: message.ID},
			Field{Key: "retry_count", Value: msg.RetryCount},
		)

		// 检查重试次数
		if msg.RetryCount < q.config.MaxRetries {
			// 重新发布消息进行重试
			msg.RetryCount++
			time.Sleep(q.config.RetryDelay)
			q.Publish(ctx, msg.Topic, &msg)
		} else {
			// 发送到死信队列
			q.sendToDeadLetterQueue(ctx, msg.Topic, &msg)
		}
	}

	// 确认消息
	q.ackMessage(ctx, streamKey, groupKey, message.ID)
}

// ackMessage 确认消息
func (q *RedisStreamsQueue) ackMessage(ctx context.Context, streamKey, _, messageID string) {
	err := q.client.XAck(ctx, streamKey, q.config.ConsumerGroup, messageID).Err()
	if err != nil {
		Error("Failed to ack message",
			Field{Key: "error", Value: err.Error()},
			Field{Key: "message_id", Value: messageID},
		)
	}
}

// sendToDeadLetterQueue 发送到死信队列
func (q *RedisStreamsQueue) sendToDeadLetterQueue(ctx context.Context, topic string, message *Message) {
	dlqKey := fmt.Sprintf("dlq:%s", topic)
	data, err := json.Marshal(message)
	if err != nil {
		Error("Failed to marshal dead letter message",
			Field{Key: "error", Value: err.Error()},
		)
		return
	}

	err = q.client.RPush(ctx, dlqKey, string(data)).Err()
	if err != nil {
		Error("Failed to send message to dead letter queue",
			Field{Key: "error", Value: err.Error()},
		)
	}

	Info("Message sent to dead letter queue",
		Field{Key: "topic", Value: topic},
		Field{Key: "message_id", Value: message.ID},
		Field{Key: "retry_count", Value: message.RetryCount},
	)
}

// CreateDefaultMessagingConfig 创建默认消息队列配置
func CreateDefaultMessagingConfig() *MessagingConfig {
	return &MessagingConfig{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       0,
		MaxRetries:    3,
		RetryDelay:    5 * time.Second,
		BatchSize:     10,
		ConsumerGroup: "jobfirst-consumer-group",
	}
}

// 全局消息队列实例
var globalMessageQueue MessageQueue

// InitGlobalMessageQueue 初始化全局消息队列
func InitGlobalMessageQueue(queue MessageQueue) {
	globalMessageQueue = queue
}

// GetMessageQueue 获取全局消息队列
func GetMessageQueue() MessageQueue {
	return globalMessageQueue
}

// 辅助函数
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

func generateConsumerID() string {
	return fmt.Sprintf("consumer_%d", time.Now().UnixNano())
}

// NoopMessageQueue 空操作消息队列（用于禁用消息队列）
type NoopMessageQueue struct{}

// Publish 空操作发布
func (q *NoopMessageQueue) Publish(ctx context.Context, topic string, message *Message) error {
	return nil
}

// Subscribe 空操作订阅
func (q *NoopMessageQueue) Subscribe(ctx context.Context, topic string, handler MessageHandler) error {
	return nil
}

// Unsubscribe 空操作取消订阅
func (q *NoopMessageQueue) Unsubscribe(ctx context.Context, topic string) error {
	return nil
}

// PublishBatch 空操作批量发布
func (q *NoopMessageQueue) PublishBatch(ctx context.Context, topic string, messages []*Message) error {
	return nil
}

// GetDeadLetterQueue 空操作获取死信队列
func (q *NoopMessageQueue) GetDeadLetterQueue(ctx context.Context, topic string) ([]*Message, error) {
	return []*Message{}, nil
}

// RetryMessage 空操作重试消息
func (q *NoopMessageQueue) RetryMessage(ctx context.Context, topic string, messageID string) error {
	return nil
}

// Close 空操作关闭
func (q *NoopMessageQueue) Close() error {
	return nil
}
