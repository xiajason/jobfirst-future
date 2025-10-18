-- 扩展职位表，添加解析结果字段
-- 检查并添加parsed_data字段
SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE TABLE_SCHEMA = DATABASE() 
     AND TABLE_NAME = 'jobs' 
     AND COLUMN_NAME = 'parsed_data') = 0,
    'ALTER TABLE jobs ADD COLUMN parsed_data TEXT COMMENT ''MinerU解析结果JSON''',
    'SELECT ''parsed_data column already exists'' as message'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 检查并添加parsing_status字段
SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE TABLE_SCHEMA = DATABASE() 
     AND TABLE_NAME = 'jobs' 
     AND COLUMN_NAME = 'parsing_status') = 0,
    'ALTER TABLE jobs ADD COLUMN parsing_status VARCHAR(20) DEFAULT ''pending'' COMMENT ''解析状态: pending, processing, completed, failed''',
    'SELECT ''parsing_status column already exists'' as message'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
