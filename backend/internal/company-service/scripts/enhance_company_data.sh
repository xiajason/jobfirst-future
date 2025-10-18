#!/bin/bash

# 企业数据完善和数据关联脚本
# 功能：完善现有企业数据的空字段，建立企业画像表的数据关联

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
echo "🏢 企业数据完善和数据关联工具"
echo "=========================================="

# 1. 检查数据库连接
log_info "检查数据库连接..."
if ! mysql -h "$DB_HOST" -u "$DB_USER" -e "USE $DB_NAME;" 2>/dev/null; then
    log_error "无法连接到数据库 $DB_NAME"
    exit 1
fi
log_success "数据库连接正常"

# 2. 检查现有数据状态
log_info "检查现有数据状态..."

# 检查企业基础信息表
COMPANY_COUNT=$(mysql -h "$DB_HOST" -u "$DB_USER" -e "USE $DB_NAME; SELECT COUNT(*) FROM company_basic_info;" 2>/dev/null | tail -n 1)
log_info "企业基础信息表记录数: $COMPANY_COUNT"

# 检查企业文档表
DOCUMENT_COUNT=$(mysql -h "$DB_HOST" -u "$DB_USER" -e "USE $DB_NAME; SELECT COUNT(*) FROM company_documents;" 2>/dev/null | tail -n 1)
log_info "企业文档表记录数: $DOCUMENT_COUNT"

# 检查结构化数据表
STRUCTURED_COUNT=$(mysql -h "$DB_HOST" -u "$DB_USER" -e "USE $DB_NAME; SELECT COUNT(*) FROM company_structured_data;" 2>/dev/null | tail -n 1)
log_info "结构化数据表记录数: $STRUCTURED_COUNT"

# 3. 创建示例企业数据
log_info "创建示例企业数据..."

# 创建企业基础信息
mysql -h "$DB_HOST" -u "$DB_USER" -e "
USE $DB_NAME;

-- 插入示例企业基础信息
INSERT INTO company_basic_info (
    company_id, report_id, company_name, used_name, 
    unified_social_credit_code, registration_date, legal_representative,
    business_status, registered_capital, currency, insured_count,
    industry_category, registration_authority, business_scope,
    tags, data_source, data_update_time, created_at, updated_at
) VALUES (
    1, 'REPORT_001', '某某科技有限公司', '某某科技',
    '91110000123456789X', '2020-01-15', '张三',
    '存续', 10000000.00, 'CNY', 150,
    '软件和信息技术服务业', '北京市市场监督管理局',
    '技术开发、技术咨询、技术服务；计算机系统服务；数据处理；软件开发；销售计算机、软件及辅助设备',
    '[\"高新技术企业\", \"软件企业\", \"科技型中小企业\"]',
    '企业年报', NOW(), NOW(), NOW()
) ON DUPLICATE KEY UPDATE
    company_name = VALUES(company_name),
    updated_at = NOW();

-- 插入企业财务信息
INSERT INTO company_financial_info (
    company_id, report_id, total_assets, total_liabilities,
    equity, annual_revenue, net_profit, financing_status,
    listing_status, financial_year, data_source, created_at, updated_at
) VALUES (
    1, 'REPORT_001', 50000000.00, 20000000.00,
    30000000.00, 25000000.00, 5000000.00, '已完成A轮融资',
    '未上市', 2024, '企业年报', NOW(), NOW()
) ON DUPLICATE KEY UPDATE
    total_assets = VALUES(total_assets),
    updated_at = NOW();

-- 插入企业风险信息
INSERT INTO company_risk_info (
    company_id, report_id, risk_level, legal_risks,
    financial_risks, operational_risks, credit_rating,
    risk_factors, created_at, updated_at
) VALUES (
    1, 'REPORT_001', '低风险', 
    '[\"法律合规风险：低\", \"知识产权风险：中\"]',
    '[\"资金链风险：低\", \"应收账款风险：中\"]',
    '[\"技术风险：低\", \"人员流失风险：中\"]',
    'AAA', '市场竞争加剧，技术更新换代快',
    NOW(), NOW()
) ON DUPLICATE KEY UPDATE
    risk_level = VALUES(risk_level),
    updated_at = NOW();

