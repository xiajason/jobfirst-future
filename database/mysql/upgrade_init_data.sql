-- JobFirst 数据库升级初始化数据脚本
-- 版本: V4.0
-- 日期: 2025年1月6日
-- 描述: 为升级后的数据库初始化基础数据

-- ==============================================
-- 角色和权限初始化
-- ==============================================

-- 插入系统角色
INSERT INTO roles (name, display_name, description, level, is_system) VALUES
('super_admin', '超级管理员', '系统最高权限管理员，拥有所有权限', 5, 1),
('system_admin', '系统管理员', '系统管理权限，负责系统维护和配置', 4, 1),
('data_admin', '数据管理员', '数据管理权限，负责数据维护和分析', 3, 1),
('hr_admin', 'HR管理员', '人力资源管理权限，负责用户和职位管理', 3, 1),
('company_admin', '企业管理员', '企业管理权限，负责企业信息管理', 2, 1),
('regular_user', '普通用户', '普通用户权限，基础功能使用', 1, 1);

-- 插入基础权限
INSERT INTO permissions (name, display_name, resource, action, level, is_system) VALUES
-- Level 4 极高敏感权限
('users.password.read', '查看用户密码', 'users', 'password.read', 4, 1),
('users.password.write', '修改用户密码', 'users', 'password.write', 4, 1),
('sessions.token.read', '查看会话令牌', 'sessions', 'token.read', 4, 1),
('sessions.token.write', '管理会话令牌', 'sessions', 'token.write', 4, 1),

-- Level 3 高敏感权限
('users.personal.read', '查看个人信息', 'users', 'personal.read', 3, 1),
('users.personal.write', '修改个人信息', 'users', 'personal.write', 3, 1),
('users.personal.delete', '删除个人信息', 'users', 'personal.delete', 3, 1),
('files.sensitive.read', '查看敏感文件', 'files', 'sensitive.read', 3, 1),
('files.sensitive.write', '管理敏感文件', 'files', 'sensitive.write', 3, 1),
('points.balance.read', '查看积分余额', 'points', 'balance.read', 3, 1),
('points.balance.write', '管理积分余额', 'points', 'balance.write', 3, 1),

-- Level 2 中敏感权限
('resumes.read', '查看简历', 'resumes', 'read', 2, 1),
('resumes.write', '修改简历', 'resumes', 'write', 2, 1),
('resumes.delete', '删除简历', 'resumes', 'delete', 2, 1),
('jobs.read', '查看职位', 'jobs', 'read', 2, 1),
('jobs.write', '修改职位', 'jobs', 'write', 2, 1),
('jobs.delete', '删除职位', 'jobs', 'delete', 2, 1),
('companies.read', '查看企业', 'companies', 'read', 2, 1),
('companies.write', '修改企业', 'companies', 'write', 2, 1),
('companies.delete', '删除企业', 'companies', 'delete', 2, 1),
('skills.read', '查看技能', 'skills', 'read', 2, 1),
('skills.write', '修改技能', 'skills', 'write', 2, 1),

-- Level 1 低敏感权限
('public.read', '公开数据读取', 'public', 'read', 1, 1),
('statistics.read', '统计数据读取', 'statistics', 'read', 1, 1),
('templates.read', '模板读取', 'templates', 'read', 1, 1),
('banners.read', '轮播图读取', 'banners', 'read', 1, 1);

-- 为角色分配权限
-- 超级管理员 - 所有权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'super_admin';

-- 系统管理员 - Level 1-4 权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'system_admin' AND p.level <= 4;

-- 数据管理员 - Level 1-3 权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'data_admin' AND p.level <= 3;

-- HR管理员 - Level 1-3 权限（用户和职位相关）
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'hr_admin' AND p.level <= 3 
AND (p.resource IN ('users', 'resumes', 'jobs', 'companies', 'skills') OR p.level = 1);

-- 企业管理员 - Level 1-2 权限（企业相关）
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'company_admin' AND p.level <= 2 
AND (p.resource IN ('companies', 'jobs') OR p.level = 1);

