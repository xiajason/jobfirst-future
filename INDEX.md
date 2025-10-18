# Zervigo Future CI/CD 文档索引

快速查找您需要的文档和资源。

## 📚 核心文档

### 🚀 快速开始
- **[README.md](README.md)** - 项目说明和快速开始指南
- **[INSTALLATION.md](INSTALLATION.md)** - 详细的安装配置步骤
- **[SUMMARY.md](SUMMARY.md)** - 项目完成情况和总结

### 📖 使用指南
- **[部署指南](docs/ALIBABA_MICROSERVICE_DEPLOYMENT_GUIDE.md)** - 完整的手动部署步骤
- **[快速参考](docs/QUICK_REFERENCE_ALIBABA_DEPLOYMENT.md)** - 常用命令速查手册
- **[实现总结](docs/ZERVIGO_FUTURE_CICD_IMPLEMENTATION_SUMMARY.md)** - CI/CD架构和实现细节

### 📝 版本信息
- **[CHANGELOG.md](CHANGELOG.md)** - 版本变更历史
- **[INDEX.md](INDEX.md)** - 本文档索引

## 🛠️ 工具和脚本

### GitHub Actions
- **[zervigo-future-deploy.yml](workflows/zervigo-future-deploy.yml)** - CI/CD自动部署流水线

### 部署脚本
| 脚本 | 功能 | 使用场景 |
|------|------|---------|
| [setup-cicd.sh](scripts/setup-cicd.sh) | CI/CD快速安装 | 首次安装时使用 |
| [quick-deploy.sh](scripts/quick-deploy.sh) | 快速部署 | 需要快速部署所有服务 |
| [microservice-deployment-manager.sh](scripts/microservice-deployment-manager.sh) | 微服务管理 | 启动、停止、重启服务 |
| [verify-microservice-deployment.sh](scripts/verify-microservice-deployment.sh) | 部署验证 | 验证部署状态 |

## 🎯 按使用场景查找

### 场景1: 首次安装CI/CD
1. 阅读 [README.md](README.md) 了解项目概况
2. 按照 [INSTALLATION.md](INSTALLATION.md) 完成安装
3. 运行 `./scripts/setup-cicd.sh` 快速配置
4. 查看 [部署指南](docs/ALIBABA_MICROSERVICE_DEPLOYMENT_GUIDE.md) 了解详细步骤

