#!/bin/bash

###############################################################################
# Zervigo Future 微服务部署管理器
# 功能: 按正确时序部署微服务，排除认证服务
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
DEPLOY_PATH="/opt/services"
LOG_DIR="$DEPLOY_PATH/logs"
BACKUP_DIR="$DEPLOY_PATH/backups"

# 微服务部署时序 (排除认证服务)
declare -A DEPLOYMENT_ORDER=(
    ["infrastructure"]="mysql,postgres,redis,neo4j,consul"
    ["gateway"]="api-gateway"
    ["auth-layer"]="user-service"
    ["business"]="resume-service,company-service,job-service"
    ["ai-services"]="ai-service"
    ["nginx"]="nginx"
)

# 微服务端口映射
declare -A SERVICE_PORTS=(
    ["api-gateway"]="8080"
    ["user-service"]="8081"
    ["resume-service"]="8082"
    ["company-service"]="8083"
    ["job-service"]="8084"
    ["ai-service"]="8100"
    ["nginx"]="80"
)

# 微服务健康检查端点
declare -A HEALTH_ENDPOINTS=(
    ["api-gateway"]="http://localhost:8080/health"
    ["user-service"]="http://localhost:8081/health"
    ["resume-service"]="http://localhost:8082/health"
    ["company-service"]="http://localhost:8083/health"
    ["job-service"]="http://localhost:8084/health"
    ["ai-service"]="http://localhost:8100/health"
    ["nginx"]="http://localhost/health"
)

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

# 函数: 等待服务启动
wait_for_service() {
    local service_name="$1"
    local health_url="$2"
    local max_attempts=30
    local attempt=1
    
    print_info "等待 $service_name 启动..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f -s "$health_url" > /dev/null 2>&1; then
            print_success "$service_name 启动成功"
            return 0
        fi
        
        echo -n "."
        sleep 2
        ((attempt++))
    done
    
    print_error "$service_name 启动超时"
    return 1
}

# 函数: 检查服务依赖
check_dependencies() {
    local service_name="$1"
    
    case "$service_name" in
        "api-gateway")
            # API Gateway依赖Consul
            if ! curl -f -s http://localhost:8500/v1/status/leader > /dev/null 2>&1; then
                print_error "Consul服务未启动，无法启动API Gateway"
                return 1
            fi
            ;;
        "user-service")
            # User Service依赖MySQL和Redis
            if ! podman exec jobfirst-mysql mysql -uroot -pjobfirst_password_2024 -e "SELECT 1" > /dev/null 2>&1; then
                print_error "MySQL服务未启动，无法启动User Service"
                return 1
            fi
            if ! podman exec jobfirst-redis redis-cli ping > /dev/null 2>&1; then
                print_error "Redis服务未启动，无法启动User Service"
                return 1
            fi
            ;;
        "resume-service"|"company-service"|"job-service")
            # 业务服务依赖User Service
            if ! curl -f -s http://localhost:8081/health > /dev/null 2>&1; then
                print_error "User Service未启动，无法启动 $service_name"
                return 1
            fi
            ;;
        "ai-service")
            # AI服务依赖PostgreSQL和User Service
            if ! podman exec jobfirst-postgres psql -U postgres -d jobfirst_ai -c "SELECT 1" > /dev/null 2>&1; then
                print_error "PostgreSQL服务未启动，无法启动AI Service"
                return 1
            fi
            if ! curl -f -s http://localhost:8081/health > /dev/null 2>&1; then
                print_error "User Service未启动，无法启动AI Service"
                return 1
            fi
            ;;
    esac
    
    return 0
}

# 函数: 启动基础设施层
deploy_infrastructure() {
    print_header "🏗️ 部署基础设施层"
    
    print_info "启动数据库服务..."
    
    # 启动MySQL
    print_info "启动MySQL..."
    podman-compose up -d mysql || docker-compose up -d mysql
    wait_for_service "MySQL" "http://localhost:3306" || true
    
    # 启动PostgreSQL
    print_info "启动PostgreSQL..."
    podman-compose up -d postgres || docker-compose up -d postgres
    wait_for_service "PostgreSQL" "http://localhost:5432" || true
    
    # 启动Redis
    print_info "启动Redis..."
    podman-compose up -d redis || docker-compose up -d redis
    wait_for_service "Redis" "http://localhost:6379" || true
    
    # 启动Neo4j
    print_info "启动Neo4j..."
    podman-compose up -d neo4j || docker-compose up -d neo4j
    wait_for_service "Neo4j" "http://localhost:7474" || true
    
    # 启动Consul
    print_info "启动Consul..."
    podman-compose up -d consul || docker-compose up -d consul
    wait_for_service "Consul" "http://localhost:8500/v1/status/leader" || true
    
    print_success "基础设施层部署完成"
}

