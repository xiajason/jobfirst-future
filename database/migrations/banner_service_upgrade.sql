-- Banner Service 升级脚本
-- 重构为内容管理服务，支持Banner、Markdown内容和评论功能

-- 删除旧的banners表（如果存在）
DROP TABLE IF EXISTS banners;

-- 创建新的banners表
CREATE TABLE banners (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(200) NOT NULL COMMENT 'Banner标题',
    description TEXT COMMENT 'Banner描述',
    image_url VARCHAR(500) COMMENT '图片URL',
    link_url VARCHAR(500) COMMENT '链接URL',
    sort_order INT DEFAULT 0 COMMENT '排序顺序',
    status VARCHAR(20) DEFAULT 'draft' COMMENT '状态: draft, active, inactive',
    start_date TIMESTAMP NULL COMMENT '开始时间',
    end_date TIMESTAMP NULL COMMENT '结束时间',
    click_count INT DEFAULT 0 COMMENT '点击次数',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否激活',
    created_by INT NOT NULL COMMENT '创建者ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_status (status),
    INDEX idx_sort_order (sort_order),
    INDEX idx_created_by (created_by),
    INDEX idx_is_active (is_active),
    INDEX idx_date_range (start_date, end_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Banner表';

-- 创建markdown_contents表
CREATE TABLE markdown_contents (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(200) NOT NULL COMMENT '内容标题',
    slug VARCHAR(200) UNIQUE NOT NULL COMMENT 'URL别名',
    content LONGTEXT NOT NULL COMMENT 'Markdown内容',
    excerpt TEXT COMMENT '内容摘要',
    category VARCHAR(100) COMMENT '分类',
    tags TEXT COMMENT '标签(JSON格式)',
    status VARCHAR(20) DEFAULT 'draft' COMMENT '状态: draft, published, archived',
    published_at TIMESTAMP NULL COMMENT '发布时间',
    view_count INT DEFAULT 0 COMMENT '浏览次数',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否激活',
    created_by INT NOT NULL COMMENT '创建者ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_slug (slug),
    INDEX idx_status (status),
    INDEX idx_category (category),
    INDEX idx_created_by (created_by),
    INDEX idx_published_at (published_at),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Markdown内容表';

-- 创建comments表
CREATE TABLE comments (
    id INT AUTO_INCREMENT PRIMARY KEY,
    content_id INT NOT NULL COMMENT '关联的Markdown内容ID',
    user_id INT NOT NULL COMMENT '用户ID',
    parent_id INT NULL COMMENT '父评论ID',
    content TEXT NOT NULL COMMENT '评论内容',
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending, approved, rejected',
    ip_address VARCHAR(45) COMMENT 'IP地址',
    user_agent VARCHAR(500) COMMENT '用户代理',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否激活',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_content_id (content_id),
    INDEX idx_user_id (user_id),
    INDEX idx_parent_id (parent_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    INDEX idx_is_active (is_active),
    FOREIGN KEY (content_id) REFERENCES markdown_contents(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES comments(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='评论表';

-- 插入示例Banner数据
INSERT INTO banners (title, description, image_url, link_url, sort_order, status, created_by, created_at, updated_at) VALUES
('欢迎使用JobFirst平台', '专业的求职招聘平台，为求职者和企业提供优质服务', 'https://example.com/banner1.jpg', 'https://jobfirst.com', 1, 'active', 1, NOW(), NOW()),
('最新职位推荐', '查看最新发布的优质职位，找到心仪的工作', 'https://example.com/banner2.jpg', 'https://jobfirst.com/jobs', 2, 'active', 1, NOW(), NOW()),
('企业招聘服务', '为企业提供专业的招聘解决方案', 'https://example.com/banner3.jpg', 'https://jobfirst.com/enterprise', 3, 'active', 1, NOW(), NOW());

-- 插入示例Markdown内容
INSERT INTO markdown_contents (title, slug, content, excerpt, category, tags, status, published_at, created_by, created_at, updated_at) VALUES
('如何写好简历', 'how-to-write-resume', 
'# 如何写好简历\n\n## 基本信息\n- 姓名要真实\n- 联系方式要准确\n- 邮箱要专业\n\n## 工作经验\n- 按时间倒序排列\n- 突出工作成果\n- 使用数据说话\n\n## 教育背景\n- 从最高学历开始\n- 包含相关课程\n- 突出学术成就', 
'本文介绍了如何写一份优秀的简历，包括基本信息、工作经验和教育背景的写作要点。', 
'求职指导', 
'["简历", "求职", "职场"]', 
'published', 
NOW(), 
1, 
NOW(), 
NOW()),

('面试技巧大全', 'interview-tips', 
'# 面试技巧大全\n\n## 面试前准备\n- 了解公司背景\n- 准备常见问题\n- 着装得体\n\n## 面试中表现\n- 保持自信\n- 回答问题要具体\n- 主动提问\n\n## 面试后跟进\n- 发送感谢邮件\n- 耐心等待结果', 
'全面的面试技巧指南，帮助求职者在面试中表现出色。', 
'求职指导', 
'["面试", "求职", "技巧"]', 
'published', 
NOW(), 
1, 
NOW(), 
NOW()),

('职场新人指南', 'newbie-guide', 
'# 职场新人指南\n\n## 第一印象很重要\n- 准时到达\n- 着装规范\n- 积极态度\n\n## 学习与成长\n- 主动学习\n- 寻求反馈\n- 建立人脉\n\n## 工作习惯\n- 保持专业\n- 及时沟通\n- 承担责任', 
'为职场新人提供实用的工作建议和成长指导。', 
'职场发展', 
'["职场", "新人", "成长"]', 
'draft', 
NULL, 
1, 
NOW(), 
NOW());

-- 插入示例评论数据
INSERT INTO comments (content_id, user_id, content, status, ip_address, user_agent, created_at, updated_at) VALUES
(1, 1, '非常实用的简历写作指南，对我帮助很大！', 'approved', '127.0.0.1', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)', NOW(), NOW()),
(1, 2, '学到了很多，特别是关于工作经验部分的建议。', 'approved', '127.0.0.1', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)', NOW(), NOW()),
(2, 1, '面试技巧很全面，准备按照这些建议来准备面试。', 'approved', '127.0.0.1', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)', NOW(), NOW()),
(2, 3, '请问如何回答"你为什么选择我们公司"这个问题？', 'pending', '127.0.0.1', 'Mozilla/5.0 (iPhone; CPU iPhone OS 14_0)', NOW(), NOW());

-- 创建内容统计视图
CREATE OR REPLACE VIEW content_statistics AS
SELECT 
    mc.id,
    mc.title,
    mc.category,
    mc.status,
    mc.view_count,
    mc.created_at,
    mc.published_at,
    u.username as created_by_name,
    COUNT(c.id) as comment_count,
    COUNT(CASE WHEN c.status = 'approved' THEN 1 END) as approved_comments
FROM markdown_contents mc
LEFT JOIN comments c ON mc.id = c.content_id
LEFT JOIN users u ON mc.created_by = u.id
GROUP BY mc.id, mc.title, mc.category, mc.status, mc.view_count, mc.created_at, mc.published_at, u.username;

-- 创建评论统计视图
CREATE OR REPLACE VIEW comment_statistics AS
SELECT 
    c.id,
    c.content_id,
    c.user_id,
    c.status,
    c.created_at,
    u.username as user_name,
    mc.title as content_title
FROM comments c
LEFT JOIN users u ON c.user_id = u.id
LEFT JOIN markdown_contents mc ON c.content_id = mc.id;
