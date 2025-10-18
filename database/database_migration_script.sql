-- JobFirst 数据库统一迁移脚本
-- 将 jobfirst 和 jobfirst_v3 合并到统一的 jobfirst 数据库

-- 第一步：备份现有数据
-- 已在阶段一完成

-- 第二步：创建统一的表结构
-- 基于 jobfirst_v3 的表结构，因为它是更完整的业务表

-- 1. 创建 users 表（统一版本）
DROP TABLE IF EXISTS users_unified;
CREATE TABLE users_unified (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    uuid VARCHAR(36) NOT NULL UNIQUE,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    avatar_url VARCHAR(500),
    email_verified TINYINT(1) DEFAULT 0,
    phone_verified TINYINT(1) DEFAULT 0,
    status ENUM('active', 'inactive', 'suspended') DEFAULT 'active',
    role ENUM('admin', 'user', 'guest') DEFAULT 'user',
    last_login_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    INDEX idx_uuid (uuid),
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. 创建 user_sessions 表（统一版本）
DROP TABLE IF EXISTS user_sessions_unified;
CREATE TABLE user_sessions_unified (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    session_token VARCHAR(255) NOT NULL UNIQUE,
    refresh_token VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    device_info JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_session_token (session_token),
    INDEX idx_refresh_token (refresh_token),
    INDEX idx_expires_at (expires_at),
    FOREIGN KEY (user_id) REFERENCES users_unified(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. 创建系统配置表
DROP TABLE IF EXISTS system_configs_unified;
CREATE TABLE system_configs_unified (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    config_key VARCHAR(100) NOT NULL UNIQUE,
    config_value TEXT,
    config_type ENUM('string', 'number', 'boolean', 'json') DEFAULT 'string',
    description TEXT,
    is_public TINYINT(1) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_config_key (config_key),
    INDEX idx_config_type (config_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. 创建操作日志表
DROP TABLE IF EXISTS operation_logs_unified;
CREATE TABLE operation_logs_unified (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT,
    operation_type VARCHAR(50) NOT NULL,
    operation_desc TEXT,
    request_method VARCHAR(10),
    request_url VARCHAR(500),
    request_ip VARCHAR(45),
    user_agent TEXT,
    request_data JSON,
    response_data JSON,
    status_code INT,
    execution_time INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_operation_type (operation_type),
    INDEX idx_created_at (created_at),
    INDEX idx_request_ip (request_ip),
    FOREIGN KEY (user_id) REFERENCES users_unified(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 第三步：数据迁移
-- 1. 迁移 users 表数据
INSERT INTO users_unified (
    id, uuid, username, email, password_hash, first_name, last_name, 
    phone, avatar_url, email_verified, phone_verified, status, role, 
    last_login_at, created_at, updated_at, deleted_at
)
SELECT 
    u.id,
    COALESCE(u3.uuid, UUID()) as uuid,
    u.username,
    u.email,
    u.password_hash,
    u3.first_name,
    u3.last_name,
    u3.phone,
    u3.avatar_url,
    COALESCE(u3.email_verified, 0) as email_verified,
    COALESCE(u3.phone_verified, 0) as phone_verified,
    u.status,
    u.role,
    COALESCE(u3.last_login_at, u.last_login) as last_login_at,
    COALESCE(u3.created_at, u.created_at) as created_at,
    COALESCE(u3.updated_at, u.updated_at) as updated_at,
    u3.deleted_at
FROM jobfirst.users u
LEFT JOIN jobfirst_v3.users u3 ON u.id = u3.id
UNION
SELECT 
    u3.id,
    u3.uuid,
    u3.username,
    u3.email,
    u3.password_hash,
    u3.first_name,
    u3.last_name,
    u3.phone,
    u3.avatar_url,
    u3.email_verified,
    u3.phone_verified,
    u3.status,
    'user' as role,  -- 默认角色
    u3.last_login_at,
    u3.created_at,
    u3.updated_at,
    u3.deleted_at
FROM jobfirst_v3.users u3
LEFT JOIN jobfirst.users u ON u3.id = u.id
WHERE u.id IS NULL;

-- 2. 迁移 user_sessions 表数据
INSERT INTO user_sessions_unified (
    id, user_id, session_token, refresh_token, expires_at, 
    ip_address, user_agent, device_info, created_at, updated_at
)
SELECT 
    us.id,
    us.user_id,
    COALESCE(us3.session_token, us.token) as session_token,
    COALESCE(us3.refresh_token, '') as refresh_token,
    us.expires_at,
    us3.ip_address,
    us3.user_agent,
    us3.device_info,
    COALESCE(us3.created_at, us.created_at) as created_at,
    COALESCE(us3.updated_at, us.created_at) as updated_at
FROM jobfirst.user_sessions us
LEFT JOIN jobfirst_v3.user_sessions us3 ON us.id = us3.id
UNION
SELECT 
    us3.id,
    us3.user_id,
    us3.session_token,
    us3.refresh_token,
    us3.expires_at,
    us3.ip_address,
    us3.user_agent,
    us3.device_info,
    us3.created_at,
    us3.updated_at
FROM jobfirst_v3.user_sessions us3
LEFT JOIN jobfirst.user_sessions us ON us3.id = us.id
WHERE us.id IS NULL;

-- 3. 迁移系统配置数据
INSERT INTO system_configs_unified (
    config_key, config_value, config_type, description, is_public, created_at, updated_at
)
SELECT 
    config_key, 
    config_value, 
    'string' as config_type,
    '' as description,
    0 as is_public,
    created_at,
    updated_at
FROM jobfirst.system_configs;

-- 4. 迁移操作日志数据
INSERT INTO operation_logs_unified (
    user_id, operation_type, operation_desc, request_method, request_url, 
    request_ip, user_agent, request_data, response_data, status_code, 
    execution_time, created_at
)
SELECT 
    user_id, 
    operation_type, 
    operation_desc, 
    request_method, 
    request_url, 
    request_ip, 
    user_agent, 
    request_data, 
    response_data, 
    status_code, 
    execution_time, 
    created_at
FROM jobfirst.operation_logs;

-- 第四步：复制 jobfirst_v3 的业务表到统一数据库
-- 这些表在 jobfirst 中不存在，直接复制

-- 复制所有业务表
CREATE TABLE certifications AS SELECT * FROM jobfirst_v3.certifications;
CREATE TABLE companies AS SELECT * FROM jobfirst_v3.companies;
CREATE TABLE educations AS SELECT * FROM jobfirst_v3.educations;
CREATE TABLE files AS SELECT * FROM jobfirst_v3.files;
CREATE TABLE point_history AS SELECT * FROM jobfirst_v3.point_history;
CREATE TABLE points AS SELECT * FROM jobfirst_v3.points;
CREATE TABLE positions AS SELECT * FROM jobfirst_v3.positions;
CREATE TABLE projects AS SELECT * FROM jobfirst_v3.projects;
CREATE TABLE resume_comments AS SELECT * FROM jobfirst_v3.resume_comments;
CREATE TABLE resume_likes AS SELECT * FROM jobfirst_v3.resume_likes;
CREATE TABLE resume_shares AS SELECT * FROM jobfirst_v3.resume_shares;
CREATE TABLE resume_skills AS SELECT * FROM jobfirst_v3.resume_skills;
CREATE TABLE resume_templates AS SELECT * FROM jobfirst_v3.resume_templates;
CREATE TABLE resume_v3 AS SELECT * FROM jobfirst_v3.resume_v3;
CREATE TABLE resumes AS SELECT * FROM jobfirst_v3.resumes;
CREATE TABLE skills AS SELECT * FROM jobfirst_v3.skills;
CREATE TABLE user_profiles AS SELECT * FROM jobfirst_v3.user_profiles;
CREATE TABLE user_settings AS SELECT * FROM jobfirst_v3.user_settings;
CREATE TABLE work_experiences AS SELECT * FROM jobfirst_v3.work_experiences;

-- 第五步：重命名表（替换原有表）
-- 备份原表
RENAME TABLE users TO users_backup;
RENAME TABLE user_sessions TO user_sessions_backup;
RENAME TABLE system_configs TO system_configs_backup;
RENAME TABLE operation_logs TO operation_logs_backup;

-- 重命名统一表
RENAME TABLE users_unified TO users;
RENAME TABLE user_sessions_unified TO user_sessions;
RENAME TABLE system_configs_unified TO system_configs;
RENAME TABLE operation_logs_unified TO operation_logs;

-- 第六步：验证数据完整性
SELECT 'Migration completed. Verification:' as status;
SELECT COUNT(*) as total_users FROM users;
SELECT COUNT(*) as total_sessions FROM user_sessions;
SELECT COUNT(*) as total_configs FROM system_configs;
SELECT COUNT(*) as total_logs FROM operation_logs;
SELECT COUNT(*) as total_business_tables FROM information_schema.tables WHERE table_schema = 'jobfirst' AND table_name NOT IN ('users', 'user_sessions', 'system_configs', 'operation_logs', 'users_backup', 'user_sessions_backup', 'system_configs_backup', 'operation_logs_backup');
