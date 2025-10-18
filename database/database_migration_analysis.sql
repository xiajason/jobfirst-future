-- JobFirst 数据库统一迁移分析脚本
-- 分析各个数据库的表结构和字段定义

-- 1. 分析 jobfirst 数据库表结构
SELECT 'jobfirst' as database_name, table_name, table_comment 
FROM information_schema.tables 
WHERE table_schema = 'jobfirst' 
ORDER BY table_name;

-- 2. 分析 jobfirst_v3 数据库表结构  
SELECT 'jobfirst_v3' as database_name, table_name, table_comment 
FROM information_schema.tables 
WHERE table_schema = 'jobfirst_v3' 
ORDER BY table_name;

-- 3. 分析表字段结构对比
-- jobfirst.users vs jobfirst_v3.users
SELECT 
    'users' as table_name,
    'jobfirst' as source_db,
    column_name,
    data_type,
    is_nullable,
    column_default,
    column_comment
FROM information_schema.columns 
WHERE table_schema = 'jobfirst' AND table_name = 'users'
UNION ALL
SELECT 
    'users' as table_name,
    'jobfirst_v3' as source_db,
    column_name,
    data_type,
    is_nullable,
    column_default,
    column_comment
FROM information_schema.columns 
WHERE table_schema = 'jobfirst_v3' AND table_name = 'users'
ORDER BY table_name, source_db;

-- 4. 分析 user_sessions 表结构对比
SELECT 
    'user_sessions' as table_name,
    'jobfirst' as source_db,
    column_name,
    data_type,
    is_nullable,
    column_default,
    column_comment
FROM information_schema.columns 
WHERE table_schema = 'jobfirst' AND table_name = 'user_sessions'
UNION ALL
SELECT 
    'user_sessions' as table_name,
    'jobfirst_v3' as source_db,
    column_name,
    data_type,
    is_nullable,
    column_default,
    column_comment
FROM information_schema.columns 
WHERE table_schema = 'jobfirst_v3' AND table_name = 'user_sessions'
ORDER BY table_name, source_db;

-- 5. 统计各数据库表数量
SELECT 
    table_schema as database_name,
    COUNT(*) as table_count
FROM information_schema.tables 
WHERE table_schema IN ('jobfirst', 'jobfirst_v3', 'jobfirst_advanced')
GROUP BY table_schema
ORDER BY table_schema;
