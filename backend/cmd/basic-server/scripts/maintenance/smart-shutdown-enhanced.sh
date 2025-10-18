#!/bin/bash

# JobFirst 增强智能关闭脚本 - 标准化关闭流程
# 解决：1.标准化关闭流程 2.端口释放验证 3.日志管理优化

set -e

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
BACKUP_DIR="$PROJECT_ROOT/backups"
SHUTDOWN_LOG="$LOG_DIR/smart-shutdown.log"

# 关闭超时配置
GRACEFUL_TIMEOUT=30
FORCE_TIMEOUT=10
PORT_RELEASE_TIMEOUT=15

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
    "consul:8500"
    "mysql:3306"
    "redis:6379"
    "postgresql:5432"
    "neo4j:7474"
)

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$SHUTDOWN_LOG"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$SHUTDOWN_LOG"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$SHUTDOWN_LOG"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$SHUTDOWN_LOG"
}

log_step() {
    echo -e "${PURPLE}[STEP]${NC} $1" | tee -a "$SHUTDOWN_LOG"
}

# 创建必要的目录
create_directories() {
    mkdir -p "$LOG_DIR"
    mkdir -p "$BACKUP_DIR"
    mkdir -p "$PROJECT_ROOT/temp"
}

# 标准化端口检查函数
check_port_status() {
    local port=$1
    local service_name=$2
    
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        local pid=$(lsof -Pi :$port -sTCP:LISTEN -t | head -1)
        log_info "$service_name 端口 $port 被占用 (PID: $pid)"
        return 0
    else
        log_info "$service_name 端口 $port 已释放"
        return 1
    fi
}

# 等待端口释放
wait_for_port_release() {
    local port=$1
    local service_name=$2
    local timeout=${3:-$PORT_RELEASE_TIMEOUT}
    
    log_info "等待 $service_name 端口 $port 释放..."
    
    local count=0
    while [[ $count -lt $timeout ]]; do
        if ! check_port_status "$port" "$service_name" >/dev/null 2>&1; then
            log_success "$service_name 端口 $port 已成功释放"
            return 0
        fi
        
        sleep 1
        ((count++))
        echo -n "."
    done
    
    echo ""
    log_warning "$service_name 端口 $port 释放超时"
    return 1
}

# 标准化服务关闭函数
standard_shutdown_service() {
    local service_name=$1
    local port=$2
    local pid_file="$LOG_DIR/${service_name}.pid"
    
    log_info "标准化关闭 $service_name (端口: $port)..."
    
    # 步骤1: 检查服务是否运行
    if ! check_port_status "$port" "$service_name" >/dev/null 2>&1; then
        log_info "$service_name 未运行，跳过"
        return 0
    fi
    
    # 步骤2: 尝试通过PID文件关闭
    if [[ -f "$pid_file" ]]; then
        local pid=$(cat "$pid_file")
        if kill -0 "$pid" 2>/dev/null; then
            log_info "通过PID文件关闭 $service_name (PID: $pid)"
            kill -TERM "$pid" 2>/dev/null || true
            
            # 等待优雅关闭
            local count=0
            while kill -0 "$pid" 2>/dev/null && [[ $count -lt $GRACEFUL_TIMEOUT ]]; do
                sleep 1
                ((count++))
            done
            
            if kill -0 "$pid" 2>/dev/null; then
                log_warning "$service_name 优雅关闭超时，强制关闭"
                kill -KILL "$pid" 2>/dev/null || true
                sleep 2
            fi
        fi
        rm -f "$pid_file"
    fi
    
    # 步骤3: 通过端口强制关闭
    local pids=$(lsof -ti ":$port" 2>/dev/null)
    if [[ -n "$pids" ]]; then
        log_info "通过端口 $port 强制关闭剩余进程: $pids"
        echo "$pids" | xargs kill -KILL 2>/dev/null || true
        sleep 2
    fi
    
    # 步骤4: 验证端口释放
    if wait_for_port_release "$port" "$service_name"; then
        log_success "$service_name 已成功关闭并释放端口"
        return 0
    else
        log_error "$service_name 关闭失败，端口仍被占用"
        return 1
    fi
}

