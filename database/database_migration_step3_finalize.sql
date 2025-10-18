-- JobFirst 数据库统一迁移 - 第三步：完成迁移
-- 执行前请确保前两步已完成并验证无误

USE jobfirst;

-- 1. 备份原表（重命名）
RENAME TABLE users TO users_backup_$(date +%Y%m%d_%H%M%S);
RENAME TABLE user_sessions TO user_sessions_backup_$(date +%Y%m%d_%H%M%S);
RENAME TABLE system_configs TO system_configs_backup_$(date +%Y%m%d_%H%M%S);
RENAME TABLE operation_logs TO operation_logs_backup_$(date +%Y%m%d_%H%M%S);

-- 2. 重命名统一表为正式表名
RENAME TABLE users_unified TO users;
RENAME TABLE user_sessions_unified TO user_sessions;
RENAME TABLE system_configs_unified TO system_configs;
RENAME TABLE operation_logs_unified TO operation_logs;

-- 3. 添加外键约束
ALTER TABLE user_sessions ADD CONSTRAINT fk_user_sessions_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE operation_logs ADD CONSTRAINT fk_operation_logs_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

-- 4. 最终验证
SELECT 'Step 3 completed: Migration finalized' as status;
SELECT COUNT(*) as total_users FROM users;
SELECT COUNT(*) as total_sessions FROM user_sessions;
SELECT COUNT(*) as total_configs FROM system_configs;
SELECT COUNT(*) as total_logs FROM operation_logs;
SELECT COUNT(*) as total_business_tables FROM information_schema.tables 
    WHERE table_schema = 'jobfirst' 
    AND table_name NOT IN ('users', 'user_sessions', 'system_configs', 'operation_logs')
    AND table_name NOT LIKE '%_backup_%';

-- 显示所有表
SELECT table_name FROM information_schema.tables 
    WHERE table_schema = 'jobfirst' 
    ORDER BY table_name;
