# Zervigo Future 阿里云微服务部署指南

**更新时间**: 2025年10月18日  
**服务器**: 阿里云 47.115.168.107  
**部署方式**: GitHub Actions CI/CD + 时序化部署

## 📋 部署概述

本指南基于阿里云服务器的实际情况，**数据库和AI服务已预部署**，本次CI/CD流水线仅部署**Go微服务**（8080-8089端口）。

### 阿里云服务器现状

根据[服务器现状报告](../../../ALIYUN_SERVER_STATUS_REPORT_20251018.md)：

✅ **已部署服务** (无需流水线部署):
- PostgreSQL (5432) - migration-postgres容器
- MySQL (3306) - migration-mysql容器  
- Redis (6379) - migration-redis容器
- MongoDB (27017) - migration-mongodb容器
- AI Service (8100) - Python服务，已运行

❌ **待部署服务** (本次流水线部署):
- 8080: API Gateway
- 8081: User Service
- 8082: Resume Service
- 8083: Company Service
- 8084: Notification Service
- 8085: Template Service
- 8086: Statistics Service
- 8087: Banner Service
- 8088: Dev Team Service
- 8089: Job Service

## 🏗️ 微服务架构

### 完整的服务端口映射

```
数据库层 (已部署):
├── MySQL (3306)
├── PostgreSQL (5432)
├── Redis (6379)
└── MongoDB (27017)

AI服务层 (已部署):
└── AI Service (8100)

微服务层 (待部署):
├── 网关层
│   └── API Gateway (8080)
│
├── 认证授权层
│   └── User Service (8081)
│
├── 核心业务层
│   ├── Resume Service (8082)
│   └── Company Service (8083)
│
├── 支撑服务层
│   ├── Notification Service (8084)
│   ├── Template Service (8085)
│   ├── Statistics Service (8086)
│   └── Banner Service (8087)
│
└── 管理服务层
    ├── Dev Team Service (8088)
    └── Job Service (8089)
```

### 服务依赖关系

```
数据库层 (PostgreSQL, MySQL, Redis, MongoDB)
    ↓
AI Service (8100) - 已运行
    ↓
API Gateway (8080) - 统一入口
    ↓
User Service (8081) - 认证授权
    ↓
├── Resume Service (8082)
├── Company Service (8083)
├── Notification Service (8084)
├── Template Service (8085)
├── Statistics Service (8086)
├── Banner Service (8087)
├── Dev Team Service (8088)
└── Job Service (8089)
```

## 🚀 快速部署

### 方式1: 自动部署 (推荐)

通过GitHub Actions自动部署：

```bash
# 1. 推送到main分支触发生产环境部署
git push origin main

# 2. 推送到develop分支触发测试环境部署
git push origin develop

# 3. 手动触发部署
# 在GitHub仓库页面 -> Actions -> Zervigo Future 微服务部署流水线 -> Run workflow
```

### 方式2: 手动部署

如需手动部署，按以下步骤操作：

## 📝 详细部署步骤

### 准备工作

#### 1. 确认服务器环境

```bash
# SSH连接到服务器
ssh root@47.115.168.107

# 检查数据库容器状态
podman ps | grep migration

# 检查AI服务状态
ps aux | grep ai_service
curl http://localhost:8100/health
```

#### 2. 确认数据库密码配置

根据服务器报告，建议统一使用强密码：
- PostgreSQL: `JobFirst2025!PG`
- MySQL: `JobFirst2025!MySQL`
- MongoDB: `JobFirst2025!Mongo`
- Redis: `JobFirst2025!Redis`

### 阶段1: 构建微服务

在本地开发环境：

