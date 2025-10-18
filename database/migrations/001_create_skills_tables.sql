-- 技能标准化数据库表结构
-- 创建时间: 2025年10月3日
-- 用途: 支持理性AI身份服务的技能标准化功能

-- 技能分类表
CREATE TABLE IF NOT EXISTS skill_categories (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL UNIQUE COMMENT '分类名称',
    display_name VARCHAR(100) NOT NULL COMMENT '显示名称',
    description TEXT COMMENT '分类描述',
    parent_id INT NULL COMMENT '父分类ID',
    sort_order INT DEFAULT 0 COMMENT '排序顺序',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_parent_id (parent_id),
    INDEX idx_sort_order (sort_order),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='技能分类表';

-- 标准化技能表
CREATE TABLE IF NOT EXISTS standardized_skills (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE COMMENT '标准化技能名称',
    category_id INT NOT NULL COMMENT '技能分类ID',
    aliases JSON COMMENT '技能别名列表',
    description TEXT COMMENT '技能描述',
    level ENUM('BEGINNER', 'INTERMEDIATE', 'ADVANCED', 'EXPERT', 'MASTER') DEFAULT 'INTERMEDIATE' COMMENT '技能等级',
    related_skills JSON COMMENT '相关技能列表',
    industry_relevance JSON COMMENT '行业相关性评分',
    popularity_score FLOAT DEFAULT 0.0 COMMENT '流行度评分',
    difficulty_score FLOAT DEFAULT 0.0 COMMENT '难度评分',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES skill_categories(id) ON DELETE RESTRICT,
    INDEX idx_category_id (category_id),
    INDEX idx_level (level),
    INDEX idx_popularity_score (popularity_score),
    INDEX idx_difficulty_score (difficulty_score),
    INDEX idx_is_active (is_active),
    FULLTEXT INDEX idx_name_description (name, description)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标准化技能表';

-- 技能别名表
CREATE TABLE IF NOT EXISTS skill_aliases (
    id INT PRIMARY KEY AUTO_INCREMENT,
    skill_id INT NOT NULL COMMENT '技能ID',
    alias VARCHAR(100) NOT NULL COMMENT '别名',
    alias_type ENUM('EXACT', 'SIMILAR', 'ABBREVIATION', 'VARIATION') DEFAULT 'EXACT' COMMENT '别名类型',
    confidence FLOAT DEFAULT 1.0 COMMENT '匹配置信度',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (skill_id) REFERENCES standardized_skills(id) ON DELETE CASCADE,
    UNIQUE KEY uk_skill_alias (skill_id, alias),
    INDEX idx_alias (alias),
    INDEX idx_skill_id (skill_id),
    INDEX idx_alias_type (alias_type),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='技能别名表';

-- 技能关系表
CREATE TABLE IF NOT EXISTS skill_relationships (
    id INT PRIMARY KEY AUTO_INCREMENT,
    source_skill_id INT NOT NULL COMMENT '源技能ID',
    target_skill_id INT NOT NULL COMMENT '目标技能ID',
    relationship_type ENUM('RELATED', 'PREREQUISITE', 'SUCCESSOR', 'SIMILAR', 'COMPLEMENTARY') NOT NULL COMMENT '关系类型',
    strength FLOAT DEFAULT 1.0 COMMENT '关系强度',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (source_skill_id) REFERENCES standardized_skills(id) ON DELETE CASCADE,
    FOREIGN KEY (target_skill_id) REFERENCES standardized_skills(id) ON DELETE CASCADE,
    UNIQUE KEY uk_skill_relationship (source_skill_id, target_skill_id, relationship_type),
    INDEX idx_source_skill_id (source_skill_id),
    INDEX idx_target_skill_id (target_skill_id),
    INDEX idx_relationship_type (relationship_type),
    INDEX idx_strength (strength),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='技能关系表';

-- 行业技能需求表
CREATE TABLE IF NOT EXISTS industry_skill_demands (
    id INT PRIMARY KEY AUTO_INCREMENT,
    industry VARCHAR(100) NOT NULL COMMENT '行业名称',
    skill_id INT NOT NULL COMMENT '技能ID',
    demand_level ENUM('LOW', 'MEDIUM', 'HIGH', 'CRITICAL') DEFAULT 'MEDIUM' COMMENT '需求等级',
    demand_score FLOAT DEFAULT 0.5 COMMENT '需求评分(0-1)',
    trend ENUM('DECREASING', 'STABLE', 'INCREASING', 'RAPIDLY_INCREASING') DEFAULT 'STABLE' COMMENT '需求趋势',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (skill_id) REFERENCES standardized_skills(id) ON DELETE CASCADE,
    UNIQUE KEY uk_industry_skill (industry, skill_id),
    INDEX idx_industry (industry),
    INDEX idx_skill_id (skill_id),
    INDEX idx_demand_level (demand_level),
    INDEX idx_demand_score (demand_score),
    INDEX idx_trend (trend),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='行业技能需求表';

-- 用户技能表
CREATE TABLE IF NOT EXISTS user_skills (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',
    skill_id INT NOT NULL COMMENT '技能ID',
    skill_name VARCHAR(100) NOT NULL COMMENT '用户输入的技能名称',
    standardized_skill_id INT NULL COMMENT '标准化后的技能ID',
    level ENUM('BEGINNER', 'INTERMEDIATE', 'ADVANCED', 'EXPERT', 'MASTER') DEFAULT 'INTERMEDIATE' COMMENT '技能等级',
    experience_years FLOAT DEFAULT 0.0 COMMENT '经验年限',
    experience_description TEXT COMMENT '经验描述',
    confidence_score FLOAT DEFAULT 1.0 COMMENT '置信度评分',
    last_used_date DATE NULL COMMENT '最后使用日期',
    is_verified BOOLEAN DEFAULT FALSE COMMENT '是否已验证',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (skill_id) REFERENCES standardized_skills(id) ON DELETE RESTRICT,
    FOREIGN KEY (standardized_skill_id) REFERENCES standardized_skills(id) ON DELETE SET NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_skill_id (skill_id),
    INDEX idx_standardized_skill_id (standardized_skill_id),
    INDEX idx_level (level),
    INDEX idx_experience_years (experience_years),
    INDEX idx_confidence_score (confidence_score),
    INDEX idx_last_used_date (last_used_date),
    INDEX idx_is_verified (is_verified),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户技能表';

-- 技能匹配记录表
CREATE TABLE IF NOT EXISTS skill_matches (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',
    job_id INT NULL COMMENT '职位ID',
    match_type ENUM('JOB_MATCH', 'SKILL_RECOMMENDATION', 'CAREER_PATH') NOT NULL COMMENT '匹配类型',
    source_skills JSON COMMENT '源技能列表',
    target_skills JSON COMMENT '目标技能列表',
    match_score FLOAT NOT NULL COMMENT '匹配评分',
    match_details JSON COMMENT '匹配详情',
    recommendations JSON COMMENT '推荐建议',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_job_id (job_id),
    INDEX idx_match_type (match_type),
    INDEX idx_match_score (match_score),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='技能匹配记录表';

-- 插入技能分类数据
INSERT INTO skill_categories (name, display_name, description, sort_order) VALUES
('programming_language', '编程语言', '各种编程语言技能', 1),
('framework', '框架技术', '各种开发框架和库', 2),
('database', '数据库技术', '各种数据库管理系统', 3),
('cloud_service', '云服务技术', '各种云平台和服务', 4),
('development_tool', '开发工具', '各种开发工具和环境', 5),
('soft_skill', '软技能', '沟通、领导等软技能', 6),
('business_skill', '业务技能', '业务理解和分析技能', 7),
('industry_skill', '行业技能', '特定行业的专业技能', 8);

-- 插入标准化技能数据
INSERT INTO standardized_skills (name, category_id, aliases, description, level, related_skills, industry_relevance, popularity_score, difficulty_score) VALUES
-- 编程语言
('Python', 1, '["python", "py", "python3"]', 'Python编程语言', 'INTERMEDIATE', '["Django", "Flask", "Pandas", "NumPy", "TensorFlow"]', '{"tech": 0.9, "finance": 0.8, "data": 0.95}', 0.95, 0.6),
('Java', 1, '["java", "jdk", "jvm"]', 'Java编程语言', 'INTERMEDIATE', '["Spring Boot", "Hibernate", "Maven", "Gradle"]', '{"tech": 0.9, "finance": 0.85, "enterprise": 0.95}', 0.9, 0.7),
('Go', 1, '["go", "golang", "go lang"]', 'Go编程语言', 'INTERMEDIATE', '["Gin", "Echo", "Docker", "Kubernetes"]', '{"tech": 0.8, "cloud": 0.9, "microservices": 0.85}', 0.8, 0.7),
('JavaScript', 1, '["javascript", "js", "ecmascript"]', 'JavaScript编程语言', 'INTERMEDIATE', '["Node.js", "React", "Vue.js", "TypeScript"]', '{"tech": 0.95, "web": 0.98, "frontend": 0.95}', 0.98, 0.5),
('TypeScript', 1, '["typescript", "ts", "tsx"]', 'TypeScript编程语言', 'INTERMEDIATE', '["JavaScript", "React", "Angular", "Node.js"]', '{"tech": 0.85, "web": 0.9, "frontend": 0.9}', 0.85, 0.6),

-- 框架技术
('React', 2, '["react", "reactjs", "react.js"]', 'React前端框架', 'INTERMEDIATE', '["JavaScript", "TypeScript", "Redux", "Next.js"]', '{"tech": 0.9, "frontend": 0.95, "web": 0.9}', 0.9, 0.6),
('Vue.js', 2, '["vue", "vuejs", "vue.js"]', 'Vue.js前端框架', 'INTERMEDIATE', '["JavaScript", "TypeScript", "Vuex", "Nuxt.js"]', '{"tech": 0.8, "frontend": 0.85, "web": 0.8}', 0.8, 0.6),
('Spring Boot', 2, '["springboot", "spring boot", "springboot framework"]', 'Spring Boot后端框架', 'INTERMEDIATE', '["Java", "Maven", "Hibernate", "REST API"]', '{"tech": 0.9, "backend": 0.95, "enterprise": 0.9}', 0.9, 0.7),
('Django', 2, '["django", "django framework"]', 'Django Python后端框架', 'INTERMEDIATE', '["Python", "PostgreSQL", "Redis", "Celery"]', '{"tech": 0.8, "backend": 0.85, "python": 0.9}', 0.8, 0.7),

-- 数据库技术
('MySQL', 3, '["mysql", "mysql database"]', 'MySQL关系型数据库', 'INTERMEDIATE', '["SQL", "InnoDB", "MySQL Workbench", "phpMyAdmin"]', '{"tech": 0.9, "database": 0.95, "web": 0.85}', 0.95, 0.6),
('PostgreSQL', 3, '["postgresql", "postgres", "pg"]', 'PostgreSQL关系型数据库', 'INTERMEDIATE', '["SQL", "pgvector", "PostGIS", "pgAdmin"]', '{"tech": 0.8, "database": 0.9, "enterprise": 0.85}', 0.8, 0.7),
('MongoDB', 3, '["mongodb", "mongo", "nosql"]', 'MongoDB文档数据库', 'INTERMEDIATE', '["NoSQL", "Mongoose", "Atlas", "Compass"]', '{"tech": 0.8, "database": 0.85, "nosql": 0.9}', 0.8, 0.6),
('Redis', 3, '["redis", "redis cache"]', 'Redis内存数据库', 'INTERMEDIATE', '["Cache", "Pub/Sub", "Lua", "Cluster"]', '{"tech": 0.8, "database": 0.85, "cache": 0.95}', 0.85, 0.5),

-- 云服务技术
('AWS', 4, '["amazon web services", "amazon aws", "aws cloud"]', 'Amazon Web Services云服务', 'ADVANCED', '["EC2", "S3", "Lambda", "RDS", "CloudFormation"]', '{"tech": 0.9, "cloud": 0.95, "enterprise": 0.9}', 0.9, 0.8),
('Kubernetes', 4, '["k8s", "kubernetes", "kube"]', 'Kubernetes容器编排平台', 'ADVANCED', '["Docker", "Helm", "Istio", "Prometheus"]', '{"tech": 0.9, "cloud": 0.95, "devops": 0.9}', 0.85, 0.8),
('Docker', 4, '["docker", "docker container", "containerization"]', 'Docker容器化技术', 'INTERMEDIATE', '["Kubernetes", "Docker Compose", "Dockerfile", "Registry"]', '{"tech": 0.9, "cloud": 0.9, "devops": 0.95}', 0.95, 0.6),

-- 开发工具
('Git', 5, '["git", "git version control", "git scm"]', 'Git版本控制系统', 'INTERMEDIATE', '["GitHub", "GitLab", "Bitbucket", "Git Flow"]', '{"tech": 0.95, "development": 0.98, "collaboration": 0.9}', 0.98, 0.4),
('Jenkins', 5, '["jenkins", "jenkins ci", "jenkins pipeline"]', 'Jenkins持续集成工具', 'ADVANCED', '["CI/CD", "Pipeline", "Docker", "Kubernetes"]', '{"tech": 0.8, "devops": 0.9, "automation": 0.85}', 0.8, 0.7),

-- 软技能
('Leadership', 6, '["leadership", "team leadership", "leadership skills"]', '领导力和团队管理', 'ADVANCED', '["Team Management", "Project Management", "Communication", "Decision Making"]', '{"management": 0.95, "business": 0.9, "tech": 0.7}', 0.9, 0.8),
('Communication', 6, '["communication", "communication skills", "verbal communication"]', '沟通和表达能力', 'INTERMEDIATE', '["Presentation", "Writing", "Public Speaking", "Interpersonal Skills"]', '{"business": 0.95, "management": 0.9, "tech": 0.8}', 0.95, 0.5),
('Problem Solving', 6, '["problem solving", "analytical thinking", "critical thinking"]', '问题解决和分析思维', 'INTERMEDIATE', '["Analytical Thinking", "Critical Thinking", "Innovation", "Creativity"]', '{"tech": 0.9, "business": 0.9, "management": 0.85}', 0.9, 0.6);

-- 插入技能关系数据
INSERT INTO skill_relationships (source_skill_id, target_skill_id, relationship_type, strength) VALUES
-- Python相关技能
(1, 7, 'RELATED', 0.8), -- Python -> Django
(1, 8, 'RELATED', 0.7), -- Python -> Flask
(1, 9, 'RELATED', 0.6), -- Python -> PostgreSQL

-- Java相关技能
(2, 9, 'RELATED', 0.8), -- Java -> Spring Boot
(2, 10, 'RELATED', 0.6), -- Java -> MySQL

-- JavaScript相关技能
(4, 6, 'RELATED', 0.9), -- JavaScript -> React
(4, 7, 'RELATED', 0.8), -- JavaScript -> Vue.js
(4, 5, 'RELATED', 0.9), -- JavaScript -> TypeScript

-- 数据库相关技能
(10, 11, 'RELATED', 0.7), -- MySQL -> PostgreSQL
(10, 12, 'RELATED', 0.6), -- MySQL -> MongoDB
(11, 12, 'RELATED', 0.6), -- PostgreSQL -> MongoDB

-- 云服务相关技能
(14, 15, 'RELATED', 0.8), -- AWS -> Kubernetes
(15, 16, 'RELATED', 0.9), -- Kubernetes -> Docker
(14, 16, 'RELATED', 0.7), -- AWS -> Docker

-- 软技能关系
(19, 20, 'RELATED', 0.8), -- Leadership -> Communication
(19, 21, 'RELATED', 0.7), -- Leadership -> Problem Solving
(20, 21, 'RELATED', 0.6); -- Communication -> Problem Solving

-- 插入行业技能需求数据
INSERT INTO industry_skill_demands (industry, skill_id, demand_level, demand_score, trend) VALUES
-- 技术行业
('tech', 1, 'CRITICAL', 0.95, 'RAPIDLY_INCREASING'), -- Python
('tech', 2, 'HIGH', 0.85, 'INCREASING'), -- Java
('tech', 4, 'CRITICAL', 0.98, 'RAPIDLY_INCREASING'), -- JavaScript
('tech', 6, 'HIGH', 0.9, 'INCREASING'), -- React
('tech', 10, 'HIGH', 0.85, 'STABLE'), -- MySQL
('tech', 14, 'HIGH', 0.9, 'INCREASING'), -- AWS

-- 金融行业
('finance', 1, 'HIGH', 0.8, 'INCREASING'), -- Python
('finance', 2, 'HIGH', 0.85, 'STABLE'), -- Java
('finance', 10, 'HIGH', 0.9, 'STABLE'), -- MySQL
('finance', 19, 'CRITICAL', 0.95, 'STABLE'), -- Leadership

-- 数据行业
('data', 1, 'CRITICAL', 0.95, 'RAPIDLY_INCREASING'), -- Python
('data', 4, 'HIGH', 0.8, 'INCREASING'), -- JavaScript
('data', 11, 'HIGH', 0.85, 'INCREASING'), -- PostgreSQL
('data', 12, 'MEDIUM', 0.7, 'INCREASING'), -- MongoDB

-- 前端开发
('frontend', 4, 'CRITICAL', 0.98, 'STABLE'), -- JavaScript
('frontend', 5, 'HIGH', 0.9, 'INCREASING'), -- TypeScript
('frontend', 6, 'CRITICAL', 0.95, 'STABLE'), -- React
('frontend', 7, 'MEDIUM', 0.8, 'STABLE'), -- Vue.js

-- 后端开发
('backend', 1, 'HIGH', 0.9, 'INCREASING'), -- Python
('backend', 2, 'HIGH', 0.9, 'STABLE'), -- Java
('backend', 3, 'MEDIUM', 0.8, 'INCREASING'), -- Go
('backend', 8, 'HIGH', 0.85, 'INCREASING'), -- Spring Boot
('backend', 9, 'HIGH', 0.85, 'INCREASING'), -- Django

-- DevOps
('devops', 3, 'HIGH', 0.8, 'INCREASING'), -- Go
('devops', 15, 'CRITICAL', 0.95, 'RAPIDLY_INCREASING'), -- Kubernetes
('devops', 16, 'CRITICAL', 0.98, 'RAPIDLY_INCREASING'), -- Docker
('devops', 14, 'HIGH', 0.9, 'INCREASING'), -- AWS
('devops', 17, 'HIGH', 0.85, 'STABLE'); -- Jenkins
