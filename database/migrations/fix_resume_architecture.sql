-- 简历存储架构修复脚本
-- 用途：修正MySQL数据库结构，实现"MySQL只存储元数据"的设计原则
-- 创建时间：2025-09-13

USE jobfirst;

-- ==============================================
-- 第一步：备份现有数据到临时表
-- ==============================================

-- 备份现有resumes表数据
CREATE TABLE IF NOT EXISTS resumes_backup AS SELECT * FROM resumes;

-- 备份现有resume_files表数据  
CREATE TABLE IF NOT EXISTS resume_files_backup AS SELECT * FROM resume_files;

-- 备份现有resume_parsing_tasks表数据
CREATE TABLE IF NOT EXISTS resume_parsing_tasks_backup AS SELECT * FROM resume_parsing_tasks;

-- ==============================================
-- 第二步：创建新的元数据表结构
-- ==============================================

-- 创建新的简历元数据表（符合设计原则）
DROP TABLE IF EXISTS resume_metadata;
CREATE TABLE resume_metadata (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,  -- 匹配users.id的类型
    file_id INT NOT NULL,
    title VARCHAR(255) NOT NULL,
    creation_mode VARCHAR(20) DEFAULT 'upload',
    status VARCHAR(20) DEFAULT 'draft',
    parsing_status VARCHAR(20) DEFAULT 'pending',
    parsing_error TEXT,
    sqlite_db_path VARCHAR(500), -- 指向用户SQLite数据库路径
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 外键约束
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (file_id) REFERENCES resume_files(id) ON DELETE CASCADE,
    
    -- 索引
    INDEX idx_user_id (user_id),
    INDEX idx_parsing_status (parsing_status),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==============================================
-- 第三步：修正现有表结构
-- ==============================================

-- 修正resume_files表（移除违规字段，保持元数据）
-- 注意：MySQL不支持DROP COLUMN IF EXISTS，需要先检查字段是否存在
SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE TABLE_SCHEMA = 'jobfirst' 
     AND TABLE_NAME = 'resume_files' 
     AND COLUMN_NAME = 'content') > 0,
    'ALTER TABLE resume_files DROP COLUMN content',
    'SELECT "resume_files.content column does not exist"'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE TABLE_SCHEMA = 'jobfirst' 
     AND TABLE_NAME = 'resume_files' 
     AND COLUMN_NAME = 'parsing_result') > 0,
    'ALTER TABLE resume_files DROP COLUMN parsing_result',
    'SELECT "resume_files.parsing_result column does not exist"'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE TABLE_SCHEMA = 'jobfirst' 
     AND TABLE_NAME = 'resume_files' 
     AND COLUMN_NAME = 'postgresql_id') > 0,
    'ALTER TABLE resume_files DROP COLUMN postgresql_id',
    'SELECT "resume_files.postgresql_id column does not exist"'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 修正resume_parsing_tasks表（移除违规字段，保持元数据）
SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE TABLE_SCHEMA = 'jobfirst' 
     AND TABLE_NAME = 'resume_parsing_tasks' 
     AND COLUMN_NAME = 'content') > 0,
    'ALTER TABLE resume_parsing_tasks DROP COLUMN content',
    'SELECT "resume_parsing_tasks.content column does not exist"'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE TABLE_SCHEMA = 'jobfirst' 
     AND TABLE_NAME = 'resume_parsing_tasks' 
     AND COLUMN_NAME = 'postgresql_id') > 0,
    'ALTER TABLE resume_parsing_tasks DROP COLUMN postgresql_id',
    'SELECT "resume_parsing_tasks.postgresql_id column does not exist"'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ==============================================
-- 第四步：数据迁移
-- ==============================================

-- 将现有resumes表数据迁移到新的resume_metadata表
INSERT INTO resume_metadata (
    id, user_id, file_id, title, creation_mode, status, 
    parsing_status, parsing_error, created_at, updated_at
)
SELECT 
    id, user_id, file_id, title, creation_mode, status,
    parsing_status, parsing_error, created_at, updated_at
FROM resumes_backup
WHERE id IS NOT NULL;

-- ==============================================
-- 第五步：清理和验证
-- ==============================================

-- 验证数据迁移结果
SELECT 
    'resume_metadata' as table_name,
    COUNT(*) as record_count
FROM resume_metadata
UNION ALL
SELECT 
    'resumes_backup' as table_name,
    COUNT(*) as record_count
FROM resumes_backup;

-- 显示新表结构
DESCRIBE resume_metadata;

-- 显示修正后的表结构
DESCRIBE resume_files;
DESCRIBE resume_parsing_tasks;

-- ==============================================
-- 第六步：创建视图（向后兼容）
-- ==============================================

-- 创建向后兼容的视图，映射到新的表结构
CREATE OR REPLACE VIEW resumes AS
SELECT 
    rm.id,
    rm.user_id,
    rm.file_id,
    rm.title,
    rm.creation_mode,
    NULL as template_id,  -- 新表中没有此字段，设为NULL
    rm.status,
    FALSE as is_public,   -- 新表中没有此字段，设为默认值
    0 as view_count,      -- 新表中没有此字段，设为默认值
    rm.parsing_status,
    rm.parsing_error,
    rm.created_at,
    rm.updated_at
FROM resume_metadata rm;

-- ==============================================
-- 完成提示
-- ==============================================

SELECT 'MySQL数据库结构修正完成！' as message;
SELECT '下一步：创建正确的SQLite表结构' as next_step;
