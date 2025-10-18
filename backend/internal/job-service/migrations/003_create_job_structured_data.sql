-- 创建职位结构化数据表
CREATE TABLE IF NOT EXISTS job_structured_data (
    id INT PRIMARY KEY AUTO_INCREMENT,
    job_id INT UNSIGNED NOT NULL,
    task_id INT NOT NULL,
    basic_info JSON,  -- 基本信息 (职位标题、类型、地点、薪资、招聘人数)
    requirements_info JSON,  -- 要求信息 (学历、经验、技能、语言、证书)
    responsibilities_info JSON,  -- 职责信息 (工作职责、内容、汇报关系、团队规模)
    benefits_info JSON,  -- 福利信息 (薪资福利、假期制度、培训机会、晋升通道)
    confidence FLOAT,  -- 解析置信度 0-1
    parsing_version VARCHAR(50) DEFAULT 'mineru-v1.0',  -- 解析器版本
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_job_id (job_id),
    INDEX idx_task_id (task_id),
    INDEX idx_confidence (confidence),
    
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE,
    FOREIGN KEY (task_id) REFERENCES job_parsing_tasks(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