# 函数: 启动网关层
deploy_gateway() {
    print_header "🌐 部署网关层"
    
    print_info "启动API Gateway..."
    
    # 检查依赖
    check_dependencies "api-gateway" || return 1
    
    # 停止现有服务
    pkill -f api-gateway || true
    
    # 启动API Gateway
    cd "$DEPLOY_PATH/backend"
    chmod +x bin/api-gateway
    nohup ./bin/api-gateway > logs/api-gateway.log 2>&1 &
    echo $! > logs/api-gateway.pid
    
    # 等待启动
    wait_for_service "API Gateway" "http://localhost:8080/health"
    
    print_success "网关层部署完成"
}

# 函数: 启动认证授权层
deploy_auth_layer() {
    print_header "🔐 部署认证授权层"
    
    print_info "启动User Service..."
    
    # 检查依赖
    check_dependencies "user-service" || return 1
    
    # 停止现有服务
    pkill -f user-service || true
    
    # 启动User Service
    cd "$DEPLOY_PATH/backend"
    chmod +x bin/user-service
    nohup ./bin/user-service > logs/user-service.log 2>&1 &
    echo $! > logs/user-service.pid
    
    # 等待启动
    wait_for_service "User Service" "http://localhost:8081/health"
    
    print_success "认证授权层部署完成"
}

# 函数: 启动业务服务层
deploy_business_services() {
    print_header "💼 部署业务服务层"
    
    # 启动Resume Service
    print_info "启动Resume Service..."
    check_dependencies "resume-service" || return 1
    pkill -f resume-service || true
    cd "$DEPLOY_PATH/backend"
    chmod +x bin/resume-service
    nohup ./bin/resume-service > logs/resume-service.log 2>&1 &
    echo $! > logs/resume-service.pid
    wait_for_service "Resume Service" "http://localhost:8082/health"
    
    # 启动Company Service
    print_info "启动Company Service..."
    check_dependencies "company-service" || return 1
    pkill -f company-service || true
    chmod +x bin/company-service
    nohup ./bin/company-service > logs/company-service.log 2>&1 &
    echo $! > logs/company-service.pid
    wait_for_service "Company Service" "http://localhost:8083/health"
    
    # 启动Job Service
    print_info "启动Job Service..."
    check_dependencies "job-service" || return 1
    pkill -f job-service || true
    chmod +x bin/job-service
    nohup ./bin/job-service > logs/job-service.log 2>&1 &
    echo $! > logs/job-service.pid
    wait_for_service "Job Service" "http://localhost:8084/health"
    
    print_success "业务服务层部署完成"
}

# 函数: 启动AI服务层
deploy_ai_services() {
    print_header "🤖 部署AI服务层"
    
    print_info "启动AI Service..."
    
    # 检查依赖
    check_dependencies "ai-service" || return 1
    
    # 停止现有服务
    pkill -f ai_service_with_zervigo.py || true
    
    # 设置AI服务环境
    cd "$DEPLOY_PATH/ai-services/ai-service"
    if [ ! -d "venv" ]; then
        python3 -m venv venv
    fi
    source venv/bin/activate
    pip install -r requirements.txt
    
    # 启动AI服务
    nohup python ai_service_with_zervigo.py > ai_service.log 2>&1 &
    echo $! > ai_service.pid
    
    # 等待启动
    wait_for_service "AI Service" "http://localhost:8100/health"
    
    print_success "AI服务层部署完成"
}

# 函数: 启动Nginx
deploy_nginx() {
    print_header "🌐 部署Nginx反向代理"
    
    print_info "启动Nginx..."
    
    # 停止现有Nginx
    podman-compose down nginx || docker-compose down nginx || true
    
    # 启动Nginx
    podman-compose up -d nginx || docker-compose up -d nginx
    
    # 等待启动
    wait_for_service "Nginx" "http://localhost/health"
    
    print_success "Nginx部署完成"
}

