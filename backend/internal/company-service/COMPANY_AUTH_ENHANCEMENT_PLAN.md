# Company服务认证机制改进方案

## 📋 方案概述

### 问题背景
当前Company服务的认证机制存在以下问题：
1. **权限模型过于简单**：只有 `CreatedBy` 字段，无法支持多用户管理
2. **业务逻辑缺陷**：统一社会信用代码是公开信息，不适合作为授权凭证
3. **安全性问题**：任何人都可以通过 `company_id` 访问企业信息
4. **缺少企业角色管理**：无法支持企业内部权限分级和授权委托

### 改进目标
1. **建立完整的企业权限管理体系**
2. **支持法定代表人、经办人等业务角色**
3. **实现基于企业角色的权限控制**
4. **支持企业委托第三方处理业务**

## 🏗️ 技术架构设计

### 1. 数据库结构改进

#### 1.1 扩展Company表结构
```sql
-- 添加新字段到Company表
ALTER TABLE companies 
ADD COLUMN unified_social_credit_code VARCHAR(50) UNIQUE,
ADD COLUMN legal_representative VARCHAR(100),
ADD COLUMN legal_representative_id VARCHAR(50),
ADD COLUMN legal_rep_user_id INT,
ADD COLUMN authorized_users JSON;
```

#### 1.2 创建企业用户关联表
```sql
-- 创建企业用户关联表
CREATE TABLE company_users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    company_id INT NOT NULL,
    user_id INT NOT NULL,
    role VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    UNIQUE KEY unique_company_user (company_id, user_id)
);

-- 创建索引
CREATE INDEX idx_company_users_company_id ON company_users(company_id);
CREATE INDEX idx_company_users_user_id ON company_users(user_id);
CREATE INDEX idx_company_users_role ON company_users(role);
```

### 2. 数据模型设计

#### 2.1 扩展Company结构体
```go
type Company struct {
    ID                uint      `json:"id" gorm:"primaryKey"`
    Name              string    `json:"name" gorm:"size:200;not null"`
    ShortName         string    `json:"short_name" gorm:"size:100"`
    LogoURL           string    `json:"logo_url" gorm:"size:500"`
    Industry          string    `json:"industry" gorm:"size:100"`
    CompanySize       string    `json:"company_size" gorm:"size:50"`
    Location          string    `json:"location" gorm:"size:200"`
    Website           string    `json:"website" gorm:"size:200"`
    Description       string    `json:"description" gorm:"type:text"`
    FoundedYear       int       `json:"founded_year"`
    
    // 企业认证信息
    UnifiedSocialCreditCode string `json:"unified_social_credit_code" gorm:"size:50;uniqueIndex"`
    LegalRepresentative     string `json:"legal_representative" gorm:"size:100"`
    LegalRepresentativeID   string `json:"legal_representative_id" gorm:"size:50"` // 身份证号
    
    // 权限管理字段
    CreatedBy         uint      `json:"created_by" gorm:"not null"`           // 创建者
    LegalRepUserID    uint      `json:"legal_rep_user_id"`                    // 法定代表人用户ID
    AuthorizedUsers   string    `json:"authorized_users" gorm:"type:json"`    // 授权用户列表
    
    Status            string    `json:"status" gorm:"size:20;default:pending"`
    VerificationLevel string    `json:"verification_level" gorm:"size:20;default:unverified"`
    JobCount          int       `json:"job_count" gorm:"default:0"`
    ViewCount         int       `json:"view_count" gorm:"default:0"`
    CreatedAt         time.Time `json:"created_at"`
    UpdatedAt         time.Time `json:"updated_at"`
}
```

#### 2.2 创建企业用户关联模型
```go
type CompanyUser struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    CompanyID uint      `json:"company_id" gorm:"not null"`
    UserID    uint      `json:"user_id" gorm:"not null"`
    Role      string    `json:"role" gorm:"size:50;not null"` // legal_rep, authorized_user, admin
    Status    string    `json:"status" gorm:"size:20;default:active"` // active, inactive, pending
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    
    // 外键关联
    Company Company `json:"company" gorm:"foreignKey:CompanyID"`
    User    User    `json:"user" gorm:"foreignKey:UserID"`
}
```

### 3. 权限控制逻辑

#### 3.1 增强权限检查函数
```go
func (api *CompanyProfileAPI) checkCompanyAccess(userID, companyID uint, c *gin.Context) bool {
    // 检查是否为系统管理员
    role := c.GetString("role")
    if role == "admin" || role == "super_admin" {
        return true
    }

    db := api.core.GetDB()
    
    // 检查企业是否存在
    var company Company
    if err := db.First(&company, companyID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "企业不存在"})
        return false
    }

    // 检查用户是否为企业创建者
    if company.CreatedBy == userID {
        return true
    }

    // 检查用户是否为法定代表人
    if company.LegalRepUserID == userID {
        return true
    }

    // 检查用户是否在授权用户列表中
    var companyUser CompanyUser
    if err := db.Where("company_id = ? AND user_id = ? AND status = ?", 
        companyID, userID, "active").First(&companyUser).Error; err == nil {
        return true
    }

    // 检查授权用户JSON字段
    if company.AuthorizedUsers != "" {
        var authorizedUsers []uint
        if err := json.Unmarshal([]byte(company.AuthorizedUsers), &authorizedUsers); err == nil {
            for _, authorizedUserID := range authorizedUsers {
                if authorizedUserID == userID {
                    return true
                }
            }
        }
    }

    c.JSON(http.StatusForbidden, gin.H{"error": "权限不足，您没有访问该企业的权限"})
    return false
}
```

#### 3.2 权限检查优先级
1. **系统管理员** (`admin`, `super_admin`) - 最高权限
2. **企业创建者** (`CreatedBy`) - 企业所有者权限
3. **法定代表人** (`LegalRepUserID`) - 企业法人权限
4. **授权用户** (`CompanyUser` 表) - 企业委托权限
5. **JSON授权用户** (`AuthorizedUsers` 字段) - 临时授权权限

