-- 经验量化分析数据库表结构
-- 创建时间: 2025年10月3日
-- 用途: 支持理性AI身份服务的经验量化分析功能

-- 项目复杂度评估表
CREATE TABLE IF NOT EXISTS project_complexity_assessments (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',
    project_id INT NULL COMMENT '项目ID',
    project_title VARCHAR(200) NOT NULL COMMENT '项目标题',
    project_description TEXT NOT NULL COMMENT '项目描述',
    technical_complexity FLOAT NOT NULL COMMENT '技术复杂度评分',
    business_complexity FLOAT NOT NULL COMMENT '业务复杂度评分',
    team_complexity FLOAT NOT NULL COMMENT '团队复杂度评分',
    overall_complexity FLOAT NOT NULL COMMENT '整体复杂度评分',
    complexity_level ENUM('LOW', 'MEDIUM', 'HIGH', 'VERY_HIGH', 'EXTREME') NOT NULL COMMENT '复杂度等级',
    complexity_factors JSON COMMENT '复杂度因子详情',
    assessment_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '评估时间戳',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_project_id (project_id),
    INDEX idx_complexity_level (complexity_level),
    INDEX idx_overall_complexity (overall_complexity),
    INDEX idx_assessment_timestamp (assessment_timestamp),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目复杂度评估表';

-- 量化成果表
CREATE TABLE IF NOT EXISTS quantified_achievements (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',
    project_id INT NULL COMMENT '项目ID',
    achievement_type ENUM('PERFORMANCE', 'EFFICIENCY', 'COST_SAVING', 'REVENUE', 'USER_GROWTH', 'QUALITY', 'INNOVATION', 'TEAM') NOT NULL COMMENT '成果类型',
    description TEXT NOT NULL COMMENT '成果描述',
    metric VARCHAR(500) NOT NULL COMMENT '指标描述',
    value DECIMAL(15,4) NOT NULL COMMENT '数值',
    unit VARCHAR(50) NOT NULL COMMENT '单位',
    impact_score FLOAT NOT NULL COMMENT '影响力评分',
    confidence FLOAT NOT NULL COMMENT '置信度',
    achievement_date DATE NULL COMMENT '成果日期',
    source_text TEXT COMMENT '来源文本',
    is_verified BOOLEAN DEFAULT FALSE COMMENT '是否已验证',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_project_id (project_id),
    INDEX idx_achievement_type (achievement_type),
    INDEX idx_value (value),
    INDEX idx_impact_score (impact_score),
    INDEX idx_achievement_date (achievement_date),
    INDEX idx_is_verified (is_verified),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='量化成果表';

-- 领导力指标表
CREATE TABLE IF NOT EXISTS leadership_indicators (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',
    project_id INT NULL COMMENT '项目ID',
    indicator_type ENUM('TEAM_MANAGEMENT', 'PROJECT_LEADERSHIP', 'MENTORING', 'DECISION_MAKING', 'STRATEGIC_THINKING', 'INNOVATION') NOT NULL COMMENT '指标类型',
    indicator_score FLOAT NOT NULL COMMENT '指标评分',
    evidence_text TEXT COMMENT '证据文本',
    confidence FLOAT NOT NULL COMMENT '置信度',
    assessment_date DATE NOT NULL COMMENT '评估日期',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_project_id (project_id),
    INDEX idx_indicator_type (indicator_type),
    INDEX idx_indicator_score (indicator_score),
    INDEX idx_assessment_date (assessment_date),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='领导力指标表';

-- 经验分析结果表
CREATE TABLE IF NOT EXISTS experience_analyses (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',
    experience_text TEXT NOT NULL COMMENT '经验文本',
    project_complexity_id INT NULL COMMENT '项目复杂度评估ID',
    experience_score FLOAT NOT NULL COMMENT '经验评分',
    growth_trajectory FLOAT NOT NULL COMMENT '成长轨迹',
    leadership_score FLOAT NOT NULL COMMENT '领导力评分',
    achievement_count INT DEFAULT 0 COMMENT '成果数量',
    analysis_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '分析时间戳',
    analysis_metadata JSON COMMENT '分析元数据',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (project_complexity_id) REFERENCES project_complexity_assessments(id) ON DELETE SET NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_project_complexity_id (project_complexity_id),
    INDEX idx_experience_score (experience_score),
    INDEX idx_growth_trajectory (growth_trajectory),
    INDEX idx_analysis_timestamp (analysis_timestamp),
    INDEX idx_is_active (is_active),
    FULLTEXT INDEX idx_experience_text (experience_text)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='经验分析结果表';

-- 经验成长轨迹表
CREATE TABLE IF NOT EXISTS experience_trajectories (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',
    trajectory_period VARCHAR(50) NOT NULL COMMENT '轨迹期间',
    start_date DATE NOT NULL COMMENT '开始日期',
    end_date DATE NOT NULL COMMENT '结束日期',
    initial_score FLOAT NOT NULL COMMENT '初始评分',
    final_score FLOAT NOT NULL COMMENT '最终评分',
    growth_rate FLOAT NOT NULL COMMENT '成长率',
    complexity_trend JSON COMMENT '复杂度趋势',
    achievement_trend JSON COMMENT '成果趋势',
    leadership_trend JSON COMMENT '领导力趋势',
    trajectory_analysis TEXT COMMENT '轨迹分析',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_start_date (start_date),
    INDEX idx_end_date (end_date),
    INDEX idx_growth_rate (growth_rate),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='经验成长轨迹表';

-- 行业经验基准表
CREATE TABLE IF NOT EXISTS industry_experience_benchmarks (
    id INT PRIMARY KEY AUTO_INCREMENT,
    industry VARCHAR(100) NOT NULL COMMENT '行业名称',
    experience_level ENUM('JUNIOR', 'MIDDLE', 'SENIOR', 'LEAD', 'PRINCIPAL', 'EXPERT') NOT NULL COMMENT '经验等级',
    avg_complexity FLOAT NOT NULL COMMENT '平均复杂度',
    avg_achievement_score FLOAT NOT NULL COMMENT '平均成果评分',
    avg_leadership_score FLOAT NOT NULL COMMENT '平均领导力评分',
    avg_experience_score FLOAT NOT NULL COMMENT '平均经验评分',
    benchmark_period VARCHAR(50) NOT NULL COMMENT '基准期间',
    sample_size INT NOT NULL COMMENT '样本数量',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_industry_level (industry, experience_level, benchmark_period),
    INDEX idx_industry (industry),
    INDEX idx_experience_level (experience_level),
    INDEX idx_avg_experience_score (avg_experience_score),
    INDEX idx_benchmark_period (benchmark_period),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='行业经验基准表';

-- 经验匹配记录表
CREATE TABLE IF NOT EXISTS experience_matches (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',
    job_id INT NULL COMMENT '职位ID',
    match_type ENUM('EXPERIENCE_MATCH', 'COMPLEXITY_MATCH', 'ACHIEVEMENT_MATCH', 'LEADERSHIP_MATCH') NOT NULL COMMENT '匹配类型',
    user_experience_score FLOAT NOT NULL COMMENT '用户经验评分',
    job_requirement_score FLOAT NOT NULL COMMENT '职位要求评分',
    match_score FLOAT NOT NULL COMMENT '匹配评分',
    match_details JSON COMMENT '匹配详情',
    recommendations JSON COMMENT '推荐建议',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_job_id (job_id),
    INDEX idx_match_type (match_type),
    INDEX idx_match_score (match_score),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='经验匹配记录表';

-- 插入行业经验基准数据
INSERT INTO industry_experience_benchmarks (industry, experience_level, avg_complexity, avg_achievement_score, avg_leadership_score, avg_experience_score, benchmark_period, sample_size) VALUES
-- 技术行业
('tech', 'JUNIOR', 2.0, 15.0, 5.0, 25.0, '2024-Q4', 1000),
('tech', 'MIDDLE', 3.2, 25.0, 12.0, 45.0, '2024-Q4', 1500),
('tech', 'SENIOR', 4.1, 35.0, 20.0, 65.0, '2024-Q4', 1200),
('tech', 'LEAD', 4.5, 45.0, 35.0, 80.0, '2024-Q4', 800),
('tech', 'PRINCIPAL', 4.8, 55.0, 50.0, 95.0, '2024-Q4', 400),
('tech', 'EXPERT', 5.0, 65.0, 60.0, 105.0, '2024-Q4', 200),

-- 金融行业
('finance', 'JUNIOR', 2.2, 12.0, 4.0, 22.0, '2024-Q4', 800),
('finance', 'MIDDLE', 3.0, 22.0, 10.0, 40.0, '2024-Q4', 1200),
('finance', 'SENIOR', 3.8, 32.0, 18.0, 60.0, '2024-Q4', 1000),
('finance', 'LEAD', 4.2, 42.0, 32.0, 75.0, '2024-Q4', 600),
('finance', 'PRINCIPAL', 4.6, 52.0, 48.0, 90.0, '2024-Q4', 300),
('finance', 'EXPERT', 4.9, 62.0, 58.0, 100.0, '2024-Q4', 150),

-- 制造业
('manufacturing', 'JUNIOR', 1.8, 10.0, 3.0, 20.0, '2024-Q4', 600),
('manufacturing', 'MIDDLE', 2.8, 20.0, 8.0, 35.0, '2024-Q4', 900),
('manufacturing', 'SENIOR', 3.6, 30.0, 15.0, 55.0, '2024-Q4', 800),
('manufacturing', 'LEAD', 4.0, 40.0, 28.0, 70.0, '2024-Q4', 500),
('manufacturing', 'PRINCIPAL', 4.4, 50.0, 45.0, 85.0, '2024-Q4', 250),
('manufacturing', 'EXPERT', 4.7, 60.0, 55.0, 95.0, '2024-Q4', 120),

-- 医疗行业
('healthcare', 'JUNIOR', 2.1, 13.0, 4.5, 23.0, '2024-Q4', 400),
('healthcare', 'MIDDLE', 3.1, 23.0, 11.0, 42.0, '2024-Q4', 600),
('healthcare', 'SENIOR', 3.9, 33.0, 19.0, 62.0, '2024-Q4', 500),
('healthcare', 'LEAD', 4.3, 43.0, 34.0, 77.0, '2024-Q4', 300),
('healthcare', 'PRINCIPAL', 4.7, 53.0, 49.0, 92.0, '2024-Q4', 150),
('healthcare', 'EXPERT', 4.9, 63.0, 59.0, 102.0, '2024-Q4', 80),

-- 教育行业
('education', 'JUNIOR', 1.9, 11.0, 3.5, 21.0, '2024-Q4', 300),
('education', 'MIDDLE', 2.9, 21.0, 9.0, 38.0, '2024-Q4', 500),
('education', 'SENIOR', 3.7, 31.0, 17.0, 58.0, '2024-Q4', 400),
('education', 'LEAD', 4.1, 41.0, 30.0, 72.0, '2024-Q4', 250),
('education', 'PRINCIPAL', 4.5, 51.0, 46.0, 87.0, '2024-Q4', 120),
('education', 'EXPERT', 4.8, 61.0, 56.0, 97.0, '2024-Q4', 60);