-- 普通用户 - Level 1-2 权限（个人数据）
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'regular_user' AND p.level <= 2 
AND (p.resource IN ('users', 'resumes', 'public', 'statistics', 'templates', 'banners') OR p.level = 1);

-- ==============================================
-- 数据分类标签初始化
-- ==============================================

-- 插入数据分类标签
INSERT INTO data_classification_tags (table_name, field_name, sensitivity_level, protection_method, retention_period, encryption_required, access_control_required, audit_required) VALUES
-- Level 4 极高敏感字段
('users', 'password_hash', 'critical', 'bcrypt_encryption', 0, 1, 1, 1),
('user_sessions', 'session_token', 'critical', 'jwt_encryption', 0, 1, 1, 1),
('user_sessions', 'refresh_token', 'critical', 'jwt_encryption', 0, 1, 1, 1),

-- Level 3 高敏感字段
('users', 'email', 'high', 'aes256_encryption', 2555, 1, 1, 1),
('users', 'phone', 'high', 'aes256_encryption', 2555, 1, 1, 1),
('users', 'first_name', 'high', 'access_control', 2555, 0, 1, 1),
('users', 'last_name', 'high', 'access_control', 2555, 0, 1, 1),
('user_profiles', 'date_of_birth', 'high', 'aes256_encryption', 2555, 1, 1, 1),
('user_profiles', 'location', 'high', 'access_control', 2555, 0, 1, 1),
('user_profiles', 'gender', 'high', 'access_control', 2555, 0, 1, 1),
('user_profiles', 'nationality', 'high', 'access_control', 2555, 0, 1, 1),
('points', 'balance', 'high', 'access_control', 2555, 0, 1, 1),
('points', 'total_earned', 'high', 'access_control', 2555, 0, 1, 1),
('points', 'total_spent', 'high', 'access_control', 2555, 0, 1, 1),
('point_history', 'points', 'high', 'access_control', 2555, 0, 1, 1),
('point_history', 'type', 'high', 'access_control', 2555, 0, 1, 1),
('point_history', 'reason', 'high', 'access_control', 2555, 0, 1, 1),
('point_history', 'description', 'high', 'access_control', 2555, 0, 1, 1),
('point_history', 'balance_after', 'high', 'access_control', 2555, 0, 1, 1),
('user_sessions', 'device_info', 'high', 'access_control', 2555, 0, 1, 1),
('user_sessions', 'ip_address', 'high', 'access_control', 2555, 0, 1, 1),
('user_sessions', 'user_agent', 'high', 'access_control', 2555, 0, 1, 1),
('files', 'original_filename', 'high', 'access_control', 2555, 0, 1, 1),
('files', 'file_path', 'high', 'access_control', 2555, 0, 1, 1),
('user_settings', 'timezone', 'high', 'access_control', 2555, 0, 1, 1),
('user_settings', 'privacy_level', 'high', 'access_control', 2555, 0, 1, 1),
('user_settings', 'resume_visibility', 'high', 'access_control', 2555, 0, 1, 1),

