-- Job服务演进功能数据库表创建脚本
-- 创建时间: 2025年9月29日
-- 功能: 支持远程工作、灵活用工、智能匹配、职业发展等新时代就业形态

-- 1. 远程工作职位表
CREATE TABLE IF NOT EXISTS remote_work_jobs (
    id BIGSERIAL PRIMARY KEY,
    job_id BIGINT NOT NULL,
    remote_type VARCHAR(50) NOT NULL, -- fully_remote, hybrid, flexible
    time_zone VARCHAR(50),
    work_hours VARCHAR(100), -- 工作时间要求
    communication_tools TEXT, -- 沟通工具
    equipment_provided TEXT, -- 设备提供
    flexibility_level VARCHAR(50), -- 灵活度级别
    location_requirement TEXT, -- 地点要求
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_job_id (job_id),
    INDEX idx_remote_type (remote_type),
    INDEX idx_flexibility_level (flexibility_level),
    
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. 灵活用工管理表
CREATE TABLE IF NOT EXISTS flexible_employments (
    id BIGSERIAL PRIMARY KEY,
    job_id BIGINT NOT NULL,
    employment_type VARCHAR(50) NOT NULL, -- contract, freelance, part_time, project_based
    duration VARCHAR(100), -- 工作期限
    payment_type VARCHAR(50), -- hourly, project, milestone
    payment_rate DECIMAL(10,2), -- 支付费率
    flexibility_level VARCHAR(50), -- 灵活度
    skill_requirements TEXT, -- 技能要求
    project_scope TEXT, -- 项目范围
    timeline VARCHAR(200), -- 时间线
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_job_id (job_id),
    INDEX idx_employment_type (employment_type),
    INDEX idx_payment_type (payment_type),
    
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. 智能匹配引擎表
CREATE TABLE IF NOT EXISTS smart_matchings (
    id BIGSERIAL PRIMARY KEY,
    job_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    match_score DECIMAL(5,2), -- 匹配分数
    skill_match DECIMAL(5,2), -- 技能匹配度
    experience_match DECIMAL(5,2), -- 经验匹配度
    location_match DECIMAL(5,2), -- 地点匹配度
    salary_match DECIMAL(5,2), -- 薪资匹配度
    culture_match DECIMAL(5,2), -- 文化匹配度
    ai_recommendation TEXT, -- AI推荐理由
    match_factors JSON, -- 匹配因子
    status VARCHAR(20) DEFAULT 'pending', -- pending, accepted, rejected
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_job_id (job_id),
    INDEX idx_user_id (user_id),
    INDEX idx_match_score (match_score),
    INDEX idx_status (status),
    
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. 个性化职业发展表
CREATE TABLE IF NOT EXISTS career_developments (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    current_level VARCHAR(50), -- 当前职业级别
    target_level VARCHAR(50), -- 目标职业级别
    skills_gap TEXT, -- 技能差距
    development_plan TEXT, -- 发展计划
    recommended_jobs JSON, -- 推荐职位
    skill_assessment JSON, -- 技能评估
    career_goals TEXT, -- 职业目标
    progress_tracking JSON, -- 进度跟踪
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user_id (user_id),
    INDEX idx_current_level (current_level),
    INDEX idx_target_level (target_level),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 5. 工作生活平衡表
CREATE TABLE IF NOT EXISTS work_life_balances (
    id BIGSERIAL PRIMARY KEY,
    job_id BIGINT NOT NULL,
    flexible_hours BOOLEAN DEFAULT FALSE, -- 灵活工作时间
    remote_work_days INT DEFAULT 0, -- 远程工作天数
    vacation_days INT DEFAULT 0, -- 假期天数
    health_benefits TEXT, -- 健康福利
    wellness_programs TEXT, -- 健康项目
    family_support TEXT, -- 家庭支持
    workload_balance VARCHAR(100), -- 工作负载平衡
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_job_id (job_id),
    INDEX idx_flexible_hours (flexible_hours),
    INDEX idx_remote_work_days (remote_work_days),
    
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 6. 技能评估表
CREATE TABLE IF NOT EXISTS skill_assessments (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    skill_name VARCHAR(100) NOT NULL,
    skill_level INT DEFAULT 1, -- 1-5级别
    assessment_method VARCHAR(50), -- 评估方法
    assessment_score DECIMAL(5,2), -- 评估分数
    certification VARCHAR(200), -- 认证
    experience_years INT DEFAULT 0, -- 经验年数
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user_id (user_id),
    INDEX idx_skill_name (skill_name),
    INDEX idx_skill_level (skill_level),
    INDEX idx_assessment_method (assessment_method),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 7. AI职位推荐表
CREATE TABLE IF NOT EXISTS job_recommendations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    job_id BIGINT NOT NULL,
    recommendation_score DECIMAL(5,2), -- 推荐分数
    recommendation_reason TEXT, -- 推荐理由
    ai_insights TEXT, -- AI洞察
    match_factors JSON, -- 匹配因子
    status VARCHAR(20) DEFAULT 'active', -- active, viewed, applied, dismissed
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user_id (user_id),
    INDEX idx_job_id (job_id),
    INDEX idx_recommendation_score (recommendation_score),
    INDEX idx_status (status),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 添加表注释
ALTER TABLE remote_work_jobs COMMENT = '远程工作职位表';
ALTER TABLE flexible_employments COMMENT = '灵活用工管理表';
ALTER TABLE smart_matchings COMMENT = '智能匹配引擎表';
ALTER TABLE career_developments COMMENT = '个性化职业发展表';
ALTER TABLE work_life_balances COMMENT = '工作生活平衡表';
ALTER TABLE skill_assessments COMMENT = '技能评估表';
ALTER TABLE job_recommendations COMMENT = 'AI职位推荐表';
