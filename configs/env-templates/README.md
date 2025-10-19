# 环境配置模板说明

## 📋 目的

提供不同云平台的环境配置模板，确保代码**100%云无关**，可以部署到任何云平台。

## 🌐 支持的云平台

| 模板文件 | 云平台 | 状态 |
|---------|--------|------|
| `aliyun.env` | 阿里云 | ✅ 已部署 |
| `aws.env` | AWS | ⭐ 可立即部署 |
| `azure.env` | Azure | ⭐ 可立即部署 |
| `huawei.env` | 华为云 | 💡 按需创建 |
| `local.env` | 本地开发 | ✅ 开发环境 |

## 🔧 使用方法

### 方式1: 环境变量文件

```bash
# 部署到阿里云
cp configs/env-templates/aliyun.env .env
source .env
./start-all-services.sh

# 部署到AWS  
cp configs/env-templates/aws.env .env
source .env
./start-all-services.sh

# 部署到Azure
cp configs/env-templates/azure.env .env
source .env
./start-all-services.sh
```

### 方式2: Docker Compose

```bash
# 部署到阿里云
docker-compose --env-file configs/env-templates/aliyun.env up -d

# 部署到AWS
docker-compose --env-file configs/env-templates/aws.env up -d
```

### 方式3: Kubernetes ConfigMap

```bash
# 创建ConfigMap
kubectl create configmap jobfirst-config \
  --from-env-file=configs/env-templates/aws.env

# 在Deployment中引用
envFrom:
  - configMapRef:
      name: jobfirst-config
```

## 🔑 关键配置项

### 跨云统一配置（必须一致）

```bash
# JWT密钥 - 所有云环境必须使用相同的密钥
JWT_SECRET=jobfirst-unified-auth-secret-key-2024

# 天翼云认证中心 - 跨云调用
AUTH_CENTER_URL=http://101.33.251.158:8207
```

**这两项确保了跨云认证的正常工作！**

### 云特定配置（每个云不同）

```bash
# 数据库连接（各云不同）
DB_HOST=...
DB_PASSWORD=...

# Redis连接（各云不同）
REDIS_HOST=...

# MinerU服务地址（各云不同）
MINERU_SERVICE_URL=...
```

## 📝 配置模板说明

### aliyun.env - 阿里云
```
- 数据库: 本地Podman容器
- Redis: 本地Podman容器
- MinerU: http://47.115.168.107:8621
- 状态: 已部署运行
```

### aws.env - AWS
```
- 数据库: RDS MySQL
- Redis: ElastiCache
- S3: 文件存储
- 状态: 模板准备就绪
```

### azure.env - Azure
```
- 数据库: Azure Database for MySQL
- Redis: Azure Cache for Redis
- Blob: 文件存储
- 状态: 模板准备就绪
```

## ⚠️ 重要提醒

### 1. 密码安全

**不要在代码库中提交真实密码！**

```bash
# ✅ 好的做法
.env
.env.*
*.env

# 添加到 .gitignore
```

### 2. JWT密钥统一

**所有云环境必须使用相同的JWT密钥！**

这样天翼云生成的Token才能在所有云平台验证。

### 3. 服务发现

不同云平台可能使用不同的服务发现机制：
- Consul（通用）
- AWS Cloud Map
- Azure Service Fabric
- K8s Service Discovery

配置时注意适配。

## 🎯 使用示例

### 部署到阿里云

```bash
# 1. 准备配置
export MINERU_SERVICE_URL=http://47.115.168.107:8621
export DB_PASSWORD=JobFirst2025!MySQL

# 2. 启动服务
./user-service -config configs/jobfirst-core-config.yaml
```

### 部署到AWS

```bash
# 1. 准备配置
export MINERU_SERVICE_URL=http://aws-mineru:8621
export DB_HOST=jobfirst.xxxxx.rds.amazonaws.com
export DB_PASSWORD=AWS_SecurePassword_2025

# 2. 启动服务（代码完全相同！）
./user-service -config configs/jobfirst-core-config.yaml
```

**代码0改动！**

---

**模板版本**: 1.0  
**更新时间**: 2025-10-19  
**支持云平台**: 阿里云, AWS, Azure, 华为云, 私有云

