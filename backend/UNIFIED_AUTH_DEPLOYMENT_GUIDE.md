# 统一认证系统部署指南

## 🎯 解决方案概述

本指南解决了以下关键问题：
1. **数据库字段问题**: 用户信息查询时NULL值处理不当
2. **JWT实现不完整**: 当前是临时实现，需要完整的JWT验证
3. **权限表缺失**: 权限检查返回false，可能缺少权限表数据
4. **角色系统复杂**: 多个地方定义了不同的角色常量

## 🏗️ 架构设计

### 统一角色系统
```
guest (访客)     - Level 1 - 只能读取公开内容
user (普通用户)   - Level 2 - 可以管理自己的内容
admin (管理员)    - Level 3 - 可以管理所有内容
super_admin (超级管理员) - Level 4 - 拥有所有权限
```

### 权限系统
```
read:public    - 读取公开内容
read:own       - 读取自己的内容
write:own      - 修改自己的内容
read:all       - 读取所有内容
write:all      - 修改所有内容
delete:own     - 删除自己的内容
delete:all     - 删除所有内容
admin:users    - 用户管理
admin:system   - 系统管理
```

## 📋 部署步骤

### 1. 数据库迁移

```bash
# 执行数据库迁移脚本
mysql -u root jobfirst < scripts/migrate_auth_system.sql
```

### 2. 编译统一认证服务

```bash
# 进入统一认证服务目录
cd backend/cmd/unified-auth

# 编译服务
go build -o unified-auth main.go

# 使脚本可执行
chmod +x ../../scripts/test_unified_auth.sh
```

### 3. 启动统一认证服务

```bash
# 设置环境变量
export JWT_SECRET="your-secure-jwt-secret-key"
export DATABASE_URL="root:@tcp(localhost:3306)/jobfirst?charset=utf8mb4&parseTime=True&loc=Local"
export AUTH_SERVICE_PORT="8207"

# 启动服务
./unified-auth
```

### 4. 验证部署

```bash
# 运行测试脚本
./scripts/test_unified_auth.sh
```

## 🔧 配置说明

### 环境变量
- `JWT_SECRET`: JWT签名密钥（必需）
- `DATABASE_URL`: 数据库连接字符串（可选，有默认值）
- `AUTH_SERVICE_PORT`: 服务端口（可选，默认8207）

### 数据库表结构

#### users 表
```sql
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    status VARCHAR(20) DEFAULT 'active',
    subscription_type VARCHAR(50) DEFAULT NULL,
    subscription_expiry TIMESTAMP NULL DEFAULT NULL,
    last_login TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

#### permissions 表
```sql
CREATE TABLE permissions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### role_permissions 表
```sql
CREATE TABLE role_permissions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    role VARCHAR(20) NOT NULL,
    permission VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_role_permission (role, permission)
);
```

## 🚀 API 端点

### 认证相关
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/validate` - JWT验证

### 权限相关
- `GET /api/v1/auth/permission` - 权限检查
- `POST /api/v1/auth/access` - 访问验证
- `GET /api/v1/auth/roles` - 获取角色列表
- `GET /api/v1/auth/permissions` - 获取权限列表

### 用户相关
- `GET /api/v1/auth/user` - 获取用户信息

### 日志相关
- `POST /api/v1/auth/log` - 访问日志记录

### 系统相关
- `GET /health` - 健康检查

## 🔒 安全特性

### JWT Token
- 使用HS256算法签名
- 24小时过期时间
- 包含用户ID、角色、权限信息
- 支持token验证和刷新

### 密码安全
- 使用bcrypt加密存储
- 默认成本因子为10
- 支持密码验证

### 访问控制
- 基于角色的访问控制(RBAC)
- 细粒度权限管理
- 支持权限继承

### 审计日志
- 记录所有访问日志
- 包含IP地址和User-Agent
- 支持结果追踪

## 🧪 测试验证

### 自动化测试
```bash
# 运行完整测试套件
./scripts/test_unified_auth.sh
```

### 手动测试
```bash
# 1. 健康检查
curl http://localhost:8207/health

