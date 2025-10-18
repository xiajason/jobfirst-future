# Company服务企业画像数据库设计文档

**创建日期**: 2025年9月15日  
**版本**: v1.0  
**状态**: 📋 设计完成，准备实施

---

## 📋 设计概述

基于公司画像文档，我们为Company服务设计了完整的企业画像数据库结构，支持PDF解析后的结构化数据存储和管理。

### 🎯 设计目标

1. **完整支持企业画像数据存储**
2. **兼容PDF解析结果存储**
3. **支持复杂JSON数据结构**
4. **保证数据一致性和完整性**
5. **优化查询性能**

---

## 🗄️ 数据库表结构设计

### 1. 企业基本信息表 (company_basic_info)

**用途**: 存储企业的核心基本信息

| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| id | INT | 主键 | PRIMARY KEY, AUTO_INCREMENT |
| company_id | BIGINT UNSIGNED | 企业ID | NOT NULL, FOREIGN KEY |
| report_id | VARCHAR(50) | 报告编号 | UNIQUE INDEX |
| company_name | VARCHAR(255) | 企业名称 | NOT NULL |
| used_name | VARCHAR(255) | 曾用名 | |
| unified_social_credit_code | VARCHAR(50) | 统一社会信用代码 | INDEX |
| registration_date | DATE | 成立日期 | |
| legal_representative | VARCHAR(100) | 法定代表人 | |
| business_status | VARCHAR(50) | 经营状态 | |
| registered_capital | DECIMAL(18,2) | 注册资本（万元） | |
| currency | VARCHAR(20) | 注册币种 | DEFAULT 'CNY' |
| insured_count | INT | 参保人数 | |
| industry_category | VARCHAR(100) | 所属行业类别 | INDEX |
| registration_authority | VARCHAR(255) | 登记机关 | |
| business_scope | TEXT | 经营范围 | |
| tags | JSON | 企业标签 | |
| data_source | VARCHAR(100) | 数据来源 | |
| data_update_time | TIMESTAMP | 数据更新时间 | AUTO UPDATE |
| created_at | TIMESTAMP | 创建时间 | DEFAULT CURRENT_TIMESTAMP |
| updated_at | TIMESTAMP | 更新时间 | AUTO UPDATE |

### 2. 资质许可表 (qualification_license)

**用途**: 存储企业的资质和许可信息

| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| id | INT | 主键 | PRIMARY KEY, AUTO_INCREMENT |
| company_id | BIGINT UNSIGNED | 企业ID | NOT NULL, FOREIGN KEY |
| report_id | VARCHAR(50) | 报告编号 | |
| type | ENUM | 类型 | ('资质', '许可', '备案') |
| name | VARCHAR(255) | 资质/许可名称 | NOT NULL |
| status | VARCHAR(20) | 状态 | DEFAULT '有效' |
| certificate_number | VARCHAR(100) | 证书编号 | INDEX |
| issue_date | DATE | 颁发时间 | |
| issuing_authority | VARCHAR(255) | 颁发单位 | |
| validity_period | DATE | 有效期限 | |
| content | TEXT | 内容描述 | |

### 3. 人员竞争力表 (personnel_competitiveness)

**用途**: 存储企业人员竞争力相关信息

| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| id | INT | 主键 | PRIMARY KEY, AUTO_INCREMENT |
| company_id | BIGINT UNSIGNED | 企业ID | NOT NULL, FOREIGN KEY |
| report_id | VARCHAR(50) | 报告编号 | |
| data_update_date | DATE | 数据更新时间 | |
| total_employees | INT | 企业人数 | |
| industry_ranking | VARCHAR(50) | 行业内排名 | |
| industry_avg_employees | INT | 行业内平均人数 | |
| turnover_rate | DECIMAL(5,2) | 离职率 | |
| entry_rate | DECIMAL(5,2) | 入职率 | |
| tenure_distribution | JSON | 司龄分布 | |

### 4. 公积金信息表 (provident_fund)

**用途**: 存储企业公积金相关信息

| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| id | INT | 主键 | PRIMARY KEY, AUTO_INCREMENT |
| company_id | BIGINT UNSIGNED | 企业ID | NOT NULL, FOREIGN KEY |
| report_id | VARCHAR(50) | 报告编号 | |
| unit_nature | VARCHAR(100) | 单位性质 | |
| opening_date | DATE | 开户日期 | |
| last_payment_month | DATE | 最近缴存月份 | |
| total_payment | VARCHAR(50) | 累计缴存总额 | |
| payment_records | JSON | 月度缴存记录 | |

