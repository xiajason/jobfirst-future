-- 统一认证系统数据库迁移脚本
-- 解决角色权限系统的复杂性问题

-- 1. 备份现有用户数据
CREATE TABLE IF NOT EXISTS users_backup AS SELECT * FROM users;

-- 2. 创建访问日志表（如果不存在）
CREATE TABLE IF NOT EXISTS access_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED,
    service VARCHAR(50) DEFAULT 'unified-auth',
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    result VARCHAR(20) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_service (service),
    INDEX idx_action (action),
    INDEX idx_resource (resource),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. 确保现有用户表有正确的角色映射
-- 现有users表已经包含Dev-Team的7种角色，无需迁移

-- 7. 插入权限数据
INSERT IGNORE INTO permissions (name, resource, action, description) VALUES
('read:public', 'public', 'read', '读取公开内容'),
('read:own', 'own', 'read', '读取自己的内容'),
('write:own', 'own', 'write', '修改自己的内容'),
('read:all', 'all', 'read', '读取所有内容'),
('write:all', 'all', 'write', '修改所有内容'),
('delete:own', 'own', 'delete', '删除自己的内容'),
('delete:all', 'all', 'delete', '删除所有内容'),
('admin:users', 'users', 'admin', '用户管理'),
('admin:system', 'system', 'admin', '系统管理');

-- 4. 为Dev-Team角色分配权限
-- 访客权限
INSERT IGNORE INTO user_permissions (user_id, permission_id, is_active, granted_at)
SELECT u.id, p.id, 1, NOW()
FROM users u, permissions p
WHERE u.role = 'guest' AND p.name = 'read:public';

-- 前端开发权限
INSERT IGNORE INTO user_permissions (user_id, permission_id, is_active, granted_at)
SELECT u.id, p.id, 1, NOW()
FROM users u, permissions p
WHERE u.role = 'frontend_dev' AND p.name IN ('read:public', 'read:own', 'write:own');

-- 后端开发权限
INSERT IGNORE INTO user_permissions (user_id, permission_id, is_active, granted_at)
SELECT u.id, p.id, 1, NOW()
FROM users u, permissions p
WHERE u.role = 'backend_dev' AND p.name IN ('read:public', 'read:own', 'write:own', 'read:all', 'write:all');

-- 测试工程师权限
INSERT IGNORE INTO user_permissions (user_id, permission_id, is_active, granted_at)
SELECT u.id, p.id, 1, NOW()
FROM users u, permissions p
WHERE u.role = 'qa_engineer' AND p.name IN ('read:public', 'read:own', 'write:own', 'read:all');

-- 开发负责人权限
INSERT IGNORE INTO user_permissions (user_id, permission_id, is_active, granted_at)
SELECT u.id, p.id, 1, NOW()
FROM users u, permissions p
WHERE u.role = 'dev_lead' AND p.name IN ('read:public', 'read:own', 'write:own', 'read:all', 'write:all', 'delete:own', 'admin:users');

-- 系统管理员权限
INSERT IGNORE INTO user_permissions (user_id, permission_id, is_active, granted_at)
SELECT u.id, p.id, 1, NOW()
FROM users u, permissions p
WHERE u.role = 'system_admin' AND p.name IN ('read:public', 'read:own', 'write:own', 'read:all', 'write:all', 'delete:own', 'admin:users', 'admin:system');

-- 超级管理员权限（所有权限）
INSERT IGNORE INTO user_permissions (user_id, permission_id, is_active, granted_at)
SELECT u.id, p.id, 1, NOW()
FROM users u, permissions p
WHERE u.role = 'super_admin';

-- 5. 创建默认超级管理员（如果不存在）
INSERT IGNORE INTO users (username, email, password_hash, role, status, created_at, updated_at)
VALUES ('admin', 'admin@jobfirst.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'super_admin', 'active', NOW(), NOW());

-- 6. 验证数据迁移
SELECT 'Migration completed successfully' as status;
SELECT COUNT(*) as total_users FROM users;
SELECT role, COUNT(*) as count FROM users GROUP BY role;
SELECT COUNT(*) as total_permissions FROM permissions;
SELECT COUNT(*) as total_user_permissions FROM user_permissions;
