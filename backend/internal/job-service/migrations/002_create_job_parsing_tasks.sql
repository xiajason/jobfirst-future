-- 创建职位解析任务表
CREATE TABLE IF NOT EXISTS job_parsing_tasks (
    id INT PRIMARY KEY AUTO_INCREMENT,
    job_id INT UNSIGNED NOT NULL,
    document_id INT NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',  -- pending, processing, completed, failed
    progress INT DEFAULT 0,  -- 解析进度 0-100
    error_message TEXT,  -- 错误信息
    result_data TEXT,  -- 解析结果JSON
    mineru_task_id VARCHAR(100),  -- MinerU任务ID
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_job_id (job_id),
    INDEX idx_document_id (document_id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_mineru_task_id (mineru_task_id),
    
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE,
    FOREIGN KEY (document_id) REFERENCES job_documents(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