### 5. 资助补贴表 (subsidy_info)

**用途**: 存储企业资助补贴信息

| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| id | INT | 主键 | PRIMARY KEY, AUTO_INCREMENT |
| company_id | BIGINT UNSIGNED | 企业ID | NOT NULL, FOREIGN KEY |
| report_id | VARCHAR(50) | 报告编号 | |
| subsidy_year | INT | 资助年份 | INDEX |
| amount | DECIMAL(18,2) | 资助金额（万元） | |
| count | INT | 资助次数 | |
| source | VARCHAR(255) | 信息来源 | |
| subsidy_list | TEXT | 资助名单 | |

### 6. 企业关系图谱表 (company_relationships)

**用途**: 存储企业间的关系信息

| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| id | INT | 主键 | PRIMARY KEY, AUTO_INCREMENT |
| company_id | BIGINT UNSIGNED | 企业ID | NOT NULL, FOREIGN KEY |
| report_id | VARCHAR(50) | 报告编号 | |
| related_company_name | VARCHAR(255) | 关联企业名称 | INDEX |
| relationship_type | ENUM | 关联类型 | ('投资', '任职', '合作', '控股', '参股') |
| investment_amount | DECIMAL(18,2) | 投资金额 | |
| investment_ratio | DECIMAL(5,2) | 投资比例 | |
| position | VARCHAR(100) | 任职职位 | |

### 7. 科创评分表 (tech_innovation_score)

**用途**: 存储企业科技创新评分信息

| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| id | INT | 主键 | PRIMARY KEY, AUTO_INCREMENT |
| company_id | BIGINT UNSIGNED | 企业ID | NOT NULL, FOREIGN KEY |
| report_id | VARCHAR(50) | 报告编号 | |
| basic_score | DECIMAL(5,2) | 基本面评分 | |
| talent_score | DECIMAL(5,2) | 人才竞争力评分 | |
| industry_ranking | VARCHAR(50) | 行业排名 | |
| strategic_industry | VARCHAR(100) | 战略新兴产业 | |
| intellectual_property | JSON | 知识产权数据 | |

### 8. 企业财务信息表 (company_financial_info)

**用途**: 存储企业财务相关信息

| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| id | INT | 主键 | PRIMARY KEY, AUTO_INCREMENT |
| company_id | BIGINT UNSIGNED | 企业ID | NOT NULL, FOREIGN KEY |
| report_id | VARCHAR(50) | 报告编号 | |
| annual_revenue | DECIMAL(18,2) | 年营业额 | |
| net_profit | DECIMAL(18,2) | 净利润 | |
| total_assets | DECIMAL(18,2) | 总资产 | |
| total_liabilities | DECIMAL(18,2) | 总负债 | |
| equity | DECIMAL(18,2) | 净资产 | |
| financing_status | VARCHAR(100) | 融资情况 | |
| listing_status | VARCHAR(50) | 上市状态 | |
| financial_year | INT | 财务年度 | INDEX |
| data_source | VARCHAR(100) | 数据来源 | |

### 9. 企业风险信息表 (company_risk_info)

**用途**: 存储企业风险相关信息

| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| id | INT | 主键 | PRIMARY KEY, AUTO_INCREMENT |
| company_id | BIGINT UNSIGNED | 企业ID | NOT NULL, FOREIGN KEY |
| report_id | VARCHAR(50) | 报告编号 | |
| risk_level | ENUM | 风险等级 | ('低风险', '中风险', '高风险') |
| legal_risks | JSON | 法律风险 | |
| financial_risks | JSON | 财务风险 | |
| operational_risks | JSON | 经营风险 | |
| credit_rating | VARCHAR(20) | 信用评级 | |
| risk_factors | TEXT | 风险因素 | |

---

## 🔗 实体关系图

```
企业基本信息表 (company_basic_info)
│
├── 资质许可表 (qualification_license)
├── 人员竞争力表 (personnel_competitiveness)
├── 公积金信息表 (provident_fund)
├── 资助补贴表 (subsidy_info)
├── 企业关系图谱表 (company_relationships)
├── 科创评分表 (tech_innovation_score)
├── 企业财务信息表 (company_financial_info)
└── 企业风险信息表 (company_risk_info)
```

---

## 📊 索引优化设计

### 主要索引

