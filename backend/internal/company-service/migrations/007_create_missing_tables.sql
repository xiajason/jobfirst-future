-- Company服务认证机制增强 - 创建缺失的表
-- 目标：创建企业用户关联表、权限审计日志表、数据同步状态表

-- 1. 创建企业用户关联表（如果不存在）
CREATE TABLE IF NOT EXISTS company_users (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    company_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    role VARCHAR(50) NOT NULL COMMENT '角色：legal_rep, authorized_user, admin',
    status VARCHAR(20) DEFAULT 'active' COMMENT '状态：active, inactive, pending',
    permissions JSON COMMENT '权限列表',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_company_user (company_id, user_id)
) COMMENT='企业用户关联表';

-- 2. 创建企业权限审计日志表（如果不存在）
CREATE TABLE IF NOT EXISTS company_permission_audit_logs (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    company_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    action VARCHAR(100) NOT NULL COMMENT '操作类型',
    resource_type VARCHAR(50) NOT NULL COMMENT '资源类型',
    resource_id BIGINT UNSIGNED COMMENT '资源ID',
    permission_result BOOLEAN NOT NULL COMMENT '权限检查结果',
    ip_address VARCHAR(45) COMMENT 'IP地址',
    user_agent TEXT COMMENT '用户代理',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) COMMENT='企业权限审计日志表';

-- 3. 创建企业数据同步状态表（如果不存在）
CREATE TABLE IF NOT EXISTS company_data_sync_status (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    company_id BIGINT UNSIGNED NOT NULL,
    sync_target VARCHAR(50) NOT NULL COMMENT '同步目标：postgresql, neo4j, redis',
    sync_status VARCHAR(20) DEFAULT 'pending' COMMENT '同步状态：pending, syncing, success, failed',
    last_sync_time TIMESTAMP NULL COMMENT '最后同步时间',
    sync_error TEXT COMMENT '同步错误信息',
    retry_count INT DEFAULT 0 COMMENT '重试次数',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    UNIQUE KEY unique_company_sync_target (company_id, sync_target)
) COMMENT='企业数据同步状态表';

-- 4. 创建索引优化查询性能
-- 企业用户关联表索引
CREATE INDEX idx_company_users_company_id ON company_users(company_id);
CREATE INDEX idx_company_users_user_id ON company_users(user_id);
CREATE INDEX idx_company_users_role ON company_users(role);
CREATE INDEX idx_company_users_status ON company_users(status);
CREATE INDEX idx_company_users_company_user ON company_users(company_id, user_id, status);
CREATE INDEX idx_company_users_user_company ON company_users(user_id, company_id, role);

-- 企业权限审计日志表索引
CREATE INDEX idx_company_audit_company_id ON company_permission_audit_logs(company_id);
CREATE INDEX idx_company_audit_user_id ON company_permission_audit_logs(user_id);
CREATE INDEX idx_company_audit_action ON company_permission_audit_logs(action);
CREATE INDEX idx_company_audit_created_at ON company_permission_audit_logs(created_at);

-- 企业数据同步状态表索引
CREATE INDEX idx_company_sync_company_id ON company_data_sync_status(company_id);
CREATE INDEX idx_company_sync_target ON company_data_sync_status(sync_target);
CREATE INDEX idx_company_sync_status ON company_data_sync_status(sync_status);

-- 5. 为企业地理位置查询创建索引
CREATE INDEX idx_companies_bd_location ON companies(bd_latitude, bd_longitude);
CREATE INDEX idx_companies_city_district ON companies(city, district);
CREATE INDEX idx_companies_unified_code ON companies(unified_social_credit_code);

-- 6. 创建企业权限检查视图
CREATE OR REPLACE VIEW company_user_permissions AS
SELECT 
    c.id as company_id,
    c.name as company_name,
    c.unified_social_credit_code,
    c.legal_representative,
    c.legal_rep_user_id,
    cu.user_id,
    cu.role,
    cu.status as user_status,
    cu.permissions,
    u.username,
    u.email,
    u.role as user_role,
    CASE 
        WHEN u.role IN ('admin', 'super_admin') THEN 'system_admin'
        WHEN c.created_by = cu.user_id THEN 'company_owner'
        WHEN c.legal_rep_user_id = cu.user_id THEN 'legal_representative'
        WHEN cu.role = 'authorized_user' THEN 'authorized_user'
        ELSE 'no_access'
    END as effective_permission_level
FROM companies c
LEFT JOIN company_users cu ON c.id = cu.company_id
LEFT JOIN users u ON cu.user_id = u.id
WHERE cu.status = 'active' OR u.role IN ('admin', 'super_admin');

-- 7. 创建企业地理位置统计视图
CREATE OR REPLACE VIEW company_location_stats AS
SELECT 
    city,
    district,
    COUNT(*) as company_count,
    AVG(bd_latitude) as avg_latitude,
    AVG(bd_longitude) as avg_longitude,
    MIN(bd_latitude) as min_latitude,
    MAX(bd_latitude) as max_latitude,
    MIN(bd_longitude) as min_longitude,
    MAX(bd_longitude) as max_longitude
FROM companies 
WHERE bd_latitude IS NOT NULL 
  AND bd_longitude IS NOT NULL 
  AND status = 'active'
GROUP BY city, district
ORDER BY company_count DESC;
