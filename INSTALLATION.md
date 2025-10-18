# Zervigo Future CI/CD 安装配置指南

本指南将帮助您快速安装和配置Zervigo Future微服务CI/CD系统。

## 📋 前置要求

### 本地开发环境

- Git
- Go 1.23+
- SSH客户端
- 访问阿里云服务器的权限

### 阿里云服务器环境

已预部署的服务（无需重新部署）：
- ✅ MySQL (3306)
- ✅ PostgreSQL (5432)
- ✅ Redis (6379)
- ✅ MongoDB (27017)
- ✅ AI Service (8100)

待部署的服务（通过CI/CD部署）：
- ❌ 10个Go微服务 (8080-8089)

## 🚀 安装步骤

### 步骤1: 克隆CI/CD套件

```bash
# 如果您已经在项目根目录
cd /Users/szjason72/szbolent/LoomaCRM

# CI/CD套件已在 zervigo_future_CICD 目录中
cd zervigo_future_CICD
```

### 步骤2: 配置GitHub Actions

#### 2.1 复制workflow文件

```bash
# 从CI/CD套件目录复制到GitHub Actions目录
mkdir -p ../.github/workflows
cp workflows/zervigo-future-deploy.yml ../.github/workflows/
```

#### 2.2 配置GitHub Secrets

在GitHub仓库设置中添加以下Secrets:

1. 访问GitHub仓库: `https://github.com/your-org/your-repo`
2. 点击 `Settings` -> `Secrets and variables` -> `Actions`
3. 点击 `New repository secret` 添加以下密钥:

| Secret名称 | 值 | 说明 |
|-----------|-----|------|
| `ALIBABA_SERVER_IP` | `47.115.168.107` | 阿里云服务器IP |
| `ALIBABA_SERVER_USER` | `root` | SSH用户名 |
| `ALIBABA_SSH_PRIVATE_KEY` | SSH私钥内容 | SSH私钥（见下方获取方法） |
| `ALIBABA_DEPLOY_PATH` | `/opt/services` | 部署目录（可选） |

#### 2.3 获取SSH私钥

```bash
# 查看现有SSH私钥
cat ~/.ssh/cross_cloud_key

# 复制输出的内容到GitHub Secret: ALIBABA_SSH_PRIVATE_KEY
# 包括 -----BEGIN ... KEY----- 和 -----END ... KEY----- 行
```

### 步骤3: 测试SSH连接

```bash
# 测试SSH连接
ssh -i ~/.ssh/cross_cloud_key root@47.115.168.107 "echo 'SSH连接正常'"

# 如果连接成功，会显示: SSH连接正常
```

### 步骤4: 准备服务器环境

```bash
# SSH到服务器
ssh -i ~/.ssh/cross_cloud_key root@47.115.168.107

# 创建部署目录
mkdir -p /opt/services/{backend/bin,configs,logs,scripts}

# 设置权限
chmod 755 /opt/services
chmod 755 /opt/services/backend
chmod 755 /opt/services/backend/bin
chmod 755 /opt/services/configs
chmod 755 /opt/services/logs
chmod 755 /opt/services/scripts

# 退出服务器
exit
```

### 步骤5: 上传部署脚本（首次部署）

```bash
# 上传部署脚本到服务器
scp -i ~/.ssh/cross_cloud_key scripts/*.sh root@47.115.168.107:/opt/services/scripts/

# 给脚本执行权限
ssh -i ~/.ssh/cross_cloud_key root@47.115.168.107 'chmod +x /opt/services/scripts/*.sh'
```

## 🧪 测试部署

### 方式1: 手动测试部署

```bash
# 1. 本地构建一个服务测试
cd ../zervigo_future/backend
go build -o bin/api-gateway ./cmd/basic-server

# 2. 上传到服务器
scp -i ~/.ssh/cross_cloud_key bin/api-gateway root@47.115.168.107:/opt/services/backend/bin/

# 3. SSH到服务器启动
ssh -i ~/.ssh/cross_cloud_key root@47.115.168.107
cd /opt/services/backend/bin
chmod +x api-gateway
nohup ./api-gateway > ../../logs/api-gateway.log 2>&1 &

# 4. 验证服务
sleep 10
curl http://localhost:8080/health

# 5. 如果成功，停止测试服务
pkill -f api-gateway
exit
```

### 方式2: 测试GitHub Actions（推荐）

```bash
# 1. 提交一个小的变更来触发CI/CD
cd /Users/szjason72/szbolent/LoomaCRM
echo "# Test deployment" >> zervigo_future/backend/README.md
git add zervigo_future/backend/README.md
git commit -m "test: trigger CI/CD deployment"

# 2. 推送到develop分支测试
git push origin develop

# 3. 在GitHub上查看Actions执行情况
# https://github.com/your-org/your-repo/actions

# 4. 如果测试成功，推送到main分支
git checkout main
git merge develop
git push origin main
```

## 📊 验证安装

### 检查GitHub Actions