-- 插入企业关系信息
INSERT INTO company_relationships (
    company_id, report_id, related_company_name, relationship_type,
    investment_amount, investment_ratio, position, created_at, updated_at
) VALUES 
(1, 'REPORT_001', '某某集团股份有限公司', '控股', 50000000.00, 60.00, '控股股东', NOW(), NOW()),
(1, 'REPORT_001', '某某科技子公司A', '投资', 10000000.00, 100.00, '全资子公司', NOW(), NOW()),
(1, 'REPORT_001', '某某科技子公司B', '投资', 5000000.00, 80.00, '控股子公司', NOW(), NOW()),
(1, 'REPORT_001', '合作伙伴A', '合作', NULL, NULL, '战略合作伙伴', NOW(), NOW()),
(1, 'REPORT_001', '合作伙伴B', '合作', NULL, NULL, '技术合作伙伴', NOW(), NOW())
ON DUPLICATE KEY UPDATE
    relationship_type = VALUES(relationship_type),
    updated_at = NOW();
"

if [ $? -eq 0 ]; then
    log_success "示例企业数据创建成功"
else
    log_error "示例企业数据创建失败"
    exit 1
fi

# 4. 建立数据关联
log_info "建立企业画像表的数据关联..."

# 更新结构化数据，关联到企业基础信息
mysql -h "$DB_HOST" -u "$DB_USER" -e "
USE $DB_NAME;

-- 更新结构化数据，添加企业关联信息
UPDATE company_structured_data 
SET 
    basic_info = JSON_SET(
        COALESCE(basic_info, '{}'),
        '$.company_id', 1,
        '$.company_name', '某某科技有限公司',
        '$.report_id', 'REPORT_001',
        '$.industry', '软件和信息技术服务业',
        '$.location', '北京市',
        '$.founded_year', 2020,
        '$.company_size', '中型企业',
        '$.website', 'https://www.example-tech.com'
    ),
    business_info = JSON_SET(
        COALESCE(business_info, '{}'),
        '$.main_business', '技术开发、技术咨询、技术服务',
        '$.products', '[\"企业管理系统\", \"数据分析平台\", \"移动应用开发\"]',
        '$.target_customers', '[\"中小企业\", \"政府机构\", \"教育机构\"]',
        '$.competitive_advantage', '[\"技术实力强\", \"服务经验丰富\", \"客户口碑好\"]'
    ),
    organization_info = JSON_SET(
        COALESCE(organization_info, '{}'),
        '$.organization_structure', '有限责任公司',
        '$.departments', '[\"技术部\", \"市场部\", \"财务部\", \"人事部\"]',
        '$.personnel_scale', '150人',
        '$.management_info', '[\"总经理：张三\", \"技术总监：李四\", \"市场总监：王五\"]'
    ),
    financial_info = JSON_SET(
        COALESCE(financial_info, '{}'),
        '$.registered_capital', '1000万元',
        '$.annual_revenue', '2500万元',
        '$.financing_status', '已完成A轮融资',
        '$.listing_status', '未上市'
    ),
    updated_at = NOW()
WHERE company_id = 1;
"

if [ $? -eq 0 ]; then
    log_success "数据关联建立成功"
else
    log_error "数据关联建立失败"
    exit 1
fi

# 5. 验证数据完整性
log_info "验证数据完整性..."

# 检查企业基础信息
BASIC_INFO_COUNT=$(mysql -h "$DB_HOST" -u "$DB_USER" -e "USE $DB_NAME; SELECT COUNT(*) FROM company_basic_info WHERE company_id = 1;" 2>/dev/null | tail -n 1)
log_info "企业基础信息记录数: $BASIC_INFO_COUNT"

# 检查企业财务信息
FINANCIAL_INFO_COUNT=$(mysql -h "$DB_HOST" -u "$DB_USER" -e "USE $DB_NAME; SELECT COUNT(*) FROM company_financial_info WHERE company_id = 1;" 2>/dev/null | tail -n 1)
log_info "企业财务信息记录数: $FINANCIAL_INFO_COUNT"

# 检查企业风险信息
RISK_INFO_COUNT=$(mysql -h "$DB_HOST" -u "$DB_USER" -e "USE $DB_NAME; SELECT COUNT(*) FROM company_risk_info WHERE company_id = 1;" 2>/dev/null | tail -n 1)
log_info "企业风险信息记录数: $RISK_INFO_COUNT"

# 检查企业关系信息
RELATIONSHIP_COUNT=$(mysql -h "$DB_HOST" -u "$DB_USER" -e "USE $DB_NAME; SELECT COUNT(*) FROM company_relationships WHERE company_id = 1;" 2>/dev/null | tail -n 1)
log_info "企业关系信息记录数: $RELATIONSHIP_COUNT"

