# Company服务PDF文档解析使用指南

## 概述

本指南介绍如何使用Company服务的PDF文档解析功能，该功能集成了MinerU服务，能够智能解析企业相关文档并提取结构化信息。

## 功能特性

### 🚀 核心功能
- **多格式支持**: PDF、DOCX、DOC、TXT
- **智能解析**: 使用MinerU服务进行文档内容提取
- **结构化存储**: 将解析结果存储为结构化数据
- **异步处理**: 支持大文档的异步解析
- **权限控制**: 基于用户角色的访问控制
- **状态跟踪**: 实时跟踪解析进度和状态

### 📊 解析内容
- **基本信息**: 企业名称、简称、成立年份、规模、行业、地址、网站
- **业务信息**: 主营业务、产品服务、目标客户、竞争优势
- **组织信息**: 组织架构、部门设置、人员规模、管理层信息
- **财务信息**: 注册资本、年营业额、融资情况、上市状态

## 架构设计

```
Company服务 ←→ MinerU服务 ←→ AI模型服务
     ↓              ↓              ↓
  文档存储        文档解析        智能分析
     ↓              ↓              ↓
  结构化数据      解析结果        业务洞察
```

## API接口

### 1. 文档上传

**接口**: `POST /api/v1/company/documents/upload`

**请求参数**:
- `file`: 上传的文件（multipart/form-data）
- `company_id`: 企业ID
- `title`: 文档标题

**响应示例**:
```json
{
  "status": "success",
  "document_id": 1,
  "message": "文档上传成功",
  "upload_time": "2025-09-15T10:30:00Z"
}
```

### 2. 文档解析

**接口**: `POST /api/v1/company/documents/{id}/parse`

**响应示例**:
```json
{
  "status": "success",
  "task_id": 1,
  "message": "解析任务已创建，正在处理中"
}
```

### 3. 解析状态查询

**接口**: `GET /api/v1/company/documents/{id}/parse/status`

**响应示例**:
```json
{
  "status": "completed",
  "task_id": 1,
  "progress": 100,
  "message": "解析完成",
  "structured_data": {
    "basic_info": {
      "name": "测试科技有限公司",
      "short_name": "测试科技",
      "founded_year": 2020,
      "company_size": "51-100人",
      "industry": "互联网/电子商务",
      "location": "北京市海淀区中关村大街1号",
      "website": "https://www.testtech.com"
    },
    "business_info": {
      "main_business": "软件开发、技术咨询",
      "products": "企业管理系统、移动应用开发",
      "target_customers": "中小企业、政府机构",
      "competitive_advantage": "技术领先、服务优质"
    },
    "organization_info": {
      "organization_structure": "技术部、市场部、人事部、财务部",
      "departments": "研发中心、销售中心、运营中心",
      "personnel_scale": "80人",
      "management_info": "CEO、CTO、CFO"
    },
    "financial_info": {
      "registered_capital": "1000万元",
      "annual_revenue": "5000万元",
      "financing_status": "A轮融资2000万元",
      "listing_status": "未上市"
    },
    "confidence": 0.85,
    "parsing_version": "mineru-v1.0"
  }
}
```

### 4. 文档列表查询

**接口**: `GET /api/v1/company/documents/`

**查询参数**:
- `company_id`: 企业ID（可选）
- `page`: 页码（默认1）
- `page_size`: 每页大小（默认10）

### 5. 文档详情查询

**接口**: `GET /api/v1/company/documents/{id}`

### 6. 文档删除

**接口**: `DELETE /api/v1/company/documents/{id}`

## 使用步骤

### 步骤1: 启动服务

确保以下服务正在运行：
- Company服务 (端口8083)
- MinerU服务 (端口8001)
- 数据库服务

```bash
# 启动Company服务
cd basic/backend/internal/company-service
go run main.go

# 启动MinerU服务
cd basic/ai-services
docker-compose up -d
```

### 步骤2: 获取认证Token

```bash
# 登录获取JWT token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "your_username",
    "password": "your_password"
  }'
```

### 步骤3: 上传文档

