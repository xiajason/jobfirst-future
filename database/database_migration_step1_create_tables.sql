-- JobFirst 数据库统一迁移 - 第一步：创建统一表结构
-- 执行前请确保已备份所有数据

USE jobfirst;

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
    INDEX idx_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. 创建系统配置表（统一版本）
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

-- 4. 创建操作日志表（统一版本）
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
    INDEX idx_request_ip (request_ip)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 5. 复制 jobfirst_v3 的业务表到统一数据库
-- 这些表在 jobfirst 中不存在，直接复制

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

-- 验证表创建
SELECT 'Step 1 completed: Tables created successfully' as status;
SELECT COUNT(*) as total_tables FROM information_schema.tables WHERE table_schema = 'jobfirst';