### 4. API接口设计

#### 4.1 企业授权管理API
```go
// 企业授权管理API路由组
auth := api.Group("/api/v1/company/auth")
auth.Use(authMiddleware)
{
    // 添加授权用户
    auth.POST("/users", api.addAuthorizedUser)
    
    // 获取企业授权用户列表
    auth.GET("/users/:company_id", api.getAuthorizedUsers)
    
    // 移除授权用户
    auth.DELETE("/users/:company_id/:user_id", api.removeAuthorizedUser)
    
    // 更新用户角色
    auth.PUT("/users/:company_id/:user_id", api.updateUserRole)
    
    // 设置法定代表人
    auth.PUT("/legal-rep/:company_id", api.setLegalRepresentative)
}
```

#### 4.2 核心API实现
```go
// 添加授权用户
func (api *CompanyProfileAPI) addAuthorizedUser(c *gin.Context) {
    userIDInterface, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
        return
    }
    userID := userIDInterface.(uint)

    var req struct {
        CompanyID uint   `json:"company_id" binding:"required"`
        UserID    uint   `json:"user_id" binding:"required"`
        Role      string `json:"role" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 检查权限：只有企业创建者或法定代表人可以添加授权用户
    if !api.checkCompanyAccess(userID, req.CompanyID, c) {
        return
    }

    db := api.core.GetDB()
    
    // 创建企业用户关联
    companyUser := CompanyUser{
        CompanyID: req.CompanyID,
        UserID:    req.UserID,
        Role:      req.Role,
        Status:    "active",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    if err := db.Create(&companyUser).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "添加授权用户失败"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status": "success",
        "message": "授权用户添加成功",
        "data": companyUser,
    })
}
```

### 5. 企业创建流程改进

#### 5.1 企业创建时的权限设置
```go
func createCompany(c *gin.Context) {
    userIDInterface, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
        return
    }
    userID := userIDInterface.(uint)

    var company Company
    if err := c.ShouldBindJSON(&company); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 设置创建者
    company.CreatedBy = userID
    
    // 如果提供了法定代表人信息，设置法定代表人用户ID
    if company.LegalRepUserID == 0 {
        company.LegalRepUserID = userID // 默认创建者为法定代表人
    }

    db := core.GetDB()
    
    // 创建企业
    if err := db.Create(&company).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "创建企业失败"})
        return
    }

    // 创建企业用户关联记录
    companyUser := CompanyUser{
        CompanyID: company.ID,
        UserID:    userID,
        Role:      "legal_rep", // 创建者默认为法定代表人
        Status:    "active",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    if err := db.Create(&companyUser).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "创建企业用户关联失败"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status": "success",
        "message": "企业创建成功",
        "data": company,
    })
}
```

## 🔄 实施计划

### 阶段1：数据库结构升级 (1-2天)
1. **创建数据库迁移脚本**
2. **执行数据库结构升级**
3. **数据迁移和验证**

### 阶段2：代码结构改进 (2-3天)
1. **更新Company结构体**
2. **创建CompanyUser模型**
3. **实现新的权限检查逻辑**

### 阶段3：API接口开发 (2-3天)
1. **实现企业授权管理API**
2. **更新现有API的权限检查**
3. **添加企业创建流程改进**

### 阶段4：测试和验证 (1-2天)
1. **单元测试**
2. **集成测试**
3. **权限控制测试**

### 阶段5：部署和监控 (1天)
1. **生产环境部署**
2. **监控和日志**
3. **性能优化**

## 📊 预期效果

### 1. 安全性提升
- **基于企业角色的权限控制**
- **支持企业内部权限分级**
- **防止未授权访问**

### 2. 业务逻辑完善
- **支持法定代表人管理**
- **支持授权用户管理**
- **支持企业委托第三方处理业务**

### 3. 扩展性增强
- **支持复杂的企业组织结构**
- **支持企业权限的动态管理**
- **支持企业间的业务协作**

## ⚠️ 风险评估

### 1. 数据迁移风险
- **现有数据兼容性**
- **数据完整性验证**
- **回滚方案准备**

### 2. 性能影响
- **权限检查性能**
- **数据库查询优化**
- **缓存策略设计**

### 3. 业务连续性
- **API兼容性**
- **渐进式升级**
- **用户培训**

## 🔧 技术细节

### 1. 数据库索引优化
```sql
-- 复合索引优化查询性能
CREATE INDEX idx_company_users_company_user ON company_users(company_id, user_id, status);
CREATE INDEX idx_company_users_user_company ON company_users(user_id, company_id, role);
```

### 2. 缓存策略
```go
// 企业权限缓存
type CompanyPermissionCache struct {
    UserID    uint
    CompanyID uint
    Permissions []string
    ExpiresAt time.Time
}
```

### 3. 日志记录
```go
// 权限检查日志
type PermissionCheckLog struct {
    UserID    uint
    CompanyID uint
    Action    string
    Result    bool
    Timestamp time.Time
}
```

## 📝 总结

本改进方案通过引入**经办人**和**法定代表人**等业务角色，建立了完整的企业权限管理体系，解决了统一社会信用代码不适合作为授权凭证的问题。方案具有以下特点：

1. **安全性**：基于企业角色的权限控制
2. **灵活性**：支持多种授权方式
3. **扩展性**：支持复杂的企业组织结构
4. **兼容性**：渐进式升级，保证业务连续性

该方案为Company服务提供了完整的认证和授权机制，为后续的业务整合和优化奠定了坚实的基础。

## 🗺️ 多数据库架构集成方案

### 数据边界定义

#### 1. MySQL - 核心业务数据存储
```sql
-- 职责：存储核心业务实体和关系数据
-- 特点：ACID事务、强一致性、结构化数据