```bash
cd zervigo_future/backend

# 设置Go代理
go env -w GOPROXY=https://goproxy.cn,direct

# 下载依赖
go mod download
go mod verify

# 创建bin目录
mkdir -p bin

# 构建所有微服务
echo "构建 API Gateway..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo \
  -ldflags="-s -w" -o bin/api-gateway ./cmd/basic-server

echo "构建 User Service..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo \
  -ldflags="-s -w" -o bin/user-service ./internal/user-service

echo "构建 Resume Service..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo \
  -ldflags="-s -w" -o bin/resume-service ./internal/resume-service

echo "构建 Company Service..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo \
  -ldflags="-s -w" -o bin/company-service ./internal/company-service

echo "构建 Notification Service..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo \
  -ldflags="-s -w" -o bin/notification-service ./internal/notification-service

echo "构建 Template Service..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo \
  -ldflags="-s -w" -o bin/template-service ./internal/template-service

echo "构建 Statistics Service..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo \
  -ldflags="-s -w" -o bin/statistics-service ./internal/statistics-service

echo "构建 Banner Service..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo \
  -ldflags="-s -w" -o bin/banner-service ./internal/banner-service

echo "构建 Dev Team Service..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo \
  -ldflags="-s -w" -o bin/dev-team-service ./internal/dev-team-service

echo "构建 Job Service..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo \
  -ldflags="-s -w" -o bin/job-service ./internal/job-service

echo "✅ 所有微服务构建完成"
ls -lh bin/
```

### 阶段2: 上传文件到服务器

```bash
# 创建远程目录
ssh root@47.115.168.107 "mkdir -p /opt/services/{backend/bin,configs,logs,scripts}"

# 上传构建产物
scp bin/* root@47.115.168.107:/opt/services/backend/bin/

# 上传配置文件
scp -r configs/* root@47.115.168.107:/opt/services/configs/

# 上传部署脚本
scp scripts/*.sh root@47.115.168.107:/opt/services/scripts/
```

### 阶段3: 部署微服务 (按时序)

SSH到服务器并执行：

