-- JobFirst AI服务数据库初始化脚本

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "vector";

-- 创建AI对话表
CREATE TABLE IF NOT EXISTS ai_conversations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(100) NOT NULL,
    session_id VARCHAR(100) NOT NULL,
    title VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_session_id (session_id),
    INDEX idx_created_at (created_at)
);

-- 创建AI消息表
CREATE TABLE IF NOT EXISTS ai_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID NOT NULL,
    role ENUM('user', 'assistant', 'system') NOT NULL,
    content TEXT NOT NULL,
    embedding vector(1536),
    tokens INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES ai_conversations(id) ON DELETE CASCADE,
    INDEX idx_conversation_id (conversation_id),
    INDEX idx_role (role),
    INDEX idx_created_at (created_at)
);

-- 创建AI知识库表
CREATE TABLE IF NOT EXISTS ai_knowledge_base (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    embedding vector(1536),
    source VARCHAR(255),
    category VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_title (title),
    INDEX idx_category (category),
    INDEX idx_created_at (created_at)
);

-- 创建AI模型配置表
CREATE TABLE IF NOT EXISTS ai_model_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    model_name VARCHAR(100) NOT NULL,
    model_type VARCHAR(50) NOT NULL,
    config JSONB NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_model_name (model_name),
    INDEX idx_model_type (model_type),
    INDEX idx_is_active (is_active)
);

-- 创建AI使用统计表
CREATE TABLE IF NOT EXISTS ai_usage_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(100) NOT NULL,
    model_name VARCHAR(100) NOT NULL,
    request_count INTEGER DEFAULT 0,
    token_count INTEGER DEFAULT 0,
    cost DECIMAL(10, 4) DEFAULT 0.0,
    date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, model_name, date),
    INDEX idx_user_id (user_id),
    INDEX idx_model_name (model_name),
    INDEX idx_date (date)
);

-- 插入默认AI模型配置
INSERT INTO ai_model_configs (model_name, model_type, config, is_active) VALUES
('gemma-2-2b', 'text-generation', '{"temperature": 0.7, "max_tokens": 2048, "top_p": 0.9}', true),
('gemma-2-9b', 'text-generation', '{"temperature": 0.7, "max_tokens": 4096, "top_p": 0.9}', true),
('text-embedding-ada-002', 'embedding', '{"dimensions": 1536}', true);

-- 创建触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 创建触发器
CREATE TRIGGER update_ai_conversations_updated_at BEFORE UPDATE ON ai_conversations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ai_knowledge_base_updated_at BEFORE UPDATE ON ai_knowledge_base FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ai_model_configs_updated_at BEFORE UPDATE ON ai_model_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ai_usage_stats_updated_at BEFORE UPDATE ON ai_usage_stats FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 创建索引优化
CREATE INDEX idx_ai_messages_embedding ON ai_messages USING ivfflat (embedding vector_cosine_ops);
CREATE INDEX idx_ai_knowledge_base_embedding ON ai_knowledge_base USING ivfflat (embedding vector_cosine_ops);