### 场景2: 日常开发部署
1. 修改代码后提交: `git push origin develop`
2. 查看GitHub Actions执行情况
3. 如需验证: `./scripts/verify-microservice-deployment.sh`
4. 遇到问题查看 [故障排除](#故障排除)

### 场景3: 生产环境发布
1. 确认测试环境正常
2. 合并到main分支: `git merge develop && git push origin main`
3. 监控部署过程
4. 运行健康检查验证

### 场景4: 紧急修复
1. 使用快速部署: `./scripts/quick-deploy.sh`
2. 或SSH到服务器手动重启服务
3. 查看 [快速参考](docs/QUICK_REFERENCE_ALIBABA_DEPLOYMENT.md)

### 场景5: 服务管理
1. 查看状态: `./scripts/microservice-deployment-manager.sh status`
2. 重启服务: `./scripts/microservice-deployment-manager.sh restart`
3. 查看日志: `ssh root@47.115.168.107 'tail -f /opt/services/logs/*.log'`

## 📖 按主题查找

### 架构设计
- [实现总结 - 架构设计](docs/ZERVIGO_FUTURE_CICD_IMPLEMENTATION_SUMMARY.md#架构设计)
- [部署指南 - 微服务架构](docs/ALIBABA_MICROSERVICE_DEPLOYMENT_GUIDE.md#微服务架构)

### 部署流程
- [README - 部署流程](README.md#部署流程)
- [部署指南 - 详细步骤](docs/ALIBABA_MICROSERVICE_DEPLOYMENT_GUIDE.md#详细部署步骤)
- [实现总结 - 部署流程](docs/ZERVIGO_FUTURE_CICD_IMPLEMENTATION_SUMMARY.md#部署流程)

### 健康检查
- [快速参考 - 健康检查](docs/QUICK_REFERENCE_ALIBABA_DEPLOYMENT.md#快速健康检查)
- [部署指南 - 服务监控](docs/ALIBABA_MICROSERVICE_DEPLOYMENT_GUIDE.md#服务监控)

### 故障排除
- [部署指南 - 故障排除](docs/ALIBABA_MICROSERVICE_DEPLOYMENT_GUIDE.md#故障排除)
- [安装指南 - 故障排除](INSTALLATION.md#故障排除)
- [快速参考 - 故障排除](docs/QUICK_REFERENCE_ALIBABA_DEPLOYMENT.md#故障排除)

### 安全配置
- [部署指南 - 安全配置](docs/ALIBABA_MICROSERVICE_DEPLOYMENT_GUIDE.md#安全配置)
- [快速参考 - 数据库密码](docs/QUICK_REFERENCE_ALIBABA_DEPLOYMENT.md#数据库密码)

### 最佳实践
- [部署指南 - 最佳实践](docs/ALIBABA_MICROSERVICE_DEPLOYMENT_GUIDE.md#最佳实践)
- [总结 - 最佳实践](SUMMARY.md#最佳实践)

## 🔍 快速查询

### 端口映射
| 端口 | 服务 | 状态 |
|------|------|------|
| 3306 | MySQL | ✅ 已部署 |
| 5432 | PostgreSQL | ✅ 已部署 |
| 6379 | Redis | ✅ 已部署 |
| 27017 | MongoDB | ✅ 已部署 |
| 8080 | API Gateway | 待部署 |
| 8081 | User Service | 待部署 |
| 8082 | Resume Service | 待部署 |
| 8083 | Company Service | 待部署 |
| 8084 | Notification Service | 待部署 |
| 8085 | Template Service | 待部署 |
| 8086 | Statistics Service | 待部署 |
| 8087 | Banner Service | 待部署 |
| 8088 | Dev Team Service | 待部署 |
| 8089 | Job Service | 待部署 |
| 8100 | AI Service | ✅ 已部署 |

详细说明: [快速参考 - 端口映射](docs/QUICK_REFERENCE_ALIBABA_DEPLOYMENT.md#服务端口映射)

### 常用命令
```bash
# 快速部署
./scripts/quick-deploy.sh

# 服务管理
./scripts/microservice-deployment-manager.sh [deploy-all|stop|restart|status]

# 部署验证
./scripts/verify-microservice-deployment.sh

# 查看日志
ssh root@47.115.168.107 'tail -f /opt/services/logs/api-gateway.log'

# 健康检查
curl http://47.115.168.107:8080/health
```

详细命令: [快速参考 - 常用命令](docs/QUICK_REFERENCE_ALIBABA_DEPLOYMENT.md#常用命令)

### GitHub Secrets配置
| Secret名称 | 值 | 必需 |
|-----------|-----|------|
| ALIBABA_SERVER_IP | 47.115.168.107 | ✅ |
| ALIBABA_SERVER_USER | root | ✅ |
| ALIBABA_SSH_PRIVATE_KEY | SSH私钥内容 | ✅ |
| ALIBABA_DEPLOY_PATH | /opt/services | ⭕ |

配置步骤: [安装指南 - GitHub Secrets](INSTALLATION.md#配置github-secrets)

### 服务器信息
- **IP地址**: 47.115.168.107
- **SSH用户**: root
- **SSH密钥**: ~/.ssh/cross_cloud_key
- **部署目录**: /opt/services
- **日志目录**: /opt/services/logs
- **配置目录**: /opt/services/configs

详细信息: [快速参考](docs/QUICK_REFERENCE_ALIBABA_DEPLOYMENT.md)

## 📚 相关资源

### 外部链接
- GitHub Actions文档: https://docs.github.com/actions
- Go语言文档: https://golang.org/doc/
- 阿里云文档: https://www.aliyun.com/

### 项目相关
- 服务器现状报告: [ALIYUN_SERVER_STATUS_REPORT_20251018.md](../ALIYUN_SERVER_STATUS_REPORT_20251018.md)
- GitHub仓库: https://github.com/your-org/your-repo
- GitHub Actions: https://github.com/your-org/your-repo/actions

## 🆘 获取帮助

### 文档查询流程
1. 先查看 [README.md](README.md) 了解基本概念
2. 根据使用场景查找对应文档
3. 遇到问题查看故障排除章节
4. 使用快速参考查询命令

### 常见问题
- **如何开始?** → [README.md](README.md)
- **如何安装?** → [INSTALLATION.md](INSTALLATION.md)
- **如何部署?** → [部署指南](docs/ALIBABA_MICROSERVICE_DEPLOYMENT_GUIDE.md)
- **服务启动失败?** → [故障排除](docs/ALIBABA_MICROSERVICE_DEPLOYMENT_GUIDE.md#故障排除)
- **需要快速查询?** → [快速参考](docs/QUICK_REFERENCE_ALIBABA_DEPLOYMENT.md)

## 📝 文档维护

### 文档更新
- 当前版本: 1.0.0
- 最后更新: 2025-10-18
- 维护者: AI Assistant

### 反馈建议
如发现文档问题或有改进建议，请通过以下方式反馈:
- 提交GitHub Issue
- 更新相关文档并提交PR
- 联系项目维护者

---

**提示**: 使用Ctrl+F或Command+F快速搜索关键词
