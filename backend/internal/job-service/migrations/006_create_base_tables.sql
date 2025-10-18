-- 创建Job服务基础表 (PostgreSQL版本)
-- 创建时间: 2025年9月29日
-- 功能: 创建jobs和users基础表

-- 1. 创建users表
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    role VARCHAR(50) DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建users表索引
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- 2. 创建jobs表
CREATE TABLE IF NOT EXISTS jobs (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    requirements TEXT,
    company_id BIGINT NOT NULL,
    industry VARCHAR(100),
    location VARCHAR(200),
    salary_min INTEGER,
    salary_max INTEGER,
    experience VARCHAR(50),
    education VARCHAR(100),
    job_type VARCHAR(50), -- full-time, part-time, contract
    status VARCHAR(20) DEFAULT 'active',
    view_count INTEGER DEFAULT 0,
    apply_count INTEGER DEFAULT 0,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建jobs表索引
CREATE INDEX IF NOT EXISTS idx_jobs_company_id ON jobs(company_id);
CREATE INDEX IF NOT EXISTS idx_jobs_created_by ON jobs(created_by);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_industry ON jobs(industry);
CREATE INDEX IF NOT EXISTS idx_jobs_location ON jobs(location);
CREATE INDEX IF NOT EXISTS idx_jobs_job_type ON jobs(job_type);

-- 创建外键约束
ALTER TABLE jobs ADD CONSTRAINT fk_jobs_company_id 
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE;
ALTER TABLE jobs ADD CONSTRAINT fk_jobs_created_by 
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE;

-- 添加表注释
COMMENT ON TABLE users IS '用户表';
COMMENT ON TABLE jobs IS '职位表';
