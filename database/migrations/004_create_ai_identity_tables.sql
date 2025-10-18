-- AI身份数据模型数据库表结构
-- 创建时间: 2025年10月3日
-- 用途: 支持AI身份数据模型集成，包括档案管理、向量存储、相似度计算等

-- AI身份档案表
CREATE TABLE IF NOT EXISTS ai_identity_profiles (
    id INT PRIMARY KEY AUTO_INCREMENT,
    profile_id VARCHAR(100) NOT NULL UNIQUE COMMENT '档案ID',
    user_id INT NOT NULL COMMENT '用户ID',
    identity_type ENUM('rational', 'emotional', 'integrated') NOT NULL DEFAULT 'rational' COMMENT 'AI身份类型',
    
    -- 基础信息
    personal_info JSON COMMENT '个人信息',
    education_background JSON COMMENT '教育背景',
    
    -- 核心评分
    overall_skill_score FLOAT DEFAULT 0.0 COMMENT '整体技能评分',
    overall_experience_score FLOAT DEFAULT 0.0 COMMENT '整体经验评分',
    overall_competency_score FLOAT DEFAULT 0.0 COMMENT '整体能力评分',
    comprehensive_score FLOAT DEFAULT 0.0 COMMENT '综合评分',
    
    -- 数据质量
    data_completeness FLOAT DEFAULT 0.0 COMMENT '数据完整性',
    data_sources JSON COMMENT '数据源列表',
    
    -- 向量化相关
    vector_dimension INT NULL COMMENT '向量维度',
    vector_model VARCHAR(100) NULL COMMENT '向量模型名称',
    has_vector BOOLEAN DEFAULT FALSE COMMENT '是否已向量化',
    
    -- 元数据
    version VARCHAR(20) DEFAULT '1.0.0' COMMENT '版本号',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_profile_id (profile_id),
    INDEX idx_user_id (user_id),
    INDEX idx_identity_type (identity_type),
    INDEX idx_comprehensive_score (comprehensive_score),
    INDEX idx_data_completeness (data_completeness),
    INDEX idx_has_vector (has_vector),
    INDEX idx_is_active (is_active),
    INDEX idx_created_at (created_at),
    INDEX idx_updated_at (updated_at)
    -- FULLTEXT INDEX idx_personal_info (personal_info) -- JSON列不支持全文索引
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI身份档案表';

-- AI身份向量表
CREATE TABLE IF NOT EXISTS ai_identity_vectors (
    id INT PRIMARY KEY AUTO_INCREMENT,
    profile_id VARCHAR(100) NOT NULL COMMENT '档案ID',
    vector_type ENUM('comprehensive', 'skill', 'experience', 'competency') NOT NULL DEFAULT 'comprehensive' COMMENT '向量类型',
    
    -- 向量数据
    vector_embedding LONGTEXT NOT NULL COMMENT '向量嵌入数据(JSON格式)',
    vector_dimension INT NOT NULL COMMENT '向量维度',
    vector_model VARCHAR(100) NOT NULL COMMENT '向量模型名称',
    
    -- 向量元数据
    embedding_source TEXT COMMENT '向量生成源数据',
    embedding_algorithm VARCHAR(100) COMMENT '嵌入算法',
    embedding_parameters JSON COMMENT '嵌入参数',
    
    -- 性能指标
    embedding_time_ms INT COMMENT '嵌入生成时间(毫秒)',
    vector_norm FLOAT COMMENT '向量范数',
    vector_magnitude FLOAT COMMENT '向量幅度',
    
    -- 索引相关
    index_type VARCHAR(50) DEFAULT 'flat' COMMENT '索引类型',
    index_id VARCHAR(100) NULL COMMENT '外部索引ID',
    
    -- 元数据
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (profile_id) REFERENCES ai_identity_profiles(profile_id) ON DELETE CASCADE,
    INDEX idx_profile_id (profile_id),
    INDEX idx_vector_type (vector_type),
    INDEX idx_vector_model (vector_model),
    INDEX idx_vector_dimension (vector_dimension),
    INDEX idx_is_active (is_active),
    INDEX idx_created_at (created_at),
    UNIQUE KEY uk_profile_vector_type (profile_id, vector_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI身份向量表';

-- AI身份相似度表
CREATE TABLE IF NOT EXISTS ai_identity_similarities (
    id INT PRIMARY KEY AUTO_INCREMENT,
    source_profile_id VARCHAR(100) NOT NULL COMMENT '源档案ID',
    target_profile_id VARCHAR(100) NOT NULL COMMENT '目标档案ID',
    similarity_type ENUM('comprehensive', 'skill', 'experience', 'competency', 'custom') NOT NULL DEFAULT 'comprehensive' COMMENT '相似度类型',
    
    -- 相似度指标
    cosine_similarity FLOAT NOT NULL COMMENT '余弦相似度',
    euclidean_distance FLOAT NOT NULL COMMENT '欧几里得距离',
    manhattan_distance FLOAT NOT NULL COMMENT '曼哈顿距离',
    pearson_correlation FLOAT NULL COMMENT '皮尔逊相关系数',
    
    -- 综合相似度评分
    overall_similarity_score FLOAT NOT NULL COMMENT '综合相似度评分',
    similarity_rank INT NULL COMMENT '相似度排名',
    
    -- 相似度详情
    similarity_components JSON COMMENT '相似度组件详情',
    matching_features JSON COMMENT '匹配特征',
    similarity_explanation TEXT COMMENT '相似度解释',
    
    -- 计算元数据
    calculation_algorithm VARCHAR(100) COMMENT '计算算法',
    calculation_parameters JSON COMMENT '计算参数',
    calculation_time_ms INT COMMENT '计算时间(毫秒)',
    
    -- 元数据
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (source_profile_id) REFERENCES ai_identity_profiles(profile_id) ON DELETE CASCADE,
    FOREIGN KEY (target_profile_id) REFERENCES ai_identity_profiles(profile_id) ON DELETE CASCADE,
    INDEX idx_source_profile_id (source_profile_id),
    INDEX idx_target_profile_id (target_profile_id),
    INDEX idx_similarity_type (similarity_type),
    INDEX idx_overall_similarity_score (overall_similarity_score),
    INDEX idx_similarity_rank (similarity_rank),
    INDEX idx_is_active (is_active),
    INDEX idx_created_at (created_at),
    UNIQUE KEY uk_profiles_similarity_type (source_profile_id, target_profile_id, similarity_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI身份相似度表';

-- AI身份元数据表
CREATE TABLE IF NOT EXISTS ai_identity_metadata (
    id INT PRIMARY KEY AUTO_INCREMENT,
    profile_id VARCHAR(100) NOT NULL COMMENT '档案ID',
    metadata_type ENUM('profile', 'vector', 'similarity', 'performance', 'usage', 'system') NOT NULL COMMENT '元数据类型',
    
    -- 元数据内容
    metadata_key VARCHAR(100) NOT NULL COMMENT '元数据键',
    metadata_value LONGTEXT COMMENT '元数据值',
    metadata_json JSON COMMENT 'JSON格式元数据',
    
    -- 数据类型
    value_type ENUM('string', 'number', 'boolean', 'json', 'array', 'object') DEFAULT 'string' COMMENT '值类型',
    data_size INT COMMENT '数据大小(字节)',
    
    -- 元数据属性
    is_indexed BOOLEAN DEFAULT FALSE COMMENT '是否已建立索引',
    is_cached BOOLEAN DEFAULT FALSE COMMENT '是否已缓存',
    cache_ttl INT NULL COMMENT '缓存TTL(秒)',
    
    -- 版本控制
    version VARCHAR(20) DEFAULT '1.0.0' COMMENT '元数据版本',
    parent_id INT NULL COMMENT '父元数据ID',
    
    -- 元数据
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (profile_id) REFERENCES ai_identity_profiles(profile_id) ON DELETE CASCADE,
    INDEX idx_profile_id (profile_id),
    INDEX idx_metadata_type (metadata_type),
    INDEX idx_metadata_key (metadata_key),
    INDEX idx_value_type (value_type),
    INDEX idx_is_indexed (is_indexed),
    INDEX idx_is_cached (is_cached),
    INDEX idx_is_active (is_active),
    INDEX idx_created_at (created_at),
    UNIQUE KEY uk_profile_type_key (profile_id, metadata_type, metadata_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI身份元数据表';

-- AI身份性能指标表
CREATE TABLE IF NOT EXISTS ai_identity_performance_metrics (
    id INT PRIMARY KEY AUTO_INCREMENT,
    profile_id VARCHAR(100) NOT NULL COMMENT '档案ID',
    operation_type ENUM('create', 'update', 'vectorize', 'similarity', 'search', 'match') NOT NULL COMMENT '操作类型',
    
    -- 性能指标
    execution_time_ms INT NOT NULL COMMENT '执行时间(毫秒)',
    memory_usage_mb FLOAT COMMENT '内存使用量(MB)',
    cpu_usage_percent FLOAT COMMENT 'CPU使用率(%)',
    
    -- 数据量指标
    input_data_size INT COMMENT '输入数据大小(字节)',
    output_data_size INT COMMENT '输出数据大小(字节)',
    processed_records INT COMMENT '处理记录数',
    
    -- 质量指标
    success_rate FLOAT DEFAULT 1.0 COMMENT '成功率',
    accuracy_score FLOAT NULL COMMENT '准确率评分',
    precision_score FLOAT NULL COMMENT '精确率评分',
    recall_score FLOAT NULL COMMENT '召回率评分',
    f1_score FLOAT NULL COMMENT 'F1评分',
    
    -- 性能详情
    performance_details JSON COMMENT '性能详情',
    bottleneck_analysis JSON COMMENT '瓶颈分析',
    optimization_suggestions JSON COMMENT '优化建议',
    
    -- 环境信息
    system_info JSON COMMENT '系统信息',
    resource_usage JSON COMMENT '资源使用情况',
    
    -- 元数据
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (profile_id) REFERENCES ai_identity_profiles(profile_id) ON DELETE CASCADE,
    INDEX idx_profile_id (profile_id),
    INDEX idx_operation_type (operation_type),
    INDEX idx_execution_time_ms (execution_time_ms),
    INDEX idx_memory_usage_mb (memory_usage_mb),
    INDEX idx_success_rate (success_rate),
    INDEX idx_accuracy_score (accuracy_score),
    INDEX idx_is_active (is_active),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI身份性能指标表';

-- AI身份使用统计表
CREATE TABLE IF NOT EXISTS ai_identity_usage_stats (
    id INT PRIMARY KEY AUTO_INCREMENT,
    profile_id VARCHAR(100) NOT NULL COMMENT '档案ID',
    stat_date DATE NOT NULL COMMENT '统计日期',
    stat_type ENUM('daily', 'weekly', 'monthly', 'yearly') DEFAULT 'daily' COMMENT '统计类型',
    
    -- 使用统计
    view_count INT DEFAULT 0 COMMENT '查看次数',
    search_count INT DEFAULT 0 COMMENT '搜索次数',
    match_count INT DEFAULT 0 COMMENT '匹配次数',
    similarity_calculation_count INT DEFAULT 0 COMMENT '相似度计算次数',
    
    -- 性能统计
    avg_response_time_ms FLOAT COMMENT '平均响应时间(毫秒)',
    max_response_time_ms INT COMMENT '最大响应时间(毫秒)',
    min_response_time_ms INT COMMENT '最小响应时间(毫秒)',
    
    -- 质量统计
    avg_similarity_score FLOAT COMMENT '平均相似度评分',
    max_similarity_score FLOAT COMMENT '最大相似度评分',
    min_similarity_score FLOAT COMMENT '最小相似度评分',
    
    -- 用户行为
    unique_users INT DEFAULT 0 COMMENT '独立用户数',
    returning_users INT DEFAULT 0 COMMENT '回访用户数',
    new_users INT DEFAULT 0 COMMENT '新用户数',
    
    -- 元数据
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (profile_id) REFERENCES ai_identity_profiles(profile_id) ON DELETE CASCADE,
    INDEX idx_profile_id (profile_id),
    INDEX idx_stat_date (stat_date),
    INDEX idx_stat_type (stat_type),
    INDEX idx_view_count (view_count),
    INDEX idx_search_count (search_count),
    INDEX idx_match_count (match_count),
    INDEX idx_is_active (is_active),
    UNIQUE KEY uk_profile_date_type (profile_id, stat_date, stat_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI身份使用统计表';

-- 创建索引优化
-- 为AI身份档案表创建复合索引
-- CREATE INDEX idx_profile_comprehensive ON ai_identity_profiles(identity_type, comprehensive_score, data_completeness, is_active); -- 重复索引
-- CREATE INDEX idx_profile_vector ON ai_identity_profiles(has_vector, vector_model, is_active); -- 重复索引

-- 为AI身份向量表创建复合索引
-- CREATE INDEX idx_vector_profile_type ON ai_identity_vectors(profile_id, vector_type, is_active); -- 重复索引
-- CREATE INDEX idx_vector_model_dim ON ai_identity_vectors(vector_model, vector_dimension, is_active); -- 重复索引

-- 为AI身份相似度表创建复合索引
-- CREATE INDEX idx_similarity_source_type ON ai_identity_similarities(source_profile_id, similarity_type, overall_similarity_score); -- 重复索引
-- CREATE INDEX idx_similarity_target_type ON ai_identity_similarities(target_profile_id, similarity_type, overall_similarity_score); -- 重复索引
-- CREATE INDEX idx_similarity_score_rank ON ai_identity_similarities(overall_similarity_score DESC, similarity_rank ASC); -- 重复索引

-- 为AI身份元数据表创建复合索引
-- CREATE INDEX idx_metadata_profile_type ON ai_identity_metadata(profile_id, metadata_type, is_active); -- 重复索引
-- CREATE INDEX idx_metadata_type_key ON ai_identity_metadata(metadata_type, metadata_key, is_active); -- 重复索引

-- 为AI身份性能指标表创建复合索引
-- CREATE INDEX idx_performance_profile_type ON ai_identity_performance_metrics(profile_id, operation_type, execution_time_ms); -- 重复索引
-- CREATE INDEX idx_performance_date ON ai_identity_performance_metrics(created_at, operation_type, execution_time_ms); -- 重复索引

-- 为AI身份使用统计表创建复合索引
-- CREATE INDEX idx_usage_profile_date ON ai_identity_usage_stats(profile_id, stat_date, stat_type); -- 重复索引
-- CREATE INDEX idx_usage_date_type ON ai_identity_usage_stats(stat_date, stat_type, view_count); -- 重复索引

-- 创建触发器：更新档案时自动更新updated_at
DELIMITER $$

CREATE TRIGGER ai_identity_profiles_update_trigger
    BEFORE UPDATE ON ai_identity_profiles
    FOR EACH ROW
BEGIN
    SET NEW.updated_at = CURRENT_TIMESTAMP;
END$$

CREATE TRIGGER ai_identity_vectors_update_trigger
    BEFORE UPDATE ON ai_identity_vectors
    FOR EACH ROW
BEGIN
    SET NEW.updated_at = CURRENT_TIMESTAMP;
END$$

CREATE TRIGGER ai_identity_similarities_update_trigger
    BEFORE UPDATE ON ai_identity_similarities
    FOR EACH ROW
BEGIN
    SET NEW.updated_at = CURRENT_TIMESTAMP;
END$$

CREATE TRIGGER ai_identity_metadata_update_trigger
    BEFORE UPDATE ON ai_identity_metadata
    FOR EACH ROW
BEGIN
    SET NEW.updated_at = CURRENT_TIMESTAMP;
END$$

CREATE TRIGGER ai_identity_performance_metrics_update_trigger
    BEFORE UPDATE ON ai_identity_performance_metrics
    FOR EACH ROW
BEGIN
    SET NEW.updated_at = CURRENT_TIMESTAMP;
END$$

CREATE TRIGGER ai_identity_usage_stats_update_trigger
    BEFORE UPDATE ON ai_identity_usage_stats
    FOR EACH ROW
BEGIN
    SET NEW.updated_at = CURRENT_TIMESTAMP;
END$$

DELIMITER ;

-- 创建视图：AI身份档案概览
CREATE VIEW ai_identity_profiles_overview AS
SELECT 
    p.id,
    p.profile_id,
    p.user_id,
    p.identity_type,
    p.comprehensive_score,
    p.data_completeness,
    p.has_vector,
    p.vector_model,
    p.version,
    p.is_active,
    p.created_at,
    p.updated_at,
    COUNT(v.id) as vector_count,
    COUNT(s.id) as similarity_count,
    COUNT(m.id) as metadata_count
FROM ai_identity_profiles p
LEFT JOIN ai_identity_vectors v ON p.profile_id = v.profile_id AND v.is_active = TRUE
LEFT JOIN ai_identity_similarities s ON p.profile_id = s.source_profile_id AND s.is_active = TRUE
LEFT JOIN ai_identity_metadata m ON p.profile_id = m.profile_id AND m.is_active = TRUE
WHERE p.is_active = TRUE
GROUP BY p.id, p.profile_id, p.user_id, p.identity_type, p.comprehensive_score, 
         p.data_completeness, p.has_vector, p.vector_model, p.version, 
         p.is_active, p.created_at, p.updated_at;

-- 创建视图：AI身份相似度排行榜
CREATE VIEW ai_identity_similarity_rankings AS
SELECT 
    s.id,
    s.source_profile_id,
    s.target_profile_id,
    s.similarity_type,
    s.overall_similarity_score,
    s.similarity_rank,
    s.cosine_similarity,
    s.euclidean_distance,
    s.created_at,
    sp.identity_type as source_identity_type,
    tp.identity_type as target_identity_type,
    sp.comprehensive_score as source_comprehensive_score,
    tp.comprehensive_score as target_comprehensive_score
FROM ai_identity_similarities s
JOIN ai_identity_profiles sp ON s.source_profile_id = sp.profile_id
JOIN ai_identity_profiles tp ON s.target_profile_id = tp.profile_id
WHERE s.is_active = TRUE 
  AND sp.is_active = TRUE 
  AND tp.is_active = TRUE
ORDER BY s.overall_similarity_score DESC, s.created_at DESC;

-- 插入初始数据
INSERT INTO ai_identity_metadata (profile_id, metadata_type, metadata_key, metadata_value, value_type, is_active) VALUES
('system', 'system', 'table_version', '1.0.0', 'string', TRUE),
('system', 'system', 'created_date', NOW(), 'string', TRUE),
('system', 'system', 'description', 'AI身份数据模型集成表结构', 'string', TRUE);

-- 创建存储过程：清理过期数据
DELIMITER $$

CREATE PROCEDURE CleanupExpiredAIIdentityData(IN days_to_keep INT)
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;
    
    START TRANSACTION;
    
    -- 清理过期的性能指标数据
    DELETE FROM ai_identity_performance_metrics 
    WHERE created_at < DATE_SUB(NOW(), INTERVAL days_to_keep DAY)
      AND is_active = FALSE;
    
    -- 清理过期的使用统计数据
    DELETE FROM ai_identity_usage_stats 
    WHERE stat_date < DATE_SUB(NOW(), INTERVAL days_to_keep DAY)
      AND is_active = FALSE;
    
    -- 清理过期的元数据
    DELETE FROM ai_identity_metadata 
    WHERE created_at < DATE_SUB(NOW(), INTERVAL days_to_keep DAY)
      AND is_active = FALSE
      AND metadata_type IN ('performance', 'usage');
    
    COMMIT;
    
    SELECT ROW_COUNT() as cleaned_records;
END$$

DELIMITER ;

-- 创建存储过程：获取AI身份统计信息
DELIMITER $$

CREATE PROCEDURE GetAIIdentityStatistics()
BEGIN
    SELECT 
        'profiles' as table_name,
        COUNT(*) as total_records,
        COUNT(CASE WHEN is_active = TRUE THEN 1 END) as active_records,
        AVG(comprehensive_score) as avg_comprehensive_score,
        AVG(data_completeness) as avg_data_completeness
    FROM ai_identity_profiles
    
    UNION ALL
    
    SELECT 
        'vectors' as table_name,
        COUNT(*) as total_records,
        COUNT(CASE WHEN is_active = TRUE THEN 1 END) as active_records,
        AVG(vector_dimension) as avg_dimension,
        COUNT(DISTINCT vector_model) as model_count
    FROM ai_identity_vectors
    
    UNION ALL
    
    SELECT 
        'similarities' as table_name,
        COUNT(*) as total_records,
        COUNT(CASE WHEN is_active = TRUE THEN 1 END) as active_records,
        AVG(overall_similarity_score) as avg_similarity_score,
        MAX(overall_similarity_score) as max_similarity_score
    FROM ai_identity_similarities;
END$$

DELIMITER ;

-- 表结构创建完成
SELECT 'AI身份数据模型表结构创建完成' as status;
