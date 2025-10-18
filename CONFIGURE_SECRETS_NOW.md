# 🔐 立即配置GitHub Secrets

代码已推送到: https://github.com/xiajason/jobfirst-future

## ⚡ 快速配置步骤

### 1. 访问Secrets配置页面

**直接链接**: https://github.com/xiajason/jobfirst-future/settings/secrets/actions

### 2. 添加4个Secrets

点击 **"New repository secret"**，依次添加：

#### Secret 1: ALIBABA_SERVER_IP
- Name: `ALIBABA_SERVER_IP`
- Value: `47.115.168.107`

#### Secret 2: ALIBABA_SERVER_USER
- Name: `ALIBABA_SERVER_USER`
- Value: `root`

#### Secret 3: ALIBABA_SSH_PRIVATE_KEY
- Name: `ALIBABA_SSH_PRIVATE_KEY`
- Value: [执行下方命令获取]

```bash
cat ~/.ssh/cross_cloud_key
```

**完整复制输出**，包括：
```
-----BEGIN OPENSSH PRIVATE KEY-----
[私钥内容]
-----END OPENSSH PRIVATE KEY-----
```

#### Secret 4: ALIBABA_DEPLOY_PATH
- Name: `ALIBABA_DEPLOY_PATH`
- Value: `/opt/services`

### 3. 验证配置

刷新页面，确认看到4个Secrets:
- ✅ ALIBABA_SERVER_IP
- ✅ ALIBABA_SERVER_USER  
- ✅ ALIBABA_SSH_PRIVATE_KEY
- ✅ ALIBABA_DEPLOY_PATH

## 🚀 触发首次部署

配置完成后，有两种方式触发部署：

### 方式1: 手动触发（推荐）

1. 访问: https://github.com/xiajason/jobfirst-future/actions
2. 选择 **"Zervigo Future 微服务部署流水线"**
3. 点击 **"Run workflow"** 按钮
4. 选择环境: **production**
5. 点击绿色的 **"Run workflow"** 按钮

### 方式2: 推送代码触发

```bash
cd /Users/szjason72/szbolent/LoomaCRM/zervigo_future_CICD
echo "# First deployment" >> README.md
git add README.md
git commit -m "trigger: first deployment with secrets configured"
git push origin main
```

## 📊 监控部署进度

访问: https://github.com/xiajason/jobfirst-future/actions

您会看到部署流程：

```
🔍 检测代码变更
    ↓
🔨 构建Go微服务 (10个服务)
    ↓
🚀 部署到阿里云
   ├── 阶段1: 网关层 (8080)
   ├── 阶段2: 认证层 (8081)
   ├── 阶段3: 核心业务层 (8082-8083)
   ├── 阶段4: 支撑服务层 (8084-8087)
   └── 阶段5: 管理服务层 (8088-8089)
    ↓
✅ 验证部署
    ↓
📢 部署通知
```

预计部署时间: **5-7分钟**

## 🎯 部署成功后

访问服务器验证：

```bash
# 检查所有微服务
for port in 8080 8081 8082 8083 8084 8085 8086 8087 8088 8089; do
    curl -f http://47.115.168.107:$port/health && echo "✅ Port $port OK" || echo "❌ Port $port Failed"
done
```

---

**下一步**: 立即配置Secrets，然后触发首次部署！
