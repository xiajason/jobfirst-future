-- 企业画像数据库表结构设计
-- 基于公司画像文档设计的PostgreSQL数据库表单和字段设计
-- 适配MySQL数据库，支持Company服务PDF解析功能

-- 1. 企业基本信息表扩展 (company_basic_info)
CREATE TABLE IF NOT EXISTS company_basic_info (
    id INT PRIMARY KEY AUTO_INCREMENT,
    company_id BIGINT UNSIGNED NOT NULL,
    report_id VARCHAR(50) UNIQUE,  -- 报告编号（FSCR202503281920000003等）
    company_name VARCHAR(255) NOT NULL,  -- 企业名称
    used_name VARCHAR(255),              -- 曾用名
    unified_social_credit_code VARCHAR(50), -- 统一社会信用代码
    registration_date DATE,               -- 成立日期
    legal_representative VARCHAR(100),    -- 法定代表人
    business_status VARCHAR(50),          -- 经营状态（登记成立/存续等）
    registered_capital DECIMAL(18,2),     -- 注册资本（万元）
    currency VARCHAR(20) DEFAULT 'CNY',   -- 注册币种
    insured_count INT,                    -- 参保人数
    industry_category VARCHAR(100),       -- 所属行业类别
    registration_authority VARCHAR(255), -- 登记机关
    business_scope TEXT,                  -- 经营范围
    tags JSON,                            -- 企业标签（JSON数组）
    data_source VARCHAR(100),             -- 数据来源
    data_update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_company_id (company_id),
    INDEX idx_report_id (report_id),
    INDEX idx_company_name (company_name),
    INDEX idx_unified_social_credit_code (unified_social_credit_code),
    INDEX idx_registration_date (registration_date),
    INDEX idx_industry_category (industry_category),
    
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. 资质许可表 (qualification_license)
CREATE TABLE IF NOT EXISTS qualification_license (
    id INT PRIMARY KEY AUTO_INCREMENT,
    company_id BIGINT UNSIGNED NOT NULL,
    report_id VARCHAR(50),
    type ENUM('资质', '许可', '备案') NOT NULL,  -- 类型
    name VARCHAR(255) NOT NULL,                    -- 资质/许可名称
    status VARCHAR(20) DEFAULT '有效',             -- 状态（有效/无效）
    certificate_number VARCHAR(100),               -- 证书编号
    issue_date DATE,                               -- 颁发时间
    issuing_authority VARCHAR(255),                -- 颁发单位
    validity_period DATE,                          -- 有效期限
    content TEXT,                                  -- 内容描述
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_company_id (company_id),
    INDEX idx_report_id (report_id),
    INDEX idx_type (type),
    INDEX idx_status (status),
    INDEX idx_certificate_number (certificate_number),
    INDEX idx_issue_date (issue_date),
    
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. 人员竞争力表 (personnel_competitiveness)
CREATE TABLE IF NOT EXISTS personnel_competitiveness (
    id INT PRIMARY KEY AUTO_INCREMENT,
    company_id BIGINT UNSIGNED NOT NULL,
    report_id VARCHAR(50),
    data_update_date DATE,                -- 数据更新时间
    total_employees INT,                  -- 企业人数
    industry_ranking VARCHAR(50),        -- 行业内排名
    industry_avg_employees INT,           -- 行业内平均人数
    turnover_rate DECIMAL(5,2),           -- 离职率
    entry_rate DECIMAL(5,2),              -- 入职率
    tenure_distribution JSON,             -- 司龄分布（JSON格式）
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_company_id (company_id),
    INDEX idx_report_id (report_id),
    INDEX idx_data_update_date (data_update_date),
    INDEX idx_total_employees (total_employees),
    
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. 公积金信息表 (provident_fund)
CREATE TABLE IF NOT EXISTS provident_fund (
    id INT PRIMARY KEY AUTO_INCREMENT,
    company_id BIGINT UNSIGNED NOT NULL,
    report_id VARCHAR(50),
    unit_nature VARCHAR(100),             -- 单位性质
    opening_date DATE,                    -- 开户日期
    last_payment_month DATE,              -- 最近缴存月份
    total_payment VARCHAR(50),            -- 累计缴存总额
    payment_records JSON,                 -- 月度缴存记录（JSON数组）
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_company_id (company_id),
    INDEX idx_report_id (report_id),
    INDEX idx_opening_date (opening_date),
    INDEX idx_last_payment_month (last_payment_month),
    
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 5. 资助补贴表 (subsidy_info)
CREATE TABLE IF NOT EXISTS subsidy_info (
    id INT PRIMARY KEY AUTO_INCREMENT,
    company_id BIGINT UNSIGNED NOT NULL,
    report_id VARCHAR(50),
    subsidy_year INT,                     -- 资助年份
    amount DECIMAL(18,2),                 -- 资助金额（万元）
    count INT,                            -- 资助次数
    source VARCHAR(255),                  -- 信息来源
    subsidy_list TEXT,                    -- 资助名单
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_company_id (company_id),
    INDEX idx_report_id (report_id),
    INDEX idx_subsidy_year (subsidy_year),
    INDEX idx_amount (amount),
    
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 6. 企业关系图谱表 (company_relationships)
CREATE TABLE IF NOT EXISTS company_relationships (
    id INT PRIMARY KEY AUTO_INCREMENT,
    company_id BIGINT UNSIGNED NOT NULL,
    report_id VARCHAR(50),
    related_company_name VARCHAR(255),    -- 关联企业名称
    relationship_type ENUM('投资', '任职', '合作', '控股', '参股') NOT NULL,  -- 关联类型
    investment_amount DECIMAL(18,2),      -- 投资金额
    investment_ratio DECIMAL(5,2),        -- 投资比例
    position VARCHAR(100),                -- 任职职位
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_company_id (company_id),
    INDEX idx_report_id (report_id),
    INDEX idx_related_company_name (related_company_name),
    INDEX idx_relationship_type (relationship_type),
    INDEX idx_investment_amount (investment_amount),
    
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 7. 科创评分表 (tech_innovation_score)
CREATE TABLE IF NOT EXISTS tech_innovation_score (
    id INT PRIMARY KEY AUTO_INCREMENT,
    company_id BIGINT UNSIGNED NOT NULL,
    report_id VARCHAR(50),
    basic_score DECIMAL(5,2),             -- 基本面评分
    talent_score DECIMAL(5,2),            -- 人才竞争力评分
    industry_ranking VARCHAR(50),         -- 行业排名
    strategic_industry VARCHAR(100),      -- 战略新兴产业
    intellectual_property JSON,           -- 知识产权数据
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_company_id (company_id),
    INDEX idx_report_id (report_id),
    INDEX idx_basic_score (basic_score),
    INDEX idx_talent_score (talent_score),
    INDEX idx_industry_ranking (industry_ranking),
    
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 8. 企业财务信息表 (company_financial_info)
CREATE TABLE IF NOT EXISTS company_financial_info (
    id INT PRIMARY KEY AUTO_INCREMENT,
    company_id BIGINT UNSIGNED NOT NULL,
    report_id VARCHAR(50),
    annual_revenue DECIMAL(18,2),         -- 年营业额
    net_profit DECIMAL(18,2),             -- 净利润
    total_assets DECIMAL(18,2),           -- 总资产
    total_liabilities DECIMAL(18,2),      -- 总负债
    equity DECIMAL(18,2),                 -- 净资产
    financing_status VARCHAR(100),        -- 融资情况
    listing_status VARCHAR(50),           -- 上市状态
    financial_year INT,                   -- 财务年度
    data_source VARCHAR(100),             -- 数据来源
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_company_id (company_id),
    INDEX idx_report_id (report_id),
    INDEX idx_financial_year (financial_year),
    INDEX idx_annual_revenue (annual_revenue),
    
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 9. 企业风险信息表 (company_risk_info)
CREATE TABLE IF NOT EXISTS company_risk_info (
    id INT PRIMARY KEY AUTO_INCREMENT,
    company_id BIGINT UNSIGNED NOT NULL,
    report_id VARCHAR(50),
    risk_level ENUM('低风险', '中风险', '高风险') DEFAULT '低风险',  -- 风险等级
    legal_risks JSON,                     -- 法律风险
    financial_risks JSON,                 -- 财务风险
    operational_risks JSON,               -- 经营风险
    credit_rating VARCHAR(20),            -- 信用评级
    risk_factors TEXT,                    -- 风险因素
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_company_id (company_id),
    INDEX idx_report_id (report_id),
    INDEX idx_risk_level (risk_level),
    INDEX idx_credit_rating (credit_rating),
    
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建复合索引优化查询性能
CREATE INDEX idx_company_basic_info_composite ON company_basic_info(company_id, report_id, industry_category);
CREATE INDEX idx_qualification_license_composite ON qualification_license(company_id, type, status);
CREATE INDEX idx_personnel_competitiveness_composite ON personnel_competitiveness(company_id, data_update_date);
CREATE INDEX idx_company_relationships_composite ON company_relationships(company_id, relationship_type);
CREATE INDEX idx_tech_innovation_score_composite ON tech_innovation_score(company_id, basic_score, talent_score);
