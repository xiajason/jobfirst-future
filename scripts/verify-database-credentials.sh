#!/bin/bash

# 数据库密码适配验证脚本
# 用于验证容器化数据库集群的密码配置

set -e

# 颜色输出
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

# 检查Docker容器状态
check_docker_containers() {
    log_info "检查Docker容器状态..."
    
    # 检查MySQL容器
    if docker ps | grep -q "jobfirst-mysql"; then
        log_success "MySQL容器运行正常"
    else
        log_error "MySQL容器未运行"
        return 1
    fi
    
    # 检查PostgreSQL容器
    if docker ps | grep -q "migration-postgres"; then
        log_success "PostgreSQL容器运行正常"
    else
        log_error "PostgreSQL容器未运行"
        return 1
    fi
    
    # 检查MongoDB容器
    if docker ps | grep -q "migration-mongodb"; then
        log_success "MongoDB容器运行正常"
    else
        log_error "MongoDB容器未运行"
        return 1
    fi
    
    # 检查Redis容器
    if docker ps | grep -q "migration-redis"; then
        log_success "Redis容器运行正常"
    else
        log_error "Redis容器未运行"
        return 1
    fi
}

# 获取容器实际密码
get_container_passwords() {
    log_info "获取容器实际密码..."
    
    # MySQL密码
    MYSQL_PASSWORD=$(docker inspect jobfirst-mysql | grep -o 'MYSQL_ROOT_PASSWORD=[^"]*' | cut -d'=' -f2)
    log_info "MySQL实际密码: $MYSQL_PASSWORD"
    
    # PostgreSQL密码
    POSTGRES_PASSWORD=$(docker inspect migration-postgres | grep -o 'POSTGRES_PASSWORD=[^"]*' | cut -d'=' -f2)
    log_info "PostgreSQL实际密码: $POSTGRES_PASSWORD"
    
    # MongoDB密码
    MONGO_PASSWORD=$(docker inspect migration-mongodb | grep -o 'MONGO_INITDB_ROOT_PASSWORD=[^"]*' | cut -d'=' -f2)
    log_info "MongoDB实际密码: $MONGO_PASSWORD"
    
    # Redis密码 (检查是否有密码配置)
    REDIS_PASSWORD=""
    if docker inspect migration-redis | grep -q 'REDIS_PASSWORD'; then
        REDIS_PASSWORD=$(docker inspect migration-redis | grep -o 'REDIS_PASSWORD=[^"]*' | cut -d'=' -f2)
    fi
    log_info "Redis实际密码: ${REDIS_PASSWORD:-'无密码'}"
}

# 测试数据库连接
test_database_connections() {
    log_info "测试数据库连接..."
    
    # 测试MySQL连接
    log_info "测试MySQL连接..."
    if docker exec jobfirst-mysql mysql -uroot -p"$MYSQL_PASSWORD" -e "SELECT 1;" >/dev/null 2>&1; then
        log_success "MySQL连接成功"
    else
        log_error "MySQL连接失败"
        return 1
    fi
    
    # 测试PostgreSQL连接
    log_info "测试PostgreSQL连接..."
    if docker exec migration-postgres psql -U postgres -d jobfirst_future -c "SELECT 1;" >/dev/null 2>&1; then
        log_success "PostgreSQL连接成功"
    else
        log_error "PostgreSQL连接失败"
        return 1
    fi
    
    # 测试MongoDB连接
    log_info "测试MongoDB连接..."
    if docker exec migration-mongodb mongosh --eval "db.runCommand('ping')" >/dev/null 2>&1; then
        log_success "MongoDB连接成功"
    else
        log_error "MongoDB连接失败"
        return 1
    fi
    
    # 测试Redis连接
    log_info "测试Redis连接..."
    if docker exec migration-redis redis-cli ping >/dev/null 2>&1; then
        log_success "Redis连接成功"
    else
        log_error "Redis连接失败"
        return 1
    fi
}

# 生成适配的配置文件
generate_adapted_config() {
    log_info "生成适配的配置文件..."
    
    # 创建适配的.env文件
    cat > configs/.env.adapted << EOF
# 适配容器化数据库集群的配置文件
# 自动生成时间: $(date)

# MySQL Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=$MYSQL_PASSWORD
DB_NAME=jobfirst_future

# PostgreSQL Database Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=$POSTGRES_PASSWORD
POSTGRES_DATABASE=jobfirst_future

# MongoDB Configuration
MONGODB_HOST=localhost
MONGODB_PORT=27017
MONGODB_USER=admin
MONGODB_PASSWORD=$MONGO_PASSWORD
MONGODB_DATABASE=jobfirst_future

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=$REDIS_PASSWORD

# Service Ports Configuration
API_GATEWAY_PORT=8080
USER_SERVICE_PORT=8081
RESUME_SERVICE_PORT=8082
COMPANY_SERVICE_PORT=8083
NOTIFICATION_SERVICE_PORT=8084
TEMPLATE_SERVICE_PORT=8085
STATISTICS_SERVICE_PORT=8086
BANNER_SERVICE_PORT=8087
DEV_TEAM_SERVICE_PORT=8088
JOB_SERVICE_PORT=8089

# JWT Configuration
JWT_SECRET=jobfirst-unified-auth-secret-key-2024

# Log Configuration
LOG_LEVEL=info

# Environment
ENVIRONMENT=production
EOF
    
    log_success "适配配置文件已生成: configs/.env.adapted"
}

# 更新jobfirst-core配置
update_jobfirst_config() {
    log_info "更新jobfirst-core配置..."
    
    # 创建适配的jobfirst-core-config.yaml
    cat > configs/jobfirst-core-config.yaml.adapted << EOF
# JobFirst Core Configuration - 适配容器化数据库集群
# 自动生成时间: $(date)

# 数据库配置
database:
  host: "localhost"
  port: 3306
  username: "root"
  password: "$MYSQL_PASSWORD"
  database: "jobfirst_future"
  charset: "utf8mb4"
  max_idle: 10
  max_open: 100
  max_lifetime: "1h"
  log_level: "warn"

# Redis配置
redis:
  host: "localhost"
  port: 6379
  password: "$REDIS_PASSWORD"
  database: 0
  pool_size: 10

# 服务器配置
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"

# 认证配置
auth:
  jwt_secret: "jobfirst-unified-auth-secret-key-2024"
  token_expiry: "168h"
  refresh_expiry: "720h"
  password_min_length: 6
  max_login_attempts: 5
  lockout_duration: "30m"

# 日志配置
log:
  level: "info"
  format: "json"
  output: "stdout"
  file: "./logs/jobfirst-core.log"
EOF
    
    log_success "jobfirst-core配置已更新: configs/jobfirst-core-config.yaml.adapted"
}

# 主函数
main() {
    log_info "开始数据库密码适配验证..."
    
    # 检查Docker容器
    if ! check_docker_containers; then
        log_error "Docker容器检查失败"
        exit 1
    fi
    
    # 获取容器密码
    get_container_passwords
    
    # 测试数据库连接
    if ! test_database_connections; then
        log_error "数据库连接测试失败"
        exit 1
    fi
    
    # 生成适配配置
    generate_adapted_config
    update_jobfirst_config
    
    log_success "数据库密码适配验证完成！"
    log_info "请使用以下文件替换现有配置："
    log_info "  - configs/.env.adapted → configs/.env"
    log_info "  - configs/jobfirst-core-config.yaml.adapted → configs/jobfirst-core-config.yaml"
}

# 执行主函数
main "$@"
