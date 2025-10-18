-- AI服务使用统计表
-- 用于跟踪用户AI服务调用次数和成本

CREATE TABLE IF NOT EXISTS ai_service_usage (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL COMMENT '用户ID',
    service_type VARCHAR(50) NOT NULL COMMENT 'AI服务类型 (document_parsing, text_analysis, etc.)',
    service_name VARCHAR(100) NOT NULL COMMENT '具体服务名称',
    request_id VARCHAR(100) COMMENT '请求ID，用于追踪',
    input_tokens INT DEFAULT 0 COMMENT '输入token数量',
    output_tokens INT DEFAULT 0 COMMENT '输出token数量',
    total_tokens INT DEFAULT 0 COMMENT '总token数量',
    cost_usd DECIMAL(10, 6) DEFAULT 0.000000 COMMENT '本次调用成本(USD)',
    processing_time_ms INT DEFAULT 0 COMMENT '处理时间(毫秒)',
    status ENUM('success', 'failed', 'limited') DEFAULT 'success' COMMENT '调用状态',
    error_message TEXT COMMENT '错误信息',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_user_id (user_id),
    INDEX idx_service_type (service_type),
    INDEX idx_created_at (created_at),
    INDEX idx_user_service_date (user_id, service_type, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI服务使用统计表';

-- 用户AI服务配额表
-- 用于管理用户AI服务使用限制

CREATE TABLE IF NOT EXISTS user_ai_quotas (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL COMMENT '用户ID',
    subscription_type ENUM('trial', 'basic', 'premium', 'enterprise') DEFAULT 'trial' COMMENT '订阅类型',
    service_type VARCHAR(50) NOT NULL COMMENT 'AI服务类型',
    daily_limit INT DEFAULT 0 COMMENT '每日调用限制',
    monthly_limit INT DEFAULT 0 COMMENT '每月调用限制',
    daily_used INT DEFAULT 0 COMMENT '今日已使用次数',
    monthly_used INT DEFAULT 0 COMMENT '本月已使用次数',
    daily_cost_limit DECIMAL(10, 6) DEFAULT 0.000000 COMMENT '每日成本限制(USD)',
    monthly_cost_limit DECIMAL(10, 6) DEFAULT 0.000000 COMMENT '每月成本限制(USD)',
    daily_cost_used DECIMAL(10, 6) DEFAULT 0.000000 COMMENT '今日已使用成本(USD)',
    monthly_cost_used DECIMAL(10, 6) DEFAULT 0.000000 COMMENT '本月已使用成本(USD)',
    quota_reset_date DATE COMMENT '配额重置日期',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否激活',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_user_service (user_id, service_type),
    INDEX idx_user_id (user_id),
    INDEX idx_subscription_type (subscription_type),
    INDEX idx_quota_reset_date (quota_reset_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户AI服务配额表';

-- 订阅限制配置表
-- 用于配置不同订阅类型的限制

CREATE TABLE IF NOT EXISTS subscription_limits (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    subscription_type ENUM('trial', 'basic', 'premium', 'enterprise') NOT NULL COMMENT '订阅类型',
    service_type VARCHAR(50) NOT NULL COMMENT 'AI服务类型',
    daily_limit INT DEFAULT 0 COMMENT '每日调用限制',
    monthly_limit INT DEFAULT 0 COMMENT '每月调用限制',
    daily_cost_limit DECIMAL(10, 6) DEFAULT 0.000000 COMMENT '每日成本限制(USD)',
    monthly_cost_limit DECIMAL(10, 6) DEFAULT 0.000000 COMMENT '每月成本限制(USD)',
    priority INT DEFAULT 0 COMMENT '优先级，数字越大优先级越高',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否激活',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_subscription_service (subscription_type, service_type),
    INDEX idx_subscription_type (subscription_type),
    INDEX idx_service_type (service_type),
    INDEX idx_priority (priority)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订阅限制配置表';

-- 插入默认订阅限制配置
INSERT INTO subscription_limits (subscription_type, service_type, daily_limit, monthly_limit, daily_cost_limit, monthly_cost_limit, priority) VALUES
-- 试用用户限制
('trial', 'document_parsing', 5, 50, 0.50, 5.00, 1),
('trial', 'text_analysis', 10, 100, 0.30, 3.00, 1),
('trial', 'ai_chat', 20, 200, 0.20, 2.00, 1),

-- 基础用户限制
('basic', 'document_parsing', 50, 500, 5.00, 50.00, 2),
('basic', 'text_analysis', 100, 1000, 3.00, 30.00, 2),
('basic', 'ai_chat', 200, 2000, 2.00, 20.00, 2),

-- 高级用户限制
('premium', 'document_parsing', 200, 2000, 20.00, 200.00, 3),
('premium', 'text_analysis', 500, 5000, 15.00, 150.00, 3),
('premium', 'ai_chat', 1000, 10000, 10.00, 100.00, 3),

-- 企业用户限制
('enterprise', 'document_parsing', 1000, 10000, 100.00, 1000.00, 4),
('enterprise', 'text_analysis', 2000, 20000, 60.00, 600.00, 4),
('enterprise', 'ai_chat', 5000, 50000, 50.00, 500.00, 4);

-- 为现有用户初始化配额
INSERT INTO user_ai_quotas (user_id, subscription_type, service_type, daily_limit, monthly_limit, daily_cost_limit, monthly_cost_limit, quota_reset_date)
SELECT 
    u.id as user_id,
    COALESCE(u.subscription_type, 'trial') as subscription_type,
    sl.service_type,
    sl.daily_limit,
    sl.monthly_limit,
    sl.daily_cost_limit,
    sl.monthly_cost_limit,
    CURDATE() as quota_reset_date
FROM users u
CROSS JOIN subscription_limits sl
WHERE sl.subscription_type = COALESCE(u.subscription_type, 'trial')
ON DUPLICATE KEY UPDATE
    daily_limit = VALUES(daily_limit),
    monthly_limit = VALUES(monthly_limit),
    daily_cost_limit = VALUES(daily_cost_limit),
    monthly_cost_limit = VALUES(monthly_cost_limit),
    updated_at = CURRENT_TIMESTAMP;
