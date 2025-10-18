package application

import (
	"context"
	"errors"
)

// ApplicationService 应用服务接口
type ApplicationService interface {
	Execute(ctx context.Context, command interface{}) (interface{}, error)
}

// BaseApplicationService 基础应用服务
type BaseApplicationService struct {
	logger Logger
}

// NewBaseApplicationService 创建基础应用服务
func NewBaseApplicationService(logger Logger) *BaseApplicationService {
	return &BaseApplicationService{
		logger: logger,
	}
}

// Logger 日志接口
type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	Warn(args ...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

// Command 命令接口
type Command interface {
	CommandID() string
	CommandType() string
}

// Query 查询接口
type Query interface {
	QueryID() string
	QueryType() string
}

// CommandHandler 命令处理器接口
type CommandHandler[C Command, R any] interface {
	Handle(ctx context.Context, command C) (R, error)
}

// QueryHandler 查询处理器接口
type QueryHandler[Q Query, R any] interface {
	Handle(ctx context.Context, query Q) (R, error)
}

// CommandBus 命令总线接口
type CommandBus interface {
	Send(ctx context.Context, command Command) (interface{}, error)
	Register(handler CommandHandler[Command, interface{}])
}

// QueryBus 查询总线接口
type QueryBus interface {
	Send(ctx context.Context, query Query) (interface{}, error)
	Register(handler QueryHandler[Query, interface{}])
}

// BaseCommandBus 基础命令总线
type BaseCommandBus struct {
	handlers map[string]CommandHandler[Command, interface{}]
}

// NewBaseCommandBus 创建基础命令总线
func NewBaseCommandBus() *BaseCommandBus {
	return &BaseCommandBus{
		handlers: make(map[string]CommandHandler[Command, interface{}]),
	}
}

// Send 发送命令
func (b *BaseCommandBus) Send(ctx context.Context, command Command) (interface{}, error) {
	handler, exists := b.handlers[command.CommandType()]
	if !exists {
		return nil, ErrHandlerNotFound
	}
	return handler.Handle(ctx, command)
}

// Register 注册命令处理器
func (b *BaseCommandBus) Register(handler CommandHandler[Command, interface{}]) {
	// 这里需要实现类型推断来获取命令类型
	// 简化实现，实际使用时需要更复杂的类型处理
}

// BaseQueryBus 基础查询总线
type BaseQueryBus struct {
	handlers map[string]QueryHandler[Query, interface{}]
}

// NewBaseQueryBus 创建基础查询总线
func NewBaseQueryBus() *BaseQueryBus {
	return &BaseQueryBus{
		handlers: make(map[string]QueryHandler[Query, interface{}]),
	}
}

// Send 发送查询
func (b *BaseQueryBus) Send(ctx context.Context, query Query) (interface{}, error) {
	handler, exists := b.handlers[query.QueryType()]
	if !exists {
		return nil, ErrHandlerNotFound
	}
	return handler.Handle(ctx, query)
}

// Register 注册查询处理器
func (b *BaseQueryBus) Register(handler QueryHandler[Query, interface{}]) {
	// 这里需要实现类型推断来获取查询类型
	// 简化实现，实际使用时需要更复杂的类型处理
}

// 错误定义
var (
	ErrHandlerNotFound = errors.New("handler not found")
)
