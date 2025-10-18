#!/bin/bash

# 企业数据关联验证和查询脚本
# 功能：验证企业画像表之间的数据关联，提供查询功能

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

# 数据库连接信息
DB_HOST="localhost"
DB_USER="root"
DB_NAME="jobfirst"

echo "=========================================="
echo "🔍 企业数据关联验证和查询工具"
echo "=========================================="

# 1. 检查数据库连接
log_info "检查数据库连接..."
if ! mysql -h "$DB_HOST" -u "$DB_USER" -e "USE $DB_NAME;" 2>/dev/null; then
    log_error "无法连接到数据库 $DB_NAME"
    exit 1
fi
log_success "数据库连接正常"

# 2. 验证数据关联完整性
log_info "验证数据关联完整性..."

# 检查企业基础信息与财务信息的关联
BASIC_FINANCIAL_MATCH=$(mysql -h "$DB_HOST" -u "$DB_USER" -e "
USE $DB_NAME;
SELECT COUNT(*) as match_count
FROM company_basic_info cbi
LEFT JOIN company_financial_info cfi ON cbi.company_id = cfi.company_id AND cbi.report_id = cfi.report_id
WHERE cbi.company_id = 1 AND cfi.company_id IS NOT NULL;
" 2>/dev/null | tail -n 1)

log_info "企业基础信息与财务信息关联匹配数: $BASIC_FINANCIAL_MATCH"

# 检查企业基础信息与风险信息的关联
BASIC_RISK_MATCH=$(mysql -h "$DB_HOST" -u "$DB_USER" -e "
USE $DB_NAME;
SELECT COUNT(*) as match_count
FROM company_basic_info cbi
LEFT JOIN company_risk_info cri ON cbi.company_id = cri.company_id AND cbi.report_id = cri.report_id
WHERE cbi.company_id = 1 AND cri.company_id IS NOT NULL;
" 2>/dev/null | tail -n 1)

log_info "企业基础信息与风险信息关联匹配数: $BASIC_RISK_MATCH"

# 检查企业基础信息与关系信息的关联
BASIC_RELATIONSHIP_MATCH=$(mysql -h "$DB_HOST" -u "$DB_USER" -e "
USE $DB_NAME;
SELECT COUNT(*) as match_count
FROM company_basic_info cbi
LEFT JOIN company_relationships cr ON cbi.company_id = cr.company_id AND cbi.report_id = cr.report_id
WHERE cbi.company_id = 1 AND cr.company_id IS NOT NULL;
" 2>/dev/null | tail -n 1)

log_info "企业基础信息与关系信息关联匹配数: $BASIC_RELATIONSHIP_MATCH"

# 3. 生成企业完整画像报告
log_info "生成企业完整画像报告..."

mysql -h "$DB_HOST" -u "$DB_USER" -e "
USE $DB_NAME;

-- 企业完整画像查询
SELECT 
    '=== 企业完整画像报告 ===' as report_section,
    '' as company_name,
    '' as industry_category,
    '' as business_status,
    '' as registered_capital,
    '' as insured_count

UNION ALL

SELECT 
    '【企业基础信息】' as report_section,
    company_name,
    industry_category,
    business_status,
    CONCAT(registered_capital, ' ', currency) as registered_capital,
    CONCAT(insured_count, '人') as insured_count
FROM company_basic_info 
WHERE company_id = 1

UNION ALL

SELECT 
    '【企业财务信息】' as report_section,
    CONCAT('总资产: ', total_assets, '万元') as company_name,
    CONCAT('年营收: ', annual_revenue, '万元') as industry_category,
    CONCAT('净利润: ', net_profit, '万元') as business_status,
    CONCAT('融资状态: ', financing_status) as registered_capital,
    CONCAT('上市状态: ', listing_status) as insured_count
FROM company_financial_info 
WHERE company_id = 1

UNION ALL

SELECT 
    '【企业风险信息】' as report_section,
    CONCAT('风险等级: ', risk_level) as company_name,
    CONCAT('信用评级: ', credit_rating) as industry_category,
    '法律风险' as business_status,
    '财务风险' as registered_capital,
    '运营风险' as insured_count
FROM company_risk_info 
WHERE company_id = 1

UNION ALL

SELECT 
    '【企业关系信息】' as report_section,
    CONCAT('关系企业数: ', COUNT(*), '家') as company_name,
    CONCAT('控股关系: ', SUM(CASE WHEN relationship_type = '控股' THEN 1 ELSE 0 END), '家') as industry_category,
    CONCAT('投资关系: ', SUM(CASE WHEN relationship_type = '投资' THEN 1 ELSE 0 END), '家') as business_status,
    CONCAT('合作关系: ', SUM(CASE WHEN relationship_type = '合作' THEN 1 ELSE 0 END), '家') as registered_capital,
    CONCAT('总投资额: ', COALESCE(SUM(investment_amount), 0), '万元') as insured_count
FROM company_relationships 
WHERE company_id = 1;
" > /tmp/company_complete_profile.txt

if [ $? -eq 0 ]; then
    log_success "企业完整画像报告生成成功"
    echo "=========================================="
    echo "📊 企业完整画像报告"
    echo "=========================================="
    cat /tmp/company_complete_profile.txt
    echo "=========================================="
else
    log_error "企业完整画像报告生成失败"
fi

# 4. 生成数据关联分析报告
log_info "生成数据关联分析报告..."

mysql -h "$DB_HOST" -u "$DB_USER" -e "
USE $DB_NAME;

-- 数据关联分析
SELECT 
    '=== 数据关联分析报告 ===' as analysis_type,
    '' as table_name,
    '' as record_count,
    '' as null_fields,
    '' as data_quality

UNION ALL

SELECT 
    '基础信息表' as analysis_type,
    'company_basic_info' as table_name,
    COUNT(*) as record_count,
    CONCAT(
        '空字段: ',
        SUM(CASE WHEN company_name IS NULL THEN 1 ELSE 0 END), '个公司名称, ',
        SUM(CASE WHEN industry_category IS NULL THEN 1 ELSE 0 END), '个行业分类, ',
        SUM(CASE WHEN registered_capital IS NULL THEN 1 ELSE 0 END), '个注册资本'
    ) as null_fields,
    CONCAT('数据完整度: ', ROUND((COUNT(*) - SUM(CASE WHEN company_name IS NULL OR industry_category IS NULL OR registered_capital IS NULL THEN 1 ELSE 0 END)) * 100.0 / COUNT(*), 2), '%') as data_quality
FROM company_basic_info
WHERE company_id = 1

UNION ALL

SELECT 
    '财务信息表' as analysis_type,
    'company_financial_info' as table_name,
    COUNT(*) as record_count,
    CONCAT(
        '空字段: ',
        SUM(CASE WHEN total_assets IS NULL THEN 1 ELSE 0 END), '个总资产, ',
        SUM(CASE WHEN annual_revenue IS NULL THEN 1 ELSE 0 END), '个年营收, ',
        SUM(CASE WHEN net_profit IS NULL THEN 1 ELSE 0 END), '个净利润'
    ) as null_fields,
    CONCAT('数据完整度: ', ROUND((COUNT(*) - SUM(CASE WHEN total_assets IS NULL OR annual_revenue IS NULL OR net_profit IS NULL THEN 1 ELSE 0 END)) * 100.0 / COUNT(*), 2), '%') as data_quality
FROM company_financial_info
WHERE company_id = 1

UNION ALL

SELECT 
    '风险信息表' as analysis_type,
    'company_risk_info' as table_name,
    COUNT(*) as record_count,
    CONCAT(
        '空字段: ',
        SUM(CASE WHEN risk_level IS NULL THEN 1 ELSE 0 END), '个风险等级, ',
        SUM(CASE WHEN credit_rating IS NULL THEN 1 ELSE 0 END), '个信用评级'
    ) as null_fields,
    CONCAT('数据完整度: ', ROUND((COUNT(*) - SUM(CASE WHEN risk_level IS NULL OR credit_rating IS NULL THEN 1 ELSE 0 END)) * 100.0 / COUNT(*), 2), '%') as data_quality
FROM company_risk_info
WHERE company_id = 1

UNION ALL

SELECT 
    '关系信息表' as analysis_type,
    'company_relationships' as table_name,
    COUNT(*) as record_count,
    CONCAT(
        '空字段: ',
        SUM(CASE WHEN related_company_name IS NULL THEN 1 ELSE 0 END), '个关联企业, ',
        SUM(CASE WHEN relationship_type IS NULL THEN 1 ELSE 0 END), '个关系类型'
    ) as null_fields,
    CONCAT('数据完整度: ', ROUND((COUNT(*) - SUM(CASE WHEN related_company_name IS NULL OR relationship_type IS NULL THEN 1 ELSE 0 END)) * 100.0 / COUNT(*), 2), '%') as data_quality
FROM company_relationships
WHERE company_id = 1;
" > /tmp/data_relationship_analysis.txt

if [ $? -eq 0 ]; then
    log_success "数据关联分析报告生成成功"
    echo "=========================================="
    echo "📈 数据关联分析报告"
    echo "=========================================="
    cat /tmp/data_relationship_analysis.txt
    echo "=========================================="
else
    log_error "数据关联分析报告生成失败"
fi

# 5. 生成企业关系网络图数据
log_info "生成企业关系网络图数据..."

mysql -h "$DB_HOST" -u "$DB_USER" -e "
USE $DB_NAME;

-- 企业关系网络数据
SELECT 
    'source' as node_type,
    company_name as node_name,
    'center' as node_position,
    'primary' as node_color
FROM company_basic_info 
WHERE company_id = 1

UNION ALL

SELECT 
    'target' as node_type,
    related_company_name as node_name,
    relationship_type as node_position,
    CASE 
        WHEN relationship_type = '控股' THEN 'red'
        WHEN relationship_type = '投资' THEN 'blue'
        WHEN relationship_type = '合作' THEN 'green'
        ELSE 'gray'
    END as node_color
FROM company_relationships 
WHERE company_id = 1;
" > /tmp/company_network_data.txt

if [ $? -eq 0 ]; then
    log_success "企业关系网络图数据生成成功"
    echo "=========================================="
    echo "🕸️ 企业关系网络图数据"
    echo "=========================================="
    cat /tmp/company_network_data.txt
    echo "=========================================="
else
    log_error "企业关系网络图数据生成失败"
fi

# 6. 创建数据查询接口
log_info "创建数据查询接口..."

cat > /tmp/company_data_queries.sql << 'EOF'
-- 企业数据查询接口
USE jobfirst;

-- 查询1: 获取企业完整信息
DELIMITER //
CREATE PROCEDURE GetCompanyCompleteInfo(IN p_company_id INT)
BEGIN
    SELECT 
        cbi.company_name,
        cbi.industry_category,
        cbi.business_status,
        cbi.registered_capital,
        cbi.insured_count,
        cfi.total_assets,
        cfi.annual_revenue,
        cfi.net_profit,
        cfi.financing_status,
        cri.risk_level,
        cri.credit_rating
    FROM company_basic_info cbi
    LEFT JOIN company_financial_info cfi ON cbi.company_id = cfi.company_id AND cbi.report_id = cfi.report_id
    LEFT JOIN company_risk_info cri ON cbi.company_id = cri.company_id AND cbi.report_id = cri.report_id
    WHERE cbi.company_id = p_company_id;
END //
DELIMITER ;

-- 查询2: 获取企业关系网络
DELIMITER //
CREATE PROCEDURE GetCompanyRelationships(IN p_company_id INT)
BEGIN
    SELECT 
        related_company_name,
        relationship_type,
        investment_amount,
        investment_ratio,
        position
    FROM company_relationships
    WHERE company_id = p_company_id
    ORDER BY relationship_type, investment_amount DESC;
END //
DELIMITER ;

-- 查询3: 获取企业风险分析
DELIMITER //
CREATE PROCEDURE GetCompanyRiskAnalysis(IN p_company_id INT)
BEGIN
    SELECT 
        risk_level,
        credit_rating,
        legal_risks,
        financial_risks,
        operational_risks,
        risk_factors
    FROM company_risk_info
    WHERE company_id = p_company_id;
END //
DELIMITER ;
EOF

mysql -h "$DB_HOST" -u "$DB_USER" < /tmp/company_data_queries.sql

if [ $? -eq 0 ]; then
    log_success "数据查询接口创建成功"
else
    log_warning "数据查询接口创建失败（可能已存在）"
fi

# 7. 测试查询接口
log_info "测试查询接口..."

mysql -h "$DB_HOST" -u "$DB_USER" -e "
USE $DB_NAME;
CALL GetCompanyCompleteInfo(1);
" > /tmp/query_test_result.txt

if [ $? -eq 0 ]; then
    log_success "查询接口测试成功"
    echo "=========================================="
    echo "🔍 查询接口测试结果"
    echo "=========================================="
    cat /tmp/query_test_result.txt
    echo "=========================================="
else
    log_error "查询接口测试失败"
fi

# 8. 清理临时文件
rm -f /tmp/company_complete_profile.txt /tmp/data_relationship_analysis.txt /tmp/company_network_data.txt /tmp/company_data_queries.sql /tmp/query_test_result.txt

log_success "企业数据关联验证和查询工作完成！"
echo "=========================================="
echo "✅ 完成的工作："
echo "1. 验证了数据关联完整性"
echo "2. 生成了企业完整画像报告"
echo "3. 生成了数据关联分析报告"
echo "4. 生成了企业关系网络图数据"
echo "5. 创建了数据查询接口"
echo "6. 测试了查询接口功能"
echo "=========================================="
echo "📋 可用的查询接口："
echo "- CALL GetCompanyCompleteInfo(1);  # 获取企业完整信息"
echo "- CALL GetCompanyRelationships(1); # 获取企业关系网络"
echo "- CALL GetCompanyRiskAnalysis(1);  # 获取企业风险分析"
echo "=========================================="
