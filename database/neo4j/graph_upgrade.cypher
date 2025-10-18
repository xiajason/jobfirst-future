// JobFirst Neo4j图数据库升级脚本
// 版本: V3.0 -> V4.0
// 日期: 2025年1月6日
// 描述: 为Neo4j数据库构建关系网络和智能推荐图谱

// ==============================================
// 清理现有数据
// ==============================================

// 删除所有现有节点和关系
MATCH (n) DETACH DELETE n;

// ==============================================
// 创建约束和索引
// ==============================================

// 创建唯一约束
CREATE CONSTRAINT user_id_unique IF NOT EXISTS FOR (u:User) REQUIRE u.id IS UNIQUE;
CREATE CONSTRAINT skill_id_unique IF NOT EXISTS FOR (s:Skill) REQUIRE s.id IS UNIQUE;
CREATE CONSTRAINT company_id_unique IF NOT EXISTS FOR (c:Company) REQUIRE c.id IS UNIQUE;
CREATE CONSTRAINT job_id_unique IF NOT EXISTS FOR (j:Job) REQUIRE j.id IS UNIQUE;

// 创建索引
CREATE INDEX user_email_index IF NOT EXISTS FOR (u:User) ON (u.email);
CREATE INDEX skill_name_index IF NOT EXISTS FOR (s:Skill) ON (s.name);
CREATE INDEX company_name_index IF NOT EXISTS FOR (c:Company) ON (c.name);
CREATE INDEX job_title_index IF NOT EXISTS FOR (j:Job) ON (j.title);

// ==============================================
// 创建用户节点
// ==============================================

// 创建用户节点
CREATE (u1:User {
    id: 1,
    name: "张三",
    email: "zhangsan@example.com",
    phone: "13800138001",
    location: "北京",
    experience_years: 5,
    created_at: "2025-01-01T00:00:00Z"
});

CREATE (u2:User {
    id: 2,
    name: "李四",
    email: "lisi@example.com",
    phone: "13800138002",
    location: "上海",
    experience_years: 3,
    created_at: "2025-01-02T00:00:00Z"
});

CREATE (u3:User {
    id: 3,
    name: "王五",
    email: "wangwu@example.com",
    phone: "13800138003",
    location: "深圳",
    experience_years: 7,
    created_at: "2025-01-03T00:00:00Z"
});

// ==============================================
// 创建技能节点
// ==============================================

// 创建技能节点
CREATE (s1:Skill {
    id: 1,
    name: "JavaScript",
    category: "编程语言",
    popularity: 95,
    difficulty: "中级"
});

CREATE (s2:Skill {
    id: 2,
    name: "React",
    category: "前端框架",
    popularity: 90,
    difficulty: "中级"
});

CREATE (s3:Skill {
    id: 3,
    name: "Node.js",
    category: "后端框架",
    popularity: 85,
    difficulty: "中级"
});

CREATE (s4:Skill {
    id: 4,
    name: "Python",
    category: "编程语言",
    popularity: 88,
    difficulty: "中级"
});

CREATE (s5:Skill {
    id: 5,
    name: "Java",
    category: "编程语言",
    popularity: 92,
    difficulty: "中级"
});

CREATE (s6:Skill {
    id: 6,
    name: "Spring",
    category: "后端框架",
    popularity: 80,
    difficulty: "高级"
});

CREATE (s7:Skill {
    id: 7,
    name: "MySQL",
    category: "数据库",
    popularity: 85,
    difficulty: "中级"
});

CREATE (s8:Skill {
    id: 8,
    name: "Redis",
    category: "缓存",
    popularity: 75,
    difficulty: "中级"
});

// ==============================================
// 创建公司节点
// ==============================================

// 创建公司节点
CREATE (c1:Company {
    id: 1,
    name: "腾讯科技",
    industry: "互联网",
    size: "大型",
    location: "深圳",
    website: "https://www.tencent.com",
    description: "中国领先的互联网科技公司"
});

CREATE (c2:Company {
    id: 2,
    name: "阿里巴巴",
    industry: "电商",
    size: "大型",
    location: "杭州",
    website: "https://www.alibaba.com",
    description: "全球领先的电子商务公司"
});