-- Level 2 中敏感字段
('users', 'username', 'medium', 'access_control', 1095, 0, 1, 0),
('users', 'avatar_url', 'medium', 'access_control', 1095, 0, 1, 0),
('users', 'last_login_at', 'medium', 'access_control', 1095, 0, 1, 0),
('user_profiles', 'bio', 'medium', 'access_control', 1095, 0, 1, 0),
('user_profiles', 'website', 'medium', 'access_control', 1095, 0, 1, 0),
('user_profiles', 'linkedin_url', 'medium', 'access_control', 1095, 0, 1, 0),
('user_profiles', 'github_url', 'medium', 'access_control', 1095, 0, 1, 0),
('user_profiles', 'twitter_url', 'medium', 'access_control', 1095, 0, 1, 0),
('user_profiles', 'languages', 'medium', 'access_control', 1095, 0, 1, 0),
('user_profiles', 'skills', 'medium', 'access_control', 1095, 0, 1, 0),
('user_profiles', 'interests', 'medium', 'access_control', 1095, 0, 1, 0),
('resumes', 'title', 'medium', 'access_control', 1095, 0, 1, 0),
('resumes', 'summary', 'medium', 'access_control', 1095, 0, 1, 0),
('resumes', 'content', 'medium', 'access_control', 1095, 0, 1, 0),
('resumes', 'visibility', 'medium', 'access_control', 1095, 0, 1, 0),
('files', 'filename', 'medium', 'access_control', 1095, 0, 1, 0),
('files', 'description', 'medium', 'access_control', 1095, 0, 1, 0),
('files', 'tags', 'medium', 'access_control', 1095, 0, 1, 0),
('files', 'is_public', 'medium', 'access_control', 1095, 0, 1, 0),
('user_settings', 'theme', 'medium', 'access_control', 1095, 0, 1, 0),
('user_settings', 'language', 'medium', 'access_control', 1095, 0, 1, 0),
('user_settings', 'email_notifications', 'medium', 'access_control', 1095, 0, 1, 0),
('user_settings', 'push_notifications', 'medium', 'access_control', 1095, 0, 1, 0),
('resume_analytics', 'ai_score', 'medium', 'access_control', 1095, 0, 1, 0),
('resume_analytics', 'ai_feedback', 'medium', 'access_control', 1095, 0, 1, 0),
('resume_analytics', 'keywords', 'medium', 'access_control', 1095, 0, 1, 0),
('resume_analytics', 'skills_matched', 'medium', 'access_control', 1095, 0, 1, 0),

