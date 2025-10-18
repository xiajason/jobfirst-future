#!/bin/bash

###############################################################################
# Zervigo Future 快速部署脚本
# 功能: 快速部署微服务到阿里云服务器
# 作者: AI Assistant
# 日期: 2025-10-18
###############################################################################

set -e

# 配置变量
SERVER_IP="47.115.168.107"
SERVER_USER="root"
DEPLOY_PATH="/opt/services"
PROJECT_PATH="/Users/szjason72/szbolent/LoomaCRM/zervigo_future"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# 函数: 打印信息消息
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# 函数: 检查SSH连接
check_ssh_connection() {
    print_info "检查SSH连接..."
    if ssh -o ConnectTimeout=10 -o BatchMode=yes root@$SERVER_IP "echo 'SSH连接正常'" 2>/dev/null; then
        print_success "SSH连接正常"
        return 0
    else
        print_error "SSH连接失败"
        return 1
    fi
}

# 函数: 上传项目文件
upload_project_files() {
    print_header "📤 上传项目文件"
    
    print_info "上传项目文件到服务器..."
    
    # 创建部署目录
    ssh root@$SERVER_IP "mkdir -p $DEPLOY_PATH"
    
    # 上传项目文件
    rsync -avz --delete \
        --exclude='.git' \
        --exclude='node_modules' \
        --exclude='venv' \
        --exclude='__pycache__' \
        --exclude='*.log' \
        --exclude='*.pid' \
        "$PROJECT_PATH/" \
        "root@$SERVER_IP:$DEPLOY_PATH/"
    
    print_success "项目文件上传完成"
}

# 函数: 部署基础设施
deploy_infrastructure() {
    print_header "🏗️ 部署基础设施"
    
    print_info "启动数据库和服务发现..."
    
    ssh root@$SERVER_IP << 'ENDSSH'
    cd /opt/services
    
    # 停止现有服务
    podman-compose down || docker-compose down || true
    
    # 启动基础设施服务
    podman-compose -f docker-compose.microservices.yml up -d mysql postgres redis neo4j consul || \
    docker-compose -f docker-compose.microservices.yml up -d mysql postgres redis neo4j consul
    
    # 等待服务启动
    echo "等待数据库服务启动..."
    sleep 30
    
    # 验证服务
    echo "验证数据库连接..."
    podman exec jobfirst-mysql mysql -uroot -pjobfirst_password_2024 -e "SELECT 1" || \
    docker exec jobfirst-mysql mysql -uroot -pjobfirst_password_2024 -e "SELECT 1"
    
    echo "验证Redis连接..."
    podman exec jobfirst-redis redis-cli ping || \
    docker exec jobfirst-redis redis-cli ping
    
    echo "验证Consul服务..."
    curl -f http://localhost:8500/v1/status/leader || echo "Consul检查跳过"
    
    echo "基础设施部署完成"
ENDSSH
    
    print_success "基础设施部署完成"
}

# 函数: 部署微服务
deploy_microservices() {
    print_header "🚀 部署微服务"
    
    print_info "启动微服务..."
    
    ssh root@$SERVER_IP << 'ENDSSH'
    cd /opt/services
    
    # 启动微服务
    podman-compose -f docker-compose.microservices.yml up -d api-gateway user-service resume-service company-service job-service ai-service nginx || \
    docker-compose -f docker-compose.microservices.yml up -d api-gateway user-service resume-service company-service job-service ai-service nginx
    
    # 等待服务启动
    echo "等待微服务启动..."
    sleep 30
    
    # 验证服务
    echo "验证微服务健康状态..."
    curl -f http://localhost:8080/health && echo "API Gateway OK"
    curl -f http://localhost:8081/health && echo "User Service OK"
    curl -f http://localhost:8082/health && echo "Resume Service OK"
    curl -f http://localhost:8083/health && echo "Company Service OK"
    curl -f http://localhost:8084/health && echo "Job Service OK"
    curl -f http://localhost:8100/health && echo "AI Service OK"
    curl -f http://localhost/health && echo "Nginx OK"
    
    echo "微服务部署完成"
ENDSSH
    
    print_success "微服务部署完成"
}

# 函数: 验证部署
verify_deployment() {
    print_header "✅ 验证部署"
    
    print_info "验证微服务部署状态..."
    
    ssh root@$SERVER_IP << 'ENDSSH'
    cd /opt/services
    
    echo "=== 容器状态 ==="
    podman ps || docker ps
    
    echo ""
    echo "=== 微服务健康检查 ==="
    curl -f http://localhost:8080/health && echo " - API Gateway"
    curl -f http://localhost:8081/health && echo " - User Service"
    curl -f http://localhost:8082/health && echo " - Resume Service"
    curl -f http://localhost:8083/health && echo " - Company Service"
    curl -f http://localhost:8084/health && echo " - Job Service"
    curl -f http://localhost:8100/health && echo " - AI Service"
    curl -f http://localhost/health && echo " - Nginx"
    
    echo ""
    echo "=== 服务发现 ==="
    curl -f http://localhost:8500/v1/agent/services && echo " - Consul"
    
    echo ""
    echo "=== 端口监听 ==="
    netstat -tlnp | grep -E "(8080|8081|8082|8083|8084|8100|80|3306|5432|6379|7474|8500)"
    
    echo "部署验证完成"
ENDSSH
    
    print_success "部署验证完成"
}

# 函数: 显示部署信息
show_deployment_info() {
    print_header "📊 部署信息"
    
    echo "服务器地址: $SERVER_IP"
    echo "部署路径: $DEPLOY_PATH"
    echo ""
    echo "访问地址:"
    echo "  - API Gateway: http://$SERVER_IP:8080"
    echo "  - User Service: http://$SERVER_IP:8081"
    echo "  - Resume Service: http://$SERVER_IP:8082"
    echo "  - Company Service: http://$SERVER_IP:8083"
    echo "  - Job Service: http://$SERVER_IP:8084"
    echo "  - AI Service: http://$SERVER_IP:8100"
    echo "  - Nginx: http://$SERVER_IP"
    echo "  - Consul UI: http://$SERVER_IP:8500"
    echo ""
    echo "管理命令:"
    echo "  - 查看状态: ssh root@$SERVER_IP 'cd $DEPLOY_PATH && podman-compose ps'"
    echo "  - 查看日志: ssh root@$SERVER_IP 'cd $DEPLOY_PATH && podman-compose logs -f'"
    echo "  - 停止服务: ssh root@$SERVER_IP 'cd $DEPLOY_PATH && podman-compose down'"
    echo "  - 重启服务: ssh root@$SERVER_IP 'cd $DEPLOY_PATH && podman-compose restart'"
}

# 主函数
main() {
    print_header "🚀 Zervigo Future 快速部署"
    
    # 检查SSH连接
    if ! check_ssh_connection; then
        print_error "无法连接到服务器，请检查SSH配置"
        exit 1
    fi
    
    # 上传项目文件
    upload_project_files
    
    # 部署基础设施
    deploy_infrastructure
    
    # 部署微服务
    deploy_microservices
    
    # 验证部署
    verify_deployment
    
    # 显示部署信息
    show_deployment_info
    
    print_success "🎉 Zervigo Future 微服务部署完成！"
}

# 执行主函数
main "$@"
