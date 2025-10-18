-- 认证服务数据库表结构
-- 创建时间: 2025-09-14
-- 用途: 支持zervigo与AI服务的认证集成

-- 权限表（使用现有表结构）
-- CREATE TABLE IF NOT EXISTS permissions (
--     id INT PRIMARY KEY AUTO_INCREMENT,
--     permission_name VARCHAR(100) NOT NULL UNIQUE,
--     resource VARCHAR(100) NOT NULL,
--     action VARCHAR(50) NOT NULL,
--     description TEXT,
--     is_active BOOLEAN DEFAULT TRUE,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
--     
--     INDEX idx_permission_name (permission_name),
--     INDEX idx_resource (resource),
--     INDEX idx_action (action)
-- );

-- 用户权限关联表
CREATE TABLE IF NOT EXISTS user_permissions (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    permission_id BIGINT UNSIGNED NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    granted_by BIGINT UNSIGNED,
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE,
    FOREIGN KEY (granted_by) REFERENCES users(id) ON DELETE SET NULL,
    
    UNIQUE KEY unique_user_permission (user_id, permission_id),
    INDEX idx_user_id (user_id),
    INDEX idx_permission_id (permission_id),
    INDEX idx_expires_at (expires_at)
);

-- 用户配额表
CREATE TABLE IF NOT EXISTS user_quotas (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    total_quota INT NOT NULL DEFAULT 0,
    used_quota INT NOT NULL DEFAULT 0,
    reset_time TIMESTAMP NOT NULL,
    is_unlimited BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    UNIQUE KEY unique_user_resource (user_id, resource_type),
    INDEX idx_user_id (user_id),
    INDEX idx_resource_type (resource_type),
    INDEX idx_reset_time (reset_time)
);

-- 访问日志表
CREATE TABLE IF NOT EXISTS access_logs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    result VARCHAR(20) NOT NULL, -- success, failed, denied
    ip_address VARCHAR(45),
    user_agent TEXT,
    request_data JSON,
    response_data JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    INDEX idx_user_id (user_id),
    INDEX idx_action (action),
    INDEX idx_resource (resource),
    INDEX idx_result (result),
    INDEX idx_created_at (created_at),
    INDEX idx_ip_address (ip_address)
);

-- 会话表（用于JWT token管理）
CREATE TABLE IF NOT EXISTS user_sessions (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    session_token VARCHAR(255) NOT NULL UNIQUE,
    refresh_token VARCHAR(255),
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    INDEX idx_user_id (user_id),
    INDEX idx_session_token (session_token),
    INDEX idx_refresh_token (refresh_token),
    INDEX idx_expires_at (expires_at),
    INDEX idx_is_active (is_active)
);

-- 插入默认权限（使用现有permissions表结构）
INSERT INTO permissions (name, display_name, resource, action, description, level, is_system, is_active) VALUES
('ai_service_access', 'AI服务访问', 'ai_service', 'access', 'AI服务访问权限', 1, 1, 1),
('ai_job_matching', 'AI职位匹配', 'ai_service', 'job_matching', 'AI职位匹配功能', 2, 1, 1),
('ai_resume_analysis', 'AI简历分析', 'ai_service', 'resume_analysis', 'AI简历分析功能', 2, 1, 1),
('ai_chat', 'AI聊天', 'ai_service', 'chat', 'AI聊天功能', 2, 1, 1),
('unlimited_ai_requests', '无限制AI请求', 'ai_service', 'unlimited', '无限制AI请求', 3, 1, 1),
('premium_ai_features', '高级AI功能', 'ai_service', 'premium', '高级AI功能', 3, 1, 1),
('admin_ai_management', 'AI服务管理', 'ai_service', 'admin', 'AI服务管理权限', 4, 1, 1),
('view_ai_logs', '查看AI日志', 'ai_service', 'view_logs', '查看AI服务日志', 3, 1, 1),
('manage_ai_config', '管理AI配置', 'ai_service', 'manage_config', '管理AI服务配置', 4, 1, 1)
ON DUPLICATE KEY UPDATE 
    resource = VALUES(resource),
    action = VALUES(action),
    description = VALUES(description),
    updated_at = CURRENT_TIMESTAMP;

-- 为现有用户分配基础权限
INSERT INTO user_permissions (user_id, permission_id, granted_by)
SELECT 
    u.id,
    p.id,
    1 -- 假设用户ID 1是管理员
FROM users u
CROSS JOIN permissions p
WHERE p.name = 'ai_service_access'
AND u.subscription_status IN ('trial', 'premium', 'enterprise')
AND NOT EXISTS (
    SELECT 1 FROM user_permissions up 
    WHERE up.user_id = u.id AND up.permission_id = p.id
);

-- 为试用用户分配基础配额
INSERT INTO user_quotas (user_id, resource_type, total_quota, used_quota, reset_time, is_unlimited)
SELECT 
    u.id,
    'ai_requests',
    100, -- 试用用户每天100次请求
    0,
    DATE_ADD(NOW(), INTERVAL 1 DAY),
    FALSE
