# 🤖 AI Service - JobFirst AI智能服务

## 📋 服务概述

AI Service是JobFirst平台的AI智能服务层，基于Python Sanic框架开发，提供AI驱动的职位匹配、简历分析、企业画像等功能。

## 🔧 技术栈

- **框架**: Sanic (异步Web框架)
- **数据库**: MySQL, PostgreSQL, Redis
- **AI模型**: DeepSeek API
- **认证**: Zervigo统一认证
- **端口**: 8100

## 📁 主要文件说明

### 核心服务文件
- `ai_service.py` - 主服务文件（推荐使用）
- `ai_service_with_zervigo.py` - 集成Zervigo认证的版本
- `ai_service_simple.py` - 简化版本
- `ai_service_deepseek.py` - DeepSeek集成版本
- `ai_service_containerized.py` - 容器化版本

### 功能模块
- `enhanced_job_matching_engine.py` - 职位匹配引擎
- `resume_analyzer.py` - 简历分析器
- `avatar_profile_engine.py` - 用户画像引擎
- `three_layer_avatar_chat.py` - 三层对话系统
- `consent_manager.py` - 用户同意管理
- `data_anonymizer.py` - 数据匿名化
- `privacy_enhanced_data_access.py` - 隐私增强数据访问

### 认证和权限
- `unified_auth_client.py` - 统一认证客户端
- `zervigo_auth_middleware.py` - Zervigo认证中间件
- `get_user_permissions.py` - 权限获取

## 🚀 本地开发

### 1. 安装依赖
```bash
# 创建虚拟环境
python3 -m venv venv

# 激活虚拟环境
source venv/bin/activate  # Linux/macOS
# venv\Scripts\activate  # Windows

# 安装依赖
pip install -r requirements.txt
```

### 2. 配置环境变量
```bash
# 复制环境变量示例
cp env.example .env

# 编辑.env文件，填入正确的配置
vi .env
```

### 3. 启动服务
```bash
# 方式1: 使用启动脚本
chmod +x start_with_env.sh
./start_with_env.sh

# 方式2: 直接启动
source .env
python ai_service.py
```

### 4. 测试服务
```bash
# 健康检查
curl http://localhost:8100/health

# 测试AI接口
curl -X POST http://localhost:8100/api/v1/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello"}'
```

## 🌐 生产部署

### 环境要求
- Python 3.11+
- 内存: 至少2GB (AI模型已禁用重量级依赖)
- 网络: 需要访问DeepSeek API

### 部署步骤
1. 上传代码到服务器
2. 创建虚拟环境并安装依赖
3. 配置.env文件
4. 使用systemd或supervisor管理服务

### Systemd服务配置示例
```ini
[Unit]
Description=JobFirst AI Service
After=network.target

[Service]
Type=simple
User=jobfirst
WorkingDirectory=/opt/services/ai-service
EnvironmentFile=/opt/services/ai-service/.env
ExecStart=/opt/services/ai-service/venv/bin/python ai_service.py
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

## 📊 CI/CD部署

AI Service已集成到统一的CI/CD流程中：

### 部署流程
1. 代码推送到main分支
2. GitHub Actions自动触发
3. 打包AI服务代码
4. 上传到阿里云服务器
5. 自动安装依赖
6. 自动重启服务
7. 健康检查验证

### 部署配置
详见: `.github/workflows/zervigo-future-deploy.yml`

## 🔐 安全说明

### API密钥管理
- DeepSeek API密钥存储在`.env`文件中
- **重要**: `.env`文件不应提交到Git
- 生产环境使用环境变量或密钥管理服务

### 数据隐私
- 支持数据匿名化
- 用户同意管理
- 隐私增强数据访问

## 📝 API文档

### 健康检查
```
GET /health
Response: {"status": "healthy", "service": "ai-service"}
```

### AI对话
```
POST /api/v1/ai/chat
Body: {"message": "用户消息"}
Response: {"response": "AI回复"}
```

## 🐛 故障排查

### 服务无法启动
1. 检查Python版本: `python --version`
2. 检查依赖安装: `pip list`
3. 检查端口占用: `netstat -tlnp | grep 8100`
4. 查看日志: `tail -f ../logs/ai-service.log`

### 数据库连接失败
1. 检查.env配置
2. 验证数据库服务运行: `systemctl status mysql`
3. 测试数据库连接: `mysql -h localhost -u root -p`

### AI API调用失败
1. 检查DeepSeek API密钥
2. 验证网络连接
3. 查看API调用日志

## 📞 联系方式

如有问题，请联系开发团队或查看项目文档。

## 📜 许可证

Copyright © 2025 JobFirst. All rights reserved.