-- Level 1 低敏感字段
('users', 'id', 'low', 'none', 365, 0, 0, 0),
('users', 'uuid', 'low', 'none', 365, 0, 0, 0),
('users', 'status', 'low', 'none', 365, 0, 0, 0),
('users', 'email_verified', 'low', 'none', 365, 0, 0, 0),
('users', 'phone_verified', 'low', 'none', 365, 0, 0, 0),
('users', 'created_at', 'low', 'none', 365, 0, 0, 0),
('users', 'updated_at', 'low', 'none', 365, 0, 0, 0),
('users', 'deleted_at', 'low', 'none', 365, 0, 0, 0),
('user_profiles', 'id', 'low', 'none', 365, 0, 0, 0),
('user_profiles', 'user_id', 'low', 'none', 365, 0, 0, 0),
('user_profiles', 'created_at', 'low', 'none', 365, 0, 0, 0),
('user_profiles', 'updated_at', 'low', 'none', 365, 0, 0, 0),
('resumes', 'id', 'low', 'none', 365, 0, 0, 0),
('resumes', 'uuid', 'low', 'none', 365, 0, 0, 0),
('resumes', 'user_id', 'low', 'none', 365, 0, 0, 0),
('resumes', 'template_id', 'low', 'none', 365, 0, 0, 0),
('resumes', 'status', 'low', 'none', 365, 0, 0, 0),
('resumes', 'view_count', 'low', 'none', 365, 0, 0, 0),
('resumes', 'download_count', 'low', 'none', 365, 0, 0, 0),
('resumes', 'share_count', 'low', 'none', 365, 0, 0, 0),
('resumes', 'is_default', 'low', 'none', 365, 0, 0, 0),
('resumes', 'created_at', 'low', 'none', 365, 0, 0, 0),
('resumes', 'updated_at', 'low', 'none', 365, 0, 0, 0),
('resumes', 'deleted_at', 'low', 'none', 365, 0, 0, 0),
('jobs', 'id', 'low', 'none', 365, 0, 0, 0),
('jobs', 'title', 'low', 'none', 365, 0, 0, 0),
('jobs', 'company', 'low', 'none', 365, 0, 0, 0),
('jobs', 'location', 'low', 'none', 365, 0, 0, 0),
('jobs', 'salary_min', 'low', 'none', 365, 0, 0, 0),
('jobs', 'salary_max', 'low', 'none', 365, 0, 0, 0),
('jobs', 'description', 'low', 'none', 365, 0, 0, 0),
('jobs', 'requirements', 'low', 'none', 365, 0, 0, 0),
('jobs', 'status', 'low', 'none', 365, 0, 0, 0),
('jobs', 'created_at', 'low', 'none', 365, 0, 0, 0),
('jobs', 'updated_at', 'low', 'none', 365, 0, 0, 0),
('points', 'id', 'low', 'none', 365, 0, 0, 0),
('points', 'user_id', 'low', 'none', 365, 0, 0, 0),
('points', 'created_at', 'low', 'none', 365, 0, 0, 0),
('points', 'updated_at', 'low', 'none', 365, 0, 0, 0),
('files', 'id', 'low', 'none', 365, 0, 0, 0),
('files', 'uuid', 'low', 'none', 365, 0, 0, 0),
('files', 'user_id', 'low', 'none', 365, 0, 0, 0),
('files', 'file_size', 'low', 'none', 365, 0, 0, 0),
('files', 'mime_type', 'low', 'none', 365, 0, 0, 0),
('files', 'file_type', 'low', 'none', 365, 0, 0, 0),
('files', 'download_count', 'low', 'none', 365, 0, 0, 0),
('files', 'created_at', 'low', 'none', 365, 0, 0, 0),
('files', 'updated_at', 'low', 'none', 365, 0, 0, 0),
('files', 'deleted_at', 'low', 'none', 365, 0, 0, 0),
('point_history', 'id', 'low', 'none', 365, 0, 0, 0),
('point_history', 'user_id', 'low', 'none', 365, 0, 0, 0),
('point_history', 'reference_type', 'low', 'none', 365, 0, 0, 0),
('point_history', 'reference_id', 'low', 'none', 365, 0, 0, 0),
('point_history', 'created_at', 'low', 'none', 365, 0, 0, 0),
('resume_analytics', 'id', 'low', 'none', 365, 0, 0, 0),
('resume_analytics', 'resume_id', 'low', 'none', 365, 0, 0, 0),
('resume_analytics', 'user_id', 'low', 'none', 365, 0, 0, 0),
('resume_analytics', 'view_count', 'low', 'none', 365, 0, 0, 0),
('resume_analytics', 'download_count', 'low', 'none', 365, 0, 0, 0),
('resume_analytics', 'share_count', 'low', 'none', 365, 0, 0, 0),
('resume_analytics', 'like_count', 'low', 'none', 365, 0, 0, 0),
('resume_analytics', 'comment_count', 'low', 'none', 365, 0, 0, 0),
('resume_analytics', 'last_analyzed_at', 'low', 'none', 365, 0, 0, 0),
('resume_analytics', 'created_at', 'low', 'none', 365, 0, 0, 0),
('resume_analytics', 'updated_at', 'low', 'none', 365, 0, 0, 0),
('user_sessions', 'id', 'low', 'none', 365, 0, 0, 0),
('user_sessions', 'user_id', 'low', 'none', 365, 0, 0, 0),
('user_sessions', 'expires_at', 'low', 'none', 365, 0, 0, 0),
('user_sessions', 'created_at', 'low', 'none', 365, 0, 0, 0),
('user_sessions', 'updated_at', 'low', 'none', 365, 0, 0, 0),
('user_settings', 'id', 'low', 'none', 365, 0, 0, 0),
('user_settings', 'user_id', 'low', 'none', 365, 0, 0, 0),
('user_settings', 'created_at', 'low', 'none', 365, 0, 0, 0),
('user_settings', 'updated_at', 'low', 'none', 365, 0, 0, 0);

-- ==============================================
-- 数据生命周期策略初始化
-- ==============================================

-- 插入数据生命周期策略
INSERT INTO data_lifecycle_policies (table_name, policy_name, retention_period, archive_period, deletion_period, archive_location) VALUES
-- 用户会话数据 - 短期保留
('user_sessions', 'session_retention', 30, 60, 90, '/archive/sessions'),

-- 审计日志 - 中期保留
('permission_audit_logs', 'audit_retention', 365, 730, 1095, '/archive/audit'),
('data_access_logs', 'access_retention', 365, 730, 1095, '/archive/access'),

-- AI服务日志 - 中期保留
('ai_service_logs', 'ai_log_retention', 180, 365, 730, '/archive/ai_logs'),
('ai_performance_metrics', 'metrics_retention', 365, 730, 1095, '/archive/metrics'),

-- AI缓存 - 短期保留
('ai_cache', 'cache_retention', 7, 14, 30, '/archive/cache'),