# 6. 生成数据关联报告
log_info "生成数据关联报告..."

mysql -h "$DB_HOST" -u "$DB_USER" -e "
USE $DB_NAME;

-- 查询企业完整信息
SELECT 
    '企业基础信息' as info_type,
    company_name,
    industry_category,
    business_status,
    registered_capital,
    insured_count
FROM company_basic_info 
WHERE company_id = 1

UNION ALL

SELECT 
    '企业财务信息' as info_type,
    CONCAT('总资产: ', total_assets, ' 万元') as company_name,
    CONCAT('净利润: ', net_profit, ' 万元') as industry_category,
    CONCAT('融资状态: ', financing_status) as business_status,
    CONCAT('上市状态: ', listing_status) as registered_capital,
    CONCAT('财务年度: ', financial_year) as insured_count
FROM company_financial_info 
WHERE company_id = 1

UNION ALL

SELECT 
    '企业风险信息' as info_type,
    risk_level as company_name,
    CONCAT('信用评级: ', credit_rating) as industry_category,
    '法律风险' as business_status,
    '财务风险' as registered_capital,
    '运营风险' as insured_count
FROM company_risk_info 
WHERE company_id = 1;
" > /tmp/company_data_report.txt

if [ $? -eq 0 ]; then
    log_success "数据关联报告生成成功"
    echo "=========================================="
    echo "📊 企业数据关联报告"
    echo "=========================================="
    cat /tmp/company_data_report.txt
    echo "=========================================="
else
    log_error "数据关联报告生成失败"
fi

# 7. 创建数据质量检查脚本
log_info "创建数据质量检查脚本..."

cat > /tmp/check_data_quality.sql << 'EOF'
-- 数据质量检查脚本
USE jobfirst;

-- 检查企业基础信息完整性
SELECT 
    'company_basic_info' as table_name,
    COUNT(*) as total_records,
    COUNT(company_name) as non_null_company_name,
    COUNT(industry_category) as non_null_industry,
    COUNT(registered_capital) as non_null_capital
FROM company_basic_info;

-- 检查企业财务信息完整性
SELECT 
    'company_financial_info' as table_name,
    COUNT(*) as total_records,
    COUNT(total_assets) as non_null_assets,
    COUNT(annual_revenue) as non_null_revenue,
    COUNT(net_profit) as non_null_profit
FROM company_financial_info;

-- 检查企业风险信息完整性
SELECT 
    'company_risk_info' as table_name,
    COUNT(*) as total_records,
    COUNT(risk_level) as non_null_risk_level,
    COUNT(credit_rating) as non_null_credit_rating
FROM company_risk_info;

-- 检查企业关系信息完整性
SELECT 
    'company_relationships' as table_name,
    COUNT(*) as total_records,
    COUNT(related_company_name) as non_null_related_company,
    COUNT(relationship_type) as non_null_relationship_type
FROM company_relationships;

-- 检查结构化数据完整性
SELECT 
    'company_structured_data' as table_name,
    COUNT(*) as total_records,
    COUNT(basic_info) as non_null_basic_info,
    COUNT(business_info) as non_null_business_info,
    COUNT(organization_info) as non_null_organization_info,
    COUNT(financial_info) as non_null_financial_info
FROM company_structured_data;
EOF

mysql -h "$DB_HOST" -u "$DB_USER" < /tmp/check_data_quality.sql > /tmp/data_quality_report.txt

if [ $? -eq 0 ]; then
    log_success "数据质量检查完成"
    echo "=========================================="
    echo "📈 数据质量检查报告"
    echo "=========================================="
    cat /tmp/data_quality_report.txt
    echo "=========================================="
else
    log_error "数据质量检查失败"
fi

# 8. 清理临时文件
rm -f /tmp/company_data_report.txt /tmp/data_quality_report.txt /tmp/check_data_quality.sql

log_success "企业数据完善和数据关联工作完成！"
echo "=========================================="
echo "✅ 完成的工作："
echo "1. 创建了示例企业基础信息"
echo "2. 创建了企业财务信息"
echo "3. 创建了企业风险信息"
echo "4. 创建了企业关系信息"
echo "5. 建立了数据关联"
echo "6. 验证了数据完整性"
echo "7. 生成了数据质量报告"
echo "=========================================="