```bash
ssh root@47.115.168.107

cd /opt/services

# 给脚本执行权限
chmod +x backend/bin/*
chmod +x scripts/*.sh

# 停止现有微服务
echo "⏸️  停止现有微服务..."
pkill -f api-gateway || true
pkill -f user-service || true
pkill -f resume-service || true
pkill -f company-service || true
pkill -f notification-service || true
pkill -f template-service || true
pkill -f statistics-service || true
pkill -f banner-service || true
pkill -f dev-team-service || true
pkill -f job-service || true

sleep 5

# 创建日志目录
mkdir -p logs

cd backend/bin

# ========================================
# 阶段1: 网关层 (8080)
# ========================================
echo "🌐 启动网关层..."

# API Gateway (8080)
echo "启动 API Gateway (8080)..."
nohup ./api-gateway > ../../logs/api-gateway.log 2>&1 &
echo $! > ../../logs/api-gateway.pid
sleep 10
curl -f http://localhost:8080/health && echo "✅ API Gateway OK" || echo "❌ API Gateway Failed"

# ========================================
# 阶段2: 认证授权层 (8081)
# ========================================
echo "🔐 启动认证授权层..."

# User Service (8081)
echo "启动 User Service (8081)..."
nohup ./user-service > ../../logs/user-service.log 2>&1 &
echo $! > ../../logs/user-service.pid
sleep 10
curl -f http://localhost:8081/health && echo "✅ User Service OK" || echo "❌ User Service Failed"

# ========================================
# 阶段3: 核心业务层 (8082-8083)
# ========================================
echo "💼 启动核心业务层..."

# Resume Service (8082)
echo "启动 Resume Service (8082)..."
nohup ./resume-service > ../../logs/resume-service.log 2>&1 &
echo $! > ../../logs/resume-service.pid
sleep 5
curl -f http://localhost:8082/health && echo "✅ Resume Service OK" || echo "❌ Resume Service Failed"

# Company Service (8083)
echo "启动 Company Service (8083)..."
nohup ./company-service > ../../logs/company-service.log 2>&1 &
echo $! > ../../logs/company-service.pid
sleep 5
curl -f http://localhost:8083/health && echo "✅ Company Service OK" || echo "❌ Company Service Failed"

# ========================================
# 阶段4: 支撑服务层 (8084-8087)
# ========================================
echo "🔧 启动支撑服务层..."

# Notification Service (8084)
echo "启动 Notification Service (8084)..."
nohup ./notification-service > ../../logs/notification-service.log 2>&1 &
echo $! > ../../logs/notification-service.pid
sleep 3
curl -f http://localhost:8084/health && echo "✅ Notification Service OK" || echo "❌ Notification Service Failed"

# Template Service (8085)
echo "启动 Template Service (8085)..."
nohup ./template-service > ../../logs/template-service.log 2>&1 &
echo $! > ../../logs/template-service.pid
sleep 3
curl -f http://localhost:8085/health && echo "✅ Template Service OK" || echo "❌ Template Service Failed"

# Statistics Service (8086)
echo "启动 Statistics Service (8086)..."
nohup ./statistics-service > ../../logs/statistics-service.log 2>&1 &
echo $! > ../../logs/statistics-service.pid
sleep 3
curl -f http://localhost:8086/health && echo "✅ Statistics Service OK" || echo "❌ Statistics Service Failed"

# Banner Service (8087)
echo "启动 Banner Service (8087)..."
nohup ./banner-service > ../../logs/banner-service.log 2>&1 &
echo $! > ../../logs/banner-service.pid
sleep 3
curl -f http://localhost:8087/health && echo "✅ Banner Service OK" || echo "❌ Banner Service Failed"

# ========================================
# 阶段5: 管理服务层 (8088-8089)
# ========================================
echo "⚙️ 启动管理服务层..."

# Dev Team Service (8088)
echo "启动 Dev Team Service (8088)..."
nohup ./dev-team-service > ../../logs/dev-team-service.log 2>&1 &
echo $! > ../../logs/dev-team-service.pid
sleep 3
curl -f http://localhost:8088/health && echo "✅ Dev Team Service OK" || echo "❌ Dev Team Service Failed"

# Job Service (8089)
echo "启动 Job Service (8089)..."
nohup ./job-service > ../../logs/job-service.log 2>&1 &
echo $! > ../../logs/job-service.pid
sleep 3
curl -f http://localhost:8089/health && echo "✅ Job Service OK" || echo "❌ Job Service Failed"

echo ""
echo "=========================================="
echo "✅ 所有微服务部署完成！"
echo "=========================================="
```

### 阶段4: 验证部署

