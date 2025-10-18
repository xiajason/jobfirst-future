-- JobFirst 数据库升级脚本
-- 版本: V3.0 -> V4.0
-- 日期: 2025年1月6日
-- 描述: 全面升级数据库架构，支持AI服务、个人信息保护、企业职位管理

-- ==============================================
-- 阶段零：基础表创建 (从jobfirst_v3复制)
-- ==============================================

-- 0. 创建companies表
CREATE TABLE IF NOT EXISTS companies (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    industry VARCHAR(100),
    size ENUM('startup','small','medium','large','enterprise') DEFAULT 'medium',
    location VARCHAR(200),
    website VARCHAR(500),
    logo_url VARCHAR(500),
    description TEXT,
    is_verified TINYINT(1) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_name (name),
    INDEX idx_industry (industry),
    INDEX idx_size (size),
    INDEX idx_is_verified (is_verified)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 0.1 创建positions表
CREATE TABLE IF NOT EXISTS positions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_name (name),
    INDEX idx_category (category)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 0.2 创建skills表
CREATE TABLE IF NOT EXISTS skills (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    category VARCHAR(50),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_name (name),
    INDEX idx_category (category)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==============================================
-- 阶段一：权限管理系统
-- ==============================================

-- 1. 角色管理表
CREATE TABLE IF NOT EXISTS roles (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    level TINYINT UNSIGNED DEFAULT 1 COMMENT '角色级别 1-5',
    pid BIGINT UNSIGNED DEFAULT 0 COMMENT '父角色ID，支持角色继承',
    is_system TINYINT(1) DEFAULT 0 COMMENT '是否系统角色',
    is_active TINYINT(1) DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_name (name),
    INDEX idx_level (level),
    INDEX idx_pid (pid),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. 权限管理表
CREATE TABLE IF NOT EXISTS permissions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    resource VARCHAR(100) NOT NULL COMMENT '资源类型',
    action VARCHAR(50) NOT NULL COMMENT '操作类型',
    level TINYINT UNSIGNED DEFAULT 1 COMMENT '权限级别 1-4',
    is_system TINYINT(1) DEFAULT 0,
    is_active TINYINT(1) DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_name (name),
    INDEX idx_resource (resource),
    INDEX idx_action (action),
    INDEX idx_level (level),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. 角色权限关联表
CREATE TABLE IF NOT EXISTS role_permissions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    role_id BIGINT UNSIGNED NOT NULL,
    permission_id BIGINT UNSIGNED NOT NULL,
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    granted_by BIGINT UNSIGNED,
    
    UNIQUE KEY unique_role_permission (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE,
    FOREIGN KEY (granted_by) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_role_id (role_id),
    INDEX idx_permission_id (permission_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. 用户角色关联表
CREATE TABLE IF NOT EXISTS user_roles (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_by BIGINT UNSIGNED,
    expires_at TIMESTAMP NULL COMMENT '角色过期时间',
    is_active TINYINT(1) DEFAULT 1,
    
    UNIQUE KEY unique_user_role (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (assigned_by) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_user_id (user_id),
    INDEX idx_role_id (role_id),
    INDEX idx_expires_at (expires_at),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 5. 权限审计日志表
CREATE TABLE IF NOT EXISTS permission_audit_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED,
    action VARCHAR(100) NOT NULL COMMENT '操作类型',
    resource VARCHAR(100) NOT NULL COMMENT '资源类型',
    resource_id BIGINT UNSIGNED COMMENT '资源ID',
    permission VARCHAR(100) COMMENT '权限名称',
    result TINYINT(1) NOT NULL COMMENT '操作结果 1成功 0失败',
    ip_address VARCHAR(45),
    user_agent TEXT,
    request_id VARCHAR(100),
    session_id VARCHAR(100),
    details JSON COMMENT '详细信息',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_user_id (user_id),
    INDEX idx_action (action),
    INDEX idx_resource (resource),
    INDEX idx_result (result),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 6. 数据访问日志表
CREATE TABLE IF NOT EXISTS data_access_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED,
    table_name VARCHAR(64) NOT NULL,
    field_name VARCHAR(64),
    operation ENUM('SELECT', 'INSERT', 'UPDATE', 'DELETE') NOT NULL,
    record_id BIGINT UNSIGNED,
    sensitivity_level TINYINT UNSIGNED COMMENT '数据敏感级别 1-4',
    ip_address VARCHAR(45),
    user_agent TEXT,
    access_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_user_id (user_id),
    INDEX idx_table_name (table_name),
    INDEX idx_operation (operation),
    INDEX idx_sensitivity_level (sensitivity_level),
    INDEX idx_access_time (access_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==============================================
-- 阶段二：数据分类标签系统
-- ==============================================

-- 7. 数据分类标签表
CREATE TABLE IF NOT EXISTS data_classification_tags (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    table_name VARCHAR(64) NOT NULL,
    field_name VARCHAR(64) NOT NULL,
    sensitivity_level ENUM('low', 'medium', 'high', 'critical') NOT NULL,
    data_type VARCHAR(32),
    protection_method VARCHAR(64),
    retention_period INT COMMENT '保留期(天)',
    encryption_required TINYINT(1) DEFAULT 0,
    access_control_required TINYINT(1) DEFAULT 1,
    audit_required TINYINT(1) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY unique_table_field (table_name, field_name),
    INDEX idx_sensitivity_level (sensitivity_level),
    INDEX idx_table_name (table_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 8. 数据生命周期策略表
CREATE TABLE IF NOT EXISTS data_lifecycle_policies (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    table_name VARCHAR(64) NOT NULL,
    policy_name VARCHAR(100) NOT NULL,
    retention_period INT COMMENT '保留期(天)',
    archive_period INT COMMENT '归档期(天)',
    deletion_period INT COMMENT '删除期(天)',
    archive_location VARCHAR(255) COMMENT '归档位置',
    is_active TINYINT(1) DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY unique_table_policy (table_name, policy_name),
    INDEX idx_table_name (table_name),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==============================================
-- 阶段三：AI服务数据库架构
-- ==============================================

-- 9. AI模型管理表
CREATE TABLE IF NOT EXISTS ai_models (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    version VARCHAR(20) NOT NULL,
    model_type ENUM('text_generation', 'embedding', 'classification', 'regression', 'recommendation') NOT NULL,
    provider ENUM('openai', 'anthropic', 'google', 'azure', 'ollama', 'local', 'custom') NOT NULL,
    model_identifier VARCHAR(200) NOT NULL,
    description TEXT,
    parameters JSON,
    performance_metrics JSON,
    cost_per_token DECIMAL(10,8) DEFAULT 0,
    is_active TINYINT(1) DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_name (name),
    INDEX idx_model_type (model_type),
    INDEX idx_provider (provider),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 10. 模型版本管理表
CREATE TABLE IF NOT EXISTS model_versions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    model_id BIGINT UNSIGNED NOT NULL,
    version VARCHAR(20) NOT NULL,
    model_path VARCHAR(500),
    config JSON,
    training_data_hash VARCHAR(64),
    performance_score DECIMAL(5,4),
    is_production TINYINT(1) DEFAULT 0,
    deployed_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE KEY unique_model_version (model_id, version),
    FOREIGN KEY (model_id) REFERENCES ai_models(id) ON DELETE CASCADE,
    
    INDEX idx_model_id (model_id),
    INDEX idx_is_production (is_production)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 11. 企业AI画像表
CREATE TABLE IF NOT EXISTS company_ai_profiles (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    company_id BIGINT UNSIGNED NOT NULL,
    profile_type ENUM('basic', 'detailed', 'competitive', 'culture', 'comprehensive') NOT NULL,
    profile_data JSON NOT NULL,
    confidence_score DECIMAL(5,4),
    generated_by_model_id BIGINT UNSIGNED,
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    is_valid TINYINT(1) DEFAULT 1,
    
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    FOREIGN KEY (generated_by_model_id) REFERENCES ai_models(id) ON DELETE SET NULL,
    
    INDEX idx_company_id (company_id),
    INDEX idx_profile_type (profile_type),
    INDEX idx_generated_at (generated_at),
    INDEX idx_is_valid (is_valid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 12. 企业嵌入向量表
CREATE TABLE IF NOT EXISTS company_embeddings (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    company_id BIGINT UNSIGNED NOT NULL,
    embedding_type ENUM('description', 'culture', 'benefits', 'overall') NOT NULL,
    embedding_vector JSON NOT NULL,
    model_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    FOREIGN KEY (model_id) REFERENCES ai_models(id) ON DELETE CASCADE,
    
    INDEX idx_company_id (company_id),
    INDEX idx_embedding_type (embedding_type),
    INDEX idx_model_id (model_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 13. 职位AI分析表
CREATE TABLE IF NOT EXISTS job_ai_analysis (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    job_id INT NOT NULL,
    analysis_type ENUM('description_enhancement', 'skill_extraction', 'salary_prediction', 'match_score', 'comprehensive') NOT NULL,
    analysis_result JSON NOT NULL,
    confidence_score DECIMAL(5,4),
    generated_by_model_id BIGINT UNSIGNED,
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    is_valid TINYINT(1) DEFAULT 1,
    
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE,
    FOREIGN KEY (generated_by_model_id) REFERENCES ai_models(id) ON DELETE SET NULL,
    
    INDEX idx_job_id (job_id),
    INDEX idx_analysis_type (analysis_type),
    INDEX idx_generated_at (generated_at),
    INDEX idx_is_valid (is_valid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 14. 职位嵌入向量表
CREATE TABLE IF NOT EXISTS job_embeddings (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    job_id INT NOT NULL,
    embedding_type ENUM('title', 'description', 'requirements', 'overall') NOT NULL,
    embedding_vector JSON NOT NULL,
    model_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE,
    FOREIGN KEY (model_id) REFERENCES ai_models(id) ON DELETE CASCADE,
    
    INDEX idx_job_id (job_id),
    INDEX idx_embedding_type (embedding_type),
    INDEX idx_model_id (model_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 15. 用户AI画像表
CREATE TABLE IF NOT EXISTS user_ai_profiles (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    profile_type ENUM('basic', 'career', 'skills', 'preferences', 'comprehensive') NOT NULL,
    profile_data JSON NOT NULL,
    confidence_score DECIMAL(5,4),
    generated_by_model_id BIGINT UNSIGNED,
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    is_valid TINYINT(1) DEFAULT 1,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (generated_by_model_id) REFERENCES ai_models(id) ON DELETE SET NULL,
    
    INDEX idx_user_id (user_id),
    INDEX idx_profile_type (profile_type),
    INDEX idx_generated_at (generated_at),
    INDEX idx_is_valid (is_valid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 16. 用户嵌入向量表
CREATE TABLE IF NOT EXISTS user_embeddings (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    embedding_type ENUM('resume', 'skills', 'experience', 'preferences', 'overall') NOT NULL,
    embedding_vector JSON NOT NULL,
    model_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (model_id) REFERENCES ai_models(id) ON DELETE CASCADE,
    
    INDEX idx_user_id (user_id),
    INDEX idx_embedding_type (embedding_type),
    INDEX idx_model_id (model_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 17. 职位推荐表
CREATE TABLE IF NOT EXISTS job_recommendations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    job_id INT NOT NULL,
    recommendation_score DECIMAL(5,4) NOT NULL,
    recommendation_reasons JSON,
    match_factors JSON,
    generated_by_model_id BIGINT UNSIGNED,
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    is_active TINYINT(1) DEFAULT 1,
    user_interaction ENUM('viewed', 'applied', 'saved', 'dismissed') NULL,
    interaction_at TIMESTAMP NULL,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE,
    FOREIGN KEY (generated_by_model_id) REFERENCES ai_models(id) ON DELETE SET NULL,
    
    INDEX idx_user_id (user_id),
    INDEX idx_job_id (job_id),
    INDEX idx_recommendation_score (recommendation_score),
    INDEX idx_generated_at (generated_at),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 18. 企业推荐表
CREATE TABLE IF NOT EXISTS company_recommendations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    company_id BIGINT UNSIGNED NOT NULL,
    recommendation_score DECIMAL(5,4) NOT NULL,
    recommendation_reasons JSON,
    match_factors JSON,
    generated_by_model_id BIGINT UNSIGNED,
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    is_active TINYINT(1) DEFAULT 1,
    user_interaction ENUM('viewed', 'followed', 'applied', 'dismissed') NULL,
    interaction_at TIMESTAMP NULL,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    FOREIGN KEY (generated_by_model_id) REFERENCES ai_models(id) ON DELETE SET NULL,
    
    INDEX idx_user_id (user_id),
    INDEX idx_company_id (company_id),
    INDEX idx_recommendation_score (recommendation_score),
    INDEX idx_generated_at (generated_at),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 19. AI对话会话表
CREATE TABLE IF NOT EXISTS ai_conversations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    conversation_type ENUM('career_advice', 'resume_review', 'interview_prep', 'skill_analysis', 'general') NOT NULL,
    session_id VARCHAR(100) NOT NULL,
    context_data JSON,
    model_id BIGINT UNSIGNED NOT NULL,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_activity_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    is_active TINYINT(1) DEFAULT 1,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (model_id) REFERENCES ai_models(id) ON DELETE CASCADE,
    
    INDEX idx_user_id (user_id),
    INDEX idx_conversation_type (conversation_type),
    INDEX idx_session_id (session_id),
    INDEX idx_model_id (model_id),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 20. AI对话消息表
CREATE TABLE IF NOT EXISTS ai_messages (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    conversation_id BIGINT UNSIGNED NOT NULL,
    message_type ENUM('user', 'assistant', 'system') NOT NULL,
    content TEXT NOT NULL,
    metadata JSON,
    tokens_used INT,
    processing_time_ms INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (conversation_id) REFERENCES ai_conversations(id) ON DELETE CASCADE,
    
    INDEX idx_conversation_id (conversation_id),
    INDEX idx_message_type (message_type),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 21. AI服务日志表
CREATE TABLE IF NOT EXISTS ai_service_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    service_name VARCHAR(100) NOT NULL,
    operation_type ENUM('embedding', 'generation', 'classification', 'recommendation', 'analysis', 'chat') NOT NULL,
    model_id BIGINT UNSIGNED,
    user_id BIGINT UNSIGNED,
    input_tokens INT,
    output_tokens INT,
    processing_time_ms INT,
    cost_usd DECIMAL(10,6),
    success TINYINT(1) NOT NULL,
    error_message TEXT,
    request_id VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (model_id) REFERENCES ai_models(id) ON DELETE SET NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_service_name (service_name),
    INDEX idx_operation_type (operation_type),
    INDEX idx_model_id (model_id),
    INDEX idx_user_id (user_id),
    INDEX idx_success (success),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 22. AI性能指标表
CREATE TABLE IF NOT EXISTS ai_performance_metrics (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    model_id BIGINT UNSIGNED NOT NULL,
    metric_type ENUM('accuracy', 'latency', 'throughput', 'cost', 'user_satisfaction') NOT NULL,
    metric_value DECIMAL(10,6) NOT NULL,
    measurement_period_start TIMESTAMP NOT NULL,
    measurement_period_end TIMESTAMP NOT NULL,
    sample_size INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (model_id) REFERENCES ai_models(id) ON DELETE CASCADE,
    
    INDEX idx_model_id (model_id),
    INDEX idx_metric_type (metric_type),
    INDEX idx_measurement_period_start (measurement_period_start)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 23. AI缓存表
CREATE TABLE IF NOT EXISTS ai_cache (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    cache_key VARCHAR(255) NOT NULL UNIQUE,
    cache_type ENUM('embedding', 'analysis', 'recommendation', 'profile', 'conversation') NOT NULL,
    cache_data JSON NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    hit_count INT DEFAULT 0,
    last_accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_cache_key (cache_key),
    INDEX idx_cache_type (cache_type),
    INDEX idx_expires_at (expires_at),
    INDEX idx_last_accessed_at (last_accessed_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==============================================
-- 阶段四：个人信息保护升级
-- ==============================================

-- 为现有表添加加密字段
ALTER TABLE users ADD COLUMN email_encrypted BLOB COMMENT '加密邮箱';
ALTER TABLE users ADD COLUMN phone_encrypted BLOB COMMENT '加密电话';
ALTER TABLE users ADD COLUMN first_name_encrypted BLOB COMMENT '加密名字';
ALTER TABLE users ADD COLUMN last_name_encrypted BLOB COMMENT '加密姓氏';

-- 为user_profiles表添加加密字段
ALTER TABLE user_profiles ADD COLUMN date_of_birth_encrypted BLOB COMMENT '加密出生日期';
ALTER TABLE user_profiles ADD COLUMN location_encrypted BLOB COMMENT '加密位置信息';

-- 为files表添加加密字段
ALTER TABLE files ADD COLUMN file_path_encrypted BLOB COMMENT '加密文件路径';
ALTER TABLE files ADD COLUMN original_filename_encrypted BLOB COMMENT '加密原始文件名';

-- ==============================================
-- 阶段五：企业职位管理升级
-- ==============================================

-- 升级jobs表结构
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS company_id BIGINT UNSIGNED COMMENT '关联公司ID';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS position_id BIGINT UNSIGNED COMMENT '关联职位ID';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS job_type ENUM('full_time', 'part_time', 'contract', 'internship', 'freelance') DEFAULT 'full_time';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS experience_level ENUM('entry', 'junior', 'mid', 'senior', 'lead', 'executive') DEFAULT 'mid';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS remote_option ENUM('no', 'hybrid', 'full_remote') DEFAULT 'no';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS benefits JSON COMMENT '福利待遇';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS application_deadline DATE COMMENT '申请截止日期';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS is_featured TINYINT(1) DEFAULT 0 COMMENT '是否推荐';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS view_count INT DEFAULT 0 COMMENT '浏览次数';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS application_count INT DEFAULT 0 COMMENT '申请次数';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS is_active TINYINT(1) DEFAULT 1 COMMENT '是否激活';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS created_by BIGINT UNSIGNED COMMENT '创建者';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS updated_by BIGINT UNSIGNED COMMENT '更新者';

-- 24. 职位技能关联表
CREATE TABLE IF NOT EXISTS job_skills (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    job_id INT NOT NULL,
    skill_id BIGINT UNSIGNED NOT NULL,
    required_level ENUM('basic', 'intermediate', 'advanced', 'expert') DEFAULT 'intermediate',
    is_required TINYINT(1) DEFAULT 1 COMMENT '是否必需技能',
    weight DECIMAL(3,2) DEFAULT 1.00 COMMENT '技能权重',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE KEY unique_job_skill (job_id, skill_id),
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE,
    FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE,
    
    INDEX idx_job_id (job_id),
    INDEX idx_skill_id (skill_id),
    INDEX idx_required_level (required_level),
    INDEX idx_is_required (is_required)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 25. 用户技能表
CREATE TABLE IF NOT EXISTS user_skills (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    skill_id BIGINT UNSIGNED NOT NULL,
    proficiency_level ENUM('beginner', 'intermediate', 'advanced', 'expert') DEFAULT 'intermediate',
    years_of_experience DECIMAL(3,1) DEFAULT 0.0 COMMENT '经验年数',
    last_used_at DATE COMMENT '最后使用时间',
    is_verified TINYINT(1) DEFAULT 0 COMMENT '是否验证',
    verified_by BIGINT UNSIGNED COMMENT '验证者',
    verified_at TIMESTAMP NULL COMMENT '验证时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY unique_user_skill (user_id, skill_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE,
    FOREIGN KEY (verified_by) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_user_id (user_id),
    INDEX idx_skill_id (skill_id),
    INDEX idx_proficiency_level (proficiency_level),
    INDEX idx_is_verified (is_verified)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 添加外键约束（如果表存在）
-- 注意：这些外键约束需要确保相关表存在
-- ALTER TABLE jobs ADD FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE SET NULL;
-- ALTER TABLE jobs ADD FOREIGN KEY (position_id) REFERENCES positions(id) ON DELETE SET NULL;
-- ALTER TABLE jobs ADD FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;
-- ALTER TABLE jobs ADD FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL;

-- 添加索引
ALTER TABLE jobs ADD INDEX IF NOT EXISTS idx_company_id (company_id);
ALTER TABLE jobs ADD INDEX IF NOT EXISTS idx_position_id (position_id);
ALTER TABLE jobs ADD INDEX IF NOT EXISTS idx_job_type (job_type);
ALTER TABLE jobs ADD INDEX IF NOT EXISTS idx_experience_level (experience_level);
ALTER TABLE jobs ADD INDEX IF NOT EXISTS idx_remote_option (remote_option);
ALTER TABLE jobs ADD INDEX IF NOT EXISTS idx_is_featured (is_featured);
ALTER TABLE jobs ADD INDEX IF NOT EXISTS idx_is_active (is_active);
ALTER TABLE jobs ADD INDEX IF NOT EXISTS idx_created_by (created_by);

-- ==============================================
-- 数据脱敏视图
-- ==============================================

-- 创建数据脱敏视图
CREATE OR REPLACE VIEW users_masked AS
SELECT 
    id,
    uuid,
    CASE 
        WHEN email IS NOT NULL THEN CONCAT(LEFT(email, 2), '***', RIGHT(email, 4))
        ELSE NULL 
    END as email_masked,
    username,
    CASE 
        WHEN first_name IS NOT NULL THEN CONCAT(LEFT(first_name, 1), '***')
        ELSE NULL 
    END as first_name_masked,
    CASE 
        WHEN last_name IS NOT NULL THEN CONCAT(LEFT(last_name, 1), '***')
        ELSE NULL 
    END as last_name_masked,
    CASE 
        WHEN phone IS NOT NULL THEN CONCAT(LEFT(phone, 3), '****', RIGHT(phone, 4))
        ELSE NULL 
    END as phone_masked,
    avatar_url,
    status,
    email_verified,
    phone_verified,
    last_login_at,
    created_at,
    updated_at
FROM users;

-- 创建高敏感数据访问视图
CREATE OR REPLACE VIEW users_sensitive AS
SELECT 
    id,
    uuid,
    email,
    first_name,
    last_name,
    phone,
    password_hash,
    avatar_url,
    status,
    email_verified,
    phone_verified,
    last_login_at,
    created_at,
    updated_at
FROM users
WHERE deleted_at IS NULL;

-- ==============================================
-- 升级完成
-- ==============================================

-- 记录升级日志
INSERT INTO permission_audit_logs (action, resource, result, details) VALUES
('DATABASE_UPGRADE', 'system', 1, '{"version": "V3.0->V4.0", "tables_created": 25, "upgrade_date": "2025-01-06"}');

-- 显示升级完成信息
SELECT 'JobFirst数据库升级完成！' as message,
       'V3.0 -> V4.0' as version,
       NOW() as upgrade_time,
       '新增25个表，支持AI服务、权限管理、个人信息保护' as description;
