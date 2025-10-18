-- 创建templates表
CREATE TABLE IF NOT EXISTS templates (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(200) NOT NULL COMMENT '模板名称',
    category VARCHAR(100) NOT NULL COMMENT '模板分类',
    description TEXT COMMENT '模板描述',
    content TEXT NOT NULL COMMENT '模板内容',
    variables JSON COMMENT '模板变量',
    preview TEXT COMMENT '预览内容',
    usage_count INT DEFAULT 0 COMMENT '使用次数',
    rating DECIMAL(3,2) DEFAULT 0.00 COMMENT '评分(0-5)',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否激活',
    created_by INT NOT NULL COMMENT '创建者ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_category (category),
    INDEX idx_usage_count (usage_count),
    INDEX idx_rating (rating),
    INDEX idx_created_by (created_by),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='模板表';

-- 创建评分表
CREATE TABLE IF NOT EXISTS ratings (
    id INT AUTO_INCREMENT PRIMARY KEY,
    template_id INT NOT NULL COMMENT '模板ID',
    user_id INT NOT NULL COMMENT '用户ID',
    rating DECIMAL(3,2) NOT NULL COMMENT '评分(0-5)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_user_template (template_id, user_id),
    INDEX idx_template_id (template_id),
    INDEX idx_user_id (user_id),
    FOREIGN KEY (template_id) REFERENCES templates(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='模板评分表';

-- 插入示例数据
INSERT INTO templates (name, category, description, content, variables, preview, usage_count, rating, is_active, created_by, created_at, updated_at) VALUES
('标准简历模板', '简历模板', '适用于大多数职位的标准简历模板', 
'# 个人简历\n\n## 基本信息\n- 姓名：{name}\n- 联系方式：{phone}\n- 邮箱：{email}\n\n## 工作经历\n{experience}\n\n## 教育背景\n{education}\n\n## 技能专长\n{skills}', 
'["name", "phone", "email", "experience", "education", "skills"]',
'适用于大多数职位的标准简历模板，包含基本信息、工作经历、教育背景和技能专长等部分。', 
0, 0.00, true, 1, NOW(), NOW()),

('求职信模板', '求职信模板', '通用的求职信模板', 
'尊敬的{hr_name}：\n\n您好！我是{name}，看到贵公司招聘{position}职位，非常感兴趣。\n\n{self_introduction}\n\n{why_company}\n\n期待您的回复。\n\n此致\n敬礼！\n\n{name}\n{date}', 
'["hr_name", "name", "position", "self_introduction", "why_company", "date"]',
'通用的求职信模板，适用于各种职位申请。', 
0, 0.00, true, 1, NOW(), NOW()),

('项目介绍模板', '项目介绍模板', '用于项目展示的模板', 
'# {project_name}\n\n## 项目概述\n{project_overview}\n\n## 技术栈\n{tech_stack}\n\n## 主要功能\n{main_features}\n\n## 项目成果\n{achievements}\n\n## 个人贡献\n{contribution}', 
'["project_name", "project_overview", "tech_stack", "main_features", "achievements", "contribution"]',
'用于项目展示的模板，包含项目概述、技术栈、主要功能等部分。', 
0, 0.00, true, 1, NOW(), NOW());