-- 推荐数据 - 中期保留
('job_recommendations', 'recommendation_retention', 90, 180, 365, '/archive/recommendations'),
('company_recommendations', 'company_recommendation_retention', 90, 180, 365, '/archive/company_recommendations'),

-- 对话数据 - 中期保留
('ai_conversations', 'conversation_retention', 180, 365, 730, '/archive/conversations'),
('ai_messages', 'message_retention', 180, 365, 730, '/archive/messages'),

-- 积分历史 - 长期保留
('point_history', 'point_history_retention', 2555, 3650, 5475, '/archive/point_history'),

-- 文件数据 - 长期保留
('files', 'file_retention', 1095, 1825, 2555, '/archive/files');

-- ==============================================
-- AI模型初始化
-- ==============================================

-- 插入基础AI模型
INSERT INTO ai_models (name, version, model_type, provider, model_identifier, description, cost_per_token, is_active) VALUES
('gemma3-4b', '1.0', 'text_generation', 'ollama', 'gemma3:4b', 'Google Gemma 3 4B模型，用于文本生成和对话', 0.000001, 1),
('text-embedding-ada-002', '1.0', 'embedding', 'openai', 'text-embedding-ada-002', 'OpenAI文本嵌入模型，用于向量化', 0.0001, 1),
('gpt-3.5-turbo', '1.0', 'text_generation', 'openai', 'gpt-3.5-turbo', 'OpenAI GPT-3.5模型，用于智能对话', 0.002, 1),
('claude-3-haiku', '1.0', 'text_generation', 'anthropic', 'claude-3-haiku', 'Anthropic Claude 3 Haiku模型，用于快速文本生成', 0.00025, 1),
('gpt-4', '1.0', 'text_generation', 'openai', 'gpt-4', 'OpenAI GPT-4模型，用于高质量文本生成', 0.03, 1),
('claude-3-sonnet', '1.0', 'text_generation', 'anthropic', 'claude-3-sonnet', 'Anthropic Claude 3 Sonnet模型，用于平衡性能和成本', 0.003, 1);

-- 插入模型版本
INSERT INTO model_versions (model_id, version, config, performance_score, is_production) VALUES
(1, '1.0', '{"temperature": 0.3, "top_p": 0.9, "max_tokens": 1000}', 0.85, 1),
(2, '1.0', '{"dimensions": 1536, "encoding_format": "float"}', 0.92, 1),
(3, '1.0', '{"temperature": 0.7, "top_p": 1.0, "max_tokens": 2000}', 0.88, 1),
(4, '1.0', '{"temperature": 0.5, "top_p": 0.95, "max_tokens": 1500}', 0.87, 1),
(5, '1.0', '{"temperature": 0.7, "top_p": 1.0, "max_tokens": 4000}', 0.95, 0),
(6, '1.0', '{"temperature": 0.6, "top_p": 0.98, "max_tokens": 3000}', 0.93, 0);

-- ==============================================
-- 为现有用户分配角色
-- ==============================================

-- 为现有用户分配普通用户角色
INSERT INTO user_roles (user_id, role_id, assigned_at)
SELECT u.id, r.id, NOW()
FROM users u, roles r
WHERE r.name = 'regular_user'
AND u.id NOT IN (SELECT user_id FROM user_roles);

-- 为特定用户分配管理员角色（根据实际需要调整）
-- 示例：为ID为1的用户分配超级管理员角色
INSERT INTO user_roles (user_id, role_id, assigned_at)
SELECT 1, r.id, NOW()
FROM roles r
WHERE r.name = 'super_admin'
AND 1 IN (SELECT id FROM users);

-- ==============================================
-- 初始化完成
-- ==============================================

-- 记录初始化日志
INSERT INTO permission_audit_logs (action, resource, result, details) VALUES
('DATA_INITIALIZATION', 'system', 1, '{"version": "V4.0", "roles_created": 6, "permissions_created": 25, "ai_models_created": 6, "init_date": "2025-01-06"}');

-- 显示初始化完成信息
SELECT 'JobFirst数据库初始化完成！' as message,
       'V4.0' as version,
       NOW() as init_time,
       '6个角色，25个权限，6个AI模型，完整的数据分类标签' as description;
