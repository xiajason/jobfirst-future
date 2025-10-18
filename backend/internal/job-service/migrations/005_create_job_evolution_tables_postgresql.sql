-- Job服务演进功能数据库表创建脚本 (PostgreSQL版本)
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
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_remote_work_jobs_job_id ON remote_work_jobs(job_id);
CREATE INDEX IF NOT EXISTS idx_remote_work_jobs_remote_type ON remote_work_jobs(remote_type);
CREATE INDEX IF NOT EXISTS idx_remote_work_jobs_flexibility_level ON remote_work_jobs(flexibility_level);

-- 创建外键约束
ALTER TABLE remote_work_jobs ADD CONSTRAINT fk_remote_work_jobs_job_id 
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE;

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
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_flexible_employments_job_id ON flexible_employments(job_id);
CREATE INDEX IF NOT EXISTS idx_flexible_employments_employment_type ON flexible_employments(employment_type);
CREATE INDEX IF NOT EXISTS idx_flexible_employments_payment_type ON flexible_employments(payment_type);

-- 创建外键约束
ALTER TABLE flexible_employments ADD CONSTRAINT fk_flexible_employments_job_id 
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE;

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
    match_factors JSONB, -- 匹配因子
    status VARCHAR(20) DEFAULT 'pending', -- pending, accepted, rejected
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_smart_matchings_job_id ON smart_matchings(job_id);
CREATE INDEX IF NOT EXISTS idx_smart_matchings_user_id ON smart_matchings(user_id);
CREATE INDEX IF NOT EXISTS idx_smart_matchings_match_score ON smart_matchings(match_score);
CREATE INDEX IF NOT EXISTS idx_smart_matchings_status ON smart_matchings(status);

-- 创建外键约束
ALTER TABLE smart_matchings ADD CONSTRAINT fk_smart_matchings_job_id 
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE;
ALTER TABLE smart_matchings ADD CONSTRAINT fk_smart_matchings_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- 4. 个性化职业发展表
CREATE TABLE IF NOT EXISTS career_developments (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    current_level VARCHAR(50), -- 当前职业级别
    target_level VARCHAR(50), -- 目标职业级别
    skills_gap TEXT, -- 技能差距
    development_plan TEXT, -- 发展计划
    recommended_jobs JSONB, -- 推荐职位
    skill_assessment JSONB, -- 技能评估
    career_goals TEXT, -- 职业目标
    progress_tracking JSONB, -- 进度跟踪
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_career_developments_user_id ON career_developments(user_id);
CREATE INDEX IF NOT EXISTS idx_career_developments_current_level ON career_developments(current_level);
CREATE INDEX IF NOT EXISTS idx_career_developments_target_level ON career_developments(target_level);

-- 创建外键约束
ALTER TABLE career_developments ADD CONSTRAINT fk_career_developments_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- 5. 工作生活平衡表
CREATE TABLE IF NOT EXISTS work_life_balances (
    id BIGSERIAL PRIMARY KEY,
    job_id BIGINT NOT NULL,
    flexible_hours BOOLEAN DEFAULT FALSE, -- 灵活工作时间
    remote_work_days INTEGER DEFAULT 0, -- 远程工作天数
    vacation_days INTEGER DEFAULT 0, -- 假期天数
    health_benefits TEXT, -- 健康福利
    wellness_programs TEXT, -- 健康项目
    family_support TEXT, -- 家庭支持
    workload_balance VARCHAR(100), -- 工作负载平衡
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_work_life_balances_job_id ON work_life_balances(job_id);
CREATE INDEX IF NOT EXISTS idx_work_life_balances_flexible_hours ON work_life_balances(flexible_hours);
CREATE INDEX IF NOT EXISTS idx_work_life_balances_remote_work_days ON work_life_balances(remote_work_days);

-- 创建外键约束
ALTER TABLE work_life_balances ADD CONSTRAINT fk_work_life_balances_job_id 
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE;

-- 6. 技能评估表
CREATE TABLE IF NOT EXISTS skill_assessments (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    skill_name VARCHAR(100) NOT NULL,
    skill_level INTEGER DEFAULT 1, -- 1-5级别
    assessment_method VARCHAR(50), -- 评估方法
    assessment_score DECIMAL(5,2), -- 评估分数
    certification VARCHAR(200), -- 认证
    experience_years INTEGER DEFAULT 0, -- 经验年数
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_skill_assessments_user_id ON skill_assessments(user_id);
CREATE INDEX IF NOT EXISTS idx_skill_assessments_skill_name ON skill_assessments(skill_name);
CREATE INDEX IF NOT EXISTS idx_skill_assessments_skill_level ON skill_assessments(skill_level);
CREATE INDEX IF NOT EXISTS idx_skill_assessments_assessment_method ON skill_assessments(assessment_method);

-- 创建外键约束
ALTER TABLE skill_assessments ADD CONSTRAINT fk_skill_assessments_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- 7. AI职位推荐表
CREATE TABLE IF NOT EXISTS job_recommendations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    job_id BIGINT NOT NULL,
    recommendation_score DECIMAL(5,2), -- 推荐分数
    recommendation_reason TEXT, -- 推荐理由
    ai_insights TEXT, -- AI洞察
    match_factors JSONB, -- 匹配因子
    status VARCHAR(20) DEFAULT 'active', -- active, viewed, applied, dismissed
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_job_recommendations_user_id ON job_recommendations(user_id);
CREATE INDEX IF NOT EXISTS idx_job_recommendations_job_id ON job_recommendations(job_id);
CREATE INDEX IF NOT EXISTS idx_job_recommendations_recommendation_score ON job_recommendations(recommendation_score);
CREATE INDEX IF NOT EXISTS idx_job_recommendations_status ON job_recommendations(status);

-- 创建外键约束
ALTER TABLE job_recommendations ADD CONSTRAINT fk_job_recommendations_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE job_recommendations ADD CONSTRAINT fk_job_recommendations_job_id 
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE;

-- 添加表注释 (PostgreSQL语法)
COMMENT ON TABLE remote_work_jobs IS '远程工作职位表';
COMMENT ON TABLE flexible_employments IS '灵活用工管理表';
COMMENT ON TABLE smart_matchings IS '智能匹配引擎表';
COMMENT ON TABLE career_developments IS '个性化职业发展表';
COMMENT ON TABLE work_life_balances IS '工作生活平衡表';
COMMENT ON TABLE skill_assessments IS '技能评估表';
COMMENT ON TABLE job_recommendations IS 'AI职位推荐表';
