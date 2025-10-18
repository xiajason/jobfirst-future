#!/bin/bash

# Basic Server集群开发环境停止脚本

set -e

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

# 项目根目录
PROJECT_ROOT="/Users/szjason72/zervi-basic"
LOGS_DIR="$PROJECT_ROOT/basic/logs"

# 停止Basic Server实例
stop_basic_server_instances() {
    log_info "停止Basic Server实例..."
    
    # 停止Node 1
    if [ -f "$LOGS_DIR/basic-server-node1.pid" ]; then
        NODE1_PID=$(cat "$LOGS_DIR/basic-server-node1.pid")
        if kill -0 $NODE1_PID 2>/dev/null; then
            kill $NODE1_PID
            log_success "Basic Server Node 1 已停止 (PID: $NODE1_PID)"
        else
            log_warning "Basic Server Node 1 进程不存在"
        fi
        rm -f "$LOGS_DIR/basic-server-node1.pid"
    else
        log_warning "Basic Server Node 1 PID文件不存在"
    fi
    
    # 停止Node 2
    if [ -f "$LOGS_DIR/basic-server-node2.pid" ]; then
        NODE2_PID=$(cat "$LOGS_DIR/basic-server-node2.pid")
        if kill -0 $NODE2_PID 2>/dev/null; then
            kill $NODE2_PID
            log_success "Basic Server Node 2 已停止 (PID: $NODE2_PID)"
        else
            log_warning "Basic Server Node 2 进程不存在"
        fi
        rm -f "$LOGS_DIR/basic-server-node2.pid"
    else
        log_warning "Basic Server Node 2 PID文件不存在"
    fi
    
    # 停止Node 3
    if [ -f "$LOGS_DIR/basic-server-node3.pid" ]; then
        NODE3_PID=$(cat "$LOGS_DIR/basic-server-node3.pid")
        if kill -0 $NODE3_PID 2>/dev/null; then
            kill $NODE3_PID
            log_success "Basic Server Node 3 已停止 (PID: $NODE3_PID)"
        else
            log_warning "Basic Server Node 3 进程不存在"
        fi
        rm -f "$LOGS_DIR/basic-server-node3.pid"
    else
        log_warning "Basic Server Node 3 PID文件不存在"
    fi
    
    # 强制停止所有basic-server进程
    pkill -f basic-server || true
    sleep 2
    
    log_success "所有Basic Server实例已停止"
}

# 显示停止状态
show_stop_status() {
    log_info "检查停止状态..."
    
    # 检查端口占用
    if lsof -i :8080 > /dev/null 2>&1; then
        log_warning "端口8080仍被占用"
    else
        log_success "端口8080已释放"
    fi
    
    if lsof -i :8081 > /dev/null 2>&1; then
        log_warning "端口8081仍被占用"
    else
        log_success "端口8081已释放"
    fi
    
    if lsof -i :8082 > /dev/null 2>&1; then
        log_warning "端口8082仍被占用"
    else
        log_success "端口8082已释放"
    fi
    
    # 检查进程
    if pgrep -f basic-server > /dev/null; then
        log_warning "仍有basic-server进程在运行"
        pgrep -f basic-server
    else
        log_success "所有basic-server进程已停止"
    fi
}

# 主函数
main() {
    log_info "开始停止Basic Server集群开发环境..."
    echo ""
    
    stop_basic_server_instances
    show_stop_status
    
    echo ""
    log_success "Basic Server集群已完全停止！"
}

# 运行主函数
main "$@"
