-- ========================================
-- JobFirst User Service 数据库表结构修复脚本
-- 基于本地数据库最佳实践的统一表结构
-- ========================================

USE jobfirst;

-- 1. 备份现有数据
CREATE TABLE IF NOT EXISTS resumes_backup AS SELECT * FROM resumes;

-- 2. 重命名现有表以保持兼容性
-- 将resumes表重命名为resume_v3，以匹配User Service的模型定义
RENAME TABLE resumes TO resume_v3;

-- 3. 创建resume_skills表（简历技能关联表）
CREATE TABLE IF NOT EXISTS resume_skills (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    resume_id BIGINT UNSIGNED NOT NULL,
    skill_id BIGINT UNSIGNED NOT NULL,
    proficiency_level ENUM('beginner','intermediate','advanced','expert') NOT NULL DEFAULT 'intermediate',
    years_of_experience DECIMAL(3,1) DEFAULT 0.0,
    is_highlighted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (resume_id) REFERENCES resume_v3(id) ON DELETE CASCADE,
    FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE,
    UNIQUE KEY uk_resume_skill (resume_id, skill_id),
    INDEX idx_resume_skills_resume_id (resume_id),
    INDEX idx_resume_skills_skill_id (skill_id),
    INDEX idx_resume_skills_proficiency (proficiency_level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. 创建Casbin权限管理表
CREATE TABLE IF NOT EXISTS casbin_rules (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    ptype VARCHAR(100) NOT NULL,
    v0 VARCHAR(100),
    v1 VARCHAR(100),
    v2 VARCHAR(100),
    v3 VARCHAR(100),
    v4 VARCHAR(100),
    v5 VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_casbin_ptype (ptype),
    INDEX idx_casbin_v0 (v0),
    INDEX idx_casbin_v1 (v1),
    INDEX idx_casbin_v2 (v2)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 5. 创建角色管理表
CREATE TABLE IF NOT EXISTS roles (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(64) NOT NULL UNIQUE,
    description VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_roles_name (name),
    INDEX idx_roles_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 6. 创建角色分配表
CREATE TABLE IF NOT EXISTS role_assignments (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    assigned_by BIGINT UNSIGNED,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL,
    is_active BOOLEAN DEFAULT TRUE,
    reason VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (assigned_by) REFERENCES users(id) ON DELETE SET NULL,
    UNIQUE KEY uk_user_role (user_id, role_id),
    INDEX idx_role_assignments_user_id (user_id),
    INDEX idx_role_assignments_role_id (role_id),
    INDEX idx_role_assignments_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 7. 创建权限审计表
CREATE TABLE IF NOT EXISTS permission_audit_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    result ENUM('allow','deny') NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    request_data JSON,
    response_data JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_audit_user_id (user_id),
    INDEX idx_audit_action (action),
    INDEX idx_audit_resource (resource),
    INDEX idx_audit_result (result),
    INDEX idx_audit_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 8. 创建用户组管理表
CREATE TABLE IF NOT EXISTS user_groups (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_user_groups_name (name),
    INDEX idx_user_groups_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 9. 创建用户组成员表
CREATE TABLE IF NOT EXISTS user_group_members (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    group_id BIGINT UNSIGNED NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES user_groups(id) ON DELETE CASCADE,
    UNIQUE KEY uk_user_group (user_id, group_id),
    INDEX idx_user_group_members_user_id (user_id),
    INDEX idx_user_group_members_group_id (group_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 10. 扩展用户表结构
-- 检查并添加status列
SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE TABLE_SCHEMA = 'jobfirst' 
     AND TABLE_NAME = 'users' 
     AND COLUMN_NAME = 'status') = 0,
    'ALTER TABLE users ADD COLUMN status ENUM(''active'',''inactive'',''suspended'',''pending'') DEFAULT ''active''',
    'SELECT ''status column already exists'' as message'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 检查并添加last_login_at列
SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE TABLE_SCHEMA = 'jobfirst' 
     AND TABLE_NAME = 'users' 
     AND COLUMN_NAME = 'last_login_at') = 0,
    'ALTER TABLE users ADD COLUMN last_login_at TIMESTAMP NULL',
    'SELECT ''last_login_at column already exists'' as message'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 检查并添加login_count列
SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE TABLE_SCHEMA = 'jobfirst' 
     AND TABLE_NAME = 'users' 
     AND COLUMN_NAME = 'login_count') = 0,
    'ALTER TABLE users ADD COLUMN login_count INT DEFAULT 0',
    'SELECT ''login_count column already exists'' as message'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 检查并添加failed_login_count列
SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE TABLE_SCHEMA = 'jobfirst' 
     AND TABLE_NAME = 'users' 
     AND COLUMN_NAME = 'failed_login_count') = 0,
    'ALTER TABLE users ADD COLUMN failed_login_count INT DEFAULT 0',
    'SELECT ''failed_login_count column already exists'' as message'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 检查并添加locked_until列
SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE TABLE_SCHEMA = 'jobfirst' 
     AND TABLE_NAME = 'users' 
     AND COLUMN_NAME = 'locked_until') = 0,
    'ALTER TABLE users ADD COLUMN locked_until TIMESTAMP NULL',
    'SELECT ''locked_until column already exists'' as message'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 检查并添加email_verified列
SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE TABLE_SCHEMA = 'jobfirst' 
     AND TABLE_NAME = 'users' 
     AND COLUMN_NAME = 'email_verified') = 0,
    'ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT FALSE',
    'SELECT ''email_verified column already exists'' as message'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 检查并添加phone_verified列
SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE TABLE_SCHEMA = 'jobfirst' 
     AND TABLE_NAME = 'users' 
     AND COLUMN_NAME = 'phone_verified') = 0,
    'ALTER TABLE users ADD COLUMN phone_verified BOOLEAN DEFAULT FALSE',
    'SELECT ''phone_verified column already exists'' as message'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 11. 插入默认角色数据
INSERT IGNORE INTO roles (name, description) VALUES
('super_admin', '超级管理员 - 拥有所有权限'),
('admin', '系统管理员 - 拥有系统管理权限'),
('content_editor', '内容编辑 - 拥有内容管理权限'),
('user', '普通用户 - 基础用户权限');

-- 12. 插入Casbin权限规则
INSERT IGNORE INTO casbin_rules (ptype, v0, v1, v2) VALUES
-- 超级管理员权限
('p', 'super_admin', '*', '*'),
-- 管理员权限
('p', 'admin', 'users', 'read'),
('p', 'admin', 'users', 'write'),
('p', 'admin', 'users', 'delete'),
('p', 'admin', 'resumes', 'read'),
('p', 'admin', 'resumes', 'write'),
('p', 'admin', 'resumes', 'delete'),
('p', 'admin', 'roles', 'read'),
('p', 'admin', 'roles', 'write'),
-- 内容编辑权限
('p', 'content_editor', 'resumes', 'read'),
('p', 'content_editor', 'resumes', 'write'),
('p', 'content_editor', 'comments', 'read'),
('p', 'content_editor', 'comments', 'write'),
('p', 'content_editor', 'templates', 'read'),
('p', 'content_editor', 'templates', 'write'),
-- 普通用户权限
('p', 'user', 'resumes', 'read'),
('p', 'user', 'resumes', 'write'),
('p', 'user', 'profile', 'read'),
('p', 'user', 'profile', 'write'),
('p', 'user', 'comments', 'read'),
('p', 'user', 'comments', 'write');

-- 13. 为用户分配默认角色
INSERT IGNORE INTO role_assignments (user_id, role_id) VALUES
(1, 4), -- testuser -> user
(2, 4), -- demouser -> user  
(8, 1); -- jobfirst -> super_admin

-- 14. 创建默认用户组
INSERT IGNORE INTO user_groups (name, description) VALUES
('developers', '开发人员组'),
('designers', '设计师组'),
('managers', '管理人员组'),
('testers', '测试人员组');

-- 15. 验证表结构
SHOW TABLES LIKE '%resume%';
SHOW TABLES LIKE '%casbin%';
SHOW TABLES LIKE '%role%';
SHOW TABLES LIKE '%user_group%';

-- 16. 显示修复结果
SELECT 'Database tables fixed successfully!' as status;
SELECT COUNT(*) as resume_count FROM resume_v3;
SELECT COUNT(*) as skills_count FROM skills;
SELECT COUNT(*) as roles_count FROM roles;
SELECT COUNT(*) as casbin_rules_count FROM casbin_rules;
