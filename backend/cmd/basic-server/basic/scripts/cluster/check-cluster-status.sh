#!/bin/bash

# Basic Server集群状态检查脚本

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

# 检查单个节点状态
check_node_status() {
    local node_id=$1
    local port=$2
    local url="http://localhost:$port"
    
    echo "检查 $node_id (端口 $port):"
    
    # 检查健康状态
    if curl -s "$url/health" > /dev/null 2>&1; then
        log_success "  健康检查: 通过"
        
        # 获取集群状态
        cluster_status=$(curl -s "$url/api/v1/cluster/status" 2>/dev/null)
        if [ $? -eq 0 ]; then
            cluster_id=$(echo "$cluster_status" | jq -r '.status.cluster_id' 2>/dev/null)
            total_nodes=$(echo "$cluster_status" | jq -r '.status.total_nodes' 2>/dev/null)
            active_nodes=$(echo "$cluster_status" | jq -r '.status.active_nodes' 2>/dev/null)
            
            echo "  集群ID: $cluster_id"
            echo "  总节点数: $total_nodes"
            echo "  活跃节点数: $active_nodes"
        else
            log_warning "  集群状态: 获取失败"
        fi
        
        # 获取节点信息
        node_info=$(curl -s "$url/api/v1/cluster/nodes" 2>/dev/null)
        if [ $? -eq 0 ]; then
            node_count=$(echo "$node_info" | jq -r '.nodes | length' 2>/dev/null)
            echo "  注册节点数: $node_count"
        else
            log_warning "  节点信息: 获取失败"
        fi
        
    else
        log_error "  健康检查: 失败"
    fi
    
    echo ""
}

# 检查进程状态
check_process_status() {
    log_info "检查进程状态..."
    
    # 检查Node 1进程
    if [ -f "/Users/szjason72/zervi-basic/basic/logs/basic-server-node1.pid" ]; then
        NODE1_PID=$(cat "/Users/szjason72/zervi-basic/basic/logs/basic-server-node1.pid")
        if kill -0 $NODE1_PID 2>/dev/null; then
            log_success "Basic Server Node 1 进程运行中 (PID: $NODE1_PID)"
        else
            log_error "Basic Server Node 1 进程不存在"
        fi
    else
        log_warning "Basic Server Node 1 PID文件不存在"
    fi
    
    # 检查Node 2进程
    if [ -f "/Users/szjason72/zervi-basic/basic/logs/basic-server-node2.pid" ]; then
        NODE2_PID=$(cat "/Users/szjason72/zervi-basic/basic/logs/basic-server-node2.pid")
        if kill -0 $NODE2_PID 2>/dev/null; then
            log_success "Basic Server Node 2 进程运行中 (PID: $NODE2_PID)"
        else
            log_error "Basic Server Node 2 进程不存在"
        fi
    else
        log_warning "Basic Server Node 2 PID文件不存在"
    fi
    
    # 检查Node 3进程
    if [ -f "/Users/szjason72/zervi-basic/basic/logs/basic-server-node3.pid" ]; then
        NODE3_PID=$(cat "/Users/szjason72/zervi-basic/basic/logs/basic-server-node3.pid")
        if kill -0 $NODE3_PID 2>/dev/null; then
            log_success "Basic Server Node 3 进程运行中 (PID: $NODE3_PID)"
        else
            log_error "Basic Server Node 3 进程不存在"
        fi
    else
        log_warning "Basic Server Node 3 PID文件不存在"
    fi
    
    echo ""
}

# 检查端口占用
check_port_status() {
    log_info "检查端口占用..."
    
    for port in 8080 8081 8082; do
        if lsof -i :$port > /dev/null 2>&1; then
            process=$(lsof -i :$port | tail -n +2 | awk '{print $1, $2}')
            log_success "端口 $port: 被占用 ($process)"
        else
            log_warning "端口 $port: 未被占用"
        fi
    done
    
    echo ""
}

# 显示集群概览
show_cluster_overview() {
    log_info "集群概览..."
    
    echo "┌─────────────────────────────────────────────────────────┐"
    echo "│                Basic Server 集群状态                    │"
    echo "├─────────────────────────────────────────────────────────┤"
    echo "│  Node 1: http://localhost:8080                         │"
    echo "│  Node 2: http://localhost:8081                         │"
    echo "│  Node 3: http://localhost:8082                         │"
    echo "├─────────────────────────────────────────────────────────┤"
    echo "│  集群API:                                               │"
    echo "│  • 健康检查: /health                                   │"
    echo "│  • 集群状态: /api/v1/cluster/status                    │"
    echo "│  • 节点信息: /api/v1/cluster/nodes                     │"
    echo "└─────────────────────────────────────────────────────────┘"
    echo ""
}

# 主函数
main() {
    log_info "Basic Server集群状态检查"
    echo ""
    
    show_cluster_overview
    check_process_status
    check_port_status
    check_node_status "Basic Server Node 1" "8080"
    check_node_status "Basic Server Node 2" "8081"
    check_node_status "Basic Server Node 3" "8082"
    
    log_success "状态检查完成！"
}

# 运行主函数
main "$@"
