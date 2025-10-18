#!/bin/bash

# Basic Server集群开发环境启动脚本
# 用于在单机上启动多个Basic Server实例

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
BACKEND_DIR="$PROJECT_ROOT/basic/backend"
LOGS_DIR="$PROJECT_ROOT/basic/logs"

# 创建日志目录
mkdir -p "$LOGS_DIR"

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        log_error "Go未安装，请先安装Go"
        exit 1
    fi
    
    log_success "依赖检查完成"
}

# 编译Basic Server
build_basic_server() {
    log_info "编译Basic Server..."
    
    cd "$BACKEND_DIR/cmd/basic-server"
    
    if go build -o basic-server main.go; then
        log_success "Basic Server编译成功"
    else
        log_error "Basic Server编译失败"
        exit 1
    fi
}

# 启动Basic Server实例
start_basic_server_instances() {
    log_info "启动Basic Server实例..."
    
    # 停止现有的Basic Server进程
    pkill -f basic-server || true
    sleep 2
    
    # 启动Node 1 (端口8080)
    log_info "启动Basic Server Node 1 (端口8080)..."
    cd "$BACKEND_DIR/cmd/basic-server"
    NODE_ID="basic-server-node-1" \
    NODE_PORT="8080" \
    ./basic-server > "$LOGS_DIR/basic-server-node1.log" 2>&1 &
    NODE1_PID=$!
    echo $NODE1_PID > "$LOGS_DIR/basic-server-node1.pid"
    log_success "Basic Server Node 1 启动成功 (PID: $NODE1_PID)"
    
    # 启动Node 2 (端口8081)
    log_info "启动Basic Server Node 2 (端口8081)..."
    NODE_ID="basic-server-node-2" \
    NODE_PORT="8081" \
    ./basic-server > "$LOGS_DIR/basic-server-node2.log" 2>&1 &
    NODE2_PID=$!
    echo $NODE2_PID > "$LOGS_DIR/basic-server-node2.pid"
    log_success "Basic Server Node 2 启动成功 (PID: $NODE2_PID)"
    
    # 启动Node 3 (端口8082)
    log_info "启动Basic Server Node 3 (端口8082)..."
    NODE_ID="basic-server-node-3" \
    NODE_PORT="8082" \
    ./basic-server > "$LOGS_DIR/basic-server-node3.log" 2>&1 &
    NODE3_PID=$!
    echo $NODE3_PID > "$LOGS_DIR/basic-server-node3.pid"
    log_success "Basic Server Node 3 启动成功 (PID: $NODE3_PID)"
}

# 等待服务启动
wait_for_services() {
    log_info "等待服务启动..."
    
    # 等待Node 1
    for i in {1..30}; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            log_success "Basic Server Node 1 健康检查通过"
            break
        fi
        if [ $i -eq 30 ]; then
            log_error "Basic Server Node 1 启动超时"
            exit 1
        fi
        sleep 1
    done
    
    # 等待Node 2
    for i in {1..30}; do
        if curl -s http://localhost:8081/health > /dev/null 2>&1; then
            log_success "Basic Server Node 2 健康检查通过"
            break
        fi
        if [ $i -eq 30 ]; then
            log_error "Basic Server Node 2 启动超时"
            exit 1
        fi
        sleep 1
    done
    
    # 等待Node 3
    for i in {1..30}; do
        if curl -s http://localhost:8082/health > /dev/null 2>&1; then
            log_success "Basic Server Node 3 健康检查通过"
            break
        fi
        if [ $i -eq 30 ]; then
            log_error "Basic Server Node 3 启动超时"
            exit 1
        fi
        sleep 1
    done
}

# 显示集群状态
show_cluster_status() {
    log_info "集群状态:"
    echo ""
    
    # Node 1状态
    echo "Basic Server Node 1 (端口8080):"
    curl -s http://localhost:8080/api/v1/cluster/status | jq '.status.cluster_id, .status.total_nodes, .status.active_nodes' 2>/dev/null || echo "  状态获取失败"
    echo ""
    
    # Node 2状态
    echo "Basic Server Node 2 (端口8081):"
    curl -s http://localhost:8081/api/v1/cluster/status | jq '.status.cluster_id, .status.total_nodes, .status.active_nodes' 2>/dev/null || echo "  状态获取失败"
    echo ""
    
    # Node 3状态
    echo "Basic Server Node 3 (端口8082):"
    curl -s http://localhost:8082/api/v1/cluster/status | jq '.status.cluster_id, .status.total_nodes, .status.active_nodes' 2>/dev/null || echo "  状态获取失败"
    echo ""
    
    log_success "集群启动完成！"
    echo ""
    echo "访问地址:"
    echo "  Node 1: http://localhost:8080"
    echo "  Node 2: http://localhost:8081"
    echo "  Node 3: http://localhost:8082"
    echo ""
    echo "集群API:"
    echo "  Node 1 集群状态: http://localhost:8080/api/v1/cluster/status"
    echo "  Node 2 集群状态: http://localhost:8081/api/v1/cluster/status"
    echo "  Node 3 集群状态: http://localhost:8082/api/v1/cluster/status"
    echo ""
    echo "日志文件:"
    echo "  Node 1: $LOGS_DIR/basic-server-node1.log"
    echo "  Node 2: $LOGS_DIR/basic-server-node2.log"
    echo "  Node 3: $LOGS_DIR/basic-server-node3.log"
}

# 主函数
main() {
    log_info "开始启动Basic Server集群开发环境..."
    echo ""
    
    check_dependencies
    build_basic_server
    start_basic_server_instances
    wait_for_services
    show_cluster_status
}

# 运行主函数
main "$@"
