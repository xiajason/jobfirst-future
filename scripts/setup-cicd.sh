#!/bin/bash

###############################################################################
# Zervigo Future CI/CD 快速安装脚本
# 功能: 一键安装和配置CI/CD系统
# 作者: AI Assistant
# 日期: 2025-10-18
###############################################################################

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
ALIBABA_SERVER_IP="47.115.168.107"
ALIBABA_SERVER_USER="root"
SSH_KEY_PATH="$HOME/.ssh/cross_cloud_key"
DEPLOY_PATH="/opt/services"
PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
CICD_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# 函数: 打印标题
print_header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

# 函数: 打印成功消息
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

# 函数: 打印错误消息
print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 函数: 打印警告消息
print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# 函数: 打印信息消息
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# 函数: 检查SSH连接
check_ssh_connection() {
    print_info "检查SSH连接..."
    if ssh -o ConnectTimeout=10 -i "$SSH_KEY_PATH" $ALIBABA_SERVER_USER@$ALIBABA_SERVER_IP "echo 'SSH连接正常'" 2>/dev/null; then
        print_success "SSH连接正常"
        return 0
    else
        print_error "SSH连接失败"
        return 1
    fi
}

# 函数: 安装GitHub Actions workflow
install_github_workflow() {
    print_header "📦 安装GitHub Actions Workflow"
    
    # 创建.github/workflows目录
    mkdir -p "$PROJECT_ROOT/.github/workflows"
    
    # 复制workflow文件
    cp "$CICD_ROOT/workflows/zervigo-future-deploy.yml" "$PROJECT_ROOT/.github/workflows/"
    
    print_success "GitHub Actions workflow已安装"
    print_info "文件位置: .github/workflows/zervigo-future-deploy.yml"
}

# 函数: 配置GitHub Secrets
configure_github_secrets() {
    print_header "🔐 配置GitHub Secrets"
    
    echo "请在GitHub仓库中配置以下Secrets:"
    echo ""
    echo "1. ALIBABA_SERVER_IP"
    echo "   值: $ALIBABA_SERVER_IP"
    echo ""
    echo "2. ALIBABA_SERVER_USER"
    echo "   值: $ALIBABA_SERVER_USER"
    echo ""
    echo "3. ALIBABA_SSH_PRIVATE_KEY"
    echo "   值: (您的SSH私钥内容)"
    if [ -f "$SSH_KEY_PATH" ]; then
        echo ""
        echo "   可以使用以下命令获取私钥:"
        echo "   cat $SSH_KEY_PATH"
    fi
    echo ""
    echo "4. ALIBABA_DEPLOY_PATH (可选)"
    echo "   值: $DEPLOY_PATH"
    echo ""
    echo "配置步骤:"
    echo "1. 访问: https://github.com/your-org/your-repo/settings/secrets/actions"
    echo "2. 点击 'New repository secret'"
    echo "3. 输入Secret名称和值"
    echo "4. 点击 'Add secret'"
    echo ""
    
    read -p "按回车键继续..."
}

# 函数: 准备服务器环境
prepare_server() {
    print_header "🖥️ 准备服务器环境"
    
    print_info "创建部署目录..."
    ssh -i "$SSH_KEY_PATH" $ALIBABA_SERVER_USER@$ALIBABA_SERVER_IP << ENDSSH
mkdir -p $DEPLOY_PATH/{backend/bin,configs,logs,scripts}
chmod 755 $DEPLOY_PATH
chmod 755 $DEPLOY_PATH/backend
chmod 755 $DEPLOY_PATH/backend/bin
chmod 755 $DEPLOY_PATH/configs
chmod 755 $DEPLOY_PATH/logs
chmod 755 $DEPLOY_PATH/scripts
echo "✅ 服务器目录结构已创建"
ENDSSH
    
    print_success "服务器环境准备完成"
}

# 函数: 上传部署脚本
upload_scripts() {
    print_header "📤 上传部署脚本"
    
    print_info "上传脚本到服务器..."
    scp -i "$SSH_KEY_PATH" "$CICD_ROOT/scripts/"*.sh \
        $ALIBABA_SERVER_USER@$ALIBABA_SERVER_IP:$DEPLOY_PATH/scripts/
    
    print_info "设置脚本执行权限..."
    ssh -i "$SSH_KEY_PATH" $ALIBABA_SERVER_USER@$ALIBABA_SERVER_IP \
        "chmod +x $DEPLOY_PATH/scripts/*.sh"
    
    print_success "部署脚本上传完成"
}