CREATE (c3:Company {
    id: 3,
    name: "字节跳动",
    industry: "互联网",
    size: "大型",
    location: "北京",
    website: "https://www.bytedance.com",
    description: "全球化的互联网科技公司"
});

CREATE (c4:Company {
    id: 4,
    name: "美团",
    industry: "生活服务",
    size: "大型",
    location: "北京",
    website: "https://www.meituan.com",
    description: "中国领先的生活服务电子商务平台"
});

// ==============================================
// 创建职位节点
// ==============================================

// 创建职位节点
CREATE (j1:Job {
    id: 1,
    title: "前端开发工程师",
    location: "深圳",
    salary_min: 15000,
    salary_max: 25000,
    job_type: "full_time",
    experience_level: "mid",
    remote_option: "hybrid",
    description: "负责前端开发工作"
});

CREATE (j2:Job {
    id: 2,
    title: "Java开发工程师",
    location: "杭州",
    salary_min: 18000,
    salary_max: 28000,
    job_type: "full_time",
    experience_level: "senior",
    remote_option: "no",
    description: "负责Java后端开发工作"
});

CREATE (j3:Job {
    id: 3,
    title: "Python开发工程师",
    location: "北京",
    salary_min: 16000,
    salary_max: 26000,
    job_type: "full_time",
    experience_level: "mid",
    remote_option: "full_remote",
    description: "负责Python开发工作"
});

CREATE (j4:Job {
    id: 4,
    title: "全栈开发工程师",
    location: "北京",
    salary_min: 20000,
    salary_max: 30000,
    job_type: "full_time",
    experience_level: "senior",
    remote_option: "hybrid",
    description: "负责全栈开发工作"
});

// ==============================================
// 创建用户-技能关系
// ==============================================

// 用户1的技能关系
CREATE (u1)-[:HAS_SKILL {
    proficiency: "expert",
    years: 5,
    last_used: "2025-01-01",
    is_verified: true
}]->(s1);

CREATE (u1)-[:HAS_SKILL {
    proficiency: "advanced",
    years: 3,
    last_used: "2025-01-01",
    is_verified: true
}]->(s2);

CREATE (u1)-[:HAS_SKILL {
    proficiency: "intermediate",
    years: 2,
    last_used: "2024-12-01",
    is_verified: false
}]->(s3);

// 用户2的技能关系
CREATE (u2)-[:HAS_SKILL {
    proficiency: "advanced",
    years: 3,
    last_used: "2025-01-02",
    is_verified: true
}]->(s5);

CREATE (u2)-[:HAS_SKILL {
    proficiency: "intermediate",
    years: 2,
    last_used: "2025-01-02",
    is_verified: true
}]->(s6);

CREATE (u2)-[:HAS_SKILL {
    proficiency: "intermediate",
    years: 2,
    last_used: "2025-01-02",
    is_verified: true
}]->(s7);

// 用户3的技能关系
CREATE (u3)-[:HAS_SKILL {
    proficiency: "expert",
    years: 7,
    last_used: "2025-01-03",
    is_verified: true
}]->(s4);

CREATE (u3)-[:HAS_SKILL {
    proficiency: "advanced",
    years: 4,
    last_used: "2025-01-03",
    is_verified: true
}]->(s1);

CREATE (u3)-[:HAS_SKILL {
    proficiency: "intermediate",
    years: 2,
    last_used: "2024-11-01",
    is_verified: false
}]->(s8);

// ==============================================
// 创建用户-公司关系
// ==============================================

// 用户1的工作经历
CREATE (u1)-[:WORKED_AT {
    position: "前端开发工程师",
    start_date: "2020-01-01",
    end_date: "2023-12-31",
    is_current: false
}]->(c1);

// 用户2的工作经历
CREATE (u2)-[:WORKED_AT {
    position: "Java开发工程师",
    start_date: "2021-06-01",
    end_date: "2024-12-31",
    is_current: true
}]->(c2);

// 用户3的工作经历
CREATE (u3)-[:WORKED_AT {
    position: "Python开发工程师",
    start_date: "2018-03-01",
    end_date: "2024-06-30",
    is_current: false
}]->(c3);