-- 企业基础信息表（扩展版）
CREATE TABLE companies (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(200) NOT NULL,
    unified_social_credit_code VARCHAR(50) UNIQUE,
    legal_representative VARCHAR(100),
    legal_representative_id VARCHAR(50),
    industry VARCHAR(100),
    company_size VARCHAR(50),
    location VARCHAR(200),  -- 基础地理位置信息
    
    -- 北斗地理位置信息
    bd_latitude DECIMAL(10,8),      -- 北斗纬度
    bd_longitude DECIMAL(11,8),     -- 北斗经度
    bd_altitude DECIMAL(8,2),       -- 北斗海拔
    bd_accuracy DECIMAL(6,2),       -- 定位精度(米)
    bd_timestamp BIGINT,            -- 定位时间戳
    
    -- 解析后的地址信息
    address VARCHAR(500),           -- 详细地址
    city VARCHAR(100),              -- 城市
    district VARCHAR(100),          -- 区县
    area VARCHAR(100),              -- 区域/街道
    postal_code VARCHAR(20),        -- 邮政编码
    
    -- 地理位置层级编码
    city_code VARCHAR(20),          -- 城市编码
    district_code VARCHAR(20),      -- 区县编码
    area_code VARCHAR(20),          -- 区域编码
    
    status VARCHAR(20) DEFAULT 'active',
    created_by INT NOT NULL,
    legal_rep_user_id INT,          -- 法定代表人用户ID
    authorized_users JSON,          -- 授权用户列表
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

#### 2. PostgreSQL - 向量数据库（职位描述和AI分析）
```sql
-- 职责：存储职位描述向量、AI分析结果、语义搜索
-- 特点：向量相似度搜索、全文搜索、AI模型支持

-- 职位描述表
CREATE TABLE job_descriptions (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL,
    job_title VARCHAR(200) NOT NULL,
    job_description TEXT NOT NULL,
    requirements TEXT,
    location VARCHAR(200),
    salary_range VARCHAR(100),
    job_type VARCHAR(50),  -- full_time, part_time, contract
    experience_level VARCHAR(50),  -- entry, mid, senior, executive
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 职位描述向量表
CREATE TABLE job_description_vectors (
    id SERIAL PRIMARY KEY,
    job_description_id INT NOT NULL,
    vector_data VECTOR(1536),  -- OpenAI embedding维度
    vector_type VARCHAR(50),   -- title, description, requirements
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (job_description_id) REFERENCES job_descriptions(id)
);

-- 创建向量索引
CREATE INDEX ON job_description_vectors USING ivfflat (vector_data vector_cosine_ops);
```

#### 3. Neo4j - 图数据库（地理位置和关系网络）
```cypher
// 职责：存储地理位置关系、企业网络关系、复杂关系分析
// 特点：图遍历、关系分析、地理位置计算

// 地理位置节点
CREATE (city:City {
    name: "北京市",
    code: "110000",
    bd_latitude: 39.9042,
    bd_longitude: 116.4074,
    level: "city"
})

CREATE (district:District {
    name: "海淀区", 
    code: "110108",
    bd_latitude: 39.9593,
    bd_longitude: 116.2983,
    level: "district"
})

CREATE (area:Area {
    name: "中关村",
    code: "110108001", 
    bd_latitude: 39.9836,
    bd_longitude: 116.3164,
    level: "area"
})

// 企业地理位置节点
CREATE (company:Company {
    id: 1,
    name: "某某科技有限公司",
    unified_social_credit_code: "91110000123456789X",
    legal_representative: "张三",
    industry: "软件和信息技术服务业",
    bd_latitude: 39.9836,
    bd_longitude: 116.3164,
    bd_altitude: 50.0,
    bd_accuracy: 3.0,
    bd_timestamp: 1695123456789,
    address: "北京市海淀区中关村大街1号",
    city: "北京市",
    district: "海淀区",
    area: "中关村"
})

// 地理位置关系
CREATE (city)-[:CONTAINS]->(district)
CREATE (district)-[:CONTAINS]->(area)
CREATE (company)-[:LOCATED_IN]->(area)
CREATE (company)-[:IN_DISTRICT]->(district)
CREATE (company)-[:IN_CITY]->(city)

// 企业网络关系
CREATE (company1:Company {id: 1, name: "公司A"})
CREATE (company2:Company {id: 2, name: "公司B"})
CREATE (company1)-[:BUSINESS_PARTNER {
    relationship_type: "合作",
    start_date: "2023-01-01",
    business_area: "技术合作"
}]->(company2)
```

### 数据同步与一致性策略

#### 数据同步服务
```go
// 数据同步服务
type DataSyncService struct {
    mysqlDB    *gorm.DB
    postgresDB *gorm.DB
    neo4jDriver neo4j.Driver
    redisClient *redis.Client
}

// 企业数据同步
func (s *DataSyncService) SyncCompanyData(companyID uint) error {
    // 1. 从MySQL获取核心企业数据
    var company Company
    if err := s.mysqlDB.First(&company, companyID).Error; err != nil {
        return err
    }

    // 2. 同步到PostgreSQL（职位相关数据）
    if err := s.syncToPostgreSQL(company); err != nil {
        return err
    }

    // 3. 同步到Neo4j（地理位置和关系数据）
    if err := s.syncToNeo4j(company); err != nil {
        return err
    }

    return nil
}
```

### 地理位置查询API

#### 基于职位需求的智能企业匹配
```go
// 基于职位需求查找匹配企业
func (api *CompanyLocationAPI) FindMatchingCompaniesByJob(jobQuery JobSearchQuery) ([]CompanyJobMatch, error) {
    session := api.neo4jService.driver.NewSession(neo4j.SessionConfig{})
    defer session.Close()

    // 构建复杂的匹配查询
    query := `
    MATCH (u:User {id: $userID})
    MATCH (c:Company)
    MATCH (c)-[:HAS_JOB]->(j:Job)
    WHERE j.job_title CONTAINS $jobTitle
       OR j.job_description CONTAINS $jobTitle
       OR j.requirements CONTAINS $jobTitle
    WITH u, c, j,
         distance(point({latitude: u.bd_latitude, longitude: u.bd_longitude}),
                  point({latitude: c.bd_latitude, longitude: c.bd_longitude})) as distance,
         // 薪资匹配度计算
         CASE 
           WHEN j.salary_min <= $expectedSalary AND j.salary_max >= $expectedSalary THEN 1.0
           WHEN j.salary_min <= $expectedSalary * 1.2 AND j.salary_max >= $expectedSalary * 0.8 THEN 0.8
           WHEN j.salary_min <= $expectedSalary * 1.5 AND j.salary_max >= $expectedSalary * 0.6 THEN 0.6
           ELSE 0.3
         END as salaryMatch,
         // 经验匹配度计算
         CASE 
           WHEN j.experience_level = $experienceLevel THEN 1.0
           WHEN j.experience_level = 'mid' AND $experienceLevel = 'senior' THEN 0.8
           WHEN j.experience_level = 'senior' AND $experienceLevel = 'mid' THEN 0.7
           ELSE 0.4
         END as experienceMatch,
         // 地理位置匹配度计算
         CASE 
           WHEN distance <= 5 THEN 1.0
           WHEN distance <= 10 THEN 0.8
           WHEN distance <= 20 THEN 0.6
           WHEN distance <= 50 THEN 0.4
           ELSE 0.2
         END as locationMatch
    WHERE distance <= $maxDistance
      AND salaryMatch >= $minSalaryMatch
      AND experienceMatch >= $minExperienceMatch
      AND locationMatch >= $minLocationMatch
    WITH c, j, distance, salaryMatch, experienceMatch, locationMatch,
         // 综合匹配度计算
         (salaryMatch * 0.4 + experienceMatch * 0.3 + locationMatch * 0.3) as overallMatch
    ORDER BY overallMatch DESC, distance ASC
    LIMIT $limit
    RETURN c.id as company_id, c.name as company_name, c.industry, c.bd_latitude, c.bd_longitude,
           j.id as job_id, j.job_title, j.salary_min, j.salary_max, j.experience_level,
           distance, salaryMatch, experienceMatch, locationMatch, overallMatch
    `
    
    result, err := session.Run(query, map[string]interface{}{
        "userID":            jobQuery.UserID,
        "jobTitle":          jobQuery.JobTitle,
        "expectedSalary":    jobQuery.ExpectedSalary,
        "experienceLevel":   jobQuery.ExperienceLevel,
        "maxDistance":       jobQuery.MaxDistance,
        "minSalaryMatch":    jobQuery.MinSalaryMatch,
        "minExperienceMatch": jobQuery.MinExperienceMatch,
        "minLocationMatch":  jobQuery.MinLocationMatch,
        "limit":             jobQuery.Limit,
    })
    
    if err != nil {
        return nil, err
    }
    
    var matches []CompanyJobMatch
    for result.Next() {
        record := result.Record()
        match := CompanyJobMatch{
            CompanyID:        record.Values[0].(int64),
            CompanyName:      record.Values[1].(string),
            Industry:         record.Values[2].(string),
            Latitude:         record.Values[3].(float64),
            Longitude:        record.Values[4].(float64),
            JobID:            record.Values[5].(int64),
            JobTitle:         record.Values[6].(string),
            SalaryMin:        record.Values[7].(int64),
            SalaryMax:        record.Values[8].(int64),
            ExperienceLevel:  record.Values[9].(string),
            Distance:         record.Values[10].(float64),
            SalaryMatch:      record.Values[11].(float64),
            ExperienceMatch:  record.Values[12].(float64),
            LocationMatch:    record.Values[13].(float64),
            OverallMatch:     record.Values[14].(float64),
            MatchedAt:        time.Now(),
        }
        matches = append(matches, match)
    }
    
    return matches, nil
}

// 职位搜索查询结构
type JobSearchQuery struct {
    UserID            uint    `json:"user_id"`
    JobTitle          string  `json:"job_title"`           // 职位名称
    ExpectedSalary    int     `json:"expected_salary"`     // 期望薪资
    ExperienceLevel   string  `json:"experience_level"`    // 经验级别
    MaxDistance       float64 `json:"max_distance"`        // 最大距离(公里)
    MinSalaryMatch    float64 `json:"min_salary_match"`    // 最低薪资匹配度
    MinExperienceMatch float64 `json:"min_experience_match"` // 最低经验匹配度
    MinLocationMatch  float64 `json:"min_location_match"`  // 最低地理位置匹配度
    Limit             int     `json:"limit"`               // 返回结果数量限制
}

// 企业职位匹配结果
type CompanyJobMatch struct {
    CompanyID       int64     `json:"company_id"`
    CompanyName     string    `json:"company_name"`
    Industry        string    `json:"industry"`
    Latitude        float64   `json:"latitude"`
    Longitude       float64   `json:"longitude"`
    JobID           int64     `json:"job_id"`
    JobTitle        string    `json:"job_title"`
    SalaryMin       int64     `json:"salary_min"`
    SalaryMax       int64     `json:"salary_max"`
    ExperienceLevel string    `json:"experience_level"`
    Distance        float64   `json:"distance"`        // 距离(公里)
    SalaryMatch     float64   `json:"salary_match"`    // 薪资匹配度(0-1)
    ExperienceMatch float64   `json:"experience_match"` // 经验匹配度(0-1)
    LocationMatch   float64   `json:"location_match"`  // 地理位置匹配度(0-1)
    OverallMatch    float64   `json:"overall_match"`   // 综合匹配度(0-1)
    MatchedAt       time.Time `json:"matched_at"`
}

// 基于技能匹配的企业推荐
func (api *CompanyLocationAPI) FindCompaniesBySkills(userID uint, skills []string, radius float64) ([]CompanySkillMatch, error) {
    session := api.neo4jService.driver.NewSession(neo4j.SessionConfig{})
    defer session.Close()

    query := `
    MATCH (u:User {id: $userID})
    MATCH (c:Company)
    MATCH (c)-[:HAS_JOB]->(j:Job)
    MATCH (j)-[:REQUIRES_SKILL]->(s:Skill)
    WHERE s.name IN $skills
    WITH u, c, j, s,
         distance(point({latitude: u.bd_latitude, longitude: u.bd_longitude}),
                  point({latitude: c.bd_latitude, longitude: c.bd_longitude})) as distance,
         // 技能匹配度计算
         size([skill IN $skills WHERE skill IN j.required_skills]) * 1.0 / size($skills) as skillMatch
    WHERE distance <= $radius
    WITH c, j, distance, skillMatch,
         // 综合评分：技能匹配度 * 0.6 + 地理位置匹配度 * 0.4
         (skillMatch * 0.6 + 
          CASE 
            WHEN distance <= 5 THEN 1.0
            WHEN distance <= 10 THEN 0.8
            WHEN distance <= 20 THEN 0.6
            ELSE 0.4
          END * 0.4) as overallScore
    ORDER BY overallScore DESC, distance ASC
    LIMIT 20
    RETURN c.id, c.name, c.industry, j.id, j.job_title, j.salary_min, j.salary_max,
           distance, skillMatch, overallScore
    `
    
    result, err := session.Run(query, map[string]interface{}{
        "userID":  userID,
        "skills":  skills,
        "radius":  radius,
    })
    
    if err != nil {
        return nil, err
    }
    
    var matches []CompanySkillMatch
    for result.Next() {
        record := result.Record()
        match := CompanySkillMatch{
            CompanyID:    record.Values[0].(int64),
            CompanyName:  record.Values[1].(string),
            Industry:     record.Values[2].(string),
            JobID:        record.Values[3].(int64),
            JobTitle:     record.Values[4].(string),
            SalaryMin:    record.Values[5].(int64),
            SalaryMax:    record.Values[6].(int64),
            Distance:     record.Values[7].(float64),
            SkillMatch:   record.Values[8].(float64),
            OverallScore: record.Values[9].(float64),
            MatchedAt:    time.Now(),
        }
        matches = append(matches, match)
    }
    
    return matches, nil
}

// 企业技能匹配结果
type CompanySkillMatch struct {
    CompanyID    int64     `json:"company_id"`
    CompanyName  string    `json:"company_name"`
    Industry     string    `json:"industry"`
    JobID        int64     `json:"job_id"`
    JobTitle     string    `json:"job_title"`
    SalaryMin    int64     `json:"salary_min"`
    SalaryMax    int64     `json:"salary_max"`
    Distance     float64   `json:"distance"`
    SkillMatch   float64   `json:"skill_match"`
    OverallScore float64   `json:"overall_score"`
    MatchedAt    time.Time `json:"matched_at"`
}

// 智能企业推荐（基于用户简历和求职历史）
func (api *CompanyLocationAPI) GetIntelligentCompanyRecommendations(userID uint, limit int) ([]IntelligentRecommendation, error) {
    session := api.neo4jService.driver.NewSession(neo4j.SessionConfig{})
    defer session.Close()

    // 复杂的智能推荐查询
    query := `
    MATCH (u:User {id: $userID})
    MATCH (u)-[:OWNS]->(r:Resume)
    MATCH (c:Company)
    MATCH (c)-[:HAS_JOB]->(j:Job)
    
    // 获取用户技能和经验
    WITH u, c, j, r,
         distance(point({latitude: u.bd_latitude, longitude: u.bd_longitude}),
                  point({latitude: c.bd_latitude, longitude: c.bd_longitude})) as distance,
         // 从简历中提取技能匹配度
         size([skill IN r.skills WHERE skill IN j.required_skills]) * 1.0 / 
         size(j.required_skills) as skillMatch,
         // 经验匹配度
         CASE 
           WHEN r.experience_level = j.experience_level THEN 1.0
           WHEN r.experience_level = 'senior' AND j.experience_level = 'mid' THEN 0.8
           WHEN r.experience_level = 'mid' AND j.experience_level = 'senior' THEN 0.6
           ELSE 0.4
         END as experienceMatch,
         // 行业匹配度
         CASE 
           WHEN r.preferred_industry = c.industry THEN 1.0
           WHEN r.preferred_industry IN c.related_industries THEN 0.8
           ELSE 0.5
         END as industryMatch,
         // 地理位置偏好匹配度
         CASE 
           WHEN distance <= 5 THEN 1.0
           WHEN distance <= 10 THEN 0.9
           WHEN distance <= 20 THEN 0.7
           WHEN distance <= 50 THEN 0.5
           ELSE 0.3
         END as locationMatch,
         // 薪资期望匹配度
         CASE 
           WHEN j.salary_min >= r.expected_salary * 0.8 AND j.salary_max <= r.expected_salary * 1.5 THEN 1.0
           WHEN j.salary_min >= r.expected_salary * 0.6 AND j.salary_max <= r.expected_salary * 2.0 THEN 0.8
           WHEN j.salary_min >= r.expected_salary * 0.4 THEN 0.6
           ELSE 0.3
         END as salaryMatch
    
    // 计算综合推荐分数
    WITH c, j, distance, skillMatch, experienceMatch, industryMatch, locationMatch, salaryMatch,
         (skillMatch * 0.3 + experienceMatch * 0.25 + industryMatch * 0.2 + 
          locationMatch * 0.15 + salaryMatch * 0.1) as recommendationScore
    
    WHERE recommendationScore >= 0.6
      AND distance <= 50  // 50公里范围内
    
    ORDER BY recommendationScore DESC, distance ASC
    LIMIT $limit
    
    RETURN c.id as company_id, c.name as company_name, c.industry, c.bd_latitude, c.bd_longitude,
           j.id as job_id, j.job_title, j.salary_min, j.salary_max, j.experience_level,
           distance, skillMatch, experienceMatch, industryMatch, locationMatch, salaryMatch, recommendationScore
    `
    
    result, err := session.Run(query, map[string]interface{}{
        "userID": userID,
        "limit":  limit,
    })
    
    if err != nil {
        return nil, err
    }
    
    var recommendations []IntelligentRecommendation
    for result.Next() {
        record := result.Record()
        recommendation := IntelligentRecommendation{
            CompanyID:           record.Values[0].(int64),
            CompanyName:         record.Values[1].(string),
            Industry:            record.Values[2].(string),
            Latitude:            record.Values[3].(float64),
            Longitude:           record.Values[4].(float64),
            JobID:               record.Values[5].(int64),
            JobTitle:            record.Values[6].(string),
            SalaryMin:           record.Values[7].(int64),
            SalaryMax:           record.Values[8].(int64),
            ExperienceLevel:     record.Values[9].(string),
            Distance:            record.Values[10].(float64),
            SkillMatch:          record.Values[11].(float64),
            ExperienceMatch:     record.Values[12].(float64),
            IndustryMatch:       record.Values[13].(float64),
            LocationMatch:       record.Values[14].(float64),
            SalaryMatch:         record.Values[15].(float64),
            RecommendationScore: record.Values[16].(float64),
            RecommendedAt:       time.Now(),
        }
        recommendations = append(recommendations, recommendation)
    }
    
    return recommendations, nil
}

// 智能推荐结果
type IntelligentRecommendation struct {
    CompanyID           int64     `json:"company_id"`
    CompanyName         string    `json:"company_name"`
    Industry            string    `json:"industry"`
    Latitude            float64   `json:"latitude"`
    Longitude           float64   `json:"longitude"`
    JobID               int64     `json:"job_id"`
    JobTitle            string    `json:"job_title"`
    SalaryMin           int64     `json:"salary_min"`
    SalaryMax           int64     `json:"salary_max"`
    ExperienceLevel     string    `json:"experience_level"`
    Distance            float64   `json:"distance"`
    SkillMatch          float64   `json:"skill_match"`          // 技能匹配度
    ExperienceMatch     float64   `json:"experience_match"`     // 经验匹配度
    IndustryMatch       float64   `json:"industry_match"`       // 行业匹配度
    LocationMatch       float64   `json:"location_match"`       // 地理位置匹配度
    SalaryMatch         float64   `json:"salary_match"`         // 薪资匹配度
    RecommendationScore float64   `json:"recommendation_score"` // 综合推荐分数
    RecommendedAt       time.Time `json:"recommended_at"`
}

// 企业竞争分析（为求职者提供市场洞察）
func (api *CompanyLocationAPI) GetCompanyCompetitionAnalysis(companyID uint, jobTitle string) (*CompetitionAnalysis, error) {
    session := api.neo4jService.driver.NewSession(neo4j.SessionConfig{})
    defer session.Close()

    query := `
    MATCH (targetCompany:Company {id: $companyID})
    MATCH (targetCompany)-[:HAS_JOB]->(targetJob:Job {job_title: $jobTitle})
    MATCH (c:Company)
    MATCH (c)-[:HAS_JOB]->(j:Job {job_title: $jobTitle})
    WHERE c.id <> $companyID
    
    WITH targetCompany, targetJob, c, j,
         distance(point({latitude: targetCompany.bd_latitude, longitude: targetCompany.bd_longitude}),
                  point({latitude: c.bd_latitude, longitude: c.bd_longitude})) as distance
    
    WHERE distance <= 20  // 20公里范围内的竞争企业
    
    WITH targetCompany, targetJob, 
         collect({
             company_id: c.id,
             company_name: c.name,
             industry: c.industry,
             job_id: j.id,
             salary_min: j.salary_min,
             salary_max: j.salary_max,
             distance: distance
         }) as competitors,
         // 计算薪资统计
         avg(j.salary_min) as avgSalaryMin,
         avg(j.salary_max) as avgSalaryMax,
         min(j.salary_min) as minSalary,
         max(j.salary_max) as maxSalary,
         count(c) as competitorCount
    
    RETURN targetCompany.id, targetCompany.name, targetJob.id, targetJob.job_title,
           targetJob.salary_min, targetJob.salary_max,
           competitors, avgSalaryMin, avgSalaryMax, minSalary, maxSalary, competitorCount
    `
    
    result, err := session.Run(query, map[string]interface{}{
        "companyID": companyID,
        "jobTitle":  jobTitle,
    })
    
    if err != nil {
        return nil, err
    }
    
    if !result.Next() {
        return nil, fmt.Errorf("未找到相关数据")
    }
    
    record := result.Record()
    analysis := &CompetitionAnalysis{
        TargetCompanyID:   record.Values[0].(int64),
        TargetCompanyName: record.Values[1].(string),
        TargetJobID:       record.Values[2].(int64),
        TargetJobTitle:    record.Values[3].(string),
        TargetSalaryMin:   record.Values[4].(int64),
        TargetSalaryMax:   record.Values[5].(int64),
        Competitors:       record.Values[6].([]interface{}),
        AvgSalaryMin:      record.Values[7].(float64),
        AvgSalaryMax:      record.Values[8].(float64),
        MinSalary:         record.Values[9].(int64),
        MaxSalary:         record.Values[10].(int64),
        CompetitorCount:   record.Values[11].(int64),
        AnalyzedAt:        time.Now(),
    }
    
    return analysis, nil
}

// 竞争分析结果
type CompetitionAnalysis struct {
    TargetCompanyID   int64         `json:"target_company_id"`
    TargetCompanyName string        `json:"target_company_name"`
    TargetJobID       int64         `json:"target_job_id"`
    TargetJobTitle    string        `json:"target_job_title"`
    TargetSalaryMin   int64         `json:"target_salary_min"`
    TargetSalaryMax   int64         `json:"target_salary_max"`
    Competitors       []interface{} `json:"competitors"`        // 竞争企业列表
    AvgSalaryMin      float64       `json:"avg_salary_min"`     // 平均最低薪资
    AvgSalaryMax      float64       `json:"avg_salary_max"`     // 平均最高薪资
    MinSalary         int64         `json:"min_salary"`         // 市场最低薪资
    MaxSalary         int64         `json:"max_salary"`         // 市场最高薪资
    CompetitorCount   int64         `json:"competitor_count"`   // 竞争企业数量
    AnalyzedAt        time.Time     `json:"analyzed_at"`
}
```

## 🚀 实施计划

### 第一阶段：数据边界定义与基础架构 (2-3天)

#### Day 1: 数据边界设计
- [ ] 编写MySQL职责文档（核心业务数据）
- [ ] 编写PostgreSQL职责文档（向量和AI数据）
- [ ] 编写Neo4j职责文档（地理位置和关系网络）
- [ ] 设计数据同步策略
- [ ] 定义数据一致性检查机制

#### Day 2: 数据模型设计
- [ ] 扩展Company表结构（添加认证字段和地理位置字段）
- [ ] 创建CompanyUser关联表
- [ ] 设计PostgreSQL向量表结构
- [ ] 设计Neo4j节点和关系模型
- [ ] 编写数据迁移脚本

#### Day 3: 基础架构实现
- [ ] 实现DataSyncService
- [ ] 实现数据一致性检查
- [ ] 实现服务间通信机制
- [ ] 添加监控和日志

### 第二阶段：核心服务实现 (3-4天)

#### Day 4: Company服务增强
- [ ] 实现企业认证机制
- [ ] 实现企业CRUD操作
- [ ] 集成数据同步服务
- [ ] 实现企业数据同步到PostgreSQL和Neo4j

#### Day 5: Job服务实现
- [ ] 实现Job服务主程序
- [ ] 实现职位描述CRUD
- [ ] 实现向量搜索功能
- [ ] 集成AI分析功能

#### Day 6: Location服务实现
- [ ] 实现Location服务主程序
- [ ] 实现地理位置节点管理
- [ ] 实现Neo4j图数据库操作
- [ ] 实现地理位置分析功能

#### Day 7: 服务集成与测试
- [ ] 实现服务间集成
- [ ] 编写单元测试
- [ ] 编写集成测试

### 第三阶段：测试与优化 (2-3天)

#### Day 8: 集成测试
- [ ] 测试企业创建流程
- [ ] 测试职位发布流程
- [ ] 测试地理位置分析
- [ ] 测试数据同步机制

#### Day 9: 监控与日志
- [ ] 实现监控系统
- [ ] 实现日志系统
- [ ] 实现告警机制

#### Day 10: 部署与文档
- [ ] 配置Docker容器
- [ ] 编写技术文档
- [ ] 编写部署文档

## 📊 数据流向图

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│      MySQL      │    │   PostgreSQL    │    │     Neo4j       │
│                 │    │                 │    │                 │
│ 核心业务数据     │    │ 向量和AI数据     │    │ 地理位置和关系   │
│ - 企业信息      │    │ - 职位描述      │    │ - 地理位置节点   │
│ - 用户权限      │    │ - 向量嵌入      │    │ - 企业关系网络   │
│ - 业务关系      │    │ - AI分析结果    │    │ - 距离计算      │
│ - 北斗位置      │    │ - 语义搜索      │    │ - 图遍历查询    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   数据同步服务   │
                    │                 │
                    │ - 数据一致性检查 │
                    │ - 增量同步      │
                    │ - 冲突解决      │
                    └─────────────────┘
```

## ⚠️ 风险评估与缓解

### 技术风险
- **数据一致性风险**：通过数据同步服务和一致性检查机制控制
- **性能风险**：通过性能测试和优化控制
- **集成风险**：通过渐进式集成和回滚机制控制

### 进度风险
- **时间风险**：预留20%缓冲时间
- **资源风险**：确保开发人员配置充足
- **依赖风险**：提前识别外部依赖

### 质量风险
- **测试风险**：确保测试覆盖率>80%
- **文档风险**：确保文档及时更新
- **部署风险**：确保部署流程标准化

## 🚨 基于Resume服务经验的风险预警

### 1. **数据一致性问题（高优先级）**
**Resume服务踩坑经验**：MySQL和SQLite之间缺少同步机制，数据更新时容易出现不一致

**Company服务预防措施**：
```go
// 实现企业数据一致性检查机制
type CompanyDataConsistencyChecker struct {
    mysqlDB    *gorm.DB
    postgresDB *gorm.DB
    neo4jDriver neo4j.Driver
    redisClient *redis.Client
}

func (c *CompanyDataConsistencyChecker) CheckCompanyConsistency(companyID uint) error {
    // 1. 检查MySQL核心数据
    var company Company
    if err := c.mysqlDB.First(&company, companyID).Error; err != nil {
        return fmt.Errorf("MySQL数据缺失: %v", err)
    }
    
    // 2. 检查PostgreSQL向量数据
    var jobCount int64
    c.postgresDB.Model(&JobDescription{}).Where("company_id = ?", companyID).Count(&jobCount)
    
    // 3. 检查Neo4j地理位置数据
    session := c.neo4jDriver.NewSession(neo4j.SessionConfig{})
    defer session.Close()
    
    result, err := session.Run("MATCH (c:Company {id: $id}) RETURN c", map[string]interface{}{"id": companyID})
    if err != nil {
        return fmt.Errorf("Neo4j数据检查失败: %v", err)
    }
    
    if !result.Next() {
        return fmt.Errorf("Neo4j数据缺失")
    }
    
    return nil
}
```

### 2. **权限管理复杂性（中优先级）**
**Resume服务踩坑经验**：用户需要在MySQL中注册登记，需要授权订阅管理，权限设定和角色关联复杂

**Company服务优化方案**：
```go
// 统一的企业权限管理
type CompanyPermissionManager struct {
    mysqlDB    *gorm.DB
    redisClient *redis.Client
    cacheTTL   time.Duration
}

func (cpm *CompanyPermissionManager) CheckCompanyAccess(userID uint, companyID uint, action string) error {
    // 1. 尝试从缓存获取权限
    cacheKey := fmt.Sprintf("company_permission:%d:%d:%s", userID, companyID, action)
    if cached, err := cpm.redisClient.Get(cacheKey).Result(); err == nil {
        if cached == "true" {
            return nil
        }
        return fmt.Errorf("权限不足")
    }
    
    // 2. 检查系统管理员权限
    var user User
    if err := cpm.mysqlDB.First(&user, userID).Error; err == nil {
        if user.Role == "admin" || user.Role == "super_admin" {
            cpm.redisClient.Set(cacheKey, "true", cpm.cacheTTL)
            return nil
        }
    }
    
    // 3. 检查企业权限
    var companyUser CompanyUser
    if err := cpm.mysqlDB.Where("company_id = ? AND user_id = ? AND status = ?", 
        companyID, userID, "active").First(&companyUser).Error; err == nil {
        cpm.redisClient.Set(cacheKey, "true", cpm.cacheTTL)
        return nil
    }
    
    cpm.redisClient.Set(cacheKey, "false", cpm.cacheTTL)
    return fmt.Errorf("权限不足")
}
```

### 3. **性能问题（高优先级）**
**Resume服务踩坑经验**：跨数据库查询性能差，SQLite并发访问限制，缺少缓存机制

**Company服务性能优化**：
```go
// 企业数据缓存管理
type CompanyCacheManager struct {
    redisClient *redis.Client
    mysqlDB     *gorm.DB
    postgresDB  *gorm.DB
}

func (ccm *CompanyCacheManager) GetCompanyWithCache(companyID uint) (*Company, error) {
    // 1. 尝试从缓存获取
    cacheKey := fmt.Sprintf("company:%d", companyID)
    if cached, err := ccm.redisClient.Get(cacheKey).Result(); err == nil {
        var company Company
        if err := json.Unmarshal([]byte(cached), &company); err == nil {
            return &company, nil
        }
    }
    
    // 2. 从数据库获取
    var company Company
    if err := ccm.mysqlDB.Preload("CompanyUsers").First(&company, companyID).Error; err != nil {
        return nil, err
    }
    
    // 3. 缓存结果
    companyJSON, _ := json.Marshal(company)
    ccm.redisClient.Set(cacheKey, companyJSON, time.Hour)
    
    return &company, nil
}

// 批量获取企业数据（避免N+1查询）
func (ccm *CompanyCacheManager) GetCompaniesBatch(companyIDs []uint) ([]Company, error) {
    var companies []Company
    if err := ccm.mysqlDB.Where("id IN ?", companyIDs).Find(&companies).Error; err != nil {
        return nil, err
    }
    
    // 批量缓存
    for _, company := range companies {
        cacheKey := fmt.Sprintf("company:%d", company.ID)
        companyJSON, _ := json.Marshal(company)
        ccm.redisClient.Set(cacheKey, companyJSON, time.Hour)
    }
    
    return companies, nil
}
```

### 4. **数据备份和恢复困难（中优先级）**
**Resume服务踩坑经验**：SQLite文件分散，备份复杂，跨数据库事务处理困难

**Company服务备份策略**：
```go
// 统一的企业数据备份机制
type CompanyBackupManager struct {
    mysqlDB    *gorm.DB
    postgresDB *gorm.DB
    neo4jDriver neo4j.Driver
}

func (cbm *CompanyBackupManager) BackupCompanyData(companyID uint) error {
    // 1. 备份MySQL数据
    var company Company
    if err := cbm.mysqlDB.Preload("CompanyUsers").First(&company, companyID).Error; err != nil {
        return err
    }
    
    // 2. 备份PostgreSQL数据
    var jobDescriptions []JobDescription
    if err := cbm.postgresDB.Where("company_id = ?", companyID).Find(&jobDescriptions).Error; err != nil {
        return err
    }
    
    // 3. 备份Neo4j数据
    session := cbm.neo4jDriver.NewSession(neo4j.SessionConfig{})
    defer session.Close()
    
    result, err := session.Run(`
        MATCH (c:Company {id: $id})
        OPTIONAL MATCH (c)-[r]-(related)
        RETURN c, r, related
    `, map[string]interface{}{"id": companyID})
    
    if err != nil {
        return err
    }
    
    // 4. 创建备份文件
    backupData := CompanyBackupData{
        CompanyID:      companyID,
        Company:        company,
        JobDescriptions: jobDescriptions,
        Neo4jData:      extractNeo4jData(result),
        BackupTime:     time.Now(),
    }
    
    return saveCompanyBackup(backupData)
}
```

## 📝 成功标准

### 功能标准
- [ ] 企业认证机制正常工作
- [ ] 职位向量搜索正常工作
- [ ] 地理位置分析正常工作
- [ ] 数据同步机制正常工作
- [ ] 北斗地理位置集成正常工作

### 性能标准
- [ ] 数据库查询响应时间<100ms
- [ ] 向量搜索响应时间<500ms
- [ ] 图查询响应时间<200ms
- [ ] 数据同步延迟<1s

### 质量标准
- [ ] 代码测试覆盖率>80%
- [ ] 文档完整性>90%
- [ ] 系统可用性>99%
- [ ] 错误率<0.1%

---

**文档版本**: v2.0  
**创建时间**: 2025-01-16  
**最后更新**: 2025-01-16  
**状态**: 待实施  
**更新内容**: 新增多数据库架构集成方案、北斗地理位置信息、数据同步策略、实施计划
