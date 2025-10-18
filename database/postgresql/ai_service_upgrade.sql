-- JobFirst PostgreSQL AI服务升级脚本
-- 版本: V3.0 -> V4.0
-- 日期: 2025年1月6日
-- 描述: 为PostgreSQL数据库添加AI服务和向量数据支持

-- ==============================================
-- 启用PostgreSQL扩展
-- ==============================================

-- 启用向量扩展 (pgvector)
CREATE EXTENSION IF NOT EXISTS vector;

-- 启用JSON扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 启用全文搜索扩展
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ==============================================
-- AI模型管理模块
-- ==============================================

-- 1. AI模型管理表
CREATE TABLE IF NOT EXISTS ai_models (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    version VARCHAR(20) NOT NULL,
    model_type VARCHAR(50) NOT NULL CHECK (model_type IN ('text_generation', 'embedding', 'classification', 'regression', 'recommendation')),
    provider VARCHAR(50) NOT NULL CHECK (provider IN ('openai', 'anthropic', 'google', 'azure', 'ollama', 'local', 'custom')),
    model_identifier VARCHAR(200) NOT NULL,
    description TEXT,
    parameters JSONB,
    performance_metrics JSONB,
    cost_per_token DECIMAL(10,8) DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. 模型版本管理表
CREATE TABLE IF NOT EXISTS model_versions (
    id BIGSERIAL PRIMARY KEY,
    model_id BIGINT NOT NULL REFERENCES ai_models(id) ON DELETE CASCADE,
    version VARCHAR(20) NOT NULL,
    model_path VARCHAR(500),
    config JSONB,
    training_data_hash VARCHAR(64),
    performance_score DECIMAL(5,4),
    is_production BOOLEAN DEFAULT FALSE,
    deployed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(model_id, version)
);

-- ==============================================
-- 企业AI分析模块
-- ==============================================

-- 3. 企业AI画像表
CREATE TABLE IF NOT EXISTS company_ai_profiles (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL,
    profile_type VARCHAR(50) NOT NULL CHECK (profile_type IN ('basic', 'detailed', 'competitive', 'culture', 'comprehensive')),
    profile_data JSONB NOT NULL,
    confidence_score DECIMAL(5,4),
    generated_by_model_id BIGINT REFERENCES ai_models(id) ON DELETE SET NULL,
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_valid BOOLEAN DEFAULT TRUE
);

-- 4. 企业嵌入向量表
CREATE TABLE IF NOT EXISTS company_embeddings (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL,
    embedding_type VARCHAR(50) NOT NULL CHECK (embedding_type IN ('description', 'culture', 'benefits', 'overall')),
    embedding_vector vector(1536), -- 使用pgvector扩展
    model_id BIGINT NOT NULL REFERENCES ai_models(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ==============================================
-- 职位AI分析模块
-- ==============================================

-- 5. 职位AI分析表
CREATE TABLE IF NOT EXISTS job_ai_analysis (
    id BIGSERIAL PRIMARY KEY,
    job_id INTEGER NOT NULL,
    analysis_type VARCHAR(50) NOT NULL CHECK (analysis_type IN ('description_enhancement', 'skill_extraction', 'salary_prediction', 'match_score', 'comprehensive')),
    analysis_result JSONB NOT NULL,
    confidence_score DECIMAL(5,4),
    generated_by_model_id BIGINT REFERENCES ai_models(id) ON DELETE SET NULL,
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_valid BOOLEAN DEFAULT TRUE
);

-- 6. 职位嵌入向量表
CREATE TABLE IF NOT EXISTS job_embeddings (
    id BIGSERIAL PRIMARY KEY,
    job_id INTEGER NOT NULL,
    embedding_type VARCHAR(50) NOT NULL CHECK (embedding_type IN ('title', 'description', 'requirements', 'overall')),
    embedding_vector vector(1536), -- 使用pgvector扩展
    model_id BIGINT NOT NULL REFERENCES ai_models(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ==============================================
-- 用户AI画像模块
-- ==============================================

-- 7. 用户AI画像表
CREATE TABLE IF NOT EXISTS user_ai_profiles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    profile_type VARCHAR(50) NOT NULL CHECK (profile_type IN ('basic', 'career', 'skills', 'preferences', 'comprehensive')),
    profile_data JSONB NOT NULL,
    confidence_score DECIMAL(5,4),
    generated_by_model_id BIGINT REFERENCES ai_models(id) ON DELETE SET NULL,
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_valid BOOLEAN DEFAULT TRUE
);

-- 8. 用户嵌入向量表
CREATE TABLE IF NOT EXISTS user_embeddings (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    embedding_type VARCHAR(50) NOT NULL CHECK (embedding_type IN ('resume', 'skills', 'experience', 'preferences', 'overall')),
    embedding_vector vector(1536), -- 使用pgvector扩展
    model_id BIGINT NOT NULL REFERENCES ai_models(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ==============================================
-- 智能推荐模块
-- ==============================================

-- 9. 职位推荐表
CREATE TABLE IF NOT EXISTS job_recommendations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    job_id INTEGER NOT NULL,
    recommendation_score DECIMAL(5,4) NOT NULL,
    recommendation_reasons JSONB,
    match_factors JSONB,
    generated_by_model_id BIGINT REFERENCES ai_models(id) ON DELETE SET NULL,
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    user_interaction VARCHAR(20) CHECK (user_interaction IN ('viewed', 'applied', 'saved', 'dismissed')),
    interaction_at TIMESTAMP WITH TIME ZONE
);

-- 10. 企业推荐表
CREATE TABLE IF NOT EXISTS company_recommendations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    company_id BIGINT NOT NULL,
    recommendation_score DECIMAL(5,4) NOT NULL,
    recommendation_reasons JSONB,
    match_factors JSONB,
    generated_by_model_id BIGINT REFERENCES ai_models(id) ON DELETE SET NULL,
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    user_interaction VARCHAR(20) CHECK (user_interaction IN ('viewed', 'followed', 'applied', 'dismissed')),
    interaction_at TIMESTAMP WITH TIME ZONE
);

-- ==============================================
-- AI对话模块
-- ==============================================

-- 11. AI对话会话表
CREATE TABLE IF NOT EXISTS ai_conversations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    conversation_type VARCHAR(50) NOT NULL CHECK (conversation_type IN ('career_advice', 'resume_review', 'interview_prep', 'skill_analysis', 'general')),
    session_id VARCHAR(100) NOT NULL,
    context_data JSONB,
    model_id BIGINT NOT NULL REFERENCES ai_models(id) ON DELETE CASCADE,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_activity_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);

-- 12. AI对话消息表
CREATE TABLE IF NOT EXISTS ai_messages (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES ai_conversations(id) ON DELETE CASCADE,
    message_type VARCHAR(20) NOT NULL CHECK (message_type IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    metadata JSONB,
    tokens_used INTEGER,
    processing_time_ms INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ==============================================
-- 向量存储模块
-- ==============================================

-- 13. 简历向量表 (扩展现有表)
CREATE TABLE IF NOT EXISTS resume_vectors (
    id BIGSERIAL PRIMARY KEY,
    resume_id BIGINT NOT NULL,
    embedding_type VARCHAR(50) NOT NULL CHECK (embedding_type IN ('content', 'skills', 'experience', 'overall')),
    embedding_vector vector(1536), -- 使用pgvector扩展
    model_id BIGINT NOT NULL REFERENCES ai_models(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 14. 技能嵌入向量表
CREATE TABLE IF NOT EXISTS skill_embeddings (
    id BIGSERIAL PRIMARY KEY,
    skill_id BIGINT NOT NULL,
    skill_name VARCHAR(100) NOT NULL,
    embedding_vector vector(1536), -- 使用pgvector扩展
    model_id BIGINT NOT NULL REFERENCES ai_models(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 15. 公司向量表
CREATE TABLE IF NOT EXISTS company_vectors (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL,
    company_name VARCHAR(200) NOT NULL,
    embedding_vector vector(1536), -- 使用pgvector扩展
    model_id BIGINT NOT NULL REFERENCES ai_models(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ==============================================
-- 索引优化
-- ==============================================

-- AI模型索引
CREATE INDEX IF NOT EXISTS idx_ai_models_name ON ai_models(name);
CREATE INDEX IF NOT EXISTS idx_ai_models_type ON ai_models(model_type);
CREATE INDEX IF NOT EXISTS idx_ai_models_provider ON ai_models(provider);
CREATE INDEX IF NOT EXISTS idx_ai_models_active ON ai_models(is_active);

-- 模型版本索引
CREATE INDEX IF NOT EXISTS idx_model_versions_model_id ON model_versions(model_id);
CREATE INDEX IF NOT EXISTS idx_model_versions_production ON model_versions(is_production);

-- 企业AI分析索引
CREATE INDEX IF NOT EXISTS idx_company_ai_profiles_company_id ON company_ai_profiles(company_id);
CREATE INDEX IF NOT EXISTS idx_company_ai_profiles_type ON company_ai_profiles(profile_type);
CREATE INDEX IF NOT EXISTS idx_company_ai_profiles_generated_at ON company_ai_profiles(generated_at);
CREATE INDEX IF NOT EXISTS idx_company_ai_profiles_valid ON company_ai_profiles(is_valid);

-- 企业嵌入向量索引
CREATE INDEX IF NOT EXISTS idx_company_embeddings_company_id ON company_embeddings(company_id);
CREATE INDEX IF NOT EXISTS idx_company_embeddings_type ON company_embeddings(embedding_type);
CREATE INDEX IF NOT EXISTS idx_company_embeddings_model_id ON company_embeddings(model_id);

-- 职位AI分析索引
CREATE INDEX IF NOT EXISTS idx_job_ai_analysis_job_id ON job_ai_analysis(job_id);
CREATE INDEX IF NOT EXISTS idx_job_ai_analysis_type ON job_ai_analysis(analysis_type);
CREATE INDEX IF NOT EXISTS idx_job_ai_analysis_generated_at ON job_ai_analysis(generated_at);
CREATE INDEX IF NOT EXISTS idx_job_ai_analysis_valid ON job_ai_analysis(is_valid);

-- 职位嵌入向量索引
CREATE INDEX IF NOT EXISTS idx_job_embeddings_job_id ON job_embeddings(job_id);
CREATE INDEX IF NOT EXISTS idx_job_embeddings_type ON job_embeddings(embedding_type);
CREATE INDEX IF NOT EXISTS idx_job_embeddings_model_id ON job_embeddings(model_id);

-- 用户AI画像索引
CREATE INDEX IF NOT EXISTS idx_user_ai_profiles_user_id ON user_ai_profiles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_ai_profiles_type ON user_ai_profiles(profile_type);
CREATE INDEX IF NOT EXISTS idx_user_ai_profiles_generated_at ON user_ai_profiles(generated_at);
CREATE INDEX IF NOT EXISTS idx_user_ai_profiles_valid ON user_ai_profiles(is_valid);

-- 用户嵌入向量索引
CREATE INDEX IF NOT EXISTS idx_user_embeddings_user_id ON user_embeddings(user_id);
CREATE INDEX IF NOT EXISTS idx_user_embeddings_type ON user_embeddings(embedding_type);
CREATE INDEX IF NOT EXISTS idx_user_embeddings_model_id ON user_embeddings(model_id);

-- 推荐表索引
CREATE INDEX IF NOT EXISTS idx_job_recommendations_user_id ON job_recommendations(user_id);
CREATE INDEX IF NOT EXISTS idx_job_recommendations_job_id ON job_recommendations(job_id);
CREATE INDEX IF NOT EXISTS idx_job_recommendations_score ON job_recommendations(recommendation_score);
CREATE INDEX IF NOT EXISTS idx_job_recommendations_generated_at ON job_recommendations(generated_at);
CREATE INDEX IF NOT EXISTS idx_job_recommendations_active ON job_recommendations(is_active);

CREATE INDEX IF NOT EXISTS idx_company_recommendations_user_id ON company_recommendations(user_id);
CREATE INDEX IF NOT EXISTS idx_company_recommendations_company_id ON company_recommendations(company_id);
CREATE INDEX IF NOT EXISTS idx_company_recommendations_score ON company_recommendations(recommendation_score);
CREATE INDEX IF NOT EXISTS idx_company_recommendations_generated_at ON company_recommendations(generated_at);
CREATE INDEX IF NOT EXISTS idx_company_recommendations_active ON company_recommendations(is_active);

-- 对话表索引
CREATE INDEX IF NOT EXISTS idx_ai_conversations_user_id ON ai_conversations(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_conversations_type ON ai_conversations(conversation_type);
CREATE INDEX IF NOT EXISTS idx_ai_conversations_session_id ON ai_conversations(session_id);
CREATE INDEX IF NOT EXISTS idx_ai_conversations_model_id ON ai_conversations(model_id);
CREATE INDEX IF NOT EXISTS idx_ai_conversations_active ON ai_conversations(is_active);

CREATE INDEX IF NOT EXISTS idx_ai_messages_conversation_id ON ai_messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_ai_messages_type ON ai_messages(message_type);
CREATE INDEX IF NOT EXISTS idx_ai_messages_created_at ON ai_messages(created_at);

-- 向量表索引
CREATE INDEX IF NOT EXISTS idx_resume_vectors_resume_id ON resume_vectors(resume_id);
CREATE INDEX IF NOT EXISTS idx_resume_vectors_type ON resume_vectors(embedding_type);
CREATE INDEX IF NOT EXISTS idx_resume_vectors_model_id ON resume_vectors(model_id);

CREATE INDEX IF NOT EXISTS idx_skill_embeddings_skill_id ON skill_embeddings(skill_id);
CREATE INDEX IF NOT EXISTS idx_skill_embeddings_name ON skill_embeddings(skill_name);
CREATE INDEX IF NOT EXISTS idx_skill_embeddings_model_id ON skill_embeddings(model_id);

CREATE INDEX IF NOT EXISTS idx_company_vectors_company_id ON company_vectors(company_id);
CREATE INDEX IF NOT EXISTS idx_company_vectors_name ON company_vectors(company_name);
CREATE INDEX IF NOT EXISTS idx_company_vectors_model_id ON company_vectors(model_id);

-- ==============================================
-- 向量相似度搜索索引
-- ==============================================

-- 使用pgvector的HNSW索引进行高效向量搜索
CREATE INDEX IF NOT EXISTS idx_company_embeddings_vector_hnsw ON company_embeddings 
USING hnsw (embedding_vector vector_cosine_ops) WITH (m = 16, ef_construction = 64);

CREATE INDEX IF NOT EXISTS idx_job_embeddings_vector_hnsw ON job_embeddings 
USING hnsw (embedding_vector vector_cosine_ops) WITH (m = 16, ef_construction = 64);

CREATE INDEX IF NOT EXISTS idx_user_embeddings_vector_hnsw ON user_embeddings 
USING hnsw (embedding_vector vector_cosine_ops) WITH (m = 16, ef_construction = 64);

CREATE INDEX IF NOT EXISTS idx_resume_vectors_vector_hnsw ON resume_vectors 
USING hnsw (embedding_vector vector_cosine_ops) WITH (m = 16, ef_construction = 64);

CREATE INDEX IF NOT EXISTS idx_skill_embeddings_vector_hnsw ON skill_embeddings 
USING hnsw (embedding_vector vector_cosine_ops) WITH (m = 16, ef_construction = 64);

CREATE INDEX IF NOT EXISTS idx_company_vectors_vector_hnsw ON company_vectors 
USING hnsw (embedding_vector vector_cosine_ops) WITH (m = 16, ef_construction = 64);

-- ==============================================
-- 触发器函数
-- ==============================================

-- 更新时间戳触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为需要更新时间戳的表添加触发器
CREATE TRIGGER update_ai_models_updated_at BEFORE UPDATE ON ai_models
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_resume_vectors_updated_at BEFORE UPDATE ON resume_vectors
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ==============================================
-- 初始化AI模型数据
-- ==============================================

-- 插入基础AI模型
INSERT INTO ai_models (name, version, model_type, provider, model_identifier, description, cost_per_token, is_active) VALUES
('gemma3-4b', '1.0', 'text_generation', 'ollama', 'gemma3:4b', 'Google Gemma 3 4B模型，用于文本生成和对话', 0.000001, TRUE),
('text-embedding-ada-002', '1.0', 'embedding', 'openai', 'text-embedding-ada-002', 'OpenAI文本嵌入模型，用于向量化', 0.0001, TRUE),
('gpt-3.5-turbo', '1.0', 'text_generation', 'openai', 'gpt-3.5-turbo', 'OpenAI GPT-3.5模型，用于智能对话', 0.002, TRUE),
('claude-3-haiku', '1.0', 'text_generation', 'anthropic', 'claude-3-haiku', 'Anthropic Claude 3 Haiku模型，用于快速文本生成', 0.00025, TRUE),
('gpt-4', '1.0', 'text_generation', 'openai', 'gpt-4', 'OpenAI GPT-4模型，用于高质量文本生成', 0.03, TRUE),
('claude-3-sonnet', '1.0', 'text_generation', 'anthropic', 'claude-3-sonnet', 'Anthropic Claude 3 Sonnet模型，用于平衡性能和成本', 0.003, TRUE)
ON CONFLICT (name) DO NOTHING;

-- 插入模型版本
INSERT INTO model_versions (model_id, version, config, performance_score, is_production) VALUES
((SELECT id FROM ai_models WHERE name = 'gemma3-4b'), '1.0', '{"temperature": 0.3, "top_p": 0.9, "max_tokens": 1000}', 0.85, TRUE),
((SELECT id FROM ai_models WHERE name = 'text-embedding-ada-002'), '1.0', '{"dimensions": 1536, "encoding_format": "float"}', 0.92, TRUE),
((SELECT id FROM ai_models WHERE name = 'gpt-3.5-turbo'), '1.0', '{"temperature": 0.7, "top_p": 1.0, "max_tokens": 2000}', 0.88, TRUE),
((SELECT id FROM ai_models WHERE name = 'claude-3-haiku'), '1.0', '{"temperature": 0.5, "top_p": 0.95, "max_tokens": 1500}', 0.87, TRUE),
((SELECT id FROM ai_models WHERE name = 'gpt-4'), '1.0', '{"temperature": 0.7, "top_p": 1.0, "max_tokens": 4000}', 0.95, FALSE),
((SELECT id FROM ai_models WHERE name = 'claude-3-sonnet'), '1.0', '{"temperature": 0.6, "top_p": 0.98, "max_tokens": 3000}', 0.93, FALSE)
ON CONFLICT (model_id, version) DO NOTHING;

-- ==============================================
-- 向量相似度搜索函数
-- ==============================================

-- 简历相似度搜索函数
CREATE OR REPLACE FUNCTION search_similar_resumes(
    query_vector vector(1536),
    similarity_threshold float DEFAULT 0.7,
    limit_count integer DEFAULT 10
)
RETURNS TABLE (
    resume_id bigint,
    similarity_score float,
    embedding_type varchar
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        rv.resume_id,
        1 - (rv.embedding_vector <=> query_vector) as similarity_score,
        rv.embedding_type
    FROM resume_vectors rv
    WHERE 1 - (rv.embedding_vector <=> query_vector) > similarity_threshold
    ORDER BY rv.embedding_vector <=> query_vector
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- 职位相似度搜索函数
CREATE OR REPLACE FUNCTION search_similar_jobs(
    query_vector vector(1536),
    similarity_threshold float DEFAULT 0.7,
    limit_count integer DEFAULT 10
)
RETURNS TABLE (
    job_id integer,
    similarity_score float,
    embedding_type varchar
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        je.job_id,
        1 - (je.embedding_vector <=> query_vector) as similarity_score,
        je.embedding_type
    FROM job_embeddings je
    WHERE 1 - (je.embedding_vector <=> query_vector) > similarity_threshold
    ORDER BY je.embedding_vector <=> query_vector
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- 企业相似度搜索函数
CREATE OR REPLACE FUNCTION search_similar_companies(
    query_vector vector(1536),
    similarity_threshold float DEFAULT 0.7,
    limit_count integer DEFAULT 10
)
RETURNS TABLE (
    company_id bigint,
    similarity_score float,
    embedding_type varchar
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ce.company_id,
        1 - (ce.embedding_vector <=> query_vector) as similarity_score,
        ce.embedding_type
    FROM company_embeddings ce
    WHERE 1 - (ce.embedding_vector <=> query_vector) > similarity_threshold
    ORDER BY ce.embedding_vector <=> query_vector
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- ==============================================
-- 升级完成
-- ==============================================

-- 记录升级日志
INSERT INTO ai_conversations (user_id, conversation_type, session_id, model_id, context_data) VALUES
(0, 'general', 'system_upgrade_' || extract(epoch from now())::text, 
 (SELECT id FROM ai_models WHERE name = 'gemma3-4b' LIMIT 1),
 '{"upgrade": "PostgreSQL AI服务升级完成", "version": "V3.0->V4.0", "tables_created": 15}');

-- 显示升级完成信息
SELECT 'PostgreSQL AI服务升级完成！' as message,
       'V3.0 -> V4.0' as version,
       NOW() as upgrade_time,
       '新增15个表，支持AI服务、向量存储、智能推荐' as description;