```bash
echo "=========================================="
echo "🔍 微服务健康检查"
echo "=========================================="

# 网关层
echo ""
echo "=== 网关层 ==="
curl -f http://localhost:8080/health && echo "✅ API Gateway (8080)" || echo "❌ API Gateway (8080)"

# 认证授权层
echo ""
echo "=== 认证授权层 ==="
curl -f http://localhost:8081/health && echo "✅ User Service (8081)" || echo "❌ User Service (8081)"

# 核心业务层
echo ""
echo "=== 核心业务层 ==="
curl -f http://localhost:8082/health && echo "✅ Resume Service (8082)" || echo "❌ Resume Service (8082)"
curl -f http://localhost:8083/health && echo "✅ Company Service (8083)" || echo "❌ Company Service (8083)"

# 支撑服务层
echo ""
echo "=== 支撑服务层 ==="
curl -f http://localhost:8084/health && echo "✅ Notification Service (8084)" || echo "❌ Notification Service (8084)"
curl -f http://localhost:8085/health && echo "✅ Template Service (8085)" || echo "❌ Template Service (8085)"
curl -f http://localhost:8086/health && echo "✅ Statistics Service (8086)" || echo "❌ Statistics Service (8086)"
curl -f http://localhost:8087/health && echo "✅ Banner Service (8087)" || echo "❌ Banner Service (8087)"

# 管理服务层
echo ""
echo "=== 管理服务层 ==="
curl -f http://localhost:8088/health && echo "✅ Dev Team Service (8088)" || echo "❌ Dev Team Service (8088)"
curl -f http://localhost:8089/health && echo "✅ Job Service (8089)" || echo "❌ Job Service (8089)"

# AI服务 (已预部署)
echo ""
echo "=== AI服务层 (预部署) ==="
curl -f http://localhost:8100/health && echo "✅ AI Service (8100)" || echo "❌ AI Service (8100)"

# 数据库 (已预部署)
echo ""
echo "=== 数据库层 (预部署) ==="
podman exec migration-mysql mysql -uroot -pJobFirst2025!MySQL -e "SELECT 'MySQL OK' as status;" 2>/dev/null && echo "✅ MySQL (3306)" || echo "❌ MySQL (3306)"
podman exec migration-postgres psql -U postgres -c "SELECT 'PostgreSQL OK' as status;" 2>/dev/null && echo "✅ PostgreSQL (5432)" || echo "❌ PostgreSQL (5432)"
podman exec migration-redis redis-cli -a JobFirst2025!Redis ping 2>/dev/null && echo "✅ Redis (6379)" || echo "❌ Redis (6379)"
podman exec migration-mongodb mongosh -u admin -p'JobFirst2025!Mongo' --authenticationDatabase admin --eval "db.version()" --quiet 2>/dev/null && echo "✅ MongoDB (27017)" || echo "❌ MongoDB (27017)"

echo ""
echo "=========================================="
echo "✅ 部署验证完成"
echo "=========================================="
```

## 📊 服务监控

### 健康检查端点

| 服务 | 端口 | 健康检查端点 | 说明 |
|------|------|-------------|------|
| API Gateway | 8080 | `/health` | 网关健康检查 |
| User Service | 8081 | `/health` | 用户服务健康检查 |
| Resume Service | 8082 | `/health` | 简历服务健康检查 |
| Company Service | 8083 | `/health` | 公司服务健康检查 |
| Notification Service | 8084 | `/health` | 通知服务健康检查 |
| Template Service | 8085 | `/health` | 模板服务健康检查 |
| Statistics Service | 8086 | `/health` | 统计服务健康检查 |
| Banner Service | 8087 | `/health` | 横幅服务健康检查 |
| Dev Team Service | 8088 | `/health` | 开发团队服务健康检查 |
| Job Service | 8089 | `/health` | 职位服务健康检查 |
| AI Service | 8100 | `/health` | AI服务健康检查 (已部署) |

### 访问地址

- **API Gateway**: http://47.115.168.107:8080
- **User Service**: http://47.115.168.107:8081
- **Resume Service**: http://47.115.168.107:8082
- **Company Service**: http://47.115.168.107:8083
- **Notification Service**: http://47.115.168.107:8084
- **Template Service**: http://47.115.168.107:8085
- **Statistics Service**: http://47.115.168.107:8086
- **Banner Service**: http://47.115.168.107:8087
- **Dev Team Service**: http://47.115.168.107:8088
- **Job Service**: http://47.115.168.107:8089
- **AI Service**: http://47.115.168.107:8100

### 日志查看

```bash
# 查看所有服务日志
ls -lh /opt/services/logs/

# 查看特定服务日志
tail -f /opt/services/logs/api-gateway.log
tail -f /opt/services/logs/user-service.log
tail -f /opt/services/logs/resume-service.log

# 查看错误日志
grep -i error /opt/services/logs/*.log
```

### 服务管理

```bash
# 查看服务进程
ps aux | grep -E "(api-gateway|user-service|resume-service|company-service|notification-service|template-service|statistics-service|banner-service|dev-team-service|job-service)"

# 查看端口监听
netstat -tlnp | grep -E "(8080|8081|8082|8083|8084|8085|8086|8087|8088|8089)"

# 重启单个服务
pkill -f user-service
cd /opt/services/backend/bin
nohup ./user-service > ../../logs/user-service.log 2>&1 &
```

