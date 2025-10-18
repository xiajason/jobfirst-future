#!/bin/bash

###############################################################################
# Zervigo Future 微服务部署验证脚本
# 功能: 验证微服务部署状态和健康检查
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

# 微服务配置
declare -A SERVICES=(
    ["api-gateway"]="http://localhost:8080/health"
    ["user-service"]="http://localhost:8081/health"
    ["resume-service"]="http://localhost:8082/health"
    ["company-service"]="http://localhost:8083/health"
    ["job-service"]="http://localhost:8084/health"
    ["ai-service"]="http://localhost:8100/health"
    ["nginx"]="http://localhost/health"
)

declare -A DATABASES=(
    ["mysql"]="mysql -h localhost -P 3306 -u jobfirst -pjobfirst_password_2024 -e 'SELECT 1'"
    ["postgres"]="psql -h localhost -p 5432 -U postgres -d jobfirst_ai -c 'SELECT 1'"
    ["redis"]="redis-cli -h localhost -p 6379 -a redis_password_2024 ping"
    ["neo4j"]="curl -f http://localhost:7474"
    ["consul"]="curl -f http://localhost:8500/v1/status/leader"
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

# 函数: 检查服务健康状态
check_service_health() {
    local service_name="$1"
    local health_url="$2"
    
    if curl -f -s "$health_url" > /dev/null 2>&1; then
        print_success "$service_name 健康检查通过"
        return 0
    else
        print_error "$service_name 健康检查失败"
        return 1
    fi
}

# 函数: 检查数据库连接
check_database_connection() {
    local db_name="$1"
    local test_command="$2"
    
    if eval "$test_command" > /dev/null 2>&1; then
        print_success "$db_name 连接正常"
        return 0
    else
        print_error "$db_name 连接失败"
        return 1
    fi
}

# 函数: 检查端口监听
check_port_listening() {
    local port="$1"
    local service_name="$2"
    
    if netstat -tlnp 2>/dev/null | grep ":$port " > /dev/null; then
        print_success "$service_name 端口 $port 正在监听"
        return 0
    else
        print_error "$service_name 端口 $port 未监听"
        return 1
    fi
}

# 函数: 检查Docker容器状态
check_docker_containers() {
    print_header "🐳 Docker容器状态"
    
    # 检查Docker是否运行
    if ! docker info > /dev/null 2>&1; then
        print_warning "Docker未运行，跳过容器检查"
        return 0
    fi
    
    # 检查相关容器
    local containers=("jobfirst-mysql" "jobfirst-postgres" "jobfirst-redis" "jobfirst-neo4j" "jobfirst-consul" "jobfirst-api-gateway" "jobfirst-user-service" "jobfirst-resume-service" "jobfirst-company-service" "jobfirst-job-service" "jobfirst-ai-service" "jobfirst-nginx")
    
    for container in "${containers[@]}"; do
        if docker ps --format "table {{.Names}}\t{{.Status}}" | grep "$container" > /dev/null; then
            local status=$(docker ps --format "{{.Status}}" --filter "name=$container")
            print_success "$container: $status"
        else
            print_error "$container 未运行"
        fi
    done
}

# 函数: 检查微服务状态
check_microservices() {
    print_header "🔍 微服务健康检查"
    
    local total_services=0
    local healthy_services=0
    
    for service in "${!SERVICES[@]}"; do
        ((total_services++))
        if check_service_health "$service" "${SERVICES[$service]}"; then
            ((healthy_services++))
        fi
    done
    
    echo ""
    print_info "微服务健康状态: $healthy_services/$total_services"
    
    if [ $healthy_services -eq $total_services ]; then
        print_success "所有微服务健康检查通过"
        return 0
    else
        print_error "部分微服务健康检查失败"
        return 1
    fi
}

# 函数: 检查数据库状态
check_databases() {
    print_header "🗄️ 数据库连接检查"
    
    local total_dbs=0
    local healthy_dbs=0
    
    for db in "${!DATABASES[@]}"; do
        ((total_dbs++))
        if check_database_connection "$db" "${DATABASES[$db]}"; then
            ((healthy_dbs++))
        fi
    done
    
    echo ""
    print_info "数据库连接状态: $healthy_dbs/$total_dbs"
    
    if [ $healthy_dbs -eq $total_dbs ]; then
        print_success "所有数据库连接正常"
        return 0
    else
        print_error "部分数据库连接失败"
        return 1
    fi
}

# 函数: 检查端口监听
check_ports() {
    print_header "🔌 端口监听检查"
    
    local ports=("3306:MySQL" "5432:PostgreSQL" "6379:Redis" "7474:Neo4j" "8500:Consul" "8080:API Gateway" "8081:User Service" "8082:Resume Service" "8083:Company Service" "8084:Job Service" "8100:AI Service" "80:Nginx")
    
    local total_ports=0
    local listening_ports=0
    
    for port_info in "${ports[@]}"; do
        local port=$(echo "$port_info" | cut -d: -f1)
        local service=$(echo "$port_info" | cut -d: -f2)
        ((total_ports++))
        if check_port_listening "$port" "$service"; then
            ((listening_ports++))
        fi
    done
    
    echo ""
    print_info "端口监听状态: $listening_ports/$total_ports"
}

# 函数: 检查服务发现
check_service_discovery() {
    print_header "🔍 服务发现检查"
    
    # 检查Consul状态
    if curl -f -s http://localhost:8500/v1/status/leader > /dev/null 2>&1; then
        print_success "Consul服务正常运行"
        
        # 检查注册的服务
        echo ""
        print_info "已注册的服务:"
        curl -s http://localhost:8500/v1/agent/services | jq -r 'keys[]' 2>/dev/null || echo "无法获取服务列表"
        
        # 检查服务健康状态
        echo ""
        print_info "服务健康状态:"
        curl -s http://localhost:8500/v1/health/state/any | jq -r '.[] | "\(.ServiceName): \(.Status)"' 2>/dev/null || echo "无法获取健康状态"
        
    else
        print_error "Consul服务未运行"
        return 1
    fi
}

# 函数: 检查系统资源
check_system_resources() {
    print_header "💻 系统资源检查"
    
    # 检查CPU使用率
    local cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)
    print_info "CPU使用率: ${cpu_usage}%"
    
    # 检查内存使用率
    local memory_usage=$(free | grep Mem | awk '{printf "%.1f", $3/$2 * 100.0}')
    print_info "内存使用率: ${memory_usage}%"
    
    # 检查磁盘使用率
    local disk_usage=$(df -h / | awk 'NR==2{printf "%s", $5}')
    print_info "磁盘使用率: $disk_usage"
    
    # 检查网络连接
    local network_connections=$(netstat -an | wc -l)
    print_info "网络连接数: $network_connections"
}

