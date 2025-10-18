-- JobFirst 数据库统一迁移 - 第二步：数据迁移
-- 执行前请确保第一步已完成

USE jobfirst;

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

-- 验证数据迁移
SELECT 'Step 2 completed: Data migration completed' as status;
SELECT COUNT(*) as users_count FROM users_unified;
SELECT COUNT(*) as sessions_count FROM user_sessions_unified;
SELECT COUNT(*) as configs_count FROM system_configs_unified;
SELECT COUNT(*) as logs_count FROM operation_logs_unified;
