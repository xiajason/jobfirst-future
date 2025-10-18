-- 创建职位文档表
CREATE TABLE IF NOT EXISTS job_documents (
    id INT PRIMARY KEY AUTO_INCREMENT,
    job_id INT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    title VARCHAR(255) NOT NULL,
    original_file TEXT NOT NULL,  -- 原始文件路径
    file_content LONGTEXT NOT NULL,  -- Base64编码的原始文件内容
    file_type VARCHAR(50) NOT NULL,  -- 文件类型 (pdf, docx, doc, txt)
    file_size BIGINT NOT NULL,  -- 文件大小
    upload_time TIMESTAMP NOT NULL,  -- 上传时间
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_job_id (job_id),
    INDEX idx_user_id (user_id),
    INDEX idx_upload_time (upload_time),
    INDEX idx_file_type (file_type),
    
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
