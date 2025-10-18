# GitHub Secrets 配置指南

**仓库**: https://github.com/xiajason/jobfirst-future  
**配置时间**: 2025-10-18

## 🔐 必需的Secrets配置

推送成功后，请立即配置以下Secrets以启用自动部署。

### 访问Secrets配置页面

1. 访问: https://github.com/xiajason/jobfirst-future/settings/secrets/actions
2. 点击 **"New repository secret"**
3. 按照下表添加每个Secret

### Secrets列表

| Secret名称 | 值 | 说明 |
|-----------|-----|------|
| `ALIBABA_SERVER_IP` | `47.115.168.107` | 阿里云服务器IP地址 |
| `ALIBABA_SERVER_USER` | `root` | SSH登录用户名 |
| `ALIBABA_SSH_PRIVATE_KEY` | [见下方] | SSH私钥内容 |
| `ALIBABA_DEPLOY_PATH` | `/opt/services` | 部署目录路径 |

### 获取SSH私钥

在本地终端执行：

```bash
cat ~/.ssh/cross_cloud_key
```

**复制完整输出**，包括：
- `-----BEGIN OPENSSH PRIVATE KEY-----`
- 中间的密钥内容
- `-----END OPENSSH PRIVATE KEY-----`

## 📝 配置步骤详解

### Secret 1: ALIBABA_SERVER_IP

1. Name: `ALIBABA_SERVER_IP`
2. Secret: `47.115.168.107`
3. 点击 **"Add secret"**

### Secret 2: ALIBABA_SERVER_USER

1. Name: `ALIBABA_SERVER_USER`
2. Secret: `root`
3. 点击 **"Add secret"**

### Secret 3: ALIBABA_SSH_PRIVATE_KEY

1. Name: `ALIBABA_SSH_PRIVATE_KEY`
2. Secret: 执行 `cat ~/.ssh/cross_cloud_key` 的完整输出
3. 点击 **"Add secret"**

⚠️ **重要**: 
- 必须包含开头和结尾的标记行
- 不要有额外的空格或换行
- 保持原始格式

### Secret 4: ALIBABA_DEPLOY_PATH

1. Name: `ALIBABA_DEPLOY_PATH`
2. Secret: `/opt/services`
3. 点击 **"Add secret"**

## ✅ 验证配置

配置完成后：

1. 访问: https://github.com/xiajason/jobfirst-future/settings/secrets/actions
2. 确认看到4个Secrets:
   - ✅ ALIBABA_SERVER_IP
   - ✅ ALIBABA_SERVER_USER
   - ✅ ALIBABA_SSH_PRIVATE_KEY
   - ✅ ALIBABA_DEPLOY_PATH

## 🚀 触发部署

### 方式1: 推送代码触发（自动）

代码已推送，但由于Secrets未配置，首次部署会失败。配置完Secrets后：

```bash
# 创建一个小的提交来重新触发
cd /Users/szjason72/szbolent/LoomaCRM/zervigo_future_CICD
echo "# Deployment trigger" >> README.md
git add README.md
git commit -m "trigger: first deployment with secrets configured"
git push origin main
```

### 方式2: 手动触发

1. 访问: https://github.com/xiajason/jobfirst-future/actions
2. 选择 **"Zervigo Future 微服务部署流水线"**
3. 点击 **"Run workflow"**
4. 选择环境: **production**
5. 点击 **"Run workflow"**

## 📊 监控部署

### 查看部署进度

访问: https://github.com/xiajason/jobfirst-future/actions

您会看到以下部署阶段：
1. 🔍 检测代码变更
2. 🔨 构建Go微服务
3. 🚀 部署到阿里云
4. ✅ 验证部署
5. 📢 部署通知

### 部署成功标志

所有阶段都显示 ✅ 绿色勾号，表示：
- 10个微服务构建成功
- 按时序部署完成
- 健康检查全部通过

## 🆘 如果遇到问题

### 问题1: 部署失败 - SSH连接失败

**原因**: SSH私钥配置错误

**解决**:
1. 重新获取私钥: `cat ~/.ssh/cross_cloud_key`
2. 检查是否包含完整的开头和结尾标记
3. 重新配置 `ALIBABA_SSH_PRIVATE_KEY`

### 问题2: 部署失败 - 服务启动失败

**原因**: 数据库连接或依赖问题

**解决**:
1. SSH到服务器检查日志
2. 查看服务器现状报告
3. 检查数据库容器状态

---

**下一步**: 配置完Secrets后，重新触发部署！
