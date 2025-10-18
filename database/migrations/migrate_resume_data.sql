-- 简历数据迁移脚本
-- 用途：将现有数据从旧架构迁移到新架构
-- 创建时间：2025-09-13
-- 版本：v1.0

-- ==============================================
-- 第一步：验证迁移前的数据状态
-- ==============================================

-- 检查现有数据
SELECT '迁移前数据统计:' as info;
SELECT 'resumes表记录数:' as table_name, COUNT(*) as record_count FROM resumes;
SELECT 'resume_files表记录数:' as table_name, COUNT(*) as record_count FROM resume_files;
SELECT 'resume_parsing_tasks表记录数:' as table_name, COUNT(*) as record_count FROM resume_parsing_tasks;

-- ==============================================
-- 第二步：数据迁移到新架构
-- ==============================================

-- 迁移resumes表数据到resume_metadata表
INSERT INTO resume_metadata (
    id, user_id, file_id, title, creation_mode, template_id, status, 
    is_public, view_count, parsing_status, parsing_error, created_at, updated_at
)
SELECT 
    id, 
    user_id, 
    file_id, 
    title, 
    COALESCE(creation_mode, 'markdown') as creation_mode,
    template_id, -- 使用备份表中的template_id
    COALESCE(status, 'draft') as status,
    COALESCE(is_public, FALSE) as is_public,
    COALESCE(view_count, 0) as view_count,
    COALESCE(parsing_status, 'pending') as parsing_status,
    parsing_error,
    created_at,
    updated_at
FROM resumes_backup
WHERE id IS NOT NULL;

-- 更新SQLite数据库路径
UPDATE resume_metadata 
SET sqlite_db_path = CONCAT('./data/users/', user_id, '/resume.db')
WHERE sqlite_db_path IS NULL;

-- ==============================================
-- 第三步：验证迁移结果
-- ==============================================

-- 验证数据迁移结果
SELECT '迁移后数据统计:' as info;
SELECT 'resume_metadata表记录数:' as table_name, COUNT(*) as record_count FROM resume_metadata;
SELECT 'resumes_backup表记录数:' as table_name, COUNT(*) as record_count FROM resumes_backup;

-- 验证数据一致性
SELECT 
    '数据一致性检查:' as info,
    (SELECT COUNT(*) FROM resume_metadata) as metadata_count,
    (SELECT COUNT(*) FROM resumes_backup) as backup_count,
    CASE 
        WHEN (SELECT COUNT(*) FROM resume_metadata) = (SELECT COUNT(*) FROM resumes_backup) 
        THEN '✅ 数据迁移成功' 
        ELSE '❌ 数据迁移失败' 
    END as migration_status;

-- 显示迁移后的表结构
SELECT 'resume_metadata表结构:' as info;
DESCRIBE resume_metadata;

-- ==============================================
-- 第四步：创建向后兼容视图
-- ==============================================

-- 创建向后兼容的视图，映射到新的表结构
CREATE OR REPLACE VIEW resumes AS
SELECT 
    rm.id,
    rm.user_id,
    rm.file_id,
    rm.title,
    rm.creation_mode,
    rm.template_id,
    rm.status,
    rm.is_public,
    rm.view_count,
    rm.parsing_status,
    rm.parsing_error,
    rm.created_at,
    rm.updated_at
FROM resume_metadata rm;

-- 验证视图创建
SELECT '向后兼容视图创建完成' as info;
SELECT COUNT(*) as view_record_count FROM resumes;

-- ==============================================
-- 第五步：清理临时表（可选）
-- ==============================================

-- 注意：建议先验证系统正常运行后再删除备份表
-- 这里只是准备清理命令，不实际执行

-- DROP TABLE IF EXISTS resumes_backup;
-- DROP TABLE IF EXISTS resume_files_backup;
-- DROP TABLE IF EXISTS resume_parsing_tasks_backup;

SELECT '数据迁移脚本执行完成！' as message;
SELECT '下一步：创建Go代码逻辑修正' as next_step;
