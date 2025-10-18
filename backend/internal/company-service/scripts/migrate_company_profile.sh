#!/bin/bash

# Company服务企业画像数据库迁移脚本
# 用于创建企业画像相关的数据库表结构

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目配置
PROJECT_ROOT="/Users/szjason72/zervi-basic/basic"
MIGRATION_DIR="$PROJECT_ROOT/backend/internal/company-service/migrations"
LOG_DIR="$PROJECT_ROOT/logs"
MIGRATION_LOG="$LOG_DIR/company_profile_migration.log"

# 数据库配置
DB_HOST="localhost"
DB_PORT="3306"
DB_NAME="jobfirst"
DB_USER="root"
DB_PASSWORD=""

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$MIGRATION_LOG"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$MIGRATION_LOG"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$MIGRATION_LOG"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$MIGRATION_LOG"
}

# 创建必要的目录
create_directories() {
    mkdir -p "$LOG_DIR"
    log_info "创建日志目录: $LOG_DIR"
}

# 检查数据库连接
check_database_connection() {
    log_info "检查数据库连接..."
    
    if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "USE $DB_NAME;" 2>/dev/null; then
        log_success "数据库连接成功"
        return 0
    else
        log_error "数据库连接失败"
        return 1
    fi
}

# 检查表是否存在
check_table_exists() {
    local table_name=$1
    local result=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -D"$DB_NAME" -e "SHOW TABLES LIKE '$table_name';" 2>/dev/null | wc -l)
    
    if [ "$result" -gt 1 ]; then
        return 0  # 表存在
    else
        return 1  # 表不存在
    fi
}

# 执行SQL文件
execute_sql_file() {
    local sql_file=$1
    local description=$2
    
    log_info "执行 $description..."
    
    if [ ! -f "$sql_file" ]; then
        log_error "SQL文件不存在: $sql_file"
        return 1
    fi
    
    if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -D"$DB_NAME" < "$sql_file" 2>>"$MIGRATION_LOG"; then
        log_success "$description 执行成功"
        return 0
    else
        log_error "$description 执行失败"
        return 1
    fi
}

# 备份现有数据
backup_existing_data() {
    log_info "备份现有企业数据..."
    
    local backup_file="$LOG_DIR/company_data_backup_$(date +%Y%m%d_%H%M%S).sql"
    
    if mysqldump -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" companies > "$backup_file" 2>/dev/null; then
        log_success "企业数据备份完成: $backup_file"
    else
        log_warning "企业数据备份失败，继续执行迁移"
    fi
}

# 执行迁移
run_migration() {
    log_info "开始执行企业画像数据库迁移..."
    
    # 检查并执行迁移文件
    local migration_files=(
        "001_create_company_documents.sql:企业文档表"
        "002_create_company_parsing_tasks.sql:企业解析任务表"
        "003_create_company_structured_data.sql:企业结构化数据表"
        "004_alter_companies_table.sql:企业表结构更新"
        "005_create_company_profile_tables.sql:企业画像表"
    )
    
    for migration_entry in "${migration_files[@]}"; do
        local file_name=$(echo "$migration_entry" | cut -d':' -f1)
        local description=$(echo "$migration_entry" | cut -d':' -f2)
        local file_path="$MIGRATION_DIR/$file_name"
        
        if [ -f "$file_path" ]; then
            execute_sql_file "$file_path" "$description"
        else
            log_warning "迁移文件不存在: $file_path"
        fi
    done
}