# 函数: 生成部署报告
generate_report() {
    print_header "📊 部署验证报告"
    
    local report_file="/tmp/microservice-deployment-report-$(date +%Y%m%d-%H%M%S).txt"
    
    {
        echo "Zervigo Future 微服务部署验证报告"
        echo "生成时间: $(date)"
        echo "=========================================="
        echo ""
        
        echo "微服务健康状态:"
        for service in "${!SERVICES[@]}"; do
            if curl -f -s "${SERVICES[$service]}" > /dev/null 2>&1; then
                echo "✅ $service: 健康"
            else
                echo "❌ $service: 不健康"
            fi
        done
        
        echo ""
        echo "数据库连接状态:"
        for db in "${!DATABASES[@]}"; do
            if eval "${DATABASES[$db]}" > /dev/null 2>&1; then
                echo "✅ $db: 连接正常"
            else
                echo "❌ $db: 连接失败"
            fi
        done
        
        echo ""
        echo "系统资源状态:"
        echo "CPU使用率: $(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)%"
        echo "内存使用率: $(free | grep Mem | awk '{printf "%.1f", $3/$2 * 100.0}')%"
        echo "磁盘使用率: $(df -h / | awk 'NR==2{printf "%s", $5}')"
        
    } > "$report_file"
    
    print_success "部署验证报告已生成: $report_file"
}

# 主函数
main() {
    print_header "🚀 Zervigo Future 微服务部署验证"
    
    local exit_code=0
    
    # 检查Docker容器状态
    check_docker_containers || exit_code=1
    
    # 检查微服务健康状态
    check_microservices || exit_code=1
    
    # 检查数据库连接
    check_databases || exit_code=1
    
    # 检查端口监听
    check_ports
    
    # 检查服务发现
    check_service_discovery || exit_code=1
    
    # 检查系统资源
    check_system_resources
    
    # 生成部署报告
    generate_report
    
    echo ""
    if [ $exit_code -eq 0 ]; then
        print_success "🎉 所有检查通过！微服务部署验证成功！"
    else
        print_error "❌ 部分检查失败！请检查上述错误信息！"
    fi
    
    exit $exit_code
}

# 执行主函数
main "$@"
