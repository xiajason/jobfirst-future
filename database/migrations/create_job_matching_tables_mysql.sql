-- 职位匹配系统数据库迁移脚本 - MySQL部分
-- 创建时间: 2025-09-13
-- 版本: 1.0.0

-- ==============================================
-- MySQL数据库 - 职位相关表
-- ==============================================

-- 职位表
CREATE TABLE IF NOT EXISTS `jobs` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `title` varchar(200) NOT NULL COMMENT '职位标题',
    `description` text COMMENT '职位描述',
    `requirements` text COMMENT '职位要求',
    `company_id` bigint unsigned NOT NULL COMMENT '公司ID',
    `industry` varchar(100) DEFAULT NULL COMMENT '行业',
    `location` varchar(200) DEFAULT NULL COMMENT '工作地点',
    `salary_min` int(11) DEFAULT NULL COMMENT '最低薪资',
    `salary_max` int(11) DEFAULT NULL COMMENT '最高薪资',
    `experience` varchar(50) DEFAULT NULL COMMENT '经验要求',
    `education` varchar(100) DEFAULT NULL COMMENT '学历要求',
    `job_type` varchar(50) DEFAULT NULL COMMENT '工作类型',
    `status` varchar(20) DEFAULT 'active' COMMENT '职位状态',
    `view_count` int(11) DEFAULT 0 COMMENT '浏览次数',
    `apply_count` int(11) DEFAULT 0 COMMENT '申请次数',
    `created_by` bigint unsigned NOT NULL COMMENT '创建者ID',
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_company_id` (`company_id`),
    KEY `idx_industry` (`industry`),
    KEY `idx_location` (`location`),
    KEY `idx_status` (`status`),
    KEY `idx_created_at` (`created_at`),
    KEY `idx_company_status` (`company_id`, `status`),
    KEY `idx_industry_location` (`industry`, `location`),
    CONSTRAINT `fk_jobs_company` FOREIGN KEY (`company_id`) REFERENCES `companies` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_jobs_created_by` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='职位表';

-- 职位申请表
CREATE TABLE IF NOT EXISTS `job_applications` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `job_id` int(11) unsigned NOT NULL COMMENT '职位ID',
    `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
    `resume_id` int(11) NOT NULL COMMENT '简历ID',
    `status` varchar(20) DEFAULT 'pending' COMMENT '申请状态',
    `cover_letter` text COMMENT '求职信',
    `applied_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '申请时间',
    `reviewed_at` timestamp NULL DEFAULT NULL COMMENT '审核时间',
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_job_user` (`job_id`, `user_id`),
    KEY `idx_job_id` (`job_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_resume_id` (`resume_id`),
    KEY `idx_status` (`status`),
    KEY `idx_applied_at` (`applied_at`),
    CONSTRAINT `fk_applications_job` FOREIGN KEY (`job_id`) REFERENCES `jobs` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_applications_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_applications_resume` FOREIGN KEY (`resume_id`) REFERENCES `resume_metadata` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='职位申请表';

