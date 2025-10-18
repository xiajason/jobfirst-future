-- 能力评估框架数据库表结构
-- 创建时间: 2025年10月3日
-- 用途: 支持理性AI身份服务的能力评估功能

-- 技术能力评估表
CREATE TABLE IF NOT EXISTS technical_competency_assessments (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',
    competency_type ENUM('PROGRAMMING', 'ALGORITHM_DESIGN', 'SYSTEM_ARCHITECTURE', 'DATABASE_DESIGN', 'TESTING', 'DEVOPS', 'SECURITY', 'PERFORMANCE') NOT NULL COMMENT '技术能力类型',
    competency_level ENUM('BEGINNER', 'INTERMEDIATE', 'ADVANCED', 'EXPERT', 'MASTER') NOT NULL COMMENT '能力等级',
    competency_score FLOAT NOT NULL COMMENT '能力评分',
    confidence_score FLOAT NOT NULL COMMENT '置信度',
    evidence_text TEXT COMMENT '证据文本',
    keywords_matched JSON COMMENT '匹配的关键词',
    assessment_details JSON COMMENT '评估详情',
    assessment_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '评估时间戳',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_competency_type (competency_type),
    INDEX idx_competency_level (competency_level),
    INDEX idx_competency_score (competency_score),
    INDEX idx_assessment_timestamp (assessment_timestamp),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='技术能力评估表';

-- 业务能力评估表
CREATE TABLE IF NOT EXISTS business_competency_assessments (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',
    competency_type ENUM('REQUIREMENTS_ANALYSIS', 'PROJECT_MANAGEMENT', 'COMMUNICATION', 'PROBLEM_SOLVING', 'TEAMWORK', 'LEADERSHIP', 'INNOVATION', 'BUSINESS_ACUMEN') NOT NULL COMMENT '业务能力类型',
    competency_level ENUM('BEGINNER', 'INTERMEDIATE', 'ADVANCED', 'EXPERT', 'MASTER') NOT NULL COMMENT '能力等级',
    competency_score FLOAT NOT NULL COMMENT '能力评分',
    confidence_score FLOAT NOT NULL COMMENT '置信度',
    evidence_text TEXT COMMENT '证据文本',
    keywords_matched JSON COMMENT '匹配的关键词',
    assessment_details JSON COMMENT '评估详情',
    assessment_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '评估时间戳',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_competency_type (competency_type),
    INDEX idx_competency_level (competency_level),
    INDEX idx_competency_score (competency_score),
    INDEX idx_assessment_timestamp (assessment_timestamp),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='业务能力评估表';

-- 综合能力评估结果表
CREATE TABLE IF NOT EXISTS competency_assessment_results (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',
    assessment_text TEXT NOT NULL COMMENT '评估文本',
    overall_technical_score FLOAT NOT NULL COMMENT '技术能力综合评分',
    overall_business_score FLOAT NOT NULL COMMENT '业务能力综合评分',
    overall_score FLOAT NOT NULL COMMENT '总体评分',
    competency_profile JSON COMMENT '能力画像',
    growth_recommendations JSON COMMENT '成长建议',
    technical_competencies_count INT DEFAULT 0 COMMENT '技术能力评估数量',
    business_competencies_count INT DEFAULT 0 COMMENT '业务能力评估数量',
    assessment_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '评估时间戳',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_overall_score (overall_score),
    INDEX idx_overall_technical_score (overall_technical_score),
    INDEX idx_overall_business_score (overall_business_score),
    INDEX idx_assessment_timestamp (assessment_timestamp),
    INDEX idx_is_active (is_active),
    FULLTEXT INDEX idx_assessment_text (assessment_text)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='综合能力评估结果表';

-- 能力成长轨迹表
CREATE TABLE IF NOT EXISTS competency_growth_trajectories (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',
    trajectory_period VARCHAR(50) NOT NULL COMMENT '轨迹期间',
    start_date DATE NOT NULL COMMENT '开始日期',
    end_date DATE NOT NULL COMMENT '结束日期',
    initial_technical_score FLOAT NOT NULL COMMENT '初始技术评分',
    final_technical_score FLOAT NOT NULL COMMENT '最终技术评分',
    initial_business_score FLOAT NOT NULL COMMENT '初始业务评分',
    final_business_score FLOAT NOT NULL COMMENT '最终业务评分',
    initial_overall_score FLOAT NOT NULL COMMENT '初始总体评分',
    final_overall_score FLOAT NOT NULL COMMENT '最终总体评分',
    technical_growth_rate FLOAT NOT NULL COMMENT '技术能力成长率',
    business_growth_rate FLOAT NOT NULL COMMENT '业务能力成长率',
    overall_growth_rate FLOAT NOT NULL COMMENT '总体成长率',
    growth_analysis TEXT COMMENT '成长分析',
    improvement_areas JSON COMMENT '改进领域',
    achievement_areas JSON COMMENT '成就领域',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_start_date (start_date),
    INDEX idx_end_date (end_date),
    INDEX idx_overall_growth_rate (overall_growth_rate),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='能力成长轨迹表';

-- 能力基准表
CREATE TABLE IF NOT EXISTS competency_benchmarks (
    id INT PRIMARY KEY AUTO_INCREMENT,
    industry VARCHAR(100) NOT NULL COMMENT '行业名称',
    role_level ENUM('JUNIOR', 'MIDDLE', 'SENIOR', 'LEAD', 'PRINCIPAL', 'EXPERT') NOT NULL COMMENT '角色等级',
    competency_type ENUM('TECHNICAL', 'BUSINESS', 'OVERALL') NOT NULL COMMENT '能力类型',
    avg_score FLOAT NOT NULL COMMENT '平均评分',
    score_std FLOAT NOT NULL COMMENT '评分标准差',
    percentile_25 FLOAT NOT NULL COMMENT '25分位数',
    percentile_50 FLOAT NOT NULL COMMENT '50分位数',
    percentile_75 FLOAT NOT NULL COMMENT '75分位数',
    percentile_90 FLOAT NOT NULL COMMENT '90分位数',
    sample_size INT NOT NULL COMMENT '样本数量',
    benchmark_period VARCHAR(50) NOT NULL COMMENT '基准期间',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_industry_role_type (industry, role_level, competency_type, benchmark_period),
    INDEX idx_industry (industry),
    INDEX idx_role_level (role_level),
    INDEX idx_competency_type (competency_type),
    INDEX idx_avg_score (avg_score),
    INDEX idx_benchmark_period (benchmark_period),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='能力基准表';

-- 能力匹配记录表
CREATE TABLE IF NOT EXISTS competency_matches (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',
    job_id INT NULL COMMENT '职位ID',
    match_type ENUM('TECHNICAL_MATCH', 'BUSINESS_MATCH', 'OVERALL_MATCH') NOT NULL COMMENT '匹配类型',
    user_competency_score FLOAT NOT NULL COMMENT '用户能力评分',
    job_requirement_score FLOAT NOT NULL COMMENT '职位要求评分',
    match_score FLOAT NOT NULL COMMENT '匹配评分',
    match_details JSON COMMENT '匹配详情',
    competency_gaps JSON COMMENT '能力差距',
    recommendations JSON COMMENT '推荐建议',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_job_id (job_id),
    INDEX idx_match_type (match_type),
    INDEX idx_match_score (match_score),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='能力匹配记录表';

-- 能力评估模板表
CREATE TABLE IF NOT EXISTS competency_assessment_templates (
    id INT PRIMARY KEY AUTO_INCREMENT,
    template_name VARCHAR(200) NOT NULL COMMENT '模板名称',
    template_type ENUM('TECHNICAL', 'BUSINESS', 'COMPREHENSIVE') NOT NULL COMMENT '模板类型',
    template_description TEXT COMMENT '模板描述',
    competency_weights JSON NOT NULL COMMENT '能力权重配置',
    assessment_criteria JSON NOT NULL COMMENT '评估标准',
    keywords_config JSON NOT NULL COMMENT '关键词配置',
    scoring_rules JSON NOT NULL COMMENT '评分规则',
    is_default BOOLEAN DEFAULT FALSE COMMENT '是否默认模板',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_template_type (template_type),
    INDEX idx_is_default (is_default),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='能力评估模板表';

-- 插入能力基准数据
INSERT INTO competency_benchmarks (industry, role_level, competency_type, avg_score, score_std, percentile_25, percentile_50, percentile_75, percentile_90, sample_size, benchmark_period) VALUES
-- 技术行业基准
('tech', 'JUNIOR', 'TECHNICAL', 2.1, 0.8, 1.5, 2.0, 2.7, 3.2, 2000, '2024-Q4'),
('tech', 'JUNIOR', 'BUSINESS', 1.8, 0.6, 1.3, 1.7, 2.2, 2.7, 2000, '2024-Q4'),
('tech', 'JUNIOR', 'OVERALL', 2.0, 0.7, 1.4, 1.9, 2.5, 3.0, 2000, '2024-Q4'),

('tech', 'MIDDLE', 'TECHNICAL', 3.2, 0.9, 2.5, 3.1, 3.8, 4.3, 3000, '2024-Q4'),
('tech', 'MIDDLE', 'BUSINESS', 2.8, 0.7, 2.2, 2.7, 3.3, 3.8, 3000, '2024-Q4'),
('tech', 'MIDDLE', 'OVERALL', 3.0, 0.8, 2.4, 2.9, 3.6, 4.1, 3000, '2024-Q4'),

('tech', 'SENIOR', 'TECHNICAL', 4.1, 0.8, 3.4, 4.0, 4.7, 5.0, 2500, '2024-Q4'),
('tech', 'SENIOR', 'BUSINESS', 3.6, 0.8, 2.9, 3.5, 4.2, 4.7, 2500, '2024-Q4'),
('tech', 'SENIOR', 'OVERALL', 3.9, 0.8, 3.2, 3.8, 4.5, 4.9, 2500, '2024-Q4'),

('tech', 'LEAD', 'TECHNICAL', 4.5, 0.7, 3.9, 4.4, 5.0, 5.0, 1500, '2024-Q4'),
('tech', 'LEAD', 'BUSINESS', 4.2, 0.8, 3.5, 4.1, 4.8, 5.0, 1500, '2024-Q4'),
('tech', 'LEAD', 'OVERALL', 4.4, 0.7, 3.7, 4.3, 4.9, 5.0, 1500, '2024-Q4'),

('tech', 'PRINCIPAL', 'TECHNICAL', 4.8, 0.5, 4.4, 4.7, 5.0, 5.0, 800, '2024-Q4'),
('tech', 'PRINCIPAL', 'BUSINESS', 4.6, 0.6, 4.1, 4.5, 5.0, 5.0, 800, '2024-Q4'),
('tech', 'PRINCIPAL', 'OVERALL', 4.7, 0.5, 4.3, 4.6, 5.0, 5.0, 800, '2024-Q4'),

('tech', 'EXPERT', 'TECHNICAL', 5.0, 0.0, 5.0, 5.0, 5.0, 5.0, 400, '2024-Q4'),
('tech', 'EXPERT', 'BUSINESS', 4.8, 0.4, 4.5, 4.8, 5.0, 5.0, 400, '2024-Q4'),
('tech', 'EXPERT', 'OVERALL', 4.9, 0.2, 4.7, 4.9, 5.0, 5.0, 400, '2024-Q4'),

-- 金融行业基准
('finance', 'JUNIOR', 'TECHNICAL', 1.9, 0.7, 1.3, 1.8, 2.4, 2.9, 1500, '2024-Q4'),
('finance', 'JUNIOR', 'BUSINESS', 2.2, 0.8, 1.5, 2.1, 2.8, 3.3, 1500, '2024-Q4'),
('finance', 'JUNIOR', 'OVERALL', 2.0, 0.7, 1.4, 1.9, 2.6, 3.1, 1500, '2024-Q4'),

('finance', 'MIDDLE', 'TECHNICAL', 2.8, 0.8, 2.1, 2.7, 3.4, 3.9, 2000, '2024-Q4'),
('finance', 'MIDDLE', 'BUSINESS', 3.2, 0.8, 2.5, 3.1, 3.8, 4.3, 2000, '2024-Q4'),
('finance', 'MIDDLE', 'OVERALL', 3.0, 0.8, 2.3, 2.9, 3.6, 4.1, 2000, '2024-Q4'),

('finance', 'SENIOR', 'TECHNICAL', 3.6, 0.8, 2.9, 3.5, 4.2, 4.7, 1800, '2024-Q4'),
('finance', 'SENIOR', 'BUSINESS', 4.0, 0.8, 3.3, 3.9, 4.6, 5.0, 1800, '2024-Q4'),
('finance', 'SENIOR', 'OVERALL', 3.8, 0.8, 3.1, 3.7, 4.4, 4.9, 1800, '2024-Q4'),

-- 制造业基准
('manufacturing', 'JUNIOR', 'TECHNICAL', 1.7, 0.6, 1.2, 1.6, 2.1, 2.6, 1200, '2024-Q4'),
('manufacturing', 'JUNIOR', 'BUSINESS', 1.9, 0.7, 1.3, 1.8, 2.4, 2.9, 1200, '2024-Q4'),
('manufacturing', 'JUNIOR', 'OVERALL', 1.8, 0.6, 1.3, 1.7, 2.3, 2.8, 1200, '2024-Q4'),

('manufacturing', 'MIDDLE', 'TECHNICAL', 2.5, 0.7, 1.9, 2.4, 3.0, 3.5, 1500, '2024-Q4'),
('manufacturing', 'MIDDLE', 'BUSINESS', 2.7, 0.7, 2.1, 2.6, 3.2, 3.7, 1500, '2024-Q4'),
('manufacturing', 'MIDDLE', 'OVERALL', 2.6, 0.7, 2.0, 2.5, 3.1, 3.6, 1500, '2024-Q4'),

('manufacturing', 'SENIOR', 'TECHNICAL', 3.2, 0.8, 2.5, 3.1, 3.8, 4.3, 1000, '2024-Q4'),
('manufacturing', 'SENIOR', 'BUSINESS', 3.4, 0.8, 2.7, 3.3, 4.0, 4.5, 1000, '2024-Q4'),
('manufacturing', 'SENIOR', 'OVERALL', 3.3, 0.8, 2.6, 3.2, 3.9, 4.4, 1000, '2024-Q4'),

-- 医疗行业基准
('healthcare', 'JUNIOR', 'TECHNICAL', 1.8, 0.6, 1.3, 1.7, 2.2, 2.7, 800, '2024-Q4'),
('healthcare', 'JUNIOR', 'BUSINESS', 2.1, 0.7, 1.5, 2.0, 2.6, 3.1, 800, '2024-Q4'),
('healthcare', 'JUNIOR', 'OVERALL', 1.9, 0.6, 1.4, 1.8, 2.4, 2.9, 800, '2024-Q4'),

('healthcare', 'MIDDLE', 'TECHNICAL', 2.7, 0.7, 2.1, 2.6, 3.2, 3.7, 1000, '2024-Q4'),
('healthcare', 'MIDDLE', 'BUSINESS', 3.0, 0.7, 2.4, 2.9, 3.5, 4.0, 1000, '2024-Q4'),
('healthcare', 'MIDDLE', 'OVERALL', 2.8, 0.7, 2.3, 2.7, 3.4, 3.9, 1000, '2024-Q4'),

('healthcare', 'SENIOR', 'TECHNICAL', 3.4, 0.8, 2.7, 3.3, 4.0, 4.5, 600, '2024-Q4'),
('healthcare', 'SENIOR', 'BUSINESS', 3.7, 0.8, 3.0, 3.6, 4.3, 4.8, 600, '2024-Q4'),
('healthcare', 'SENIOR', 'OVERALL', 3.5, 0.8, 2.9, 3.5, 4.2, 4.7, 600, '2024-Q4'),

-- 教育行业基准
('education', 'JUNIOR', 'TECHNICAL', 1.6, 0.5, 1.2, 1.5, 2.0, 2.5, 600, '2024-Q4'),
('education', 'JUNIOR', 'BUSINESS', 2.0, 0.7, 1.4, 1.9, 2.5, 3.0, 600, '2024-Q4'),
('education', 'JUNIOR', 'OVERALL', 1.8, 0.6, 1.3, 1.7, 2.3, 2.8, 600, '2024-Q4'),

('education', 'MIDDLE', 'TECHNICAL', 2.3, 0.6, 1.8, 2.2, 2.7, 3.2, 800, '2024-Q4'),
('education', 'MIDDLE', 'BUSINESS', 2.7, 0.7, 2.1, 2.6, 3.2, 3.7, 800, '2024-Q4'),
('education', 'MIDDLE', 'OVERALL', 2.5, 0.6, 2.0, 2.4, 3.0, 3.5, 800, '2024-Q4'),

('education', 'SENIOR', 'TECHNICAL', 2.9, 0.7, 2.3, 2.8, 3.4, 3.9, 500, '2024-Q4'),
('education', 'SENIOR', 'BUSINESS', 3.3, 0.7, 2.7, 3.2, 3.8, 4.3, 500, '2024-Q4'),
('education', 'SENIOR', 'OVERALL', 3.1, 0.7, 2.5, 3.0, 3.6, 4.1, 500, '2024-Q4');

-- 插入默认评估模板
INSERT INTO competency_assessment_templates (template_name, template_type, template_description, competency_weights, assessment_criteria, keywords_config, scoring_rules, is_default) VALUES
('默认技术能力评估模板', 'TECHNICAL', '基于HireVue标准的技术能力评估模板', 
'{"PROGRAMMING": 0.25, "ALGORITHM_DESIGN": 0.20, "SYSTEM_ARCHITECTURE": 0.20, "DATABASE_DESIGN": 0.15, "TESTING": 0.10, "DEVOPS": 0.05, "SECURITY": 0.03, "PERFORMANCE": 0.02}',
'{"BEGINNER": 1, "INTERMEDIATE": 2, "ADVANCED": 3, "EXPERT": 4, "MASTER": 5}',
'{"technical_keywords": "已配置的技术关键词"}',
'{"level_weights": {"beginner": 1.0, "intermediate": 2.0, "advanced": 3.0, "expert": 4.0, "master": 5.0}}',
TRUE),

('默认业务能力评估模板', 'BUSINESS', '基于HireVue标准的业务能力评估模板',
'{"REQUIREMENTS_ANALYSIS": 0.20, "PROJECT_MANAGEMENT": 0.20, "COMMUNICATION": 0.15, "PROBLEM_SOLVING": 0.15, "TEAMWORK": 0.10, "LEADERSHIP": 0.10, "INNOVATION": 0.05, "BUSINESS_ACUMEN": 0.05}',
'{"BEGINNER": 1, "INTERMEDIATE": 2, "ADVANCED": 3, "EXPERT": 4, "MASTER": 5}',
'{"business_keywords": "已配置的业务关键词"}',
'{"level_weights": {"beginner": 1.0, "intermediate": 2.0, "advanced": 3.0, "expert": 4.0, "master": 5.0}}',
TRUE),

('综合能力评估模板', 'COMPREHENSIVE', '技术能力和业务能力的综合评估模板',
'{"technical_weight": 0.6, "business_weight": 0.4}',
'{"BEGINNER": 1, "INTERMEDIATE": 2, "ADVANCED": 3, "EXPERT": 4, "MASTER": 5}',
'{"comprehensive_keywords": "已配置的综合关键词"}',
'{"level_weights": {"beginner": 1.0, "intermediate": 2.0, "advanced": 3.0, "expert": 4.0, "master": 5.0}}',
TRUE);