CREATE (u3)-[:WORKED_AT {
    position: "全栈开发工程师",
    start_date: "2024-07-01",
    end_date: null,
    is_current: true
}]->(c4);

// ==============================================
// 创建公司-职位关系
// ==============================================

// 腾讯科技的职位
CREATE (c1)-[:OFFERS]->(j1);

// 阿里巴巴的职位
CREATE (c2)-[:OFFERS]->(j2);

// 字节跳动的职位
CREATE (c3)-[:OFFERS]->(j3);

// 美团的职位
CREATE (c4)-[:OFFERS]->(j4);

// ==============================================
// 创建职位-技能关系
// ==============================================

// 前端开发工程师职位要求
CREATE (j1)-[:REQUIRES {
    level: "advanced",
    weight: 0.9,
    is_required: true
}]->(s1);

CREATE (j1)-[:REQUIRES {
    level: "intermediate",
    weight: 0.8,
    is_required: true
}]->(s2);

CREATE (j1)-[:REQUIRES {
    level: "basic",
    weight: 0.6,
    is_required: false
}]->(s3);

// Java开发工程师职位要求
CREATE (j2)-[:REQUIRES {
    level: "advanced",
    weight: 0.9,
    is_required: true
}]->(s5);

CREATE (j2)-[:REQUIRES {
    level: "intermediate",
    weight: 0.8,
    is_required: true
}]->(s6);

CREATE (j2)-[:REQUIRES {
    level: "intermediate",
    weight: 0.7,
    is_required: true
}]->(s7);

// Python开发工程师职位要求
CREATE (j3)-[:REQUIRES {
    level: "advanced",
    weight: 0.9,
    is_required: true
}]->(s4);

CREATE (j3)-[:REQUIRES {
    level: "intermediate",
    weight: 0.7,
    is_required: false
}]->(s1);

CREATE (j3)-[:REQUIRES {
    level: "basic",
    weight: 0.5,
    is_required: false
}]->(s8);

// 全栈开发工程师职位要求
CREATE (j4)-[:REQUIRES {
    level: "advanced",
    weight: 0.9,
    is_required: true
}]->(s1);

CREATE (j4)-[:REQUIRES {
    level: "advanced",
    weight: 0.9,
    is_required: true
}]->(s5);

CREATE (j4)-[:REQUIRES {
    level: "intermediate",
    weight: 0.8,
    is_required: true
}]->(s7);

// ==============================================
// 创建用户-用户关系
// ==============================================

// 用户1和用户2是同事关系
CREATE (u1)-[:COLLEAGUE {
    company: "腾讯科技",
    period: "2020-2023",
    relationship_type: "former_colleague"
}]->(u2);

// 用户2和用户3是朋友关系
CREATE (u2)-[:FRIEND {
    relationship_type: "professional",
    connection_strength: 0.8
}]->(u3);

// ==============================================
// 创建技能-技能关系
// ==============================================

// 相关技能关系
CREATE (s1)-[:RELATED_TO {
    relationship_type: "complementary",
    strength: 0.8
}]->(s2);

CREATE (s1)-[:RELATED_TO {
    relationship_type: "complementary",
    strength: 0.7
}]->(s3);

CREATE (s5)-[:RELATED_TO {
    relationship_type: "complementary",
    strength: 0.9
}]->(s6);

CREATE (s4)-[:RELATED_TO {
    relationship_type: "alternative",
    strength: 0.6
}]->(s1);

// ==============================================
// 创建公司-公司关系
// ==============================================

// 竞争关系
CREATE (c1)-[:COMPETES_WITH {
    competition_level: "high",
    market_overlap: 0.8
}]->(c3);

CREATE (c2)-[:COMPETES_WITH {
    competition_level: "medium",
    market_overlap: 0.5
}]->(c4);

// 合作关系
CREATE (c1)-[:PARTNERS_WITH {
    partnership_type: "technology",
    strength: 0.7
}]->(c2);

// ==============================================
// 创建推荐关系
// ==============================================

