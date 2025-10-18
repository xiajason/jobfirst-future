-- JobFirst V3.0 Database Initialization
-- MySQL 8.0+ Compatible (Local Installation)
-- 基于 DATABASE_MAPPING_V3.md 的完整数据库结构

-- 创建数据库
CREATE DATABASE IF NOT EXISTS jobfirst_v3 CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE jobfirst_v3;

-- ==================== 用户相关表 ====================

-- 用户表（保持与V1.0兼容）
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    uuid VARCHAR(36) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    avatar_url VARCHAR(500),
    status ENUM('active', 'inactive', 'suspended') DEFAULT 'active',
    email_verified BOOLEAN DEFAULT FALSE,
    phone_verified BOOLEAN DEFAULT FALSE,
    last_login_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_email (email),
    INDEX idx_username (username),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 用户资料表
CREATE TABLE IF NOT EXISTS user_profiles (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    bio TEXT,
    location VARCHAR(255),
    website VARCHAR(500),
    linkedin_url VARCHAR(500),
    github_url VARCHAR(500),
    twitter_url VARCHAR(500),
    date_of_birth DATE,
    gender ENUM('male', 'female', 'other', 'prefer_not_to_say'),
    nationality VARCHAR(100),
    languages JSON,
    skills JSON,
    interests JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==================== 简历相关表 ====================

-- 简历主表（V3.0）
CREATE TABLE IF NOT EXISTS resumes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    uuid VARCHAR(36) NOT NULL UNIQUE,
    user_id BIGINT UNSIGNED NOT NULL,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE,                    -- URL友好的标识符
    summary TEXT,                                -- 简历摘要
    template_id BIGINT UNSIGNED,
    content TEXT NOT NULL,                       -- Markdown格式内容
    content_vector JSON,                         -- AI解析后的向量数据
    status ENUM('draft','published','archived') DEFAULT 'draft',
    visibility ENUM('public','friends','private') DEFAULT 'private',
    can_comment BOOLEAN DEFAULT TRUE,            -- 是否允许评论
    view_count INT UNSIGNED DEFAULT 0,
    download_count INT UNSIGNED DEFAULT 0,
    share_count INT UNSIGNED DEFAULT 0,
    comment_count INT UNSIGNED DEFAULT 0,        -- 评论数量
    like_count INT UNSIGNED DEFAULT 0,           -- 点赞数量
    is_default BOOLEAN DEFAULT FALSE,
    published_at TIMESTAMP NULL,                 -- 发布时间
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_template_id (template_id),
    INDEX idx_status (status),
    INDEX idx_visibility (visibility),
    INDEX idx_slug (slug),
    INDEX idx_published_at (published_at),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 简历模板表
CREATE TABLE IF NOT EXISTS resume_templates (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    uuid VARCHAR(36) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category ENUM('professional', 'creative', 'minimal', 'modern', 'classic') DEFAULT 'professional',
    thumbnail_url VARCHAR(500),
    preview_url VARCHAR(500),
    template_data JSON NOT NULL,
    css_styles TEXT,
    js_scripts TEXT,
    is_free BOOLEAN DEFAULT TRUE,
    price DECIMAL(10,2) DEFAULT 0.00,
    popularity_score INT DEFAULT 0,
    download_count INT UNSIGNED DEFAULT 0,
    rating DECIMAL(3,2) DEFAULT 0.00,
    review_count INT UNSIGNED DEFAULT 0,
    status ENUM('active', 'inactive', 'deprecated') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_category (category),
    INDEX idx_is_free (is_free),
    INDEX idx_popularity (popularity_score),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==================== 标准化数据表 ====================

-- 标准化技能表
CREATE TABLE IF NOT EXISTS skills (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    category VARCHAR(50) NOT NULL,               -- 技能分类
    description TEXT,
    icon VARCHAR(100),                           -- 技能图标
    is_popular BOOLEAN DEFAULT FALSE,            -- 是否热门技能
    search_count INT UNSIGNED DEFAULT 0,         -- 搜索次数
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_category (category),
    INDEX idx_is_popular (is_popular),
    INDEX idx_search_count (search_count)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 标准化公司表
CREATE TABLE IF NOT EXISTS companies (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    industry VARCHAR(100),                       -- 行业
    size ENUM('startup','small','medium','large','enterprise') DEFAULT 'medium',
    location VARCHAR(200),                       -- 公司地址
    website VARCHAR(500),                        -- 公司网站
    logo_url VARCHAR(500),                       -- 公司Logo
    description TEXT,                            -- 公司描述
    is_verified BOOLEAN DEFAULT FALSE,           -- 是否认证公司
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name (name),
    INDEX idx_industry (industry),
    INDEX idx_size (size),
    INDEX idx_is_verified (is_verified)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 标准化职位表
CREATE TABLE IF NOT EXISTS positions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL,               -- 职位分类
    level ENUM('entry','junior','mid','senior','lead','executive') DEFAULT 'mid',
    description TEXT,                            -- 职位描述
    requirements TEXT,                           -- 职位要求
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_title (title),
    INDEX idx_category (category),
    INDEX idx_level (level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==================== 简历关联表 ====================

-- 简历技能关联表
CREATE TABLE IF NOT EXISTS resume_skills (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    resume_id BIGINT UNSIGNED NOT NULL,
    skill_id BIGINT UNSIGNED NOT NULL,
    proficiency_level ENUM('beginner','intermediate','advanced','expert') NOT NULL,
    years_of_experience DECIMAL(3,1) DEFAULT 0,  -- 使用小数支持0.5年
    is_highlighted BOOLEAN DEFAULT FALSE,         -- 是否突出显示
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (resume_id) REFERENCES resumes(id) ON DELETE CASCADE,
    FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE,
    UNIQUE KEY uk_resume_skill (resume_id, skill_id),
    INDEX idx_resume_id (resume_id),
    INDEX idx_skill_id (skill_id),
    INDEX idx_proficiency_level (proficiency_level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 工作经历表
CREATE TABLE IF NOT EXISTS work_experiences (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    resume_id BIGINT UNSIGNED NOT NULL,
    company_id BIGINT UNSIGNED,
    position_id BIGINT UNSIGNED,
    title VARCHAR(100) NOT NULL,                 -- 实际职位名称
    start_date DATE NOT NULL,
    end_date DATE NULL,
    is_current BOOLEAN DEFAULT FALSE,
    location VARCHAR(200),                       -- 工作地点
    description TEXT,                            -- 工作描述
    achievements TEXT,                           -- 主要成就
    technologies TEXT,                           -- 使用技术
    salary_range VARCHAR(50),                    -- 薪资范围（可选）
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (resume_id) REFERENCES resumes(id) ON DELETE CASCADE,
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE SET NULL,
    FOREIGN KEY (position_id) REFERENCES positions(id) ON DELETE SET NULL,
    INDEX idx_resume_id (resume_id),
    INDEX idx_company_id (company_id),
    INDEX idx_position_id (position_id),
    INDEX idx_start_date (start_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 项目经验表
CREATE TABLE IF NOT EXISTS projects (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    resume_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    start_date DATE,
    end_date DATE,
    status ENUM('planning','in_progress','completed','cancelled') DEFAULT 'completed',
    technology_stack TEXT,                       -- 技术栈
    project_url VARCHAR(500),                    -- 项目链接
    github_url VARCHAR(500),                     -- GitHub链接
    demo_url VARCHAR(500),                       -- 演示链接
    company_id BIGINT UNSIGNED,                  -- 关联公司
    is_highlighted BOOLEAN DEFAULT FALSE,        -- 是否突出显示
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (resume_id) REFERENCES resumes(id) ON DELETE CASCADE,
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE SET NULL,
    INDEX idx_resume_id (resume_id),
    INDEX idx_company_id (company_id),
    INDEX idx_status (status),
    INDEX idx_is_highlighted (is_highlighted)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 教育背景表
CREATE TABLE IF NOT EXISTS educations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    resume_id BIGINT UNSIGNED NOT NULL,
    school VARCHAR(200) NOT NULL,
    degree VARCHAR(50) NOT NULL,                 -- 学位
    major VARCHAR(100) NOT NULL,                 -- 专业
    start_date DATE,
    end_date DATE,
    gpa DECIMAL(3,2),                           -- GPA成绩
    location VARCHAR(200),                       -- 学校地点
    description TEXT,                            -- 教育描述
    is_highlighted BOOLEAN DEFAULT FALSE,        -- 是否突出显示
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (resume_id) REFERENCES resumes(id) ON DELETE CASCADE,
    INDEX idx_resume_id (resume_id),
    INDEX idx_school (school),
    INDEX idx_degree (degree)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 证书认证表
CREATE TABLE IF NOT EXISTS certifications (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    resume_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(200) NOT NULL,
    issuer VARCHAR(200) NOT NULL,                -- 颁发机构
    issue_date DATE NOT NULL,
    expiry_date DATE,                            -- 过期日期
    credential_id VARCHAR(100),                  -- 证书编号
    credential_url VARCHAR(500),                 -- 证书链接
    description TEXT,                            -- 证书描述
    is_highlighted BOOLEAN DEFAULT FALSE,        -- 是否突出显示
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (resume_id) REFERENCES resumes(id) ON DELETE CASCADE,
    INDEX idx_resume_id (resume_id),
    INDEX idx_issuer (issuer),
    INDEX idx_issue_date (issue_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==================== 社交功能表 ====================

-- 简历评论表
CREATE TABLE IF NOT EXISTS resume_comments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    resume_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    parent_id BIGINT UNSIGNED NULL,              -- 支持回复
    content TEXT NOT NULL,
    is_approved BOOLEAN DEFAULT FALSE,           -- 审核状态
    like_count INT UNSIGNED DEFAULT 0,           -- 点赞数
    reply_count INT UNSIGNED DEFAULT 0,          -- 回复数
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (resume_id) REFERENCES resumes(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES resume_comments(id) ON DELETE CASCADE,
    INDEX idx_resume_id (resume_id),
    INDEX idx_user_id (user_id),
    INDEX idx_parent_id (parent_id),
    INDEX idx_is_approved (is_approved),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 简历点赞表
CREATE TABLE IF NOT EXISTS resume_likes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    resume_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (resume_id) REFERENCES resumes(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY uk_resume_user_like (resume_id, user_id),
    INDEX idx_resume_id (resume_id),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 简历分享表
CREATE TABLE IF NOT EXISTS resume_shares (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    resume_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    platform VARCHAR(50) NOT NULL,               -- 分享平台
    share_url VARCHAR(500),                      -- 分享链接
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (resume_id) REFERENCES resumes(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_resume_id (resume_id),
    INDEX idx_user_id (user_id),
    INDEX idx_platform (platform)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==================== 其他功能表 ====================

-- 文件表
CREATE TABLE IF NOT EXISTS files (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    uuid VARCHAR(36) NOT NULL UNIQUE,
    user_id BIGINT UNSIGNED NOT NULL,
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT UNSIGNED NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    file_type ENUM('resume', 'avatar', 'document', 'image', 'other') DEFAULT 'other',
    description TEXT,
    tags JSON,
    is_public BOOLEAN DEFAULT FALSE,
    download_count INT UNSIGNED DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_file_type (file_type),
    INDEX idx_is_public (is_public),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 积分表
CREATE TABLE IF NOT EXISTS points (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    balance INT NOT NULL DEFAULT 100,
    total_earned INT NOT NULL DEFAULT 100,
    total_spent INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_balance (balance)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 积分历史表
CREATE TABLE IF NOT EXISTS point_history (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    points INT NOT NULL,
    type ENUM('earn', 'spend') NOT NULL,
    reason VARCHAR(255) NOT NULL,
    description TEXT,
    reference_type VARCHAR(50),
    reference_id BIGINT UNSIGNED,
    balance_after INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_type (type),
    INDEX idx_reference (reference_type, reference_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 用户会话表
CREATE TABLE IF NOT EXISTS user_sessions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    session_token VARCHAR(255) NOT NULL UNIQUE,
    refresh_token VARCHAR(255) NOT NULL UNIQUE,
    device_info JSON,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_session_token (session_token),
    INDEX idx_refresh_token (refresh_token),
    INDEX idx_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 用户设置表
CREATE TABLE IF NOT EXISTS user_settings (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    theme ENUM('light', 'dark', 'auto') DEFAULT 'auto',
    language VARCHAR(10) DEFAULT 'zh-CN',
    timezone VARCHAR(50) DEFAULT 'Asia/Shanghai',
    email_notifications BOOLEAN DEFAULT TRUE,
    push_notifications BOOLEAN DEFAULT TRUE,
    privacy_level ENUM('public', 'friends', 'private') DEFAULT 'public',
    resume_visibility ENUM('public', 'friends', 'private') DEFAULT 'public',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

COMMIT;