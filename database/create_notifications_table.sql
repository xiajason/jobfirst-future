-- 创建通知数据库表
-- 用于存储用户通知信息

USE jobfirst;

-- 创建通知表
CREATE TABLE IF NOT EXISTS notifications (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL COMMENT '用户ID',
    type VARCHAR(50) NOT NULL COMMENT '通知类型',
    title VARCHAR(255) NOT NULL COMMENT '通知标题',
    content TEXT NOT NULL COMMENT '通知内容',
    category VARCHAR(50) DEFAULT 'system' COMMENT '通知分类',
    priority ENUM('low', 'normal', 'high', 'urgent') DEFAULT 'normal' COMMENT '优先级',
    status ENUM('unread', 'read', 'archived') DEFAULT 'unread' COMMENT '状态',
    is_read BOOLEAN DEFAULT FALSE COMMENT '是否已读',
    read_at TIMESTAMP NULL COMMENT '阅读时间',
    expires_at TIMESTAMP NULL COMMENT '过期时间',
    metadata JSON NULL COMMENT '元数据',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_user_id (user_id),
    INDEX idx_type (type),
    INDEX idx_category (category),
    INDEX idx_status (status),
    INDEX idx_is_read (is_read),
    INDEX idx_created_at (created_at),
    INDEX idx_user_status (user_id, status),
    INDEX idx_user_type (user_id, type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户通知表';

-- 创建通知统计表
CREATE TABLE IF NOT EXISTS notification_stats (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL COMMENT '用户ID',
    total_notifications INT UNSIGNED DEFAULT 0 COMMENT '总通知数',
    unread_notifications INT UNSIGNED DEFAULT 0 COMMENT '未读通知数',
    read_notifications INT UNSIGNED DEFAULT 0 COMMENT '已读通知数',
    archived_notifications INT UNSIGNED DEFAULT 0 COMMENT '归档通知数',
    last_notification_at TIMESTAMP NULL COMMENT '最后通知时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    UNIQUE KEY uk_user_id (user_id),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知统计表';

-- 创建通知类型配置表
CREATE TABLE IF NOT EXISTS notification_types (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    type VARCHAR(50) NOT NULL UNIQUE COMMENT '通知类型',
    name VARCHAR(100) NOT NULL COMMENT '类型名称',
    description TEXT COMMENT '类型描述',
    category VARCHAR(50) DEFAULT 'system' COMMENT '分类',
    priority ENUM('low', 'normal', 'high', 'urgent') DEFAULT 'normal' COMMENT '默认优先级',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    template TEXT COMMENT '模板内容',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_type (type),
    INDEX idx_category (category),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知类型配置表';

-- 插入默认通知类型
INSERT INTO notification_types (type, name, description, category, priority, template) VALUES
('subscription_expiring', '订阅即将到期', '用户订阅即将到期提醒', 'subscription', 'high', '您的订阅将在{days_left}天后到期，请及时续费'),
('subscription_upgraded', '订阅升级成功', '用户订阅升级成功通知', 'subscription', 'normal', '恭喜！您的订阅已成功升级为{to_status}'),
('subscription_expired', '订阅已过期', '用户订阅已过期通知', 'subscription', 'urgent', '您的订阅已过期，部分功能将受限'),
('subscription_renewed', '订阅续费成功', '用户订阅续费成功通知', 'subscription', 'normal', '您的订阅已成功续费，感谢您的支持'),
('ai_service_limit_warning', 'AI服务使用限制警告', 'AI服务使用量接近限制警告', 'ai_service', 'high', '您的AI服务使用量已达到{percentage}%，剩余{remaining}次'),
('ai_service_limit_exceeded', 'AI服务使用限制超出', 'AI服务使用量超出限制通知', 'ai_service', 'urgent', '您的AI服务使用量已超出限制，请升级订阅或等待重置'),
('ai_service_quota_reset', 'AI服务配额重置', 'AI服务配额重置通知', 'ai_service', 'normal', '您的AI服务配额已重置，可以继续使用'),
('cost_limit_warning', '成本限制警告', '使用成本接近限制警告', 'cost_control', 'high', '您的使用成本已达到{percentage}%，当前成本：${current_cost}'),
('cost_limit_exceeded', '成本限制超出', '使用成本超出限制通知', 'cost_control', 'urgent', '您的使用成本已超出限制，请升级订阅或等待重置'),
('cost_optimization', '成本优化建议', '成本优化建议通知', 'cost_control', 'normal', '为您推荐以下成本优化方案：{suggestions}')
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    description = VALUES(description),
    category = VALUES(category),
    priority = VALUES(priority),
    template = VALUES(template),
    updated_at = CURRENT_TIMESTAMP;

-- 创建触发器：自动更新通知统计
DELIMITER $$

CREATE TRIGGER tr_notifications_after_insert
AFTER INSERT ON notifications
FOR EACH ROW
BEGIN
    INSERT INTO notification_stats (user_id, total_notifications, unread_notifications, last_notification_at)
    VALUES (NEW.user_id, 1, 1, NEW.created_at)
    ON DUPLICATE KEY UPDATE
        total_notifications = total_notifications + 1,
        unread_notifications = unread_notifications + 1,
        last_notification_at = NEW.created_at,
        updated_at = CURRENT_TIMESTAMP;
END$$

CREATE TRIGGER tr_notifications_after_update
AFTER UPDATE ON notifications
FOR EACH ROW
BEGIN
    IF OLD.is_read != NEW.is_read THEN
        IF NEW.is_read = TRUE THEN
            UPDATE notification_stats 
            SET unread_notifications = unread_notifications - 1,
                read_notifications = read_notifications + 1,
                updated_at = CURRENT_TIMESTAMP
            WHERE user_id = NEW.user_id;
        ELSE
            UPDATE notification_stats 
            SET unread_notifications = unread_notifications + 1,
                read_notifications = read_notifications - 1,
                updated_at = CURRENT_TIMESTAMP
            WHERE user_id = NEW.user_id;
        END IF;
    END IF;
    
    IF OLD.status != NEW.status THEN
        IF NEW.status = 'archived' AND OLD.status != 'archived' THEN
            UPDATE notification_stats 
            SET archived_notifications = archived_notifications + 1,
                updated_at = CURRENT_TIMESTAMP
            WHERE user_id = NEW.user_id;
        ELSEIF OLD.status = 'archived' AND NEW.status != 'archived' THEN
            UPDATE notification_stats 
            SET archived_notifications = archived_notifications - 1,
                updated_at = CURRENT_TIMESTAMP
            WHERE user_id = NEW.user_id;
        END IF;
    END IF;
END$$

DELIMITER ;

-- 创建视图：用户通知概览
CREATE OR REPLACE VIEW v_user_notifications AS
SELECT 
    n.id,
    n.user_id,
    n.type,
    n.title,
    n.content,
    n.category,
    n.priority,
    n.status,
    n.is_read,
    n.read_at,
    n.expires_at,
    n.created_at,
    n.updated_at,
    nt.name as type_name,
    nt.description as type_description
FROM notifications n
LEFT JOIN notification_types nt ON n.type = nt.type
WHERE n.expires_at IS NULL OR n.expires_at > NOW()
ORDER BY n.created_at DESC;

-- 创建视图：用户通知统计
CREATE OR REPLACE VIEW v_user_notification_stats AS
SELECT 
    ns.user_id,
    ns.total_notifications,
    ns.unread_notifications,
    ns.read_notifications,
    ns.archived_notifications,
    ns.last_notification_at,
    CASE 
        WHEN ns.unread_notifications = 0 THEN 'all_read'
        WHEN ns.unread_notifications <= 5 THEN 'few_unread'
        WHEN ns.unread_notifications <= 20 THEN 'some_unread'
        ELSE 'many_unread'
    END as unread_status
FROM notification_stats ns;

-- 插入测试数据（可选）
INSERT INTO notifications (user_id, type, title, content, category, priority) VALUES
(4, 'subscription_expiring', '订阅即将到期提醒', '您的订阅将在3天后到期，请及时续费以继续享受服务', 'subscription', 'high'),
(4, 'ai_service_limit_warning', 'AI服务使用限制警告', '您的AI服务使用量已达到80%，剩余2次调用', 'ai_service', 'high'),
(4, 'cost_limit_warning', '成本限制警告', '您的使用成本已达到85%，当前成本：$8.50', 'cost_control', 'high')
ON DUPLICATE KEY UPDATE updated_at = CURRENT_TIMESTAMP;

-- 显示创建结果
SELECT 'Notifications table created successfully' as result;
SELECT COUNT(*) as notification_count FROM notifications;
SELECT COUNT(*) as stats_count FROM notification_stats;
SELECT COUNT(*) as types_count FROM notification_types;
