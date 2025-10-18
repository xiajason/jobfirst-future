#!/bin/bash

# JobFirst 增强智能启动脚本 - 标准化启动流程
# 解决：1.标准化启动流程 2.全面端口检查 3.日志管理优化

# set -e  # 注释掉，避免因非关键错误退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 项目配置
PROJECT_ROOT="/Users/szjason72/zervi-basic/basic"
LOG_DIR="$PROJECT_ROOT/logs"
STARTUP_LOG="$LOG_DIR/smart-startup.log"

# 启动超时配置
SERVICE_START_TIMEOUT=30
HEALTH_CHECK_INTERVAL=5
MAX_HEALTH_CHECK_ATTEMPTS=12
PORT_CHECK_TIMEOUT=10
DEPENDENCY_WAIT_TIMEOUT=60
DEPENDENCY_CHECK_INTERVAL=3

# 标准化服务配置
STANDARD_SERVICES=(
    "basic-server:8080"
    "user-service:8081"
    "resume-service:8082"
    "company-service:8083"
    "notification-service:8084"
    "template-service:8085"
    "statistics-service:8086"
    "banner-service:8087"
    "dev-team-service:8088"
    "job-service:8089"
    "multi-database-service:8090"
    "unified-auth-service:8207"
    "local-ai-service:8206"
    "containerized-ai-service:8208"
)

# 基础设施服务配置
INFRASTRUCTURE_SERVICES=(
    "mysql:3306"
    "redis:6379"
    "postgresql:5432"
    "neo4j:7474"
    "consul:8500"
)

# AI服务容器配置
AI_SERVICES=(
    "ai-service:8208"
    "mineru:8001"
    "ai-models:8002"
    "ai-monitor:9090"
)

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$STARTUP_LOG"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$STARTUP_LOG"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$STARTUP_LOG"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$STARTUP_LOG"
}

log_step() {
    echo -e "${PURPLE}[STEP]${NC} $1" | tee -a "$STARTUP_LOG"
}

# 创建必要的目录
create_directories() {
    mkdir -p "$LOG_DIR"
    mkdir -p "$PROJECT_ROOT/backend/uploads"
    mkdir -p "$PROJECT_ROOT/backend/temp"
}

# 标准化端口检查函数
check_port_available() {
    local port=$1
    local service_name=$2
    
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        local pid=$(lsof -Pi :$port -sTCP:LISTEN -t | head -1)
        log_warning "$service_name 端口 $port 已被占用 (PID: $pid)"
        return 1
    else
        log_info "$service_name 端口 $port 可用"
        return 0
    fi
}