-- 职位匹配日志表
CREATE TABLE IF NOT EXISTS `job_matching_logs` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
    `resume_id` int(11) NOT NULL COMMENT '简历ID',
    `matches_count` int(11) DEFAULT 0 COMMENT '匹配数量',
    `filters_applied` json DEFAULT NULL COMMENT '应用的筛选条件',
    `processing_time` int(11) DEFAULT 0 COMMENT '处理时间(毫秒)',
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_resume_id` (`resume_id`),
    KEY `idx_created_at` (`created_at`),
    KEY `idx_user_created` (`user_id`, `created_at`),
    CONSTRAINT `fk_matching_logs_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_matching_logs_resume` FOREIGN KEY (`resume_id`) REFERENCES `resume_metadata` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='职位匹配日志表';

-- ==============================================
-- 初始数据
-- ==============================================

-- 插入示例职位数据
INSERT IGNORE INTO `jobs` (`id`, `title`, `description`, `requirements`, `company_id`, `industry`, `location`, `salary_min`, `salary_max`, `experience`, `education`, `job_type`, `status`, `created_by`) VALUES
(1, '高级前端开发工程师', '负责公司核心产品的前端开发工作，参与产品架构设计和技术选型。', '熟练掌握React、Vue等前端框架，具备3年以上前端开发经验，熟悉TypeScript、Webpack等工具。', 1, 'technology', '北京', 15000, 25000, 'senior', 'bachelor', 'full-time', 'active', 1),
(2, 'Python后端开发工程师', '负责后端API开发，数据库设计，微服务架构设计。', '熟练掌握Python、Django/Flask框架，具备MySQL/PostgreSQL数据库经验，了解Docker、Kubernetes。', 1, 'technology', '上海', 12000, 20000, 'mid', 'bachelor', 'full-time', 'active', 1),
(3, '产品经理', '负责产品规划和设计，协调开发团队，推进产品迭代。', '具备产品设计经验，熟悉敏捷开发流程，具备良好的沟通协调能力。', 1, 'technology', '深圳', 18000, 30000, 'senior', 'bachelor', 'full-time', 'active', 1),
(4, 'UI/UX设计师', '负责产品界面设计和用户体验优化。', '熟练掌握Figma、Sketch等设计工具，具备良好的审美能力和用户体验意识。', 1, 'design', '杭州', 10000, 18000, 'mid', 'bachelor', 'full-time', 'active', 1),
(5, '数据分析师', '负责业务数据分析，建立数据模型，提供决策支持。', '熟练掌握SQL、Python，具备统计学基础，熟悉数据可视化工具。', 1, 'technology', '广州', 13000, 22000, 'mid', 'master', 'full-time', 'active', 1);

-- 更新公司职位数量
UPDATE `companies` SET `job_count` = (
    SELECT COUNT(*) FROM `jobs` WHERE `company_id` = `companies`.`id` AND `status` = 'active'
) WHERE `id` = 1;

-- ==============================================
-- 视图和存储过程
-- ==============================================

-- 职位详情视图
CREATE OR REPLACE VIEW job_details_view AS
SELECT 
    j.id,
    j.title,
    j.description,
    j.requirements,
    j.company_id,
    c.name as company_name,
    c.short_name as company_short_name,
    c.logo_url as company_logo,
    j.industry,
    j.location,
    j.salary_min,
    j.salary_max,
    j.experience,
    j.education,
    j.job_type,
    j.status,
    j.view_count,
    j.apply_count,
    j.created_by,
    j.created_at,
    j.updated_at
FROM jobs j
LEFT JOIN companies c ON j.company_id = c.id
WHERE j.status = 'active';

-- 用户申请历史视图
CREATE OR REPLACE VIEW user_applications_view AS
SELECT 
    ja.id,
    ja.job_id,
    j.title as job_title,
    ja.user_id,
    ja.resume_id,
    rm.title as resume_title,
    ja.status,
    ja.cover_letter,
    ja.applied_at,
    ja.reviewed_at,
    c.name as company_name,
    j.industry,
    j.location
FROM job_applications ja
LEFT JOIN jobs j ON ja.job_id = j.id
LEFT JOIN resume_metadata rm ON ja.resume_id = rm.id
LEFT JOIN companies c ON j.company_id = c.id;

-- ==============================================
-- 触发器
-- ==============================================

-- 更新职位申请数量触发器
DROP TRIGGER IF EXISTS update_job_apply_count;
DELIMITER $$
CREATE TRIGGER update_job_apply_count 
AFTER INSERT ON job_applications
FOR EACH ROW
BEGIN
    UPDATE jobs 
    SET apply_count = apply_count + 1 
    WHERE id = NEW.job_id;
END$$
DELIMITER ;

-- 更新职位浏览次数触发器
DROP TRIGGER IF EXISTS update_job_view_count;
DELIMITER $$
CREATE TRIGGER update_job_view_count 
AFTER UPDATE ON jobs
FOR EACH ROW
BEGIN
    IF NEW.view_count != OLD.view_count THEN
        UPDATE companies 
        SET view_count = view_count + (NEW.view_count - OLD.view_count)
        WHERE id = NEW.company_id;
    END IF;
END$$
DELIMITER ;

-- ==============================================
-- 权限设置
-- ==============================================

-- 创建职位匹配相关权限
INSERT IGNORE INTO `permissions` (`name`, `display_name`, `description`, `resource`, `action`) VALUES
('job.create', '创建职位', '创建新职位', 'job', 'create'),
('job.read', '查看职位', '查看职位信息', 'job', 'read'),
('job.update', '更新职位', '更新职位信息', 'job', 'update'),
('job.delete', '删除职位', '删除职位', 'job', 'delete'),
('job.apply', '申请职位', '申请职位', 'job', 'apply'),
('job.matching', '职位匹配', '职位匹配功能', 'job', 'matching'),
('job.matching.admin', '职位匹配管理', '职位匹配管理功能', 'job', 'matching_admin');

-- 为管理员角色分配权限
INSERT IGNORE INTO `role_permissions` (`role_id`, `permission_id`)
SELECT r.id, p.id 
FROM `roles` r, `permissions` p 
WHERE r.name = 'admin' AND p.name IN (
    'job.create', 'job.read', 'job.update', 'job.delete', 
    'job.apply', 'job.matching', 'job.matching.admin'
);

-- 为普通用户角色分配权限
INSERT IGNORE INTO `role_permissions` (`role_id`, `permission_id`)
SELECT r.id, p.id 
FROM `roles` r, `permissions` p 
WHERE r.name = 'user' AND p.name IN (
    'job.read', 'job.apply', 'job.matching'
);

-- ==============================================
-- 完成标记
-- ==============================================

-- 迁移完成
-- 创建时间: 2025-09-13
-- 版本: 1.0.0
-- 描述: 创建职位匹配系统MySQL相关数据表
