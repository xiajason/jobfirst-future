-- SQLite用户数据库架构创建脚本
-- 用途：为每个用户创建独立的SQLite数据库，存储简历内容和用户数据
-- 创建时间：2025-09-13
-- 版本：v1.0

-- 注意：此脚本为每个用户创建独立的SQLite数据库
-- 数据库路径：./data/users/{user_id}/resume.db

-- ==============================================
-- 简历内容表 - 存储实际的简历内容和用户数据
-- ==============================================
CREATE TABLE IF NOT EXISTS resume_content (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    resume_metadata_id INTEGER NOT NULL, -- 对应MySQL中的resume_metadata.id
    title TEXT NOT NULL,
    content TEXT, -- Markdown格式的简历内容
    raw_content TEXT, -- 原始文件内容（如果是上传的文件）
    content_hash TEXT, -- 内容哈希，用于去重和版本控制
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(resume_metadata_id) -- 确保一个元数据记录对应一个内容记录
);

-- ==============================================
-- 解析结果表 - 存储结构化的解析数据
-- ==============================================
CREATE TABLE IF NOT EXISTS parsed_resume_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    resume_content_id INTEGER NOT NULL,
    personal_info TEXT, -- JSON格式的个人信息
    work_experience TEXT, -- JSON格式的工作经历
    education TEXT, -- JSON格式的教育背景
    skills TEXT, -- JSON格式的技能列表
    projects TEXT, -- JSON格式的项目经验
    certifications TEXT, -- JSON格式的证书认证
    keywords TEXT, -- JSON格式的关键词
    confidence REAL, -- 解析置信度 0-1
    parsing_version TEXT, -- 解析器版本
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (resume_content_id) REFERENCES resume_content(id) ON DELETE CASCADE
);

-- ==============================================
-- 用户隐私设置表 - 详细的隐私控制
-- ==============================================
CREATE TABLE IF NOT EXISTS user_privacy_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    resume_content_id INTEGER NOT NULL,
    is_public BOOLEAN DEFAULT FALSE, -- 是否公开
    share_with_companies BOOLEAN DEFAULT FALSE, -- 是否允许公司查看
    allow_search BOOLEAN DEFAULT TRUE, -- 是否允许被搜索
    allow_download BOOLEAN DEFAULT FALSE, -- 是否允许下载
    view_permissions TEXT, -- JSON格式的查看权限设置
    download_permissions TEXT, -- JSON格式的下载权限设置
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (resume_content_id) REFERENCES resume_content(id) ON DELETE CASCADE,
    UNIQUE(resume_content_id) -- 每个简历内容对应一个隐私设置
);

-- ==============================================
-- 简历版本历史表 - 版本管理和历史追踪
-- ==============================================
CREATE TABLE IF NOT EXISTS resume_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    resume_content_id INTEGER NOT NULL,
    version_number INTEGER NOT NULL,
    content_snapshot TEXT, -- 版本快照
    change_description TEXT, -- 变更描述
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (resume_content_id) REFERENCES resume_content(id) ON DELETE CASCADE,
    UNIQUE(resume_content_id, version_number) -- 确保版本号唯一
);

-- ==============================================
-- 用户自定义字段表 - 扩展字段支持
-- ==============================================
CREATE TABLE IF NOT EXISTS user_custom_fields (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    resume_content_id INTEGER NOT NULL,
    field_name TEXT NOT NULL,
    field_value TEXT,
    field_type TEXT DEFAULT 'text', -- text, number, date, json
    is_public BOOLEAN DEFAULT FALSE, -- 是否在公开简历中显示
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (resume_content_id) REFERENCES resume_content(id) ON DELETE CASCADE,
    UNIQUE(resume_content_id, field_name) -- 每个简历的字段名唯一
);

-- ==============================================
-- 简历访问日志表 - 访问追踪和统计
-- ==============================================
CREATE TABLE IF NOT EXISTS resume_access_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    resume_content_id INTEGER NOT NULL,
    access_type TEXT NOT NULL, -- view, download, share, edit
    access_source TEXT, -- 访问来源
    user_agent TEXT, -- 用户代理
    ip_address TEXT, -- IP地址（脱敏后）
    access_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (resume_content_id) REFERENCES resume_content(id) ON DELETE CASCADE
);

-- ==============================================
-- 创建索引优化查询性能
-- ==============================================

-- resume_content表索引
CREATE INDEX IF NOT EXISTS idx_resume_content_metadata_id ON resume_content(resume_metadata_id);
CREATE INDEX IF NOT EXISTS idx_resume_content_updated_at ON resume_content(updated_at);

-- parsed_resume_data表索引
CREATE INDEX IF NOT EXISTS idx_parsed_data_content_id ON parsed_resume_data(resume_content_id);
CREATE INDEX IF NOT EXISTS idx_parsed_data_confidence ON parsed_resume_data(confidence);

-- user_privacy_settings表索引
CREATE INDEX IF NOT EXISTS idx_privacy_content_id ON user_privacy_settings(resume_content_id);
CREATE INDEX IF NOT EXISTS idx_privacy_is_public ON user_privacy_settings(is_public);

-- resume_versions表索引
CREATE INDEX IF NOT EXISTS idx_versions_content_id ON resume_versions(resume_content_id);
CREATE INDEX IF NOT EXISTS idx_versions_number ON resume_versions(version_number);

-- user_custom_fields表索引
CREATE INDEX IF NOT EXISTS idx_custom_fields_content_id ON user_custom_fields(resume_content_id);
CREATE INDEX IF NOT EXISTS idx_custom_fields_name ON user_custom_fields(field_name);

-- resume_access_logs表索引
CREATE INDEX IF NOT EXISTS idx_access_logs_content_id ON resume_access_logs(resume_content_id);
CREATE INDEX IF NOT EXISTS idx_access_logs_type ON resume_access_logs(access_type);
CREATE INDEX IF NOT EXISTS idx_access_logs_time ON resume_access_logs(access_time);

-- ==============================================
-- 创建触发器自动更新时间戳
-- ==============================================

-- resume_content表更新时间触发器
CREATE TRIGGER IF NOT EXISTS update_resume_content_timestamp 
    AFTER UPDATE ON resume_content
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE resume_content SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- parsed_resume_data表更新时间触发器
CREATE TRIGGER IF NOT EXISTS update_parsed_data_timestamp 
    AFTER UPDATE ON parsed_resume_data
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE parsed_resume_data SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- user_privacy_settings表更新时间触发器
CREATE TRIGGER IF NOT EXISTS update_privacy_timestamp 
    AFTER UPDATE ON user_privacy_settings
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE user_privacy_settings SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- user_custom_fields表更新时间触发器
CREATE TRIGGER IF NOT EXISTS update_custom_fields_timestamp 
    AFTER UPDATE ON user_custom_fields
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE user_custom_fields SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- ==============================================
-- 插入默认数据（可选）
-- ==============================================

-- 注意：实际使用时，这些数据会在代码中动态创建
-- 这里只是为了演示表结构的完整性

SELECT 'SQLite用户数据库架构创建完成！' as message;
SELECT '数据库路径：./data/users/{user_id}/resume.db' as db_path;
SELECT '下一步：创建数据迁移脚本' as next_step;