# 函数: 验证安装
verify_installation() {
    print_header "✅ 验证安装"
    
    echo "检查项目:"
    
    # 检查workflow文件
    if [ -f "$PROJECT_ROOT/.github/workflows/zervigo-future-deploy.yml" ]; then
        print_success "GitHub Actions workflow文件存在"
    else
        print_error "GitHub Actions workflow文件不存在"
    fi
    
    # 检查SSH连接
    if check_ssh_connection; then
        print_success "SSH连接正常"
    else
        print_error "SSH连接失败"
    fi
    
    # 检查服务器目录
    print_info "检查服务器目录结构..."
    ssh -i "$SSH_KEY_PATH" $ALIBABA_SERVER_USER@$ALIBABA_SERVER_IP << ENDSSH
echo "服务器目录结构:"
ls -la $DEPLOY_PATH/
echo ""
echo "部署脚本:"
ls -la $DEPLOY_PATH/scripts/
ENDSSH
    
    # 检查数据库容器
    print_info "检查数据库容器..."
    ssh -i "$SSH_KEY_PATH" $ALIBABA_SERVER_USER@$ALIBABA_SERVER_IP \
        "podman ps | grep migration"
    
    # 检查AI服务
    print_info "检查AI服务..."
    ssh -i "$SSH_KEY_PATH" $ALIBABA_SERVER_USER@$ALIBABA_SERVER_IP \
        "curl -f http://localhost:8100/health" && echo ""
    
    print_success "安装验证完成"
}

# 函数: 显示后续步骤
show_next_steps() {
    print_header "🎯 后续步骤"
    
    echo "安装完成！接下来您需要:"
    echo ""
    echo "1. 配置GitHub Secrets"
    echo "   访问: https://github.com/your-org/your-repo/settings/secrets/actions"
    echo ""
    echo "2. 测试部署"
    echo "   git add ."
    echo "   git commit -m 'test: setup CI/CD'"
    echo "   git push origin develop  # 先推送到develop分支测试"
    echo ""
    echo "3. 查看GitHub Actions执行情况"
    echo "   访问: https://github.com/your-org/your-repo/actions"
    echo ""
    echo "4. 如果测试成功，推送到main分支"
    echo "   git checkout main"
    echo "   git merge develop"
    echo "   git push origin main"
    echo ""
    echo "5. 查看部署文档"
    echo "   - README.md: 使用说明"
    echo "   - INSTALLATION.md: 安装配置指南"
    echo "   - docs/ALIBABA_MICROSERVICE_DEPLOYMENT_GUIDE.md: 详细部署指南"
    echo ""
}

# 主函数
main() {
    print_header "🚀 Zervigo Future CI/CD 快速安装"
    
    echo "本脚本将帮助您快速安装和配置CI/CD系统"
    echo ""
    echo "安装内容:"
    echo "- GitHub Actions workflow"
    echo "- 服务器环境准备"
    echo "- 部署脚本上传"
    echo ""
    echo "服务器信息:"
    echo "- IP: $ALIBABA_SERVER_IP"
    echo "- 用户: $ALIBABA_SERVER_USER"
    echo "- 部署路径: $DEPLOY_PATH"
    echo ""
    
    read -p "按回车键开始安装，或按Ctrl+C取消..."
    
    # 检查SSH密钥
    if [ ! -f "$SSH_KEY_PATH" ]; then
        print_error "SSH密钥不存在: $SSH_KEY_PATH"
        print_info "请确认SSH密钥路径是否正确"
        exit 1
    fi
    
    # 检查SSH连接
    if ! check_ssh_connection; then
        print_error "无法连接到服务器，请检查SSH配置"
        exit 1
    fi
    
    # 安装GitHub Actions workflow
    install_github_workflow
    
    # 配置GitHub Secrets
    configure_github_secrets
    
    # 准备服务器环境
    prepare_server
    
    # 上传部署脚本
    upload_scripts
    
    # 验证安装
    verify_installation
    
    # 显示后续步骤
    show_next_steps
    
    print_header "🎉 安装完成"
    print_success "CI/CD系统已成功安装！"
}

# 执行主函数
main "$@"
