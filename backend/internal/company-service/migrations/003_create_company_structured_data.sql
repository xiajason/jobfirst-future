-- 创建企业结构化数据表
CREATE TABLE IF NOT EXISTS company_structured_data (
    id INT PRIMARY KEY AUTO_INCREMENT,
    company_id BIGINT UNSIGNED NOT NULL,
    task_id INT NOT NULL,
    basic_info JSON,  -- 基本信息 (企业名称、简称、成立年份、规模、行业、地址、网站)
    business_info JSON,  -- 业务信息 (主营业务、产品服务、目标客户、竞争优势)
    organization_info JSON,  -- 组织信息 (组织架构、部门设置、人员规模、管理层信息)
    financial_info JSON,  -- 财务信息 (注册资本、年营业额、融资情况、上市状态)
    confidence FLOAT,  -- 解析置信度 0-1
    parsing_version VARCHAR(50) DEFAULT 'mineru-v1.0',  -- 解析器版本
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_company_id (company_id),
    INDEX idx_task_id (task_id),
    INDEX idx_confidence (confidence),
    
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    FOREIGN KEY (task_id) REFERENCES company_parsing_tasks(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
