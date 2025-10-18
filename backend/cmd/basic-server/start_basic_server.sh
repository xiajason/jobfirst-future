#!/bin/bash

# Basic-Server 启动脚本
# 确保正确的进程名称和PID文件管理

set -e

# 配置
SERVICE_NAME="basic-server"
SERVICE_PORT="8080"
PROJECT_ROOT="/Users/szjason72/zervi-basic/basic"
LOG_DIR="$PROJECT_ROOT/logs"
PID_FILE="$LOG_DIR/${SERVICE_NAME}.pid"
LOG_FILE="$LOG_DIR/${SERVICE_NAME}.log"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查服务是否已运行
check_service_running() {
    if [[ -f "$PID_FILE" ]]; then
        local pid=$(cat "$PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            log_warning "$SERVICE_NAME 已在运行 (PID: $pid)"
            return 0
        else
            log_info "清理过期的PID文件"
            rm -f "$PID_FILE"
        fi
    fi
    
    if lsof -i ":$SERVICE_PORT" >/dev/null 2>&1; then
        log_error "端口 $SERVICE_PORT 已被占用"
        return 1
    fi
    
    return 0
}

# 启动服务
start_service() {
    log_info "启动 $SERVICE_NAME 服务..."
    
    # 创建日志目录
    mkdir -p "$LOG_DIR"
    
    # 切换到服务目录
    cd "$PROJECT_ROOT/backend/cmd/basic-server"
    
    # 设置环境变量
    export SERVICE_NAME="$SERVICE_NAME"
    export SERVICE_PORT="$SERVICE_PORT"
    
    # 启动服务并重定向输出
    nohup ./basic-server > "$LOG_FILE" 2>&1 &
    local service_pid=$!
    
    # 保存PID
    echo $service_pid > "$PID_FILE"
    
    # 等待服务启动
    log_info "等待服务启动..."
    sleep 3
    
    # 验证服务是否启动成功
    if kill -0 "$service_pid" 2>/dev/null; then
        log_success "$SERVICE_NAME 启动成功 (PID: $service_pid)"
        
        # 健康检查
        local health_check_url="http://localhost:$SERVICE_PORT/health"
        local max_attempts=10
        local attempt=1
        
        while [ $attempt -le $max_attempts ]; do
            if curl -s "$health_check_url" >/dev/null 2>&1; then
                log_success "$SERVICE_NAME 健康检查通过"
                return 0
            fi
            
            log_info "健康检查尝试 $attempt/$max_attempts..."
            sleep 2
            attempt=$((attempt + 1))
        done
        
        log_warning "$SERVICE_NAME 健康检查超时，但服务已启动"
        return 0
    else
        log_error "$SERVICE_NAME 启动失败"
        rm -f "$PID_FILE"
        return 1
    fi
}

# 停止服务
stop_service() {
    log_info "停止 $SERVICE_NAME 服务..."
    
    if [[ -f "$PID_FILE" ]]; then
        local pid=$(cat "$PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            log_info "发送SIGTERM信号到 $SERVICE_NAME (PID: $pid)"
            kill -TERM "$pid"
            
            # 等待优雅关闭
            local max_wait=10
            local wait_time=0
            
            while [ $wait_time -lt $max_wait ]; do
                if ! kill -0 "$pid" 2>/dev/null; then
                    log_success "$SERVICE_NAME 已优雅关闭"
                    rm -f "$PID_FILE"
                    return 0
                fi
                
                sleep 1
                wait_time=$((wait_time + 1))
            done
            
            # 强制关闭
            log_warning "强制关闭 $SERVICE_NAME"
            kill -KILL "$pid" 2>/dev/null || true
            rm -f "$PID_FILE"
            log_success "$SERVICE_NAME 已强制关闭"
        else
            log_warning "PID文件存在但进程不存在，清理PID文件"
            rm -f "$PID_FILE"
        fi
    else
        log_warning "PID文件不存在，尝试通过端口关闭"
        local pids=$(lsof -ti ":$SERVICE_PORT" 2>/dev/null)
        if [[ -n "$pids" ]]; then
            echo "$pids" | xargs kill -TERM 2>/dev/null || true
            sleep 2
            echo "$pids" | xargs kill -KILL 2>/dev/null || true
            log_success "$SERVICE_NAME 已通过端口关闭"
        else
            log_info "$SERVICE_NAME 未运行"
        fi
    fi
}

# 显示服务状态
show_status() {
    log_info "$SERVICE_NAME 服务状态:"
    
    if [[ -f "$PID_FILE" ]]; then
        local pid=$(cat "$PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            echo "  ✅ 运行中 (PID: $pid)"
            
            # 显示进程信息
            echo "  📊 进程信息:"
            ps -p "$pid" -o pid,ppid,pcpu,pmem,etime,command 2>/dev/null || echo "    无法获取进程信息"
            
            # 健康检查
            local health_check_url="http://localhost:$SERVICE_PORT/health"
            if curl -s "$health_check_url" >/dev/null 2>&1; then
                echo "  🏥 健康状态: 正常"
            else
                echo "  🏥 健康状态: 异常"
            fi
        else
            echo "  ❌ PID文件存在但进程不存在"
        fi
    else
        if lsof -i ":$SERVICE_PORT" >/dev/null 2>&1; then
            echo "  ⚠️  端口被占用但PID文件不存在"
            local pids=$(lsof -ti ":$SERVICE_PORT")
            echo "    占用进程: $pids"
        else
            echo "  ❌ 未运行"
        fi
    fi
}

# 显示帮助信息
show_help() {
    echo "Basic-Server 启动脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  start     启动服务"
    echo "  stop      停止服务"
    echo "  restart   重启服务"
    echo "  status    显示服务状态"
    echo "  logs      显示服务日志"
    echo "  help      显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 start     # 启动服务"
    echo "  $0 stop      # 停止服务"
    echo "  $0 status    # 查看状态"
}

# 显示日志
show_logs() {
    if [[ -f "$LOG_FILE" ]]; then
        log_info "显示 $SERVICE_NAME 日志 (最后50行):"
        echo "----------------------------------------"
        tail -n 50 "$LOG_FILE"
        echo "----------------------------------------"
    else
        log_warning "日志文件不存在: $LOG_FILE"
    fi
}

# 主函数
main() {
    case "${1:-help}" in
        start)
            if check_service_running; then
                start_service
            else
                exit 1
            fi
            ;;
        stop)
            stop_service
            ;;
        restart)
            stop_service
            sleep 2
            start_service
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