```bash
# 上传PDF文档
curl -X POST http://localhost:8083/api/v1/company/documents/upload \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@company_document.pdf" \
  -F "company_id=1" \
  -F "title=企业介绍文档"
```

### 步骤4: 解析文档

```bash
# 开始解析文档
curl -X POST http://localhost:8083/api/v1/company/documents/1/parse \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 步骤5: 查询解析状态

```bash
# 查询解析状态
curl -X GET http://localhost:8083/api/v1/company/documents/1/parse/status \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 测试脚本

使用提供的测试脚本进行功能验证：

```bash
# 检查服务状态
./test_pdf_parsing.sh --check

# 运行完整测试
./test_pdf_parsing.sh --test
```

## 数据库表结构

### company_documents 表
存储上传的文档信息：
- `id`: 主键
- `company_id`: 企业ID
- `user_id`: 用户ID
- `title`: 文档标题
- `original_file`: 原始文件路径
- `file_content`: Base64编码的文件内容
- `file_type`: 文件类型
- `file_size`: 文件大小
- `upload_time`: 上传时间

### company_parsing_tasks 表
存储解析任务信息：
- `id`: 主键
- `company_id`: 企业ID
- `document_id`: 文档ID
- `user_id`: 用户ID
- `status`: 任务状态（pending/processing/completed/failed）
- `progress`: 解析进度（0-100）
- `error_message`: 错误信息
- `result_data`: 解析结果JSON
- `mineru_task_id`: MinerU任务ID

### company_structured_data 表
存储结构化解析结果：
- `id`: 主键
- `company_id`: 企业ID
- `task_id`: 任务ID
- `basic_info`: 基本信息JSON
- `business_info`: 业务信息JSON
- `organization_info`: 组织信息JSON
- `financial_info`: 财务信息JSON
- `confidence`: 解析置信度
- `parsing_version`: 解析器版本

## 错误处理

### 常见错误码
- `400`: 请求参数错误
- `401`: 未授权访问
- `403`: 权限不足
- `404`: 资源不存在
- `409`: 资源冲突
- `500`: 服务器内部错误
- `503`: 服务不可用

### 错误响应示例
```json
{
  "error": "文档上传失败: 不支持的文件类型，仅支持PDF、DOCX、DOC、TXT"
}
```

## 性能优化

### 1. 并发控制
- MinerU服务支持最大并发解析任务数配置
- 默认最大并发数为2，可通过环境变量调整

### 2. 文件大小限制
- 建议单个文档大小不超过50MB
- 大文档建议分页处理

### 3. 缓存策略
- 解析结果缓存到数据库
- 避免重复解析相同文档

## 监控和日志

### 1. 服务监控
- 使用Prometheus监控服务状态
- 监控解析任务队列长度
- 监控解析成功率

### 2. 日志记录
- 记录文档上传、解析、状态变更
- 记录错误信息和异常堆栈
- 支持结构化日志输出

## 安全考虑

### 1. 文件安全
- 文件类型验证
- 文件大小限制
- 恶意文件检测

### 2. 访问控制
- JWT token验证
- 基于角色的权限控制
- 用户数据隔离

### 3. 数据保护
- 敏感信息脱敏
- 数据加密存储
- 访问日志记录

## 故障排除

### 1. 服务不可用
- 检查服务端口是否被占用
- 检查数据库连接是否正常
- 检查MinerU服务是否启动

### 2. 解析失败
- 检查文档格式是否支持
- 检查文档内容是否可读
- 查看错误日志获取详细信息

### 3. 权限问题
- 检查JWT token是否有效
- 检查用户角色权限
- 检查资源访问权限

## 扩展功能

### 1. 批量处理
- 支持批量上传文档
- 支持批量解析任务
- 支持批量状态查询

### 2. 模板匹配
- 支持自定义解析模板
- 支持行业特定解析规则
- 支持多语言文档解析

### 3. 智能分析
- 集成AI模型进行内容分析
- 提供业务洞察和建议
- 支持数据可视化展示

## 联系支持

如有问题或建议，请联系开发团队或提交Issue。