FROM users u
WHERE u.subscription_status = 'trial'
AND NOT EXISTS (
    SELECT 1 FROM user_quotas uq 
    WHERE uq.user_id = u.id AND uq.resource_type = 'ai_requests'
);

-- 为付费用户分配更高配额
INSERT INTO user_quotas (user_id, resource_type, total_quota, used_quota, reset_time, is_unlimited)
SELECT 
    u.id,
    'ai_requests',
    CASE 
        WHEN u.subscription_type = 'premium' THEN 1000
        WHEN u.subscription_type = 'enterprise' THEN -1 -- 无限制
        ELSE 100
    END,
    0,
    DATE_ADD(NOW(), INTERVAL 1 DAY),
    CASE 
        WHEN u.subscription_type = 'enterprise' THEN TRUE
        ELSE FALSE
    END
FROM users u
WHERE u.subscription_status IN ('premium', 'enterprise')
AND NOT EXISTS (
    SELECT 1 FROM user_quotas uq 
    WHERE uq.user_id = u.id AND uq.resource_type = 'ai_requests'
);

-- 创建触发器：用户创建时自动分配权限
DELIMITER //
CREATE TRIGGER IF NOT EXISTS tr_user_created_permissions
AFTER INSERT ON users
FOR EACH ROW
BEGIN
    -- 为试用用户分配基础权限
    IF NEW.subscription_status = 'trial' THEN
        INSERT INTO user_permissions (user_id, permission_id, granted_by)
        SELECT NEW.id, id, 1
        FROM permissions 
        WHERE name IN ('ai_service_access', 'ai_job_matching');
        
        -- 分配基础配额
        INSERT INTO user_quotas (user_id, resource_type, total_quota, used_quota, reset_time, is_unlimited)
        VALUES (NEW.id, 'ai_requests', 100, 0, DATE_ADD(NOW(), INTERVAL 1 DAY), FALSE);
    END IF;
    
    -- 为付费用户分配完整权限
    IF NEW.subscription_status IN ('premium', 'enterprise') THEN
        INSERT INTO user_permissions (user_id, permission_id, granted_by)
        SELECT NEW.id, id, 1
        FROM permissions 
        WHERE name IN ('ai_service_access', 'ai_job_matching', 'ai_resume_analysis', 'premium_ai_features');
        
        -- 分配更高配额
        INSERT INTO user_quotas (user_id, resource_type, total_quota, used_quota, reset_time, is_unlimited)
        VALUES (
            NEW.id, 
            'ai_requests', 
            CASE WHEN NEW.subscription_type = 'enterprise' THEN -1 ELSE 1000 END,
            0, 
            DATE_ADD(NOW(), INTERVAL 1 DAY),
            CASE WHEN NEW.subscription_type = 'enterprise' THEN TRUE ELSE FALSE END
        );
    END IF;
END//
DELIMITER ;

-- 创建视图：用户权限概览
CREATE VIEW user_permission_overview AS
SELECT 
    u.id as user_id,
    u.username,
    u.email,
    u.subscription_status,
    u.subscription_type,
    COUNT(up.permission_id) as permission_count,
    GROUP_CONCAT(p.name) as permissions,
    uq.total_quota,
    uq.used_quota,
    uq.remaining_quota,
    uq.reset_time,
    uq.is_unlimited
FROM users u
LEFT JOIN user_permissions up ON u.id = up.user_id AND up.is_active = 1
LEFT JOIN permissions p ON up.permission_id = p.id AND p.is_active = 1
LEFT JOIN (
    SELECT 
        user_id,
        resource_type,
        total_quota,
        used_quota,
        (total_quota - used_quota) as remaining_quota,
        reset_time,
        is_unlimited
    FROM user_quotas 
    WHERE resource_type = 'ai_requests' AND is_active = 1
) uq ON u.id = uq.user_id
WHERE u.status = 'active'
GROUP BY u.id, u.username, u.email, u.subscription_status, u.subscription_type, 
         uq.total_quota, uq.used_quota, uq.remaining_quota, uq.reset_time, uq.is_unlimited;

-- 创建索引优化查询性能
CREATE INDEX idx_users_subscription ON users(subscription_status, subscription_type);
CREATE INDEX idx_user_permissions_active ON user_permissions(user_id, is_active);
CREATE INDEX idx_user_quotas_active ON user_quotas(user_id, resource_type, is_active);
CREATE INDEX idx_access_logs_user_time ON access_logs(user_id, created_at);

-- 添加注释
ALTER TABLE permissions COMMENT = '系统权限定义表';
ALTER TABLE user_permissions COMMENT = '用户权限关联表';
ALTER TABLE user_quotas COMMENT = '用户资源配额表';
ALTER TABLE access_logs COMMENT = '用户访问日志表';
ALTER TABLE user_sessions COMMENT = '用户会话管理表';