## 🚨 故障排除

### 常见问题

#### 1. 服务启动失败

**问题**: 微服务启动后立即退出
**解决方案**:
```bash
# 查看日志
tail -100 /opt/services/logs/[service-name].log

# 检查端口占用
netstat -tlnp | grep [port]

# 检查配置文件
cat /opt/services/configs/config.yaml
```

#### 2. 数据库连接失败

**问题**: 微服务无法连接数据库
**解决方案**:
```bash
# 检查数据库容器状态
podman ps | grep migration

# 检查数据库连接
podman exec migration-mysql mysql -uroot -pJobFirst2025!MySQL -e "SELECT 1;"
podman exec migration-postgres psql -U postgres -c "SELECT 1;"
podman exec migration-redis redis-cli -a JobFirst2025!Redis ping
podman exec migration-mongodb mongosh -u admin -p'JobFirst2025!Mongo' --authenticationDatabase admin --eval "db.version()"

# 更新配置文件中的数据库密码
nano /opt/services/configs/config.yaml
```

#### 3. 服务间调用失败

**问题**: 服务A无法调用服务B
**解决方案**:
```bash
# 检查服务B是否运行
curl http://localhost:[port]/health

# 检查服务A的日志
tail -f /opt/services/logs/[service-a].log

# 检查网络连接
netstat -an | grep [port]
```

## 🔒 安全配置

### 防火墙配置

```bash
# 开放必要端口
sudo ufw allow 22      # SSH
sudo ufw allow 80      # HTTP
sudo ufw allow 443     # HTTPS
sudo ufw allow 8080    # API Gateway
sudo ufw allow 8081    # User Service
sudo ufw allow 8082    # Resume Service
sudo ufw allow 8083    # Company Service
sudo ufw allow 8084    # Notification Service
sudo ufw allow 8085    # Template Service
sudo ufw allow 8086    # Statistics Service
sudo ufw allow 8087    # Banner Service
sudo ufw allow 8088    # Dev Team Service
sudo ufw allow 8089    # Job Service
sudo ufw allow 8100    # AI Service
sudo ufw enable
```

### 数据库密码安全

根据[服务器现状报告](../../../ALIYUN_SERVER_STATUS_REPORT_20251018.md)，建议使用强密码：

- PostgreSQL: `JobFirst2025!PG`
- MySQL: `JobFirst2025!MySQL`
- MongoDB: `JobFirst2025!Mongo`
- Redis: `JobFirst2025!Redis`

## 📈 性能优化

### 资源监控

```bash
# CPU使用率
top -bn1 | head -20

# 内存使用率
free -h

# 磁盘使用率
df -h

# 网络连接数
netstat -an | wc -l
```

### 日志轮转

创建日志轮转配置：

```bash
cat > /etc/logrotate.d/zervigo-future << 'EOF'
/opt/services/logs/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 root root
    postrotate
        # 重启服务以释放日志文件句柄（可选）
    endscript
}
EOF
```

## 🎯 最佳实践

### 1. 时序化部署

严格按照以下顺序启动服务：
1. 网关层 (8080)
2. 认证授权层 (8081)
3. 核心业务层 (8082-8083)
4. 支撑服务层 (8084-8087)
5. 管理服务层 (8088-8089)

### 2. 健康检查

每个服务启动后，等待并验证健康检查通过再启动下一个服务。

### 3. 日志管理

- 定期清理日志文件
- 配置日志轮转
- 监控日志中的错误信息

### 4. 监控告警

- 配置服务健康监控
- 设置告警规则
- 定期检查系统资源使用情况

---

**维护人员**: AI Assistant  
**联系方式**: 通过项目文档  
**更新频率**: 随架构变更更新
