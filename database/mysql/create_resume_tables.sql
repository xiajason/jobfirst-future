-- 简历相关数据库表创建脚本
-- 支持文件上传和解析的完整架构

USE jobfirst;

-- 1. 简历文件表 - 存储原始上传的文件信息
CREATE TABLE IF NOT EXISTS resume_files (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    file_type VARCHAR(50) NOT NULL, -- pdf, docx, doc
    mime_type VARCHAR(100) NOT NULL,
    upload_status VARCHAR(20) DEFAULT 'uploaded', -- uploaded, processing, completed, failed
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_upload_status (upload_status),
    INDEX idx_file_type (file_type)
);

-- 2. 简历表 - 存储简历基本信息和元数据
CREATE TABLE IF NOT EXISTS resumes (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    file_id INT, -- 关联到resume_files表
    title VARCHAR(200) NOT NULL,
    content TEXT, -- 解析后的文本内容
    creation_mode VARCHAR(20) DEFAULT 'markdown', -- markdown, upload
    template_id INT,
    status VARCHAR(20) DEFAULT 'draft', -- draft, published, archived
    is_public BOOLEAN DEFAULT FALSE,
    view_count INT DEFAULT 0,
    
    -- 解析状态和PostgreSQL关联
    parsing_status VARCHAR(20) DEFAULT 'pending', -- pending, processing, completed, failed
    parsing_error TEXT, -- 解析错误信息
    postgresql_id INT, -- 关联到PostgreSQL中的resume_data表
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_user_id (user_id),
    INDEX idx_file_id (file_id),
    INDEX idx_status (status),
    INDEX idx_creation_mode (creation_mode),
    INDEX idx_parsing_status (parsing_status),
    INDEX idx_postgresql_id (postgresql_id),
    
    FOREIGN KEY (file_id) REFERENCES resume_files(id) ON DELETE SET NULL
);

-- 3. 简历解析任务表 - 管理异步解析任务
CREATE TABLE IF NOT EXISTS resume_parsing_tasks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    resume_id INT NOT NULL,
    file_id INT NOT NULL,
    task_type VARCHAR(50) NOT NULL, -- file_parsing, ai_analysis, content_extraction
    status VARCHAR(20) DEFAULT 'pending', -- pending, processing, completed, failed
    progress INT DEFAULT 0, -- 0-100
    error_message TEXT,
    result_data JSON, -- 解析结果数据
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_resume_id (resume_id),
    INDEX idx_file_id (file_id),
    INDEX idx_status (status),
    INDEX idx_task_type (task_type),
    
    FOREIGN KEY (resume_id) REFERENCES resumes(id) ON DELETE CASCADE,
    FOREIGN KEY (file_id) REFERENCES resume_files(id) ON DELETE CASCADE
);

-- 4. 简历模板表
CREATE TABLE IF NOT EXISTS resume_templates (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    template_data JSON NOT NULL, -- 模板结构数据
    preview_image VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_category (category),
    INDEX idx_is_active (is_active)
);

-- 5. 简历分析结果表
CREATE TABLE IF NOT EXISTS resume_analyses (
    id INT AUTO_INCREMENT PRIMARY KEY,
    resume_id INT NOT NULL,
    analysis_type VARCHAR(50) NOT NULL, -- content_analysis, skill_extraction, keyword_analysis
    score INT, -- 分析评分 0-100
    suggestions JSON, -- 改进建议
    keywords JSON, -- 关键词提取
    analysis_data JSON, -- 详细分析数据
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_resume_id (resume_id),
    INDEX idx_analysis_type (analysis_type),
    
    FOREIGN KEY (resume_id) REFERENCES resumes(id) ON DELETE CASCADE
);

-- 插入一些默认模板数据
INSERT INTO resume_templates (name, description, category, template_data, is_active) VALUES
('现代简约', '简洁现代的简历模板，适合技术岗位', 'technical', '{"sections":["personal_info","work_experience","education","skills","projects"],"layout":"modern"}', TRUE),
('传统商务', '传统商务风格简历模板，适合管理岗位', 'business', '{"sections":["personal_info","work_experience","education","skills","certifications"],"layout":"traditional"}', TRUE),
('创意设计', '创意设计风格简历模板，适合设计岗位', 'creative', '{"sections":["personal_info","work_experience","education","skills","projects","certifications"],"layout":"creative"}', TRUE)
ON DUPLICATE KEY UPDATE name=VALUES(name);
