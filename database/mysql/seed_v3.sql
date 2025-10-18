-- JobFirst V3.0 模拟数据填充脚本
-- 基于 DATABASE_MAPPING_V3.md 的完整数据结构

USE jobfirst_v3;

-- ==================== 用户数据 ====================

-- 插入测试用户
INSERT INTO users (uuid, email, username, password_hash, first_name, last_name, phone, avatar_url, status) VALUES
('user-uuid-001', 'zhangsan@jobfirst.com', 'zhangsan', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj3ZxQQxq3Hy', '张三', '张', '13800138001', 'https://example.com/avatar1.jpg', 'active'),
('user-uuid-002', 'lisi@jobfirst.com', 'lisi', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj3ZxQQxq3Hy', '李四', '李', '13800138002', 'https://example.com/avatar2.jpg', 'active'),
('user-uuid-003', 'wangwu@jobfirst.com', 'wangwu', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj3ZxQQxq3Hy', '王五', '王', '13800138003', 'https://example.com/avatar3.jpg', 'active'),
('user-uuid-004', 'zhaoliu@jobfirst.com', 'zhaoliu', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj3ZxQQxq3Hy', '赵六', '赵', '13800138004', 'https://example.com/avatar4.jpg', 'active'),
('user-uuid-005', 'qianqi@jobfirst.com', 'qianqi', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj3ZxQQxq3Hy', '钱七', '钱', '13800138005', 'https://example.com/avatar5.jpg', 'active');

-- 插入用户资料
INSERT INTO user_profiles (user_id, bio, location, website, linkedin_url, github_url, date_of_birth, gender, nationality, languages, skills, interests) VALUES
(1, '5年前端开发经验，精通React、Vue等现代前端技术栈，热爱开源项目', '北京', 'https://zhangsan.dev', 'https://linkedin.com/in/zhangsan', 'https://github.com/zhangsan', '1990-05-15', 'male', '中国', '["中文", "英文"]', '["React", "Vue", "JavaScript", "TypeScript"]', '["编程", "开源", "技术分享"]'),
(2, '3年产品管理经验，擅长用户研究和产品设计，有丰富的B端产品经验', '上海', 'https://lisi.design', 'https://linkedin.com/in/lisi', 'https://github.com/lisi', '1988-08-20', 'female', '中国', '["中文", "英文", "日文"]', '["产品设计", "用户研究", "数据分析"]', '["设计", "心理学", "商业分析"]'),
(3, '7年后端开发经验，专注于分布式系统和微服务架构', '深圳', 'https://wangwu.tech', 'https://linkedin.com/in/wangwu', 'https://github.com/wangwu', '1985-12-10', 'male', '中国', '["中文", "英文"]', '["Go", "Java", "Python", "Docker", "Kubernetes"]', '["系统架构", "开源", "技术管理"]'),
(4, '4年全栈开发经验，熟悉前后端技术栈，有丰富的项目经验', '杭州', 'https://zhaoliu.dev', 'https://linkedin.com/in/zhaoliu', 'https://github.com/zhaoliu', '1992-03-25', 'male', '中国', '["中文", "英文"]', '["React", "Node.js", "Python", "MySQL", "Redis"]', '["全栈开发", "技术分享", "创业"]'),
(5, '2年UI/UX设计经验，专注于移动端和Web端界面设计', '广州', 'https://qianqi.design', 'https://linkedin.com/in/qianqi', 'https://github.com/qianqi', '1995-07-18', 'female', '中国', '["中文", "英文", "韩文"]', '["UI设计", "UX设计", "Figma", "Sketch", "Photoshop"]', '["设计", "艺术", "摄影"]');

-- ==================== 标准化数据 ====================

-- 插入技能数据
INSERT INTO skills (name, category, description, icon, is_popular, search_count) VALUES
-- 前端技能
('React', '前端框架', '用于构建用户界面的JavaScript库', 'react-icon', TRUE, 1250),
('Vue.js', '前端框架', '渐进式JavaScript框架', 'vue-icon', TRUE, 980),
('Angular', '前端框架', 'Google开发的Web应用框架', 'angular-icon', TRUE, 750),
('JavaScript', '编程语言', 'Web开发的核心编程语言', 'js-icon', TRUE, 2000),
('TypeScript', '编程语言', 'JavaScript的超集，提供静态类型检查', 'ts-icon', TRUE, 850),
('HTML', '标记语言', '超文本标记语言', 'html-icon', TRUE, 1500),
('CSS', '样式语言', '层叠样式表', 'css-icon', TRUE, 1200),
('Sass', '样式语言', 'CSS预处理器', 'sass-icon', FALSE, 300),
('Less', '样式语言', 'CSS预处理器', 'less-icon', FALSE, 200),
('Webpack', '构建工具', '模块打包器', 'webpack-icon', TRUE, 600),
('Vite', '构建工具', '下一代前端构建工具', 'vite-icon', TRUE, 400),
('Next.js', '前端框架', 'React全栈框架', 'nextjs-icon', TRUE, 500),
('Nuxt.js', '前端框架', 'Vue全栈框架', 'nuxtjs-icon', FALSE, 250),

-- 后端技能
('Go', '编程语言', 'Google开发的编程语言', 'go-icon', TRUE, 800),
('Golang', '编程语言', 'Go语言的别名', 'golang-icon', TRUE, 600),
('Java', '编程语言', '面向对象的编程语言', 'java-icon', TRUE, 1800),
('Python', '编程语言', '高级编程语言', 'python-icon', TRUE, 1600),
('Node.js', '运行时', 'JavaScript运行时环境', 'nodejs-icon', TRUE, 900),
('PHP', '编程语言', '服务器端脚本语言', 'php-icon', TRUE, 700),
('Ruby', '编程语言', '面向对象的编程语言', 'ruby-icon', FALSE, 200),
('C#', '编程语言', '微软开发的编程语言', 'csharp-icon', TRUE, 500),
('Spring', '框架', 'Java企业级应用框架', 'spring-icon', TRUE, 600),
('Django', '框架', 'Python Web框架', 'django-icon', TRUE, 400),
('Flask', '框架', 'Python轻量级Web框架', 'flask-icon', TRUE, 300),
('Express', '框架', 'Node.js Web框架', 'express-icon', TRUE, 500),
('Gin', '框架', 'Go Web框架', 'gin-icon', TRUE, 300),
('Echo', '框架', 'Go高性能Web框架', 'echo-icon', FALSE, 150),

-- 数据库技能
('MySQL', '数据库', '关系型数据库管理系统', 'mysql-icon', TRUE, 1000),
('PostgreSQL', '数据库', '开源关系型数据库', 'postgresql-icon', TRUE, 600),
('MongoDB', '数据库', 'NoSQL文档数据库', 'mongodb-icon', TRUE, 500),
('Redis', '数据库', '内存数据结构存储', 'redis-icon', TRUE, 700),
('Elasticsearch', '数据库', '分布式搜索和分析引擎', 'elasticsearch-icon', TRUE, 300),
('SQL', '数据库', '结构化查询语言', 'sql-icon', TRUE, 1200),
('NoSQL', '数据库', '非关系型数据库', 'nosql-icon', TRUE, 400),
('Oracle', '数据库', '企业级关系型数据库', 'oracle-icon', TRUE, 300),
('SQLite', '数据库', '轻量级关系型数据库', 'sqlite-icon', TRUE, 200),

-- 运维技能
('Docker', '容器化', '应用容器引擎', 'docker-icon', TRUE, 800),
('Kubernetes', '容器编排', '容器编排平台', 'kubernetes-icon', TRUE, 500),
('AWS', '云服务', '亚马逊云服务', 'aws-icon', TRUE, 600),
('Azure', '云服务', '微软云服务', 'azure-icon', TRUE, 300),
('GCP', '云服务', '谷歌云平台', 'gcp-icon', TRUE, 250),
('Jenkins', 'CI/CD', '持续集成工具', 'jenkins-icon', TRUE, 400),
('GitLab', 'CI/CD', 'DevOps平台', 'gitlab-icon', TRUE, 300),
('Linux', '操作系统', '开源操作系统', 'linux-icon', TRUE, 800),
('Nginx', 'Web服务器', '高性能Web服务器', 'nginx-icon', TRUE, 500),
('Apache', 'Web服务器', 'Web服务器软件', 'apache-icon', TRUE, 300),

-- 设计技能
('Photoshop', '设计工具', '图像处理软件', 'photoshop-icon', TRUE, 600),
('Illustrator', '设计工具', '矢量图形软件', 'illustrator-icon', TRUE, 400),
('Figma', '设计工具', '协作界面设计工具', 'figma-icon', TRUE, 500),
('Sketch', '设计工具', 'Mac界面设计工具', 'sketch-icon', TRUE, 300),
('UI设计', '设计领域', '用户界面设计', 'ui-icon', TRUE, 800),
('UX设计', '设计领域', '用户体验设计', 'ux-icon', TRUE, 700),
('设计', '设计领域', '视觉设计', 'design-icon', TRUE, 600),
('Adobe', '设计工具', 'Adobe创意套件', 'adobe-icon', TRUE, 400);

-- 插入公司数据
INSERT INTO companies (name, industry, size, location, website, logo_url, description, is_verified) VALUES
('腾讯科技', '互联网', 'enterprise', '深圳', 'https://www.tencent.com', 'https://example.com/tencent-logo.png', '中国领先的互联网综合服务提供商', TRUE),
('字节跳动', '互联网', 'enterprise', '北京', 'https://www.bytedance.com', 'https://example.com/bytedance-logo.png', '全球化的移动互联网公司', TRUE),
('阿里巴巴', '互联网', 'enterprise', '杭州', 'https://www.alibaba.com', 'https://example.com/alibaba-logo.png', '全球领先的电子商务公司', TRUE),
('百度', '互联网', 'enterprise', '北京', 'https://www.baidu.com', 'https://example.com/baidu-logo.png', '全球最大的中文搜索引擎', TRUE),
('美团', '互联网', 'enterprise', '北京', 'https://www.meituan.com', 'https://example.com/meituan-logo.png', '中国领先的生活服务电子商务平台', TRUE),
('滴滴出行', '互联网', 'enterprise', '北京', 'https://www.didiglobal.com', 'https://example.com/didi-logo.png', '全球领先的一站式出行平台', TRUE),
('小米科技', '互联网', 'enterprise', '北京', 'https://www.mi.com', 'https://example.com/xiaomi-logo.png', '以手机、智能硬件和IoT平台为核心的互联网公司', TRUE),
('华为技术', '通信', 'enterprise', '深圳', 'https://www.huawei.com', 'https://example.com/huawei-logo.png', '全球领先的信息与通信技术解决方案供应商', TRUE),
('京东', '互联网', 'enterprise', '北京', 'https://www.jd.com', 'https://example.com/jd-logo.png', '中国领先的技术驱动的电商和零售基础设施服务商', TRUE),
('网易', '互联网', 'enterprise', '杭州', 'https://www.163.com', 'https://example.com/netease-logo.png', '中国领先的互联网技术公司', TRUE),
('创新工场', '投资', 'medium', '北京', 'https://www.chuangxin.com', 'https://example.com/chuangxin-logo.png', '早期投资机构', FALSE),
('红杉资本', '投资', 'large', '北京', 'https://www.sequoiacap.com', 'https://example.com/sequoia-logo.png', '全球顶级风险投资公司', FALSE);

-- 插入职位数据
INSERT INTO positions (title, category, level, description, requirements) VALUES
-- 技术开发类
('前端开发工程师', '技术开发', 'senior', '负责前端应用开发和维护', '熟悉React、Vue等前端框架，有3年以上开发经验'),
('后端开发工程师', '技术开发', 'senior', '负责后端服务开发和维护', '熟悉Go、Java等后端技术，有3年以上开发经验'),
('全栈开发工程师', '技术开发', 'senior', '负责前后端全栈开发', '熟悉前后端技术栈，有3年以上开发经验'),
('移动端开发工程师', '技术开发', 'senior', '负责移动应用开发', '熟悉React Native、Flutter等移动开发技术'),
('算法工程师', '技术开发', 'senior', '负责算法研发和优化', '熟悉机器学习、深度学习算法，有相关项目经验'),
('架构师', '技术开发', 'lead', '负责系统架构设计', '有大型系统架构经验，熟悉分布式系统设计'),
('技术总监', '技术开发', 'executive', '负责技术团队管理', '有技术团队管理经验，熟悉技术发展趋势'),

-- 产品设计类
('产品经理', '产品设计', 'senior', '负责产品规划和设计', '有产品管理经验，熟悉用户研究方法'),
('UI设计师', '产品设计', 'senior', '负责用户界面设计', '熟悉设计工具，有丰富的UI设计经验'),
('UX设计师', '产品设计', 'senior', '负责用户体验设计', '熟悉用户体验设计流程，有相关项目经验'),
('交互设计师', '产品设计', 'senior', '负责交互设计', '熟悉交互设计原则，有相关项目经验'),

-- 运营管理类
('运营经理', '运营管理', 'senior', '负责产品运营', '有运营管理经验，熟悉数据分析'),
('市场经理', '运营管理', 'senior', '负责市场推广', '有市场营销经验，熟悉市场分析'),
('商务经理', '运营管理', 'senior', '负责商务合作', '有商务合作经验，熟悉商务谈判'),

-- 其他类
('数据分析师', '数据分析', 'senior', '负责数据分析', '熟悉数据分析工具，有相关项目经验'),
('测试工程师', '质量保证', 'senior', '负责软件测试', '熟悉测试方法和工具，有相关项目经验'),
('运维工程师', '运维部署', 'senior', '负责系统运维', '熟悉Linux系统，有运维经验');

-- ==================== 简历模板数据 ====================

INSERT INTO resume_templates (uuid, name, description, category, template_data, is_free, popularity_score, download_count, rating, review_count) VALUES
('template-uuid-001', '现代简约模板', '简洁现代的设计风格，适合技术岗位', 'modern', '{"sections": ["personal", "summary", "experience", "education", "skills"], "layout": "modern", "colors": ["#2563eb", "#64748b"]}', TRUE, 95, 1250, 4.8, 156),
('template-uuid-002', '专业商务模板', '传统商务风格，适合管理岗位', 'professional', '{"sections": ["personal", "objective", "experience", "education", "skills", "references"], "layout": "professional", "colors": ["#1f2937", "#6b7280"]}', TRUE, 90, 980, 4.6, 120),
('template-uuid-003', '创意设计模板', '富有创意的设计风格，适合设计岗位', 'creative', '{"sections": ["personal", "portfolio", "experience", "skills"], "layout": "creative", "colors": ["#7c3aed", "#a78bfa"]}', TRUE, 85, 750, 4.5, 89),
('template-uuid-004', '极简风格模板', '极简设计，突出内容', 'minimal', '{"sections": ["personal", "experience", "education"], "layout": "minimal", "colors": ["#000000", "#666666"]}', TRUE, 88, 650, 4.7, 95),
('template-uuid-005', '经典传统模板', '传统经典风格，适合保守行业', 'classic', '{"sections": ["personal", "objective", "experience", "education", "skills", "references"], "layout": "classic", "colors": ["#1f2937", "#9ca3af"]}', TRUE, 80, 520, 4.4, 78);

-- ==================== 简历数据 ====================

INSERT INTO resumes (uuid, user_id, title, slug, summary, template_id, content, status, visibility, can_comment, view_count, download_count, share_count, comment_count, like_count, is_default, published_at) VALUES
('resume-uuid-001', 1, '前端开发工程师简历', 'frontend-developer-resume', '5年前端开发经验，精通React、Vue等现代前端技术栈，有丰富的项目经验和团队协作能力', 1, '# 前端开发工程师简历

## 个人信息
- **姓名：** 张三
- **邮箱：** zhangsan@jobfirst.com
- **电话：** 13800138001
- **地址：** 北京市朝阳区
- **GitHub：** https://github.com/zhangsan

## 个人简介
5年前端开发经验，精通React、Vue等现代前端技术栈，有丰富的项目经验和团队协作能力。热爱开源项目，积极参与技术社区。

## 工作经历

### 高级前端开发工程师
**公司：** 腾讯科技
**时间：** 2020年1月 - 至今
**描述：** 负责公司核心产品的前端开发工作，使用React构建高性能的用户界面

### 前端开发工程师
**公司：** 字节跳动
**时间：** 2018年6月 - 2019年12月
**描述：** 参与多个移动端和Web端项目开发，使用Vue.js构建用户界面

## 教育背景
- **学校：** 北京理工大学
- **专业：** 计算机科学与技术
- **学位：** 学士
- **时间：** 2014年 - 2018年

## 技能
- React (专家)
- Vue.js (高级)
- JavaScript (专家)
- TypeScript (高级)
- HTML/CSS (专家)', 'published', 'public', TRUE, 156, 23, 8, 5, 12, TRUE, '2024-01-15 10:30:00'),

('resume-uuid-002', 2, '产品经理简历', 'product-manager-resume', '3年产品管理经验，擅长用户研究和产品设计，有丰富的B端产品经验', 2, '# 产品经理简历

## 个人信息
- **姓名：** 李四
- **邮箱：** lisi@jobfirst.com
- **电话：** 13800138002
- **地址：** 上海市浦东新区

## 个人简介
3年产品管理经验，擅长用户研究和产品设计，有丰富的B端产品经验。注重用户体验，善于数据驱动决策。

## 工作经历

### 高级产品经理
**公司：** 阿里巴巴
**时间：** 2021年3月 - 至今
**描述：** 负责电商平台产品规划和设计，提升用户体验和业务指标

### 产品经理
**公司：** 美团
**时间：** 2019年7月 - 2021年2月
**描述：** 负责生活服务类产品设计，通过用户研究优化产品功能

## 教育背景
- **学校：** 复旦大学
- **专业：** 工商管理
- **学位：** 硕士
- **时间：** 2017年 - 2019年

## 技能
- 产品设计 (高级)
- 用户研究 (高级)
- 数据分析 (中级)
- 项目管理 (高级)', 'published', 'public', TRUE, 89, 15, 3, 2, 8, TRUE, '2024-01-20 14:20:00'),

('resume-uuid-003', 3, '后端开发工程师简历', 'backend-developer-resume', '7年后端开发经验，专注于分布式系统和微服务架构', 1, '# 后端开发工程师简历

## 个人信息
- **姓名：** 王五
- **邮箱：** wangwu@jobfirst.com
- **电话：** 13800138003
- **地址：** 深圳市南山区

## 个人简介
7年后端开发经验，专注于分布式系统和微服务架构。有丰富的系统设计和性能优化经验。

## 工作经历

### 技术专家
**公司：** 华为技术
**时间：** 2019年5月 - 至今
**描述：** 负责分布式系统架构设计，使用Go和Java构建高性能后端服务

### 高级后端开发工程师
**公司：** 腾讯科技
**时间：** 2017年3月 - 2019年4月
**描述：** 参与多个大型项目的后端开发，负责系统架构设计

## 教育背景
- **学校：** 华南理工大学
- **专业：** 软件工程
- **学位：** 硕士
- **时间：** 2015年 - 2017年

## 技能
- Go (专家)
- Java (专家)
- Python (高级)
- Docker (高级)
- Kubernetes (高级)', 'published', 'private', TRUE, 67, 12, 2, 1, 5, TRUE, '2024-01-25 09:15:00');

-- ==================== 简历关联数据 ====================

-- 简历技能关联
INSERT INTO resume_skills (resume_id, skill_id, proficiency_level, years_of_experience, is_highlighted) VALUES
-- 张三的技能
(1, 1, 'expert', 5.0, TRUE),  -- React
(1, 2, 'advanced', 3.0, TRUE),  -- Vue.js
(1, 4, 'expert', 5.0, TRUE),  -- JavaScript
(1, 5, 'advanced', 2.0, FALSE),  -- TypeScript
(1, 6, 'expert', 5.0, FALSE),  -- HTML
(1, 7, 'expert', 5.0, FALSE),  -- CSS
(1, 10, 'advanced', 3.0, FALSE),  -- Webpack
-- 李四的技能
(2, 35, 'advanced', 3.0, TRUE),  -- 产品设计
(2, 36, 'advanced', 3.0, TRUE),  -- 用户研究
(2, 37, 'intermediate', 2.0, FALSE),  -- 数据分析
(2, 38, 'advanced', 3.0, FALSE),  -- 项目管理
-- 王五的技能
(3, 14, 'expert', 4.0, TRUE),  -- Go
(3, 16, 'expert', 7.0, TRUE),  -- Java
(3, 17, 'advanced', 3.0, FALSE),  -- Python
(3, 30, 'advanced', 3.0, TRUE),  -- Docker
(3, 31, 'advanced', 2.0, TRUE);  -- Kubernetes

-- 工作经历
INSERT INTO work_experiences (resume_id, company_id, position_id, title, start_date, end_date, is_current, location, description, achievements, technologies) VALUES
-- 张三的工作经历
(1, 1, 1, '高级前端开发工程师', '2020-01-01', NULL, TRUE, '深圳', '负责公司核心产品的前端开发工作，使用React构建高性能的用户界面', '成功重构了用户界面，提升了30%的页面加载速度', 'React, TypeScript, Webpack'),
(1, 2, 1, '前端开发工程师', '2018-06-01', '2019-12-31', FALSE, '北京', '参与多个移动端和Web端项目开发，使用Vue.js构建用户界面', '完成了3个重要项目的开发，获得团队认可', 'Vue.js, JavaScript, HTML/CSS'),
-- 李四的工作经历
(2, 3, 8, '高级产品经理', '2021-03-01', NULL, TRUE, '杭州', '负责电商平台产品规划和设计，提升用户体验和业务指标', '通过产品优化，提升了20%的用户留存率', '产品设计, 用户研究, 数据分析'),
(2, 5, 8, '产品经理', '2019-07-01', '2021-02-28', FALSE, '北京', '负责生活服务类产品设计，通过用户研究优化产品功能', '成功推出了2个新产品功能，获得用户好评', '产品设计, 用户研究'),
-- 王五的工作经历
(3, 8, 2, '技术专家', '2019-05-01', NULL, TRUE, '深圳', '负责分布式系统架构设计，使用Go和Java构建高性能后端服务', '设计了新的微服务架构，提升了系统性能50%', 'Go, Java, Docker, Kubernetes'),
(3, 1, 2, '高级后端开发工程师', '2017-03-01', '2019-04-30', FALSE, '深圳', '参与多个大型项目的后端开发，负责系统架构设计', '完成了多个核心模块的开发，获得技术奖项', 'Java, Python, MySQL, Redis');

-- 教育背景
INSERT INTO educations (resume_id, school, degree, major, start_date, end_date, gpa, location, description, is_highlighted) VALUES
(1, '北京理工大学', '学士', '计算机科学与技术', '2014-09-01', '2018-06-30', 3.8, '北京', '主修计算机科学与技术，辅修软件工程', TRUE),
(2, '复旦大学', '硕士', '工商管理', '2017-09-01', '2019-06-30', 3.9, '上海', '主修工商管理，专注于产品管理和市场营销', TRUE),
(3, '华南理工大学', '硕士', '软件工程', '2015-09-01', '2017-06-30', 3.7, '广州', '主修软件工程，专注于分布式系统设计', TRUE);

-- 项目经验
INSERT INTO projects (resume_id, name, description, start_date, end_date, status, technology_stack, project_url, github_url, demo_url, company_id, is_highlighted) VALUES
-- 张三的项目
(1, '企业级管理系统', '基于React构建的企业级管理系统，支持多租户和权限管理', '2020-03-01', '2020-12-31', 'completed', 'React, TypeScript, Ant Design, Node.js', 'https://example.com/project1', 'https://github.com/zhangsan/project1', 'https://demo.example.com/project1', 1, TRUE),
(1, '移动端H5应用', '使用Vue.js开发的移动端H5应用，支持离线使用', '2019-01-01', '2019-06-30', 'completed', 'Vue.js, Vant, PWA', 'https://example.com/project2', 'https://github.com/zhangsan/project2', 'https://demo.example.com/project2', 2, FALSE),
-- 李四的项目
(2, '电商平台优化', '通过用户研究和数据分析优化电商平台用户体验', '2021-06-01', '2021-12-31', 'completed', '用户研究, 数据分析, A/B测试', 'https://example.com/project3', NULL, 'https://demo.example.com/project3', 3, TRUE),
-- 王五的项目
(3, '微服务架构重构', '将单体应用重构为微服务架构，提升系统可扩展性', '2020-01-01', '2020-12-31', 'completed', 'Go, Docker, Kubernetes, gRPC', 'https://example.com/project4', 'https://github.com/wangwu/project4', NULL, 8, TRUE);

-- 证书认证
INSERT INTO certifications (resume_id, name, issuer, issue_date, expiry_date, credential_id, credential_url, description, is_highlighted) VALUES
(1, 'AWS认证解决方案架构师', 'Amazon Web Services', '2023-06-15', '2026-06-15', 'AWS-SAA-001', 'https://aws.amazon.com/certification', 'AWS云服务架构设计认证', TRUE),
(1, 'React开发者认证', 'Meta', '2023-03-20', NULL, 'REACT-001', 'https://react.dev', 'React框架开发认证', FALSE),
(2, '产品经理认证', 'Google', '2022-09-10', '2025-09-10', 'PM-001', 'https://grow.google', 'Google产品管理认证', TRUE),
(3, 'Kubernetes管理员认证', 'CNCF', '2023-01-15', '2026-01-15', 'CKA-001', 'https://cncf.io', 'Kubernetes集群管理认证', TRUE);

-- ==================== 社交功能数据 ====================

-- 简历评论
INSERT INTO resume_comments (resume_id, user_id, content, is_approved, like_count, reply_count) VALUES
(1, 2, '简历写得很好，技能描述很详细！', TRUE, 3, 0),
(1, 3, '前端技术栈很全面，项目经验丰富', TRUE, 2, 0),
(2, 1, '产品思维很清晰，用户体验考虑周到', TRUE, 4, 0),
(2, 3, '数据分析能力很强，值得学习', TRUE, 1, 0),
(3, 1, '后端架构设计经验丰富，技术深度很好', TRUE, 2, 0),
(3, 2, '微服务架构设计很棒，有参考价值', TRUE, 3, 0);

-- 简历点赞
INSERT INTO resume_likes (resume_id, user_id) VALUES
(1, 2), (1, 3), (1, 4), (1, 5),
(2, 1), (2, 3), (2, 4),
(3, 1), (3, 2), (3, 4);

-- 简历分享
INSERT INTO resume_shares (resume_id, user_id, platform, share_url) VALUES
(1, 1, 'LinkedIn', 'https://linkedin.com/in/zhangsan'),
(1, 1, 'GitHub', 'https://github.com/zhangsan'),
(2, 2, 'LinkedIn', 'https://linkedin.com/in/lisi'),
(3, 3, 'LinkedIn', 'https://linkedin.com/in/wangwu');

-- ==================== 其他功能数据 ====================

-- 用户设置
INSERT INTO user_settings (user_id, theme, language, timezone, email_notifications, push_notifications, privacy_level, resume_visibility) VALUES
(1, 'light', 'zh-CN', 'Asia/Shanghai', TRUE, TRUE, 'public', 'public'),
(2, 'dark', 'zh-CN', 'Asia/Shanghai', TRUE, FALSE, 'friends', 'public'),
(3, 'auto', 'zh-CN', 'Asia/Shanghai', TRUE, TRUE, 'public', 'private'),
(4, 'light', 'zh-CN', 'Asia/Shanghai', FALSE, TRUE, 'private', 'friends'),
(5, 'dark', 'zh-CN', 'Asia/Shanghai', TRUE, TRUE, 'friends', 'public');

-- 积分数据
INSERT INTO points (user_id, balance, total_earned, total_spent) VALUES
(1, 150, 200, 50),
(2, 120, 150, 30),
(3, 180, 220, 40),
(4, 100, 120, 20),
(5, 90, 100, 10);

-- 积分历史
INSERT INTO point_history (user_id, points, type, reason, description, reference_type, reference_id, balance_after) VALUES
(1, 50, 'earn', '注册奖励', '新用户注册奖励', 'registration', 1, 150),
(1, 30, 'earn', '简历发布', '发布简历获得积分', 'resume', 1, 180),
(1, 20, 'earn', '简历被点赞', '简历获得点赞奖励', 'like', 1, 200),
(1, 50, 'spend', '下载模板', '下载付费简历模板', 'template', 1, 150),
(2, 50, 'earn', '注册奖励', '新用户注册奖励', 'registration', 2, 150),
(2, 30, 'earn', '简历发布', '发布简历获得积分', 'resume', 2, 180),
(2, 30, 'spend', '下载模板', '下载付费简历模板', 'template', 2, 150);

COMMIT;