# 2. 用户登录
curl -X POST http://localhost:8207/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 3. JWT验证
curl -X POST http://localhost:8207/api/v1/auth/validate \
  -H "Content-Type: application/json" \
  -d '{"token":"YOUR_JWT_TOKEN"}'

# 4. 权限检查
curl "http://localhost:8207/api/v1/auth/permission?user_id=1&permission=admin:users"
```

## 🔄 迁移策略

### 从现有系统迁移
1. **备份现有数据**: 执行迁移脚本会自动备份
2. **逐步迁移**: 可以并行运行新旧系统
3. **数据验证**: 使用测试脚本验证数据完整性
4. **切换服务**: 更新其他服务的认证端点

### 回滚计划
1. **保留备份表**: 迁移脚本创建了`users_backup`表
2. **快速回滚**: 可以快速恢复到原始表结构
3. **服务回滚**: 停止新服务，启动旧服务

## 📊 监控和维护

### 健康检查
- 定期检查`/health`端点
- 监控数据库连接状态
- 检查JWT token有效性

### 性能监控
- 监控API响应时间
- 检查数据库查询性能
- 跟踪访问日志量

### 安全监控
- 监控异常登录尝试
- 检查权限提升尝试
- 跟踪敏感操作日志

## 🆘 故障排除

### 常见问题

#### 1. 数据库连接失败
```bash
# 检查数据库服务状态
brew services list | grep mysql

# 检查连接字符串
echo $DATABASE_URL
```

#### 2. JWT验证失败
```bash
# 检查JWT密钥
echo $JWT_SECRET

# 验证token格式
echo "YOUR_TOKEN" | base64 -d
```

#### 3. 权限检查失败
```bash
# 检查权限表数据
mysql -u root -e "USE jobfirst; SELECT * FROM role_permissions WHERE role='admin';"

# 检查用户角色
mysql -u root -e "USE jobfirst; SELECT id, username, role FROM users WHERE id=1;"
```

### 日志分析
```bash
# 查看访问日志
mysql -u root -e "USE jobfirst; SELECT * FROM access_logs ORDER BY created_at DESC LIMIT 10;"

# 分析失败登录
mysql -u root -e "USE jobfirst; SELECT * FROM access_logs WHERE result='failed' ORDER BY created_at DESC;"
```

## 📈 性能优化

### 数据库优化
- 添加适当的索引
- 定期清理访问日志
- 使用连接池

### 缓存策略
- 缓存用户权限信息
- 缓存角色配置
- 使用Redis存储会话

### 负载均衡
- 支持多实例部署
- 使用Consul进行服务发现
- 实现健康检查

## 🔮 未来扩展

### 功能扩展
- 支持OAuth2集成
- 添加多因素认证
- 实现单点登录(SSO)

### 安全增强
- 添加IP白名单
- 实现设备管理
- 支持生物识别

### 管理界面
- 开发Web管理界面
- 添加用户管理功能
- 实现权限可视化

---

## 🔗 Dev-Team集成

### 与Dev-Team服务集成

统一认证系统与Dev-Team服务完美集成，提供：

#### 角色映射
- **Dev-Team角色** → **统一认证角色**
- `super_admin` → `super_admin` (Level 4)
- `system_admin` → `admin` (Level 3)  
- `dev_lead` → `admin` (Level 3)
- `frontend_dev` → `user` (Level 2)
- `backend_dev` → `user` (Level 2)
- `qa_engineer` → `user` (Level 2)
- `guest` → `guest` (Level 1)

#### 集成部署
```bash
# 部署统一认证与Dev-Team集成
chmod +x scripts/deploy_unified_auth_dev_team.sh
./scripts/deploy_unified_auth_dev_team.sh
```

#### 验证集成
```bash
# 测试Dev-Team服务认证
curl -X GET http://localhost:8088/api/v1/dev-team/admin/members \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

详细集成方案请参考: `UNIFIED_AUTH_DEV_TEAM_INTEGRATION_PLAN.md`

## 📞 支持

如有问题，请参考：
1. 测试脚本输出
2. 服务日志
3. 数据库状态
4. 网络连接
5. Dev-Team集成状态

**默认管理员账户**:
- 用户名: `admin`
- 密码: `admin123`
- 角色: `super_admin`
- Dev-Team角色: `super_admin`