1. 访问: `https://github.com/your-org/your-repo/actions`
2. 查看 `Zervigo Future 微服务部署流水线` workflow
3. 确认workflow文件存在且配置正确

### 检查Secrets

1. 访问: `https://github.com/your-org/your-repo/settings/secrets/actions`
2. 确认所有必需的Secrets已配置:
   - ✅ ALIBABA_SERVER_IP
   - ✅ ALIBABA_SERVER_USER
   - ✅ ALIBABA_SSH_PRIVATE_KEY
   - ✅ ALIBABA_DEPLOY_PATH (可选)

### 检查服务器环境

```bash
# SSH到服务器
ssh -i ~/.ssh/cross_cloud_key root@47.115.168.107

# 检查目录结构
ls -la /opt/services/
# 应该看到: backend/, configs/, logs/, scripts/

# 检查数据库容器
podman ps | grep migration
# 应该看到: migration-mysql, migration-postgres, migration-redis, migration-mongodb

# 检查AI服务
ps aux | grep ai_service
curl http://localhost:8100/health
# 应该返回健康状态

# 退出
exit
```

## 🔧 配置微调

### 修改部署路径

如需修改部署路径，更新GitHub Secret:

```bash
# 默认: /opt/services
# 如需修改: 更新 ALIBABA_DEPLOY_PATH Secret
```

### 修改环境分支映射

编辑 `.github/workflows/zervigo-future-deploy.yml`:

```yaml
- name: 确定部署环境
  id: env
  run: |
    if [[ "${{ github.ref }}" == "refs/heads/main" ]]; then
      echo "environment=production" >> $GITHUB_OUTPUT
    elif [[ "${{ github.ref }}" == "refs/heads/develop" ]]; then
      echo "environment=staging" >> $GITHUB_OUTPUT
    else
      echo "environment=development" >> $GITHUB_OUTPUT
    fi
```

### 修改服务端口

如需修改服务端口，编辑workflow文件中的环境变量:

```yaml
env:
  API_GATEWAY_PORT: 8080      # 修改为您需要的端口
  USER_SERVICE_PORT: 8081
  # ... 其他端口
```

## 🚨 故障排除

### 问题1: GitHub Actions无法连接服务器

**错误**: `Permission denied (publickey)`

**解决方案**:
```bash
# 1. 确认SSH密钥正确
cat ~/.ssh/cross_cloud_key

# 2. 确认密钥已添加到服务器
ssh-copy-id -i ~/.ssh/cross_cloud_key.pub root@47.115.168.107

# 3. 测试连接
ssh -i ~/.ssh/cross_cloud_key root@47.115.168.107 "echo 'OK'"

# 4. 重新复制私钥内容到GitHub Secret
```

### 问题2: 服务构建失败

**错误**: `go: module not found`

**解决方案**:
```bash
# 1. 确认Go版本
go version  # 应该是 1.23+

# 2. 更新依赖
cd zervigo_future/backend
go mod tidy
go mod download

# 3. 重新提交
git add go.mod go.sum
git commit -m "fix: update go modules"
git push
```

### 问题3: 服务启动失败

**错误**: 服务健康检查失败

**解决方案**:
```bash
# SSH到服务器
ssh -i ~/.ssh/cross_cloud_key root@47.115.168.107

# 查看服务日志
tail -100 /opt/services/logs/api-gateway.log

# 检查端口占用
netstat -tlnp | grep 8080

# 检查数据库连接
podman exec migration-mysql mysql -uroot -pJobFirst2025!MySQL -e "SELECT 1;"

# 手动重启服务
cd /opt/services/backend/bin
pkill -f api-gateway
nohup ./api-gateway > ../../logs/api-gateway.log 2>&1 &
```

## ✅ 安装完成检查清单

安装完成后，请确认以下项目:

- [ ] GitHub Actions workflow文件已复制到 `.github/workflows/`
- [ ] 所有必需的GitHub Secrets已配置
- [ ] SSH连接测试成功
- [ ] 服务器目录结构已创建
- [ ] 部署脚本已上传并有执行权限
- [ ] 数据库容器运行正常
- [ ] AI服务运行正常
- [ ] 测试部署成功（至少一个服务）

## 🎉 后续步骤

安装完成后，您可以:

1. **阅读使用指南**: 查看 [README.md](README.md)
2. **查看部署指南**: 查看 [docs/ALIBABA_MICROSERVICE_DEPLOYMENT_GUIDE.md](docs/ALIBABA_MICROSERVICE_DEPLOYMENT_GUIDE.md)
3. **开始自动部署**: `git push origin main`
4. **监控部署状态**: 访问GitHub Actions页面

## 📞 获取帮助

如遇到问题:
1. 查看 [故障排除指南](docs/ALIBABA_MICROSERVICE_DEPLOYMENT_GUIDE.md#故障排除)
2. 查看 [服务器现状报告](../ALIYUN_SERVER_STATUS_REPORT_20251018.md)
3. 查看服务器日志: `/opt/services/logs/`

---

**最后更新**: 2025-10-18  
**维护者**: AI Assistant
