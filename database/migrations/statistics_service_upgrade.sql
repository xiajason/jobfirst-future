-- Statistics Service 升级脚本
-- 创建统计相关的视图和索引

-- 创建用户统计视图
CREATE OR REPLACE VIEW user_statistics AS
SELECT 
    u.id,
    u.username,
    u.email,
    u.created_at,
    u.status,
    COUNT(t.id) as template_count,
    SUM(COALESCE(t.usage_count, 0)) as total_usage,
    AVG(CASE WHEN t.rating > 0 THEN t.rating END) as avg_rating,
    COUNT(CASE WHEN t.created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY) THEN 1 END) as recent_templates
FROM users u
LEFT JOIN templates t ON u.id = t.created_by AND t.is_active = 1
GROUP BY u.id, u.username, u.email, u.created_at, u.status;

-- 创建模板统计视图
CREATE OR REPLACE VIEW template_statistics AS
SELECT 
    t.id,
    t.name,
    t.category,
    t.created_at,
    t.is_active,
    t.usage_count,
    t.rating,
    u.username as created_by_name,
    DATEDIFF(NOW(), t.created_at) as days_since_creation
FROM templates t
LEFT JOIN users u ON t.created_by = u.id
WHERE t.is_active = 1;

-- 创建分类统计视图
CREATE OR REPLACE VIEW category_statistics AS
SELECT 
    category,
    COUNT(*) as template_count,
    SUM(usage_count) as total_usage,
    AVG(rating) as avg_rating,
    COUNT(CASE WHEN created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY) THEN 1 END) as recent_count
FROM templates
WHERE is_active = 1
GROUP BY category;

-- 创建用户增长趋势视图
CREATE OR REPLACE VIEW user_growth_trend AS
SELECT 
    DATE(created_at) as date,
    COUNT(*) as count,
    SUM(COUNT(*)) OVER (ORDER BY DATE(created_at)) as cumulative_count
FROM users
GROUP BY DATE(created_at)
ORDER BY date;

-- 创建模板使用趋势视图
CREATE OR REPLACE VIEW template_usage_trend AS
SELECT 
    DATE(created_at) as date,
    COUNT(*) as templates_created,
    SUM(usage_count) as total_usage
FROM templates
WHERE is_active = 1
GROUP BY DATE(created_at)
ORDER BY date;

-- 索引可能已经存在，跳过创建
-- CREATE INDEX idx_users_created_at ON users(created_at);
-- CREATE INDEX idx_users_status ON users(status);
-- CREATE INDEX idx_templates_created_at ON templates(created_at);
-- CREATE INDEX idx_templates_category ON templates(category);
-- CREATE INDEX idx_templates_created_by ON templates(created_by);
-- CREATE INDEX idx_companies_created_at ON companies(created_at);
-- CREATE INDEX idx_companies_status ON companies(status);

-- 创建统计表用于缓存常用统计数据
CREATE TABLE IF NOT EXISTS statistics_cache (
    id INT AUTO_INCREMENT PRIMARY KEY,
    cache_key VARCHAR(100) NOT NULL UNIQUE,
    cache_value JSON NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_cache_key (cache_key),
    INDEX idx_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='统计数据缓存表';

-- 插入一些示例统计配置
INSERT IGNORE INTO statistics_cache (cache_key, cache_value, expires_at) VALUES
('overview_stats', '{}', DATE_ADD(NOW(), INTERVAL 1 HOUR)),
('user_trend_30d', '{}', DATE_ADD(NOW(), INTERVAL 30 MINUTE)),
('template_usage_top10', '{}', DATE_ADD(NOW(), INTERVAL 1 HOUR)),
('category_popular', '{}', DATE_ADD(NOW(), INTERVAL 2 HOUR));

-- 创建统计任务表（用于定时统计任务）
CREATE TABLE IF NOT EXISTS statistics_tasks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    task_name VARCHAR(100) NOT NULL,
    task_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    parameters JSON,
    result JSON,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_task_name (task_name),
    INDEX idx_status (status),
    INDEX idx_task_type (task_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='统计任务表';

-- 创建统计报告表
CREATE TABLE IF NOT EXISTS statistics_reports (
    id INT AUTO_INCREMENT PRIMARY KEY,
    report_name VARCHAR(100) NOT NULL,
    report_type VARCHAR(50) NOT NULL,
    report_data JSON NOT NULL,
    generated_by INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_report_name (report_name),
    INDEX idx_report_type (report_type),
    INDEX idx_generated_by (generated_by),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='统计报告表';