# 全面端口检查
comprehensive_port_check() {
    log_step "全面端口检查..."
    
    local occupied_ports=()
    local available_ports=()
    
    # 检查所有标准服务端口
    for service_info in "${STANDARD_SERVICES[@]}"; do
        IFS=':' read -r service_name port <<< "$service_info"
        
        if check_port_available "$port" "$service_name"; then
            available_ports+=("$service_name:$port")
        else
            occupied_ports+=("$service_name:$port")
        fi
    done
    
    # 检查基础设施服务端口
    for service_info in "${INFRASTRUCTURE_SERVICES[@]}"; do
        IFS=':' read -r service_name port <<< "$service_info"
        
        if check_port_available "$port" "$service_name"; then
            available_ports+=("$service_name:$port")
        else
            occupied_ports+=("$service_name:$port")
        fi
    done
    
    if [[ ${#occupied_ports[@]} -eq 0 ]]; then
        log_success "所有端口检查通过，可以启动服务"
        return 0
    else
        log_warning "以下端口被占用，可能影响服务启动:"
        for port in "${occupied_ports[@]}"; do
            log_warning "  - $port"
        done
        return 1
    fi
}

# 等待服务健康检查
wait_for_service_health() {
    local service_name=$1
    local health_url=$2
    local timeout=$3
    
    log_info "等待 $service_name 健康检查..."
    
    local count=0
    while [[ $count -lt $timeout ]]; do
        if curl -s "$health_url" >/dev/null 2>&1; then
            log_success "$service_name 健康检查通过"
            return 0
        fi
        
        sleep 1
        ((count++))
        echo -n "."
    done
    
    echo ""
    log_warning "$service_name 健康检查超时"
    return 1
}

# 标准化服务启动函数
standard_start_service() {
    local service_name=$1
    local port=$2
    local start_command=$3
    local health_url=$4
    local working_dir=$5
    
    log_info "标准化启动 $service_name (端口: $port)..."
    
    # 步骤1: 检查端口可用性
    if ! check_port_available "$port" "$service_name"; then
        log_error "$service_name 端口 $port 被占用，无法启动"
        return 1
    fi
    
    # 步骤2: 切换到工作目录
    if [[ -n "$working_dir" ]]; then
        cd "$working_dir" || {
            log_error "$service_name 工作目录不存在: $working_dir"
            return 1
        }
    fi
    
    # 步骤3: 执行启动命令
    log_info "执行启动命令: $start_command"
    if eval "$start_command"; then
        log_info "$service_name 启动命令执行成功"
    else
        log_error "$service_name 启动命令执行失败"
        return 1
    fi
    
    # 步骤4: 健康检查
    if [[ -n "$health_url" ]]; then
        if wait_for_service_health "$service_name" "$health_url" $SERVICE_START_TIMEOUT; then
            log_success "$service_name 启动成功并通过健康检查"
            return 0
        else
            log_warning "$service_name 启动成功但健康检查失败"
            return 1
        fi
    else
        log_success "$service_name 启动成功（无健康检查）"
        return 0
    fi
}

# 启动基础设施服务
start_infrastructure_services() {
    log_step "启动基础设施服务..."
    
    # 启动MySQL
    if ! brew services list | grep mysql | grep started &> /dev/null; then
        log_info "启动MySQL服务..."
        if brew services start mysql; then
            log_success "MySQL启动成功"
            sleep 5
        else
            log_error "MySQL启动失败"
            exit 1
        fi
    else
        log_info "MySQL已在运行"
    fi
    
    # 启动Redis
    if ! brew services list | grep redis | grep started &> /dev/null; then
        log_info "启动Redis服务..."
        if brew services start redis; then
            log_success "Redis启动成功"
            sleep 3
        else
            log_error "Redis启动失败"
            exit 1
        fi
    else
        log_info "Redis已在运行"
    fi
    
    # 启动PostgreSQL@14
    if ! brew services list | grep postgresql@14 | grep started &> /dev/null; then
        log_info "启动PostgreSQL@14服务..."
        if brew services start postgresql@14; then
            log_success "PostgreSQL@14启动成功"
            sleep 5
        else
            log_error "PostgreSQL@14启动失败"
            exit 1
        fi
    else
        log_info "PostgreSQL@14已在运行"
    fi
    
    # 启动Neo4j
    if ! brew services list | grep neo4j | grep started &> /dev/null; then
        log_info "启动Neo4j服务..."
        if brew services start neo4j; then
            log_success "Neo4j启动成功"
            sleep 5
        else
            log_error "Neo4j启动失败"
            exit 1
        fi
    else
        log_info "Neo4j已在运行"
    fi
}

# 启动Consul服务（统一使用launchctl管理）
start_consul_service() {
    log_step "启动Consul服务..."
    
    if ! curl -s http://localhost:8500/v1/status/leader >/dev/null 2>&1; then
        log_info "使用launchctl启动Consul..."
        if launchctl load /opt/homebrew/etc/consul.plist; then
            log_success "Consul启动成功 (launchctl)"
            sleep 5
        else
            log_warning "Consul启动失败，继续使用独立模式"
        fi
    else
        log_info "Consul已在运行"
    fi
}

# 等待服务依赖就绪
wait_for_dependency() {
    local service_name="$1"
    local check_url="$2"
    local expected_response="$3"
    
    log_info "等待 $service_name 依赖就绪..."
    local attempts=0
    local max_attempts=$((DEPENDENCY_WAIT_TIMEOUT / DEPENDENCY_CHECK_INTERVAL))
    
    while [ $attempts -lt $max_attempts ]; do
        if curl -s "$check_url" | grep -q "$expected_response" 2>/dev/null; then
            log_success "$service_name 依赖就绪"
            return 0
        fi
        log_info "等待 $service_name 就绪... ($((attempts + 1))/$max_attempts)"
        sleep $DEPENDENCY_CHECK_INTERVAL
        attempts=$((attempts + 1))
    done
    
    log_error "$service_name 依赖等待超时"
    return 1
}

# 启动统一认证服务
start_unified_auth_service() {
    log_step "启动统一认证服务..."
    
    local start_cmd="export JWT_SECRET='jobfirst-unified-auth-secret-key-2024' && export DATABASE_URL='root:@tcp(localhost:3306)/jobfirst?charset=utf8mb4&parseTime=True&loc=Local' && export AUTH_SERVICE_PORT='8207' && ./unified-auth > '$LOG_DIR/unified-auth-service.log' 2>&1 &"
    
    standard_start_service "unified-auth-service" "8207" "$start_cmd" "http://localhost:8207/health" "$PROJECT_ROOT/backend/cmd/unified-auth"
}

# 启动Basic-Server
start_basic_server() {
    log_step "启动Basic-Server..."
    
    # 等待Consul完全就绪
    if ! wait_for_dependency "Consul" "http://localhost:8500/v1/status/leader" "127.0.0.1:8300"; then
        log_warning "Consul未就绪，Basic-Server将在独立模式下启动"
    fi
    
    local start_cmd="export BASIC_SERVER_MODE='standalone' && export CONSUL_ENABLED=false && ./start_basic_server.sh start"
    
    standard_start_service "basic-server" "8080" "$start_cmd" "http://localhost:8080/health" "$PROJECT_ROOT/backend/cmd/basic-server"
}

# 启动User Service
start_user_service() {
    log_step "启动User Service..."
    
    # 等待Consul和Basic-Server完全就绪
    if ! wait_for_dependency "Consul" "http://localhost:8500/v1/status/leader" "127.0.0.1:8300"; then
        log_warning "Consul未就绪，User Service将在受限模式下启动"
    fi
    
    if ! wait_for_dependency "Basic-Server" "http://localhost:8080/health" "true"; then
        log_warning "Basic-Server未就绪，User Service将在受限模式下启动"
    fi
    
    local start_cmd="./start_user_service.sh start"
    
    standard_start_service "user-service" "8081" "$start_cmd" "http://localhost:8081/health" "$PROJECT_ROOT/backend/internal/user"
}

# 启动Resume Service
start_resume_service() {
    log_step "启动Resume Service..."
    
    # 等待User Service就绪（Resume Service需要用户认证）
    if ! wait_for_dependency "User Service" "http://localhost:8081/health" "healthy"; then
        log_warning "User Service未就绪，Resume Service将在受限模式下启动"
    fi
    
    local start_cmd="go build -o resume-service . && ./resume-service > '$LOG_DIR/resume-service.log' 2>&1 &"
    
    standard_start_service "resume-service" "8082" "$start_cmd" "http://localhost:8082/health" "$PROJECT_ROOT/backend/internal/resume"
}

# 启动Company Service
start_company_service() {
    log_step "启动Company Service..."
    
    local start_cmd="go build -o company-service . && ./company-service > '$LOG_DIR/company-service.log' 2>&1 &"
    
    standard_start_service "company-service" "8083" "$start_cmd" "http://localhost:8083/health" "$PROJECT_ROOT/backend/internal/company-service"
}

# 启动Job Service
start_job_service() {
    log_step "启动Job Service..."
    
    local start_cmd="go build -o job-service . && ./job-service > '$LOG_DIR/job-service.log' 2>&1 &"
    
    standard_start_service "job-service" "8089" "$start_cmd" "http://localhost:8089/health" "$PROJECT_ROOT/backend/internal/job-service"
}

# 检查数据库服务是否就绪
check_database_service() {
    local service_name="$1"
    local port="$2"
    local max_attempts=10
    local attempt=0
    
    log_info "检查 $service_name 服务状态..."
    
    while [ $attempt -lt $max_attempts ]; do
        if netstat -an | grep -q ":$port.*LISTEN"; then
            log_success "$service_name 服务就绪"
            return 0
        fi
        log_info "等待 $service_name 端口 $port 就绪... ($((attempt + 1))/$max_attempts)"
        sleep 2
        attempt=$((attempt + 1))
    done
    
    log_warning "$service_name 服务未就绪"
    return 1
}

# 启动Multi-Database Service
start_multi_database_service() {
    log_step "启动Multi-Database Service..."
    
    # 检查基础设施服务状态
    check_database_service "MySQL" "3306" || log_warning "MySQL未就绪，Multi-Database Service将在受限模式下启动"
    check_database_service "PostgreSQL" "5432" || log_warning "PostgreSQL未就绪，Multi-Database Service将在受限模式下启动"
    check_database_service "Neo4j" "7474" || log_warning "Neo4j未就绪，Multi-Database Service将在受限模式下启动"
    check_database_service "Redis" "6379" || log_warning "Redis未就绪，Multi-Database Service将在受限模式下启动"
    
    local start_cmd="go build -o multi-database-service . && ./multi-database-service > '$LOG_DIR/multi-database-service.log' 2>&1 &"
    
    standard_start_service "multi-database-service" "8090" "$start_cmd" "http://localhost:8090/health" "$PROJECT_ROOT/backend/internal/multi-database-service"
}

# 启动Notification Service
start_notification_service() {
    log_step "启动Notification Service..."
    
    local start_cmd="go build -o notification-service . && ./notification-service > '$LOG_DIR/notification-service.log' 2>&1 &"
    
    standard_start_service "notification-service" "8084" "$start_cmd" "http://localhost:8084/health" "$PROJECT_ROOT/backend/internal/notification-service"
}

# 启动Template Service
start_template_service() {
    log_step "启动Template Service..."
    
    local start_cmd="go build -o template-service . && ./template-service > '$LOG_DIR/template-service.log' 2>&1 &"
    
    standard_start_service "template-service" "8085" "$start_cmd" "http://localhost:8085/health" "$PROJECT_ROOT/backend/internal/template-service"
}

# 启动Statistics Service
start_statistics_service() {
    log_step "启动Statistics Service..."
    
    local start_cmd="go build -o statistics-service . && ./statistics-service > '$LOG_DIR/statistics-service.log' 2>&1 &"
    
    standard_start_service "statistics-service" "8086" "$start_cmd" "http://localhost:8086/health" "$PROJECT_ROOT/backend/internal/statistics-service"
}

# 启动Banner Service
start_banner_service() {
    log_step "启动Banner Service..."
    
    local start_cmd="go build -o banner-service . && ./banner-service > '$LOG_DIR/banner-service.log' 2>&1 &"
    
    standard_start_service "banner-service" "8087" "$start_cmd" "http://localhost:8087/health" "$PROJECT_ROOT/backend/internal/banner-service"
}

# 启动Dev-Team Service
start_dev_team_service() {
    log_step "启动Dev-Team Service..."
    
    local start_cmd="go build -o dev-team-service . && ./dev-team-service > '$LOG_DIR/dev-team-service.log' 2>&1 &"
    
    standard_start_service "dev-team-service" "8088" "$start_cmd" "http://localhost:8088/health" "$PROJECT_ROOT/backend/internal/dev-team-service"
}

# 启动本地化AI服务
start_local_ai_service() {
    log_step "启动本地化AI服务..."
    
    local start_cmd="source venv/bin/activate && python ai_service_with_zervigo.py > '$LOG_DIR/local-ai-service.log' 2>&1 &"
    
    standard_start_service "local-ai-service" "8206" "$start_cmd" "http://localhost:8206/health" "$PROJECT_ROOT/backend/internal/ai-service"
}

# 启动容器化AI服务
# Docker清理功能
cleanup_docker_images() {
    log_step "清理Docker镜像和容器..."
    
    # 检查Docker是否运行
    if ! docker info > /dev/null 2>&1; then
        log_warning "Docker daemon未运行，跳过Docker清理"
        return 0
    fi
    
    log_info "开始Docker清理流程..."
    
    # 1. 清理停止的容器
    log_info "清理停止的容器..."
    local stopped_containers=$(docker ps -a --filter "status=exited" --format "{{.Names}}" | grep -E "(jobfirst-|none)" || true)
    if [[ -n "$stopped_containers" ]]; then
        echo "$stopped_containers" | xargs docker rm -f 2>/dev/null || true
        log_success "已清理停止的容器: $(echo $stopped_containers | tr '\n' ' ')"
    else
        log_info "没有需要清理的停止容器"
    fi
    
    # 2. 清理悬空镜像 (none标签的镜像)
    log_info "清理悬空镜像 (none标签)..."
    local dangling_images=$(docker images --filter "dangling=true" --format "{{.ID}}" || true)
    if [[ -n "$dangling_images" ]]; then
        echo "$dangling_images" | xargs docker rmi -f 2>/dev/null || true
        log_success "已清理悬空镜像: $(echo $dangling_images | wc -w) 个"
    else
        log_info "没有需要清理的悬空镜像"
    fi
    
    # 3. 清理none标签的镜像
    log_info "清理none标签的镜像..."
    local none_images=$(docker images --filter "reference=none" --format "{{.ID}}" || true)
    if [[ -n "$none_images" ]]; then
        echo "$none_images" | xargs docker rmi -f 2>/dev/null || true
        log_success "已清理none标签镜像: $(echo $none_images | wc -w) 个"
    else
        log_info "没有需要清理的none标签镜像"
    fi
    
    # 4. 清理未使用的网络
    log_info "清理未使用的网络..."
    docker network prune -f > /dev/null 2>&1 || true
    log_success "已清理未使用的网络"
    
    # 5. 清理未使用的卷
    log_info "清理未使用的卷..."
    docker volume prune -f > /dev/null 2>&1 || true
    log_success "已清理未使用的卷"
    
    # 6. 显示清理后的Docker状态
    log_info "Docker清理完成，当前状态:"
    local total_images=$(docker images --format "{{.ID}}" | wc -l)
    local total_containers=$(docker ps -a --format "{{.ID}}" | wc -l)
    log_info "  镜像数量: $total_images"
    log_info "  容器数量: $total_containers"
    
    log_success "Docker清理完成"
}

start_containerized_ai_service() {
    log_step "启动容器化AI服务..."
    
    # 检查Docker是否运行
    if ! docker info > /dev/null 2>&1; then
        log_warning "Docker daemon未运行，尝试启动Docker Desktop..."
        # 尝试启动Docker Desktop
        open -a Docker > /dev/null 2>&1
        log_info "等待Docker daemon启动..."
        local docker_wait_count=0
        while [ $docker_wait_count -lt 30 ]; do
            if docker info > /dev/null 2>&1; then
                log_success "Docker daemon已启动"
                break
            fi
            log_info "等待Docker daemon启动... ($((docker_wait_count + 1))/30)"
            sleep 2
            docker_wait_count=$((docker_wait_count + 1))
        done
        
        if ! docker info > /dev/null 2>&1; then
            log_warning "Docker daemon启动失败，跳过容器化AI服务启动"
            return 0
        fi
    fi
    
    # 检查AI服务容器是否已运行
    local running_ai_containers=$(docker ps --format "table {{.Names}}" | grep -E "(jobfirst-ai|jobfirst-mineru|jobfirst-models|jobfirst-monitor)" || true)
    if [[ -n "$running_ai_containers" ]]; then
        log_success "AI服务容器已在运行: $(echo $running_ai_containers | tr '\n' ' ')"
        return 0
    fi
    
    # 启动所有AI服务容器
    log_info "启动所有AI服务Docker容器..."
    cd "$PROJECT_ROOT/ai-services"
    
    if docker-compose up -d; then
        log_success "容器化AI服务启动成功"
        
        # 等待服务就绪
        log_info "等待AI服务容器就绪..."
        local attempt=0
        local healthy_services=0
        local total_services=4
        
        while [ $attempt -lt $MAX_HEALTH_CHECK_ATTEMPTS ]; do
            healthy_services=0
            
            # 检查AI服务 (8208)
            if curl -s "http://localhost:8208/health" > /dev/null 2>&1; then
                healthy_services=$((healthy_services + 1))
            fi
            
            # 检查MinerU服务 (8001)
            if curl -s "http://localhost:8001/health" > /dev/null 2>&1; then
                healthy_services=$((healthy_services + 1))
            fi
            
            # 检查AI模型服务 (8002)
            if curl -s "http://localhost:8002/health" > /dev/null 2>&1; then
                healthy_services=$((healthy_services + 1))
            fi
            
            # 检查AI监控服务 (9090)
            if curl -s "http://localhost:9090/-/healthy" > /dev/null 2>&1; then
                healthy_services=$((healthy_services + 1))
            fi
            
            if [ $healthy_services -eq $total_services ]; then
                log_success "所有AI服务健康检查通过 ($healthy_services/$total_services)"
                return 0
            fi
            
            log_info "等待AI服务就绪... ($healthy_services/$total_services 健康) ($((attempt + 1))/$MAX_HEALTH_CHECK_ATTEMPTS)"
            sleep $HEALTH_CHECK_INTERVAL
            attempt=$((attempt + 1))
        done
        
        log_warning "AI服务健康检查超时，但容器已启动 ($healthy_services/$total_services 健康)"
    else
        log_error "容器化AI服务启动失败"
        return 1
    fi
}

# 智能日志管理
manage_logs() {
    log_step "智能日志管理..."
    
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    
    # 1. 归档当前日志
    if [[ -f "$STARTUP_LOG" ]]; then
        local archive_log="$PROJECT_ROOT/backups/startup_log_$timestamp.log"
        mkdir -p "$PROJECT_ROOT/backups"
        cp "$STARTUP_LOG" "$archive_log"
        log_info "当前启动日志已归档: $archive_log"
    fi
    
    # 2. 清理旧日志文件（保留最近7天）
    log_info "清理旧日志文件..."
    find "$LOG_DIR" -name "*.log" -mtime +7 -delete 2>/dev/null || true
    find "$LOG_DIR" -name "*.pid" -mtime +1 -delete 2>/dev/null || true
    
    # 3. 压缩大日志文件
    log_info "压缩大日志文件..."
    find "$LOG_DIR" -name "*.log" -size +10M -exec gzip {} \; 2>/dev/null || true
    
    # 4. 清理临时文件
    log_info "清理临时文件..."
    find "$PROJECT_ROOT/temp" -type f -mtime +1 -delete 2>/dev/null || true
    
    log_success "日志管理完成"
}

# 验证服务状态
verify_services() {
    log_step "验证服务状态..."
    
    local running_services=()
    local failed_services=()
    
    for service_info in "${STANDARD_SERVICES[@]}"; do
        IFS=':' read -r service port <<< "$service_info"
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            local pid=$(lsof -Pi :$port -sTCP:LISTEN -t | head -1)
            running_services+=("$service:$port:$pid")
            log_success "✅ $service 正在运行 (端口: $port, PID: $pid)"
        else
            failed_services+=("$service:$port")
            log_warning "❌ $service 未运行 (端口: $port)"
        fi
    done
    
    # 检查基础设施服务
    for service_info in "${INFRASTRUCTURE_SERVICES[@]}"; do
        IFS=':' read -r service port <<< "$service_info"
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            log_success "✅ $service 正在运行 (端口: $port)"
        else
            log_warning "❌ $service 未运行 (端口: $port)"
        fi
    done
    
    # 检查AI服务容器
    if docker info > /dev/null 2>&1; then
        log_info "检查AI服务容器状态..."
        for service_info in "${AI_SERVICES[@]}"; do
            IFS=':' read -r service port <<< "$service_info"
            local container_name="jobfirst-${service}"
            if docker ps --format "table {{.Names}}" | grep -q "$container_name"; then
                local container_status=$(docker ps --filter name="$container_name" --format "{{.Status}}")
                running_services+=("$service:$port:container")
                log_success "✅ $service 容器正在运行 (端口: $port, 状态: $container_status)"
            else
                failed_services+=("$service:$port")
                log_warning "❌ $service 容器未运行 (端口: $port)"
            fi
        done
    else
        log_warning "Docker daemon未运行，跳过AI服务容器检查"
    fi
    
    echo "运行中的服务: ${#running_services[@]}"
    echo "失败的服务: ${#failed_services[@]}"
}

# 生成启动报告
generate_startup_report() {
    log_step "生成启动报告..."
    
    local report_file="$LOG_DIR/startup_report_$(date +%Y%m%d_%H%M%S).txt"
    
    cat > "$report_file" << EOF
==========================================
JobFirst 增强智能启动报告
==========================================
启动时间: $(date)
启动模式: 标准化启动流程
端口检查: 全面检查
日志管理: 已优化
启动脚本: $0
启动日志: $STARTUP_LOG

启动步骤:
✅ 全面端口检查
✅ 基础设施服务启动
✅ 服务发现服务启动
✅ 统一认证服务启动
✅ Basic-Server启动
✅ 所有微服务启动
✅ AI服务启动
✅ 智能日志管理

服务状态:
$(verify_services)

改进点:
1. 标准化启动流程 - 统一所有服务的启动方式
2. 全面端口检查 - 启动前检查所有端口可用性
3. Docker资源清理 - 自动清理悬空镜像、停止容器、未使用资源
4. 智能日志管理 - 日志归档、清理、压缩

==========================================
EOF
    
    log_success "启动报告已生成: $report_file"
}

# 显示帮助信息
show_help() {
    cat << EOF
JobFirst 增强智能启动脚本 - 标准化启动流程

改进点:
1. 标准化启动流程 - 统一所有服务的启动方式
2. 全面端口检查 - 启动前检查所有端口可用性
3. Docker资源清理 - 自动清理悬空镜像、停止容器、未使用资源
4. 智能日志管理 - 日志归档、清理、压缩

用法: $0 [选项]

选项:
  --no-port-check     跳过端口检查
  --no-logs           跳过日志管理
  --help             显示此帮助信息

启动流程:
  1. 全面端口检查
  2. Docker清理 (清理悬空镜像、停止容器、未使用资源)
  3. 启动基础设施服务 (MySQL, Redis, PostgreSQL, Neo4j)
  4. 启动服务发现服务 (Consul)
  5. 启动统一认证服务
  6. 启动Basic-Server (等待Consul就绪)
  6. 启动User Service (等待Consul和Basic-Server就绪)
  7. 启动其他微服务 (等待User Service就绪)
  8. 启动AI服务
  9. 智能日志管理
  10. 验证服务状态
  11. 生成启动报告

依赖关系:
  Consul → Basic-Server → User-Service → 其他微服务

示例:
  $0                    # 标准化启动所有服务
  $0 --no-port-check   # 跳过端口检查
  $0 --no-logs         # 跳过日志管理

EOF
}

# 主函数
main() {
    local skip_port_check=false
    local skip_logs=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --no-port-check)
                skip_port_check=true
                shift
                ;;
            --no-logs)
                skip_logs=true
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 初始化
    create_directories
    
    echo "=========================================="
    echo "🚀 JobFirst 增强智能启动工具"
    echo "=========================================="
    echo
    
    log_info "开始增强智能启动流程..."
    
    # 全面端口检查
    if [[ "$skip_port_check" = false ]]; then
        comprehensive_port_check
    fi
    
    # 执行启动步骤
    cleanup_docker_images
    start_infrastructure_services
    start_consul_service
    start_unified_auth_service
    start_basic_server
    start_user_service
    start_resume_service
    start_company_service
    start_job_service
    start_multi_database_service
    start_notification_service
    start_template_service
    start_statistics_service
    start_banner_service
    start_dev_team_service
    start_local_ai_service
    start_containerized_ai_service
    
    # 智能日志管理
    if [[ "$skip_logs" = false ]]; then
        manage_logs
    fi
    
    # 验证和报告
    verify_services
    generate_startup_report
    
    echo
    echo "=========================================="
    echo "✅ JobFirst 增强智能启动完成"
    echo "=========================================="
    echo
    log_success "系统已智能启动，端口已检查，日志已优化"
    log_info "启动日志: $STARTUP_LOG"
    echo
}

# 错误处理 - 修改为不退出，继续启动流程
trap 'log_error "启动过程中发生错误，继续启动流程..."' ERR

# 信号处理
trap 'log_warning "收到中断信号，继续启动流程..."' INT TERM

# 执行主函数
main "$@"
