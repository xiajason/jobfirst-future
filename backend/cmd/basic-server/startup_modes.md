# Basic-Server 启动模式设计

## 问题分析
Basic-Server作为微服务架构中的API Gateway，与Consul的关系确实很密切：
- 需要Consul进行服务发现
- 需要将自己注册到服务发现系统
- 需要为其他服务提供路由和代理功能

但同时，Basic-Server作为基础服务，应该能够独立运行。

## 建议的解决方案：渐进式启动模式

### 模式1：独立模式 (Standalone Mode)
- **适用场景**：开发环境、测试环境、Consul不可用时
- **特点**：
  - 不依赖Consul
  - 提供基本的API功能
  - 可以独立运行
  - 适合快速开发和测试

### 模式2：服务发现模式 (Service Discovery Mode)
- **适用场景**：生产环境、完整的微服务架构
- **特点**：
  - 依赖Consul进行服务发现
  - 自动注册到服务发现系统
  - 提供服务路由和代理
  - 支持健康检查和负载均衡

### 模式3：混合模式 (Hybrid Mode)
- **适用场景**：Consul部分可用时
- **特点**：
  - 优先使用Consul
  - Consul不可用时降级到独立模式
  - 自动重连Consul
  - 提供最佳的用户体验

## 启动顺序建议

### 标准启动顺序：
1. **基础设施服务** (MySQL, Redis, PostgreSQL, Neo4j)
2. **服务发现服务** (Consul)
3. **API Gateway** (Basic-Server)
4. **核心微服务** (User Service, Resume Service)
5. **业务微服务** (Company Service, Job Service)
6. **AI服务**

### 容错启动顺序：
1. **基础设施服务** (MySQL, Redis, PostgreSQL, Neo4j)
2. **API Gateway** (Basic-Server - 独立模式)
3. **服务发现服务** (Consul)
4. **API Gateway升级** (Basic-Server - 服务发现模式)
5. **其他微服务**

## 配置建议

### 环境变量控制：
```bash
# 独立模式
BASIC_SERVER_MODE=standalone

# 服务发现模式
BASIC_SERVER_MODE=service-discovery

# 混合模式
BASIC_SERVER_MODE=hybrid
```

### 配置文件控制：
```yaml
# config.yaml
basic_server:
  mode: "standalone"  # standalone, service-discovery, hybrid
  consul:
    enabled: false    # 独立模式
    enabled: true     # 服务发现模式
    fallback: true    # 混合模式
```

## 实现建议

1. **检测Consul可用性**：
   - 启动时检测Consul是否可用
   - 根据可用性选择启动模式

2. **动态模式切换**：
   - 运行时检测Consul状态
   - 自动在模式间切换

3. **优雅降级**：
   - Consul不可用时自动降级到独立模式
   - 保持基本功能可用

4. **健康检查**：
   - 不同模式下的健康检查策略
   - 模式状态报告

## 优势

1. **灵活性**：适应不同的部署环境
2. **可靠性**：即使Consul不可用也能运行
3. **可扩展性**：支持从简单到复杂的部署
4. **维护性**：清晰的模式分离，易于维护