1. **主键索引**: 所有表的 `id` 字段
2. **外键索引**: 所有表的 `company_id` 字段
3. **业务索引**: 
   - `company_basic_info.company_name`
   - `company_basic_info.unified_social_credit_code`
   - `company_basic_info.industry_category`
   - `qualification_license.certificate_number`
   - `company_relationships.related_company_name`

### 复合索引

1. `idx_company_basic_info_composite` ON (company_id, report_id, industry_category)
2. `idx_qualification_license_composite` ON (company_id, type, status)
3. `idx_personnel_competitiveness_composite` ON (company_id, data_update_date)
4. `idx_company_relationships_composite` ON (company_id, relationship_type)
5. `idx_tech_innovation_score_composite` ON (company_id, basic_score, talent_score)

---

## 🛠️ 技术特性

### 1. JSON字段支持
- **用途**: 存储复杂结构化数据
- **示例**: 企业标签、司龄分布、知识产权数据
- **优势**: 灵活的数据结构，支持动态字段

### 2. 数据类型优化
- **金额字段**: 使用 `DECIMAL(18,2)` 保证精度
- **日期字段**: 统一使用 `DATE` 或 `TIMESTAMP`
- **大文本字段**: 使用 `TEXT` 类型

### 3. 约束设计
- **主外键关联**: 确保数据一致性
- **CHECK约束**: 保证数据类型正确
- **ENUM约束**: 限制枚举值范围

### 4. 性能优化
- **索引策略**: 基于查询模式设计索引
- **分区支持**: 支持按时间或企业ID分区
- **查询优化**: 优化常用查询路径

---

## 🚀 实施步骤

### 1. 数据库迁移
```bash
# 执行数据库迁移脚本
cd /Users/szjason72/zervi-basic/basic/backend/internal/company-service/scripts
./migrate_company_profile.sh --full
```

### 2. 数据模型集成
```bash
# 更新Company服务数据模型
# 文件: models/company_profile_models.go
# 文件: api/company_profile_api.go
```

### 3. 功能测试
```bash
# 执行数据库结构测试
./test_company_profile_db.sh --full
```

### 4. PDF解析集成
```bash
# 测试PDF解析功能
cd /Users/szjason72/zervi-basic/basic/backend/internal/company-service
./test_pdf_parsing.sh --test
```

---

## 📈 扩展性设计

### 1. 水平扩展
- 支持分库分表
- 支持读写分离
- 支持缓存层

### 2. 功能扩展
- 支持更多企业画像维度
- 支持实时数据更新
- 支持数据版本管理

### 3. 性能扩展
- 支持索引优化
- 支持查询优化
- 支持数据压缩

---

## 🔒 安全考虑

### 1. 数据安全
- 敏感数据加密存储
- 访问权限控制
- 数据备份策略

### 2. 访问控制
- 基于角色的权限控制
- API访问限制
- 数据脱敏处理

### 3. 审计日志
- 数据变更记录
- 访问日志记录
- 异常操作监控

---

## 📋 验证清单

### 数据库结构验证
- [ ] 所有表创建成功
- [ ] 索引创建正确
- [ ] 外键约束生效
- [ ] JSON字段功能正常

### 功能验证
- [ ] 数据插入功能正常
- [ ] 数据查询功能正常
- [ ] 数据更新功能正常
- [ ] 数据删除功能正常

### 性能验证
- [ ] 查询性能满足要求
- [ ] 索引使用正确
- [ ] 并发访问正常
- [ ] 大数据量处理正常

### 集成验证
- [ ] PDF解析数据存储正常
- [ ] API接口功能正常
- [ ] 前端集成正常
- [ ] 端到端流程正常

---

## 📝 总结

这个数据库设计完全基于公司画像文档的需求，提供了：

1. **完整的企业画像数据存储能力**
2. **灵活的JSON字段支持**
3. **优化的查询性能**
4. **良好的扩展性**
5. **完善的安全机制**

通过这个设计，Company服务可以：
- 完整存储PDF解析后的企业画像数据
- 支持复杂的企业关系分析
- 提供高效的数据查询服务
- 支持未来的功能扩展

---

**文档创建者**: AI Assistant & Development Team  
**最后更新**: 2025年9月15日  
**状态**: 📋 设计完成，准备实施

---

*"Good design is not just what looks good and feels good. Good design is good business."* - 好的设计不仅仅是看起来好、感觉好。好的设计就是好的商业。