# 验证迁移结果
verify_migration() {
    log_info "验证迁移结果..."
    
    local tables=(
        "company_documents"
        "company_parsing_tasks"
        "company_structured_data"
        "company_basic_info"
        "qualification_license"
        "personnel_competitiveness"
        "provident_fund"
        "subsidy_info"
        "company_relationships"
        "tech_innovation_score"
        "company_financial_info"
        "company_risk_info"
    )
    
    local success_count=0
    local total_count=${#tables[@]}
    
    for table in "${tables[@]}"; do
        if check_table_exists "$table"; then
            log_success "✅ 表 $table 创建成功"
            ((success_count++))
        else
            log_error "❌ 表 $table 创建失败"
        fi
    done
    
    echo
    log_info "迁移验证结果: $success_count/$total_count 个表创建成功"
    
    if [ "$success_count" -eq "$total_count" ]; then
        log_success "所有表创建成功！"
        return 0
    else
        log_error "部分表创建失败，请检查日志"
        return 1
    fi
}

# 生成迁移报告
generate_migration_report() {
    log_info "生成迁移报告..."
    
    local report_file="$LOG_DIR/company_profile_migration_report_$(date +%Y%m%d_%H%M%S).txt"
    
    cat > "$report_file" << EOF
==========================================
Company服务企业画像数据库迁移报告
==========================================
迁移时间: $(date)
数据库: $DB_NAME@$DB_HOST:$DB_PORT
迁移脚本: $0

迁移内容:
✅ 企业文档表 (company_documents)
✅ 企业解析任务表 (company_parsing_tasks)
✅ 企业结构化数据表 (company_structured_data)
✅ 企业画像基本信息表 (company_basic_info)
✅ 资质许可表 (qualification_license)
✅ 人员竞争力表 (personnel_competitiveness)
✅ 公积金信息表 (provident_fund)
✅ 资助补贴表 (subsidy_info)
✅ 企业关系图谱表 (company_relationships)
✅ 科创评分表 (tech_innovation_score)
✅ 企业财务信息表 (company_financial_info)
✅ 企业风险信息表 (company_risk_info)

表结构特点:
- 支持企业画像完整数据存储
- 兼容PDF解析结果存储
- 支持JSON格式的复杂数据
- 完整的索引优化
- 外键约束保证数据一致性

详细日志: $MIGRATION_LOG
==========================================
EOF
    
    log_success "迁移报告已生成: $report_file"
}

# 显示帮助信息
show_help() {
    cat << EOF
Company服务企业画像数据库迁移脚本

用法: $0 [选项]

选项:
  --help             显示此帮助信息
  --check            仅检查数据库连接
  --backup           仅备份现有数据
  --migrate          执行迁移
  --verify           验证迁移结果
  --full             执行完整迁移流程

环境变量:
  DB_HOST            数据库主机 (默认: localhost)
  DB_PORT            数据库端口 (默认: 3306)
  DB_NAME            数据库名称 (默认: jobfirst)
  DB_USER            数据库用户 (默认: root)
  DB_PASSWORD        数据库密码 (默认: 空)

示例:
  $0 --check          # 检查数据库连接
  $0 --full           # 执行完整迁移流程
  $0 --migrate        # 仅执行迁移
  $0 --verify         # 仅验证迁移结果

EOF
}

# 主函数
main() {
    # 解析命令行参数
    case "${1:-}" in
        --help)
            show_help
            exit 0
            ;;
        --check)
            create_directories
            check_database_connection
            ;;
        --backup)
            create_directories
            check_database_connection && backup_existing_data
            ;;
        --migrate)
            create_directories
            check_database_connection && run_migration
            ;;
        --verify)
            create_directories
            verify_migration
            ;;
        --full)
            create_directories
            echo "=========================================="
            echo "🚀 Company服务企业画像数据库迁移"
            echo "=========================================="
            echo
            
            if check_database_connection; then
                backup_existing_data
                run_migration
                verify_migration
                generate_migration_report
                
                echo
                echo "=========================================="
                echo "✅ 企业画像数据库迁移完成"
                echo "=========================================="
                echo
                log_success "迁移完成，详细结果请查看迁移报告"
            else
                log_error "数据库连接失败，迁移终止"
                exit 1
            fi
            ;;
        *)
            show_help
            exit 1
            ;;
    esac
}

# 错误处理
trap 'log_error "迁移过程中发生错误"; exit 1' ERR

# 执行主函数
main "$@"
