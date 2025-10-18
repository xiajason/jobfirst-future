# ✅ 部署操作清单

**仓库**: https://github.com/xiajason/jobfirst-future  
**状态**: 代码已推送，等待配置Secrets

## 📋 立即执行清单

### ☑️ 第1步: 配置GitHub Secrets（5分钟）

访问: https://github.com/xiajason/jobfirst-future/settings/secrets/actions

添加以下4个Secrets（点击 "New repository secret"）:

#### 1. ALIBABA_SERVER_IP
```
Name: ALIBABA_SERVER_IP
Value: 47.115.168.107
```

#### 2. ALIBABA_SERVER_USER
```
Name: ALIBABA_SERVER_USER
Value: root
```

#### 3. ALIBABA_SSH_PRIVATE_KEY
```
Name: ALIBABA_SSH_PRIVATE_KEY
Value: [SSH私钥已在终端显示，完整复制]
```

提示: 私钥应该以这样的格式开始和结束:
```
-----BEGIN OPENSSH PRIVATE KEY-----
...
-----END OPENSSH PRIVATE KEY-----
```

#### 4. ALIBABA_DEPLOY_PATH
```
Name: ALIBABA_DEPLOY_PATH
Value: /opt/services
```

### ☑️ 第2步: 验证Secrets配置（1分钟）

刷新页面，确认看到4个Secrets:
- ✅ ALIBABA_SERVER_IP
- ✅ ALIBABA_SERVER_USER
- ✅ ALIBABA_SSH_PRIVATE_KEY
- ✅ ALIBABA_DEPLOY_PATH

### ☑️ 第3步: 触发首次部署（1分钟）

**方式1: 手动触发（推荐）**

1. 访问: https://github.com/xiajason/jobfirst-future/actions
2. 点击左侧 **"Zervigo Future 微服务部署流水线"**
3. 点击右侧 **"Run workflow"** 按钮
4. 选择:
   - Branch: `main`
   - 部署环境: `production`
5. 点击绿色 **"Run workflow"** 按钮

**方式2: 推送代码触发**

```bash
cd /Users/szjason72/szbolent/LoomaCRM/zervigo_future_CICD
echo "# Trigger first deployment" >> README.md
git add README.md
git commit -m "trigger: first automated deployment"
git push origin main
```

### ☑️ 第4步: 监控部署进度（5-7分钟）

访问: https://github.com/xiajason/jobfirst-future/actions

查看部署流程：

```
🔍 检测代码变更 (30秒)
    ↓
🔨 构建Go微服务 (2-3分钟)
    - 编译10个Go服务
    - 上传构建产物
    ↓
🚀 部署到阿里云 (2-3分钟)
    - 上传服务文件
    - 阶段1: 网关层 (8080)
    - 阶段2: 认证层 (8081)
    - 阶段3: 核心业务层 (8082-8083)
    - 阶段4: 支撑服务层 (8084-8087)
    - 阶段5: 管理服务层 (8088-8089)
    ↓
✅ 验证部署 (1分钟)
    - 健康检查所有服务
    - 验证数据库连接
    ↓
📢 部署通知
```

### ☑️ 第5步: 验证部署成功（2分钟）

部署完成后，执行健康检查：

```bash
# 检查所有微服务
for port in 8080 8081 8082 8083 8084 8085 8086 8087 8088 8089; do
    curl -f http://47.115.168.107:$port/health && echo "✅ Port $port OK" || echo "❌ Port $port Failed"
done

# 检查AI服务（预部署）
curl http://47.115.168.107:8100/health
```

或SSH到服务器验证：

```bash
ssh root@47.115.168.107 "cd /opt/services && ps aux | grep -E '(api-gateway|user-service|resume-service|company-service|notification-service|template-service|statistics-service|banner-service|dev-team-service|job-service)'"
```

## 🎯 预期结果

### 部署成功标志

GitHub Actions页面显示:
- ✅ 所有步骤绿色勾号
- ✅ 构建10个微服务成功
- ✅ 部署到阿里云成功
- ✅ 验证部署成功

### 服务器状态

所有10个微服务健康检查通过：
- ✅ API Gateway (8080)
- ✅ User Service (8081)
- ✅ Resume Service (8082)
- ✅ Company Service (8083)
- ✅ Notification Service (8084)
- ✅ Template Service (8085)
- ✅ Statistics Service (8086)
- ✅ Banner Service (8087)
- ✅ Dev Team Service (8088)
- ✅ Job Service (8089)

## 🆘 如果遇到问题

### 问题1: GitHub Actions找不到workflow

**解决**: 
- 检查 `.github/workflows/zervigo-future-deploy.yml` 文件是否存在
- 刷新GitHub Actions页面

### 问题2: 部署失败 - SSH连接失败

**解决**:
1. 检查 `ALIBABA_SSH_PRIVATE_KEY` 是否包含完整内容
2. 确认私钥格式正确（包含开头和结尾标记）
3. 测试本地SSH连接: `ssh root@47.115.168.107`

### 问题3: 服务启动失败

**解决**:
1. 查看GitHub Actions日志
2. SSH到服务器查看日志: `tail -f /opt/services/logs/*.log`
3. 检查数据库容器状态: `podman ps | grep migration`

## 📊 部署时间估算

- **配置Secrets**: 5分钟
- **触发部署**: 1分钟
- **等待部署**: 5-7分钟
- **验证部署**: 2分钟
- **总计**: 约15分钟

---

**开始时间**: ___________  
**完成时间**: ___________  
**实际耗时**: ___________

**准备开始！** 🚀