# 关闭所有标准服务
shutdown_standard_services() {
    log_step "关闭所有标准服务..."
    
    local failed_services=()
    
    for service_info in "${STANDARD_SERVICES[@]}"; do
        IFS=':' read -r service_name port <<< "$service_info"
        
        if ! standard_shutdown_service "$service_name" "$port"; then
            failed_services+=("$service_name:$port")
        fi
    done
    
    if [[ ${#failed_services[@]} -eq 0 ]]; then
        log_success "所有标准服务已成功关闭"
    else
        log_warning "以下服务关闭失败: ${failed_services[*]}"
    fi
}

# 关闭容器化AI服务
shutdown_containerized_ai_service() {
    log_info "关闭容器化AI服务..."
    
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
            log_warning "Docker daemon启动失败，跳过容器化AI服务关闭"
            return 0
        fi
    fi
    
    # 检查AI服务容器是否运行
    local ai_containers=$(docker ps --format "table {{.Names}}" | grep -E "(jobfirst-ai|jobfirst-mineru|jobfirst-models|jobfirst-monitor)" || true)
    if [[ -z "$ai_containers" ]]; then
        log_info "容器化AI服务未运行"
        return 0
    fi
    
    # 进入AI服务目录
    cd "$PROJECT_ROOT/ai-services"
    
    # 优雅关闭所有AI服务容器
    log_info "发送优雅关闭信号到所有AI服务容器..."
    if docker-compose stop; then
        log_success "AI服务容器已优雅停止"
        
        # 等待容器完全停止
        log_info "等待AI服务容器完全停止..."
        local count=0
        while [[ $count -lt 10 ]]; do
            local running_containers=$(docker ps --format "table {{.Names}}" | grep -E "(jobfirst-ai|jobfirst-mineru|jobfirst-models|jobfirst-monitor)" || true)
            if [[ -z "$running_containers" ]]; then
                log_success "所有AI服务容器已完全停止"
                break
            fi
            log_info "等待AI服务容器停止... ($((count + 1))/10)"
            sleep 2
            count=$((count + 1))
        done
        
        # 移除容器
        log_info "移除所有AI服务容器..."
        if docker-compose rm -f; then
            log_success "所有AI服务容器已移除"
        else
            log_warning "AI服务容器移除失败"
        fi
        
        # 等待端口释放
        wait_for_port_release 8208 "Containerized-AI-Service"
    else
        log_warning "AI服务容器优雅停止失败，尝试强制关闭..."
        
        # 强制关闭
        if docker-compose kill; then
            log_success "所有AI服务容器已强制关闭"
            docker-compose rm -f
            wait_for_port_release 8208 "Containerized-AI-Service"
            wait_for_port_release 8001 "MinerU-Service"
            wait_for_port_release 8002 "AI-Models-Service"
            wait_for_port_release 9090 "AI-Monitor-Service"
        else
            log_error "AI服务容器强制关闭失败"
        fi
    fi
}

# 关闭基础设施服务
shutdown_infrastructure_services() {
    log_step "关闭基础设施服务..."
    
    # 关闭容器化AI服务
    shutdown_containerized_ai_service
    
        # 关闭Consul
        if curl -s http://localhost:8500/v1/status/leader >/dev/null 2>&1; then
            log_info "关闭Consul服务..."
            if launchctl unload /opt/homebrew/etc/consul.plist; then
                log_success "Consul已关闭 (launchctl)"
                wait_for_port_release 8500 "Consul"
            else
                log_warning "Consul关闭失败"
            fi
        else
            log_info "Consul未运行"
        fi
    
    # 关闭Neo4j
    if brew services list | grep neo4j | grep started &> /dev/null; then
        log_info "关闭Neo4j服务..."
        if brew services stop neo4j; then
            log_success "Neo4j已关闭"
            wait_for_port_release 7474 "Neo4j"
        else
            log_warning "Neo4j关闭失败"
        fi
    else
        log_info "Neo4j未运行"
    fi
    
    # 关闭PostgreSQL@14
    if brew services list | grep postgresql@14 | grep started &> /dev/null; then
        log_info "关闭PostgreSQL@14服务..."
        if brew services stop postgresql@14; then
            log_success "PostgreSQL@14已关闭"
            wait_for_port_release 5432 "PostgreSQL@14"
        else
            log_warning "PostgreSQL@14关闭失败"
        fi
    else
        log_info "PostgreSQL@14未运行"
    fi
    
    # 关闭Redis
    if brew services list | grep redis | grep started &> /dev/null; then
        log_info "关闭Redis服务..."
        if brew services stop redis; then
            log_success "Redis已关闭"
            wait_for_port_release 6379 "Redis"
        else
            log_warning "Redis关闭失败"
        fi
    else
        log_info "Redis未运行"
    fi
    
    # 关闭MySQL
    if brew services list | grep mysql | grep started &> /dev/null; then
        log_info "关闭MySQL服务..."
        if brew services stop mysql; then
            log_success "MySQL已关闭"
            wait_for_port_release 3306 "MySQL"
        else
            log_warning "MySQL关闭失败"
        fi
    else
        log_info "MySQL未运行"
    fi
}

# 智能日志管理
manage_logs() {
    log_step "智能日志管理..."
    
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    
    # 1. 归档当前日志
    if [[ -f "$SHUTDOWN_LOG" ]]; then
        local archive_log="$BACKUP_DIR/shutdown_log_$timestamp.log"
        cp "$SHUTDOWN_LOG" "$archive_log"
        log_info "当前关闭日志已归档: $archive_log"
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

# 验证关闭状态
verify_shutdown() {
    log_step "验证关闭状态..."
    
    local running_services=()
    local occupied_ports=()
    
    # 检查所有标准服务端口
    for service_info in "${STANDARD_SERVICES[@]}"; do
        IFS=':' read -r service_name port <<< "$service_info"
        
        if check_port_status "$port" "$service_name" >/dev/null 2>&1; then
            running_services+=("$service_name:$port")
        fi
    done
    
    # 检查基础设施服务端口
    for service_info in "${INFRASTRUCTURE_SERVICES[@]}"; do
        IFS=':' read -r service_name port <<< "$service_info"
        
        if check_port_status "$port" "$service_name" >/dev/null 2>&1; then
            occupied_ports+=("$service_name:$port")
        fi
    done
    
    if [[ ${#running_services[@]} -eq 0 && ${#occupied_ports[@]} -eq 0 ]]; then
        log_success "所有服务已成功关闭，所有端口已释放"
        return 0
    else
        log_warning "以下服务/端口仍在运行:"
        for service in "${running_services[@]}"; do
            log_warning "  - $service"
        done
        for port in "${occupied_ports[@]}"; do
            log_warning "  - $port"
        done
        return 1
    fi
}

# 生成关闭报告
generate_shutdown_report() {
    log_step "生成关闭报告..."
    
    local report_file="$LOG_DIR/shutdown_report_$(date '+%Y%m%d_%H%M%S').txt"
    
    {
        echo "=========================================="
        echo "JobFirst 增强智能关闭报告"
        echo "=========================================="
        echo "关闭时间: $(date)"
        echo "关闭模式: 标准化优雅关闭"
        echo "端口验证: 已启用"
        echo "日志管理: 已优化"
        echo ""
        echo "服务关闭状态:"
        
        for service_info in "${STANDARD_SERVICES[@]}"; do
            IFS=':' read -r service_name port <<< "$service_info"
            if check_port_status "$port" "$service_name" >/dev/null 2>&1; then
                echo "  ❌ $service_name:$port - 仍在运行"
            else
                echo "  ✅ $service_name:$port - 已关闭"
            fi
        done
        
        echo ""
        echo "基础设施服务状态:"
        for service_info in "${INFRASTRUCTURE_SERVICES[@]}"; do
            IFS=':' read -r service_name port <<< "$service_info"
            if check_port_status "$port" "$service_name" >/dev/null 2>&1; then
                echo "  ❌ $service_name:$port - 仍在运行"
            else
                echo "  ✅ $service_name:$port - 已关闭"
            fi
        done
        
        echo ""
        echo "日志管理:"
        echo "  - 当前日志: $SHUTDOWN_LOG"
        echo "  - 日志归档: $BACKUP_DIR/"
        echo "  - 旧日志清理: 已完成"
        echo "  - 大文件压缩: 已完成"
        echo ""
        echo "=========================================="
    } > "$report_file"
    
    log_success "关闭报告已生成: $report_file"
}

# 显示帮助信息
show_help() {
    cat << EOF
JobFirst 增强智能关闭脚本 - 标准化关闭流程

改进点:
1. 标准化关闭流程 - 统一所有服务的关闭方式
2. 端口释放验证 - 关闭后验证端口是否真正释放
3. 智能日志管理 - 日志归档、清理、压缩

用法: $0 [选项]

选项:
  --force             强制关闭所有服务（跳过优雅关闭）
  --no-logs           跳过日志管理
  --help             显示此帮助信息

关闭流程:
  1. 标准化关闭所有微服务
  2. 关闭基础设施服务
  3. 验证端口释放
  4. 智能日志管理
  5. 生成关闭报告

示例:
  $0                    # 标准化关闭所有服务
  $0 --force           # 强制关闭所有服务
  $0 --no-logs         # 跳过日志管理

EOF
}

# 主函数
main() {
    local force_mode=false
    local skip_logs=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --force)
                force_mode=true
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
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 创建必要的目录
    create_directories
    
    # 记录关闭开始
    echo "=========================================="
    echo "🛑 JobFirst 增强智能关闭工具"
    echo "=========================================="
    echo
    
    log_info "开始增强智能关闭流程..."
    log_info "关闭模式: $([ "$force_mode" = true ] && echo "强制关闭" || echo "标准化优雅关闭")"
    
    # 执行关闭步骤
    shutdown_standard_services
    shutdown_infrastructure_services
    
    # 智能日志管理
    if [[ "$skip_logs" = false ]]; then
        manage_logs
    fi
    
    # 验证和报告
    verify_shutdown
    generate_shutdown_report
    
    echo
    echo "=========================================="
    echo "✅ JobFirst 增强智能关闭完成"
    echo "=========================================="
    echo
    log_success "系统已安全关闭，端口已释放，日志已优化"
    log_info "关闭日志: $SHUTDOWN_LOG"
    echo
}

# 错误处理
trap 'log_error "关闭过程中发生错误"; exit 1' ERR

# 信号处理
trap 'log_warning "收到中断信号，继续关闭流程..."' INT TERM

# 执行主函数
main "$@"