// 用户1的职位推荐
CREATE (u1)-[:RECOMMENDED_FOR {
    score: 0.95,
    reason: "技能匹配度高",
    generated_at: "2025-01-06T10:00:00Z"
}]->(j1);

CREATE (u1)-[:RECOMMENDED_FOR {
    score: 0.85,
    reason: "技能部分匹配",
    generated_at: "2025-01-06T10:00:00Z"
}]->(j4);

// 用户2的职位推荐
CREATE (u2)-[:RECOMMENDED_FOR {
    score: 0.92,
    reason: "技能完全匹配",
    generated_at: "2025-01-06T10:00:00Z"
}]->(j2);

// 用户3的职位推荐
CREATE (u3)-[:RECOMMENDED_FOR {
    score: 0.88,
    reason: "技能匹配度高",
    generated_at: "2025-01-06T10:00:00Z"
}]->(j3);

CREATE (u3)-[:RECOMMENDED_FOR {
    score: 0.90,
    reason: "全栈技能匹配",
    generated_at: "2025-01-06T10:00:00Z"
}]->(j4);

// ==============================================
// 创建职业路径关系
// ==============================================

// 技能发展路径
CREATE (s1)-[:LEADS_TO {
    difficulty: "medium",
    time_required: "6个月"
}]->(s2);

CREATE (s2)-[:LEADS_TO {
    difficulty: "medium",
    time_required: "3个月"
}]->(s3);

CREATE (s5)-[:LEADS_TO {
    difficulty: "high",
    time_required: "12个月"
}]->(s6);

// ==============================================
// 创建行业关系
// ==============================================

// 创建行业节点
CREATE (i1:Industry {
    name: "互联网",
    description: "互联网科技行业",
    growth_rate: 0.15
});

CREATE (i2:Industry {
    name: "电商",
    description: "电子商务行业",
    growth_rate: 0.12
});

CREATE (i3:Industry {
    name: "生活服务",
    description: "生活服务行业",
    growth_rate: 0.08
});

// 公司-行业关系
CREATE (c1)-[:BELONGS_TO]->(i1);
CREATE (c2)-[:BELONGS_TO]->(i2);
CREATE (c3)-[:BELONGS_TO]->(i1);
CREATE (c4)-[:BELONGS_TO]->(i3);

// ==============================================
// 创建地理位置关系
// ==============================================

// 创建城市节点
CREATE (city1:City {
    name: "北京",
    province: "北京",
    population: 21540000,
    gdp: 36102.6
});

CREATE (city2:City {
    name: "上海",
    province: "上海",
    population: 24280000,
    gdp: 38700.6
});

CREATE (city3:City {
    name: "深圳",
    province: "广东",
    population: 17560000,
    gdp: 27670.2
});

CREATE (city4:City {
    name: "杭州",
    province: "浙江",
    population: 11940000,
    gdp: 16106.0
});

// 用户-城市关系
CREATE (u1)-[:LIVES_IN]->(city1);
CREATE (u2)-[:LIVES_IN]->(city2);
CREATE (u3)-[:LIVES_IN]->(city3);

// 公司-城市关系
CREATE (c1)-[:LOCATED_IN]->(city3);
CREATE (c2)-[:LOCATED_IN]->(city4);
CREATE (c3)-[:LOCATED_IN]->(city1);
CREATE (c4)-[:LOCATED_IN]->(city1);

// 职位-城市关系
CREATE (j1)-[:LOCATED_IN]->(city3);
CREATE (j2)-[:LOCATED_IN]->(city4);
CREATE (j3)-[:LOCATED_IN]->(city1);
CREATE (j4)-[:LOCATED_IN]->(city1);

// ==============================================
// 创建升级完成标记
// ==============================================

// 创建升级记录节点
CREATE (upgrade:UpgradeRecord {
    version: "V4.0",
    completed_at: "2025-01-06T13:00:00Z",
    description: "Neo4j图数据库升级完成，构建了完整的关系网络"
});

// ==============================================
// 升级完成
// ==============================================

// 显示升级完成信息
RETURN "Neo4j图数据库升级完成！" as message,
       "V3.0 -> V4.0" as version,
       datetime() as upgrade_time,
       "构建了完整的关系网络，支持智能推荐和网络分析" as description;
