# Git提交指南

本文档说明哪些文件应该提交到GitHub仓库。

## ✅ 应该提交的文件

### 核心文件（必需）
```bash
git add zervigo_future_CICD/.gitignore
git add zervigo_future_CICD/README.md
git add zervigo_future_CICD/INSTALLATION.md
```

### GitHub Actions工作流（必需）
```bash
git add zervigo_future_CICD/workflows/zervigo-future-deploy.yml
```

### 部署脚本（必需）
```bash
git add zervigo_future_CICD/scripts/setup-cicd.sh
git add zervigo_future_CICD/scripts/quick-deploy.sh
git add zervigo_future_CICD/scripts/microservice-deployment-manager.sh
git add zervigo_future_CICD/scripts/verify-microservice-deployment.sh
```

### 文档（推荐）
```bash
git add zervigo_future_CICD/docs/ALIBABA_MICROSERVICE_DEPLOYMENT_GUIDE.md
git add zervigo_future_CICD/docs/ZERVIGO_FUTURE_CICD_IMPLEMENTATION_SUMMARY.md
git add zervigo_future_CICD/docs/QUICK_REFERENCE_ALIBABA_DEPLOYMENT.md
```

### 辅助文档（可选，但推荐）
```bash
git add zervigo_future_CICD/CHANGELOG.md
git add zervigo_future_CICD/INDEX.md
git add zervigo_future_CICD/SUMMARY.md
git add zervigo_future_CICD/GIT_ADD_GUIDE.md
```

## ❌ 不应该提交的文件

以下文件不应提交（已在.gitignore中）:
- `*.log` - 日志文件
- `*.tmp` - 临时文件
- `*.bak` - 备份文件
- `.DS_Store` - macOS系统文件
- `test_*.sh` - 测试脚本
- `.env.local` - 本地配置

## 📦 一键添加所有必需文件

```bash
# 方式1: 逐个添加（推荐，更清晰）
cd /Users/szjason72/szbolent/LoomaCRM

# 核心文件
git add zervigo_future_CICD/.gitignore
git add zervigo_future_CICD/README.md
git add zervigo_future_CICD/INSTALLATION.md

# GitHub Actions
git add zervigo_future_CICD/workflows/

# 脚本
git add zervigo_future_CICD/scripts/

# 文档
git add zervigo_future_CICD/docs/

# 辅助文档
git add zervigo_future_CICD/CHANGELOG.md
git add zervigo_future_CICD/INDEX.md
git add zervigo_future_CICD/SUMMARY.md
git add zervigo_future_CICD/GIT_ADD_GUIDE.md

# 方式2: 一次性添加整个目录（注意会包含所有文件）
git add zervigo_future_CICD/

# 查看将要提交的文件
git status

# 提交
git commit -m "feat: add Zervigo Future CI/CD deployment suite"
```

## 📋 提交前检查清单

提交前请确认：

- [ ] 没有包含敏感信息（密码、密钥等）
- [ ] 没有包含临时文件（.log, .tmp等）
- [ ] 没有包含测试脚本（test_*.sh）
- [ ] 脚本文件有执行权限（chmod +x）
- [ ] .gitignore文件已配置
- [ ] 所有文档路径引用正确

## 🔍 检查提交内容

```bash
# 查看即将提交的文件
git status

# 查看具体改动
git diff --cached

# 如果需要移除某个文件
git reset HEAD <file>
```

## ✨ 推荐的提交信息格式

```bash
# 新功能
git commit -m "feat: add Zervigo Future CI/CD deployment suite"

# 文档更新
git commit -m "docs: update deployment guide"

# 脚本优化
git commit -m "refactor: improve deployment scripts"

# Bug修复
git commit -m "fix: resolve health check timeout issue"
```

## 📂 目录结构说明

```
zervigo_future_CICD/          # 整个CI/CD套件
├── .gitignore               # ✅ 必需 - Git忽略规则
├── README.md                # ✅ 必需 - 项目说明
├── INSTALLATION.md          # ✅ 必需 - 安装指南
├── CHANGELOG.md             # ⭕ 可选 - 版本历史
├── INDEX.md                 # ⭕ 可选 - 文档索引
├── SUMMARY.md               # ⭕ 可选 - 项目总结
├── GIT_ADD_GUIDE.md         # ⭕ 可选 - 本指南
├── workflows/               # ✅ 必需
│   └── zervigo-future-deploy.yml
├── scripts/                 # ✅ 必需
│   ├── setup-cicd.sh
│   ├── quick-deploy.sh
│   ├── microservice-deployment-manager.sh
│   └── verify-microservice-deployment.sh
├── docs/                    # ✅ 推荐
│   ├── ALIBABA_MICROSERVICE_DEPLOYMENT_GUIDE.md
│   ├── ZERVIGO_FUTURE_CICD_IMPLEMENTATION_SUMMARY.md
│   └── QUICK_REFERENCE_ALIBABA_DEPLOYMENT.md
└── configs/                 # ⭕ 保留（空目录）
```

## 🎯 最小化提交（如果需要精简）

如果您想保持最小化，只提交核心文件：

```bash
cd /Users/szjason72/szbolent/LoomaCRM

# 最小化提交 - 只包含必需文件
git add zervigo_future_CICD/.gitignore
git add zervigo_future_CICD/README.md
git add zervigo_future_CICD/workflows/
git add zervigo_future_CICD/scripts/
git add zervigo_future_CICD/docs/ALIBABA_MICROSERVICE_DEPLOYMENT_GUIDE.md

git commit -m "feat: add minimal CI/CD deployment suite"
```

---

**提示**: 使用`.gitignore`确保不会意外提交不必要的文件。