# 函数: 验证部署
verify_deployment() {
    print_header "✅ 验证部署"
    
    print_info "验证微服务部署状态..."
    
    # 检查基础设施层
    echo "=== 基础设施层 ==="
    podman ps | grep -E "(mysql|postgres|redis|neo4j|consul)" || docker ps | grep -E "(mysql|postgres|redis|neo4j|consul)"
    
    # 检查网关层
    echo "=== 网关层 ==="
    curl -f http://localhost:8080/health && echo "API Gateway OK"
    curl -f http://localhost/health && echo "Nginx OK"
    
    # 检查认证授权层
    echo "=== 认证授权层 ==="
    curl -f http://localhost:8081/health && echo "User Service OK"
    
    # 检查业务服务层
    echo "=== 业务服务层 ==="
    curl -f http://localhost:8082/health && echo "Resume Service OK"
    curl -f http://localhost:8083/health && echo "Company Service OK"
    curl -f http://localhost:8084/health && echo "Job Service OK"
    
    # 检查AI服务层
    echo "=== AI服务层 ==="
    curl -f http://localhost:8100/health && echo "AI Service OK"
    
    # 检查服务发现
    echo "=== 服务发现 ==="
    curl -f http://localhost:8500/v1/agent/services && echo "Consul OK"
    
    print_success "所有微服务部署验证完成！"
}

# 函数: 停止所有服务
stop_all_services() {
    print_header "🛑 停止所有服务"
    
    print_info "停止微服务..."
    pkill -f api-gateway || true
    pkill -f user-service || true
    pkill -f resume-service || true
    pkill -f company-service || true
    pkill -f job-service || true
    pkill -f ai_service_with_zervigo.py || true
    
    print_info "停止容器服务..."
    podman-compose down || docker-compose down || true
    
    print_success "所有服务已停止"
}

# 函数: 显示服务状态
show_status() {
    print_header "📊 服务状态"
    
    echo "=== 微服务进程 ==="
    ps aux | grep -E "(api-gateway|user-service|resume-service|company-service|job-service|ai_service)" | grep -v grep || echo "无微服务进程运行"
    
    echo ""
    echo "=== 容器服务 ==="
    podman ps || docker ps || echo "无容器服务运行"
    
    echo ""
    echo "=== 端口监听 ==="
    netstat -tlnp | grep -E "(8080|8081|8082|8083|8084|8100|80|3306|5432|6379|7474|8500)" || echo "无相关端口监听"
}

# 函数: 显示帮助信息
show_help() {
    echo "Zervigo Future 微服务部署管理器"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  deploy-all     部署所有微服务 (按正确时序)"
    echo "  deploy-infra   仅部署基础设施层"
    echo "  deploy-gateway 仅部署网关层"
    echo "  deploy-auth    仅部署认证授权层"
    echo "  deploy-business 仅部署业务服务层"
    echo "  deploy-ai     仅部署AI服务层"
    echo "  deploy-nginx   仅部署Nginx"
    echo "  stop           停止所有服务"
    echo "  restart        重启所有服务"
    echo "  status         显示服务状态"
    echo "  verify         验证部署"
    echo "  help           显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 deploy-all    # 完整部署"
    echo "  $0 status        # 查看状态"
    echo "  $0 stop          # 停止服务"
}

# 主函数
main() {
    case "${1:-help}" in
        "deploy-all")
            deploy_infrastructure
            deploy_gateway
            deploy_auth_layer
            deploy_business_services
            deploy_ai_services
            deploy_nginx
            verify_deployment
            ;;
        "deploy-infra")
            deploy_infrastructure
            ;;
        "deploy-gateway")
            deploy_gateway
            ;;
        "deploy-auth")
            deploy_auth_layer
            ;;
        "deploy-business")
            deploy_business_services
            ;;
        "deploy-ai")
            deploy_ai_services
            ;;
        "deploy-nginx")
            deploy_nginx
            ;;
        "stop")
            stop_all_services
            ;;
        "restart")
            stop_all_services
            sleep 5
            main deploy-all
            ;;
        "status")
            show_status
            ;;
        "verify")
            verify_deployment
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
