-- 职位匹配系统数据库迁移脚本 - PostgreSQL部分
-- 创建时间: 2025-09-13
-- 版本: 1.0.0

-- ==============================================
-- PostgreSQL数据库 - 向量数据表
-- ==============================================

-- 职位向量表
CREATE TABLE IF NOT EXISTS job_vectors (
    id SERIAL PRIMARY KEY,
    job_id INTEGER NOT NULL UNIQUE,
    title_vector vector(1536),
    description_vector vector(1536),
    requirements_vector vector(1536),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建向量索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_vectors_title_hnsw 
ON job_vectors USING hnsw (title_vector vector_cosine_ops) 
WITH (m = 16, ef_construction = 64);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_vectors_description_hnsw 
ON job_vectors USING hnsw (description_vector vector_cosine_ops) 
WITH (m = 16, ef_construction = 64);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_vectors_requirements_hnsw 
ON job_vectors USING hnsw (requirements_vector vector_cosine_ops) 
WITH (m = 16, ef_construction = 64);

-- 创建复合索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_vectors_job_id 
ON job_vectors (job_id);

-- ==============================================
-- 完成标记
-- ==============================================

-- 迁移完成
-- 创建时间: 2025-09-13
-- 版本: 1.0.0
-- 描述: 创建职位匹配系统PostgreSQL向量数据表
