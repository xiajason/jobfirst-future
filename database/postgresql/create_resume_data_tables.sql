-- 简历数据结构化存储表 (PostgreSQL)
-- 存储解析后的结构化数据和向量数据

-- 1. 简历数据表 - 存储解析后的结构化数据
CREATE TABLE IF NOT EXISTS resume_data (
    id SERIAL PRIMARY KEY,
    mysql_resume_id INT NOT NULL, -- 关联到MySQL中的resumes表
    user_id INT NOT NULL,
    
    -- 解析后的结构化数据 (JSON格式)
    personal_info JSONB, -- 个人信息 (姓名、电话、邮箱等)
    work_experience JSONB, -- 工作经历
    education JSONB, -- 教育背景
    skills JSONB, -- 技能列表
    projects JSONB, -- 项目经历
    certifications JSONB, -- 证书资质
    
    -- 向量化数据 (用于AI分析和搜索)
    content_vector VECTOR(1536), -- 简历内容向量
    skills_vector VECTOR(1536), -- 技能向量
    experience_vector VECTOR(1536), -- 经验向量
    
    -- AI分析结果
    ai_analysis JSONB, -- AI分析结果
    keywords TEXT[], -- 关键词数组
    confidence FLOAT DEFAULT 0.0, -- 解析置信度 0-1
    
    -- 元数据
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_resume_data_mysql_id ON resume_data(mysql_resume_id);
CREATE INDEX IF NOT EXISTS idx_resume_data_user_id ON resume_data(user_id);
CREATE INDEX IF NOT EXISTS idx_resume_data_confidence ON resume_data(confidence);

-- 创建向量索引 (使用pgvector扩展)
CREATE INDEX IF NOT EXISTS idx_resume_data_content_vector ON resume_data 
USING ivfflat (content_vector vector_cosine_ops) WITH (lists = 100);

CREATE INDEX IF NOT EXISTS idx_resume_data_skills_vector ON resume_data 
USING ivfflat (skills_vector vector_cosine_ops) WITH (lists = 100);

CREATE INDEX IF NOT EXISTS idx_resume_data_experience_vector ON resume_data 
USING ivfflat (experience_vector vector_cosine_ops) WITH (lists = 100);

-- 创建GIN索引用于JSONB查询
CREATE INDEX IF NOT EXISTS idx_resume_data_personal_info ON resume_data 
USING gin (personal_info);

CREATE INDEX IF NOT EXISTS idx_resume_data_work_experience ON resume_data 
USING gin (work_experience);

CREATE INDEX IF NOT EXISTS idx_resume_data_education ON resume_data 
USING gin (education);

CREATE INDEX IF NOT EXISTS idx_resume_data_skills ON resume_data 
USING gin (skills);

CREATE INDEX IF NOT EXISTS idx_resume_data_projects ON resume_data 
USING gin (projects);

CREATE INDEX IF NOT EXISTS idx_resume_data_certifications ON resume_data 
USING gin (certifications);

-- 2. 简历分析结果表 - 存储AI分析结果
CREATE TABLE IF NOT EXISTS resume_analyses (
    id SERIAL PRIMARY KEY,
    resume_data_id INT NOT NULL,
    mysql_resume_id INT NOT NULL,
    user_id INT NOT NULL,
    
    -- 分析类型和结果
    analysis_type VARCHAR(50) NOT NULL, -- content_analysis, skill_extraction, keyword_analysis, matching_score
    score INT, -- 分析评分 0-100
    suggestions JSONB, -- 改进建议
    keywords TEXT[], -- 关键词提取
    analysis_data JSONB, -- 详细分析数据
    
    -- 向量相似度分析
    similarity_vector VECTOR(1536), -- 相似度向量
    matching_jobs JSONB, -- 匹配的职位信息
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (resume_data_id) REFERENCES resume_data(id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_resume_analyses_resume_data_id ON resume_analyses(resume_data_id);
CREATE INDEX IF NOT EXISTS idx_resume_analyses_mysql_resume_id ON resume_analyses(mysql_resume_id);
CREATE INDEX IF NOT EXISTS idx_resume_analyses_user_id ON resume_analyses(user_id);
CREATE INDEX IF NOT EXISTS idx_resume_analyses_type ON resume_analyses(analysis_type);
CREATE INDEX IF NOT EXISTS idx_resume_analyses_score ON resume_analyses(score);

-- 3. 简历搜索历史表 - 记录搜索和匹配历史
CREATE TABLE IF NOT EXISTS resume_search_history (
    id SERIAL PRIMARY KEY,
    resume_data_id INT NOT NULL,
    user_id INT NOT NULL,
    
    -- 搜索信息
    search_query TEXT NOT NULL,
    search_type VARCHAR(50) NOT NULL, -- keyword_search, vector_search, hybrid_search
    search_filters JSONB, -- 搜索过滤条件
    
    -- 搜索结果
    results_count INT DEFAULT 0,
    search_results JSONB, -- 搜索结果数据
    
    -- 性能指标
    search_time_ms INT, -- 搜索耗时(毫秒)
    vector_search_time_ms INT, -- 向量搜索耗时
    keyword_search_time_ms INT, -- 关键词搜索耗时
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (resume_data_id) REFERENCES resume_data(id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_resume_search_history_resume_data_id ON resume_search_history(resume_data_id);
CREATE INDEX IF NOT EXISTS idx_resume_search_history_user_id ON resume_search_history(user_id);
CREATE INDEX IF NOT EXISTS idx_resume_search_history_type ON resume_search_history(search_type);
CREATE INDEX IF NOT EXISTS idx_resume_search_history_created_at ON resume_search_history(created_at);

-- 4. 简历向量更新任务表 - 管理向量更新任务
CREATE TABLE IF NOT EXISTS resume_vector_tasks (
    id SERIAL PRIMARY KEY,
    resume_data_id INT NOT NULL,
    mysql_resume_id INT NOT NULL,
    
    -- 任务信息
    task_type VARCHAR(50) NOT NULL, -- vector_generation, vector_update, similarity_calculation
    status VARCHAR(20) DEFAULT 'pending', -- pending, processing, completed, failed
    progress INT DEFAULT 0, -- 0-100
    
    -- 任务参数
    model_name VARCHAR(100), -- 使用的AI模型名称
    embedding_model VARCHAR(100), -- 嵌入模型名称
    
    -- 结果数据
    result_data JSONB,
    error_message TEXT,
    
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (resume_data_id) REFERENCES resume_data(id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_resume_vector_tasks_resume_data_id ON resume_vector_tasks(resume_data_id);
CREATE INDEX IF NOT EXISTS idx_resume_vector_tasks_mysql_resume_id ON resume_vector_tasks(mysql_resume_id);
CREATE INDEX IF NOT EXISTS idx_resume_vector_tasks_status ON resume_vector_tasks(status);
CREATE INDEX IF NOT EXISTS idx_resume_vector_tasks_type ON resume_vector_tasks(task_type);

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为resume_data表添加更新时间触发器
DROP TRIGGER IF EXISTS update_resume_data_updated_at ON resume_data;
CREATE TRIGGER update_resume_data_updated_at
    BEFORE UPDATE ON resume_data
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 插入示例数据 (可选)
INSERT INTO resume_data (
    mysql_resume_id, 
    user_id, 
    personal_info, 
    work_experience, 
    education, 
    skills, 
    keywords,
    confidence
) VALUES (
    1,
    4,
    '{"name": "张三", "phone": "138-0000-0000", "email": "zhangsan@example.com", "address": "北京市朝阳区"}',
    '[{"company": "某科技公司", "position": "软件工程师", "start_date": "2020-01", "end_date": "2023-12", "description": "负责后端开发，使用Go语言开发微服务"}]',
    '[{"school": "北京大学", "major": "计算机科学与技术", "degree": "学士", "start_date": "2016-09", "end_date": "2020-06"}]',
    '["Go", "微服务", "数据库", "Docker"]',
    ARRAY['Go', '微服务', '后端开发', '数据库'],
    0.85
) ON CONFLICT (mysql_resume_id) DO NOTHING;
