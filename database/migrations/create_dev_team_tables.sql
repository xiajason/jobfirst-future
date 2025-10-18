-- 开发团队管理系统数据库迁移脚本
-- 创建时间: 2025-09-06
-- 描述: 创建开发团队管理相关的数据表

-- 创建开发团队成员表
CREATE TABLE IF NOT EXISTS `dev_team_users` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `user_id` bigint unsigned NOT NULL,
    `team_role` enum('super_admin','system_admin','dev_lead','frontend_dev','backend_dev','qa_engineer','guest') NOT NULL DEFAULT 'guest',
    `ssh_public_key` text,
    `server_access_level` enum('full','limited','readonly','none') NOT NULL DEFAULT 'limited',
    `code_access_modules` json DEFAULT NULL,
    `database_access` json DEFAULT NULL,
    `service_restart_permissions` json DEFAULT NULL,
    `last_login_at` timestamp NULL DEFAULT NULL,
    `status` enum('active','inactive','suspended') NOT NULL DEFAULT 'active',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` timestamp NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_dev_team_users_user_id` (`user_id`),
    KEY `idx_dev_team_users_team_role` (`team_role`),
    KEY `idx_dev_team_users_status` (`status`),
    KEY `idx_dev_team_users_created_at` (`created_at`),
    CONSTRAINT `fk_dev_team_users_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建开发操作日志表
CREATE TABLE IF NOT EXISTS `dev_operation_logs` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `user_id` bigint unsigned NOT NULL,
    `operation_type` varchar(100) NOT NULL,
    `operation_target` varchar(255) DEFAULT NULL,
    `operation_details` json DEFAULT NULL,
    `ip_address` varchar(45) DEFAULT NULL,
    `user_agent` text,
    `status` enum('success','failed','blocked') NOT NULL DEFAULT 'success',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_dev_operation_logs_user_id` (`user_id`),
    KEY `idx_dev_operation_logs_operation_type` (`operation_type`),
    KEY `idx_dev_operation_logs_status` (`status`),
    KEY `idx_dev_operation_logs_created_at` (`created_at`),
    KEY `idx_dev_operation_logs_composite` (`user_id`, `operation_type`, `created_at`),
    CONSTRAINT `fk_dev_operation_logs_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建权限配置表
CREATE TABLE IF NOT EXISTS `team_permission_configs` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `role_name` varchar(50) NOT NULL,
    `permissions` json NOT NULL,
    `description` text,
    `is_active` tinyint(1) NOT NULL DEFAULT '1',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_team_permission_configs_role_name` (`role_name`),
    KEY `idx_team_permission_configs_is_active` (`is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入默认权限配置
INSERT INTO `team_permission_configs` (`role_name`, `permissions`, `description`, `is_active`) VALUES
('super_admin', '{"server_access_level": "full", "code_access_modules": ["frontend", "backend", "database", "config"], "database_access": ["all"], "service_restart_permissions": ["all"]}', '超级管理员，拥有所有权限', 1),
('system_admin', '{"server_access_level": "full", "code_access_modules": ["frontend", "backend", "database", "config"], "database_access": ["system"], "service_restart_permissions": ["system"]}', '系统管理员，拥有系统管理权限', 1),
('dev_lead', '{"server_access_level": "limited", "code_access_modules": ["frontend", "backend"], "database_access": ["development"], "service_restart_permissions": ["backend"]}', '开发负责人，拥有项目开发权限', 1),
('frontend_dev', '{"server_access_level": "limited", "code_access_modules": ["frontend"], "database_access": [], "service_restart_permissions": []}', '前端开发，拥有前端代码权限', 1),
('backend_dev', '{"server_access_level": "limited", "code_access_modules": ["backend"], "database_access": ["development"], "service_restart_permissions": ["backend"]}', '后端开发，拥有后端代码权限', 1),
('qa_engineer', '{"server_access_level": "limited", "code_access_modules": ["test"], "database_access": ["test"], "service_restart_permissions": []}', '测试工程师，拥有测试权限', 1),
('guest', '{"server_access_level": "readonly", "code_access_modules": [], "database_access": [], "service_restart_permissions": []}', '访客用户，只读权限', 1)
ON DUPLICATE KEY UPDATE 
    `permissions` = VALUES(`permissions`),
    `description` = VALUES(`description`),
    `updated_at` = CURRENT_TIMESTAMP;

-- 创建视图：团队成员详细信息
CREATE OR REPLACE VIEW `dev_team_members_view` AS
SELECT 
    dtu.id,
    dtu.user_id,
    u.username,
    u.email,
    u.first_name,
    u.last_name,
    u.avatar_url,
    dtu.team_role,
    dtu.ssh_public_key,
    dtu.server_access_level,
    dtu.code_access_modules,
    dtu.database_access,
    dtu.service_restart_permissions,
    dtu.last_login_at,
    dtu.status,
    dtu.created_at,
    dtu.updated_at
FROM `dev_team_users` dtu
LEFT JOIN `users` u ON dtu.user_id = u.id
WHERE dtu.deleted_at IS NULL;

-- 创建存储过程：获取团队成员统计信息
DELIMITER //
CREATE PROCEDURE GetDevTeamStats()
BEGIN
    SELECT 
        COUNT(*) as total_members,
        COUNT(CASE WHEN status = 'active' THEN 1 END) as active_members,
        COUNT(CASE WHEN status != 'active' THEN 1 END) as inactive_members,
        COUNT(CASE WHEN team_role = 'super_admin' THEN 1 END) as super_admin_count,
        COUNT(CASE WHEN team_role = 'system_admin' THEN 1 END) as system_admin_count,
        COUNT(CASE WHEN team_role = 'dev_lead' THEN 1 END) as dev_lead_count,
        COUNT(CASE WHEN team_role = 'frontend_dev' THEN 1 END) as frontend_dev_count,
        COUNT(CASE WHEN team_role = 'backend_dev' THEN 1 END) as backend_dev_count,
        COUNT(CASE WHEN team_role = 'qa_engineer' THEN 1 END) as qa_engineer_count,
        COUNT(CASE WHEN team_role = 'guest' THEN 1 END) as guest_count
    FROM dev_team_users
    WHERE deleted_at IS NULL;
END //
DELIMITER ;

-- 创建触发器：记录团队成员变更日志
DELIMITER //
CREATE TRIGGER dev_team_users_audit_insert
AFTER INSERT ON dev_team_users
FOR EACH ROW
BEGIN
    INSERT INTO dev_operation_logs (
        user_id, 
        operation_type, 
        operation_target, 
        operation_details, 
        status, 
        created_at
    ) VALUES (
        NEW.user_id,
        'team_member_added',
        'dev_team_users',
        JSON_OBJECT('member_id', NEW.id, 'team_role', NEW.team_role, 'status', NEW.status),
        'success',
        NOW()
    );
END //
DELIMITER ;

DELIMITER //
CREATE TRIGGER dev_team_users_audit_update
AFTER UPDATE ON dev_team_users
FOR EACH ROW
BEGIN
    INSERT INTO dev_operation_logs (
        user_id, 
        operation_type, 
        operation_target, 
        operation_details, 
        status, 
        created_at
    ) VALUES (
        NEW.user_id,
        'team_member_updated',
        'dev_team_users',
        JSON_OBJECT(
            'member_id', NEW.id, 
            'old_team_role', OLD.team_role, 
            'new_team_role', NEW.team_role,
            'old_status', OLD.status,
            'new_status', NEW.status
        ),
        'success',
        NOW()
    );
END //
DELIMITER ;

-- 创建索引优化查询性能
CREATE INDEX idx_dev_team_users_composite ON dev_team_users (status, team_role, created_at);
CREATE INDEX idx_dev_operation_logs_composite ON dev_operation_logs (user_id, operation_type, created_at);

-- 添加注释
ALTER TABLE `dev_team_users` COMMENT = '开发团队成员表';
ALTER TABLE `dev_operation_logs` COMMENT = '开发操作日志表';
ALTER TABLE `team_permission_configs` COMMENT = '团队权限配置表';

-- 创建超级管理员初始化存储过程
DELIMITER //
CREATE PROCEDURE InitializeSuperAdmin(
    IN p_username VARCHAR(100),
    IN p_email VARCHAR(255),
    IN p_password_hash VARCHAR(255),
    IN p_first_name VARCHAR(100),
    IN p_last_name VARCHAR(100)
)
BEGIN
    DECLARE v_user_id BIGINT UNSIGNED;
    DECLARE v_uuid VARCHAR(36);
    
    -- 生成UUID
    SET v_uuid = UUID();
    
    -- 检查是否已有超级管理员
    IF EXISTS (SELECT 1 FROM dev_team_users WHERE team_role = 'super_admin' AND deleted_at IS NULL) THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '超级管理员已存在';
    END IF;
    
    -- 检查用户名和邮箱是否已存在
    IF EXISTS (SELECT 1 FROM users WHERE username = p_username OR email = p_email) THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '用户名或邮箱已存在';
    END IF;
    
    -- 创建用户
    INSERT INTO users (
        uuid, username, email, password_hash, first_name, last_name, 
        status, email_verified, created_at, updated_at
    ) VALUES (
        v_uuid, p_username, p_email, p_password_hash, p_first_name, p_last_name,
        'active', 1, NOW(), NOW()
    );
    
    -- 获取新创建的用户ID
    SET v_user_id = LAST_INSERT_ID();
    
    -- 创建超级管理员记录
    INSERT INTO dev_team_users (
        user_id, team_role, server_access_level, code_access_modules, 
        database_access, service_restart_permissions, status, created_at, updated_at
    ) VALUES (
        v_user_id, 'super_admin', 'full', 
        '["frontend", "backend", "database", "config"]',
        '["all"]',
        '["all"]',
        'active', NOW(), NOW()
    );
    
    -- 记录初始化日志
    INSERT INTO dev_operation_logs (
        user_id, operation_type, operation_target, operation_details, 
        status, created_at
    ) VALUES (
        v_user_id, 'super_admin_initialized', 'dev_team', 
        JSON_OBJECT('username', p_username, 'email', p_email),
        'success', NOW()
    );
    
    SELECT 'Super admin initialized successfully' as message, v_user_id as user_id;
END //
DELIMITER ;

-- 创建检查超级管理员状态的存储过程
DELIMITER //
CREATE PROCEDURE CheckSuperAdminStatus()
BEGIN
    SELECT 
        CASE 
            WHEN COUNT(*) > 0 THEN 'EXISTS'
            ELSE 'NOT_EXISTS'
        END as status,
        COUNT(*) as count
    FROM dev_team_users 
    WHERE team_role = 'super_admin' AND deleted_at IS NULL;
END //
DELIMITER ;

-- 完成迁移
SELECT 'Dev Team Management tables created successfully!' as message;