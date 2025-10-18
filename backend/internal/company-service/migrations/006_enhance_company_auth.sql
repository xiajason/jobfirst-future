-- Company服务认证机制增强 - 数据库迁移脚本
-- 目标：建立完整的企业权限管理体系，支持法定代表人、经办人等业务角色

-- 1. 扩展Company表结构，添加认证和地理位置字段
ALTER TABLE companies 
ADD COLUMN unified_social_credit_code VARCHAR(50) UNIQUE COMMENT '统一社会信用代码',
ADD COLUMN legal_representative VARCHAR(100) COMMENT '法定代表人姓名',
ADD COLUMN legal_representative_id VARCHAR(50) COMMENT '法定代表人身份证号',
ADD COLUMN legal_rep_user_id INT COMMENT '法定代表人用户ID',
ADD COLUMN authorized_users JSON COMMENT '授权用户列表',
ADD COLUMN bd_latitude DECIMAL(10,8) COMMENT '北斗纬度',
ADD COLUMN bd_longitude DECIMAL(11,8) COMMENT '北斗经度',
ADD COLUMN bd_altitude DECIMAL(8,2) COMMENT '北斗海拔',
ADD COLUMN bd_accuracy DECIMAL(6,2) COMMENT '定位精度(米)',
ADD COLUMN bd_timestamp BIGINT COMMENT '定位时间戳',
ADD COLUMN address VARCHAR(500) COMMENT '详细地址',
ADD COLUMN city VARCHAR(100) COMMENT '城市',
ADD COLUMN district VARCHAR(100) COMMENT '区县',
ADD COLUMN area VARCHAR(100) COMMENT '区域/街道',
ADD COLUMN postal_code VARCHAR(20) COMMENT '邮政编码',
ADD COLUMN city_code VARCHAR(20) COMMENT '城市编码',
ADD COLUMN district_code VARCHAR(20) COMMENT '区县编码',
ADD COLUMN area_code VARCHAR(20) COMMENT '区域编码';

-- 2. 创建企业用户关联表
CREATE TABLE company_users (
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

-- 3. 创建索引优化查询性能
CREATE INDEX idx_company_users_company_id ON company_users(company_id);
CREATE INDEX idx_company_users_user_id ON company_users(user_id);
CREATE INDEX idx_company_users_role ON company_users(role);
CREATE INDEX idx_company_users_status ON company_users(status);
CREATE INDEX idx_company_users_company_user ON company_users(company_id, user_id, status);
CREATE INDEX idx_company_users_user_company ON company_users(user_id, company_id, role);

-- 4. 为企业地理位置查询创建索引
CREATE INDEX idx_companies_bd_location ON companies(bd_latitude, bd_longitude);
CREATE INDEX idx_companies_city_district ON companies(city, district);
CREATE INDEX idx_companies_unified_code ON companies(unified_social_credit_code);

-- 5. 创建企业权限审计日志表
CREATE TABLE company_permission_audit_logs (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    company_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    action VARCHAR(100) NOT NULL COMMENT '操作类型',
    resource_type VARCHAR(50) NOT NULL COMMENT '资源类型',
    resource_id INT COMMENT '资源ID',
    permission_result BOOLEAN NOT NULL COMMENT '权限检查结果',
    ip_address VARCHAR(45) COMMENT 'IP地址',
    user_agent TEXT COMMENT '用户代理',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) COMMENT='企业权限审计日志表';

-- 6. 创建索引
CREATE INDEX idx_company_audit_company_id ON company_permission_audit_logs(company_id);
CREATE INDEX idx_company_audit_user_id ON company_permission_audit_logs(user_id);
CREATE INDEX idx_company_audit_action ON company_permission_audit_logs(action);
CREATE INDEX idx_company_audit_created_at ON company_permission_audit_logs(created_at);

-- 7. 创建企业数据同步状态表
CREATE TABLE company_data_sync_status (
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

-- 8. 创建索引
CREATE INDEX idx_company_sync_company_id ON company_data_sync_status(company_id);
CREATE INDEX idx_company_sync_target ON company_data_sync_status(sync_target);
CREATE INDEX idx_company_sync_status ON company_data_sync_status(sync_status);

-- 9. 插入默认数据（如果需要）
-- 为现有企业设置默认的法定代表人用户ID
UPDATE companies 
SET legal_rep_user_id = created_by 
WHERE legal_rep_user_id IS NULL;

-- 10. 创建企业权限检查视图
CREATE VIEW company_user_permissions AS
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

-- 11. 创建企业地理位置统计视图
CREATE VIEW company_location_stats AS
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

-- 12. 添加注释
ALTER TABLE companies COMMENT = '企业信息表（增强版）- 支持认证机制和地理位置';
ALTER TABLE company_users COMMENT = '企业用户关联表 - 支持多用户管理和权限控制';
ALTER TABLE company_permission_audit_logs COMMENT = '企业权限审计日志表 - 记录所有权限检查操作';
ALTER TABLE company_data_sync_status COMMENT = '企业数据同步状态表 - 跟踪多数据库同步状态';
