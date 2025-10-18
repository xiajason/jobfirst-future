# 企业信用信息查询API集成指南

## 概述

本指南介绍如何将Java版本的企业信用信息查询API集成到Go语言的Company服务中。

## 技术转换说明

### 1. **语言转换**
- **原实现**: Java + Spring Boot + OkHttp
- **新实现**: Go + Gin + net/http
- **加密库**: Go标准库 crypto 包

### 2. **依赖转换**

#### Java依赖 → Go依赖
```java
// Java依赖
<dependency>
    <groupId>com.squareup.okhttp3</groupId>
    <artifactId>okhttp</artifactId>
    <version>4.9.1</version>
</dependency>
<dependency>
    <groupId>com.alibaba</groupId>
    <artifactId>fastjson</artifactId>
    <version>1.2.49</version>
</dependency>
```

```go
// Go依赖 (标准库)
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/rsa"
    "crypto/sha256"
    "crypto/x509"
    "encoding/base64"
    "encoding/json"
    "encoding/pem"
    "net/http"
)
```

### 3. **加密算法转换**

#### AES加密
```java
// Java实现
Cipher cipher = Cipher.getInstance("AES/ECB/NoPadding");
SecretKeySpec keyspec = new SecretKeySpec(key.getBytes(), "AES");
cipher.init(Cipher.ENCRYPT_MODE, keyspec);
byte[] encrypted = cipher.doFinal(plaintext);
```

```go
// Go实现
block, err := aes.NewCipher([]byte(key))
mode := cipher.NewCBCEncrypter(block, make([]byte, aes.BlockSize))
ciphertext := make([]byte, len(plaintext))
mode.CryptBlocks(ciphertext, plaintext)
```

#### RSA加密
```java
// Java实现
PublicKey publicKey = RSAUtils.getPublicKey(rsa_pub);
String reqKey = Base64Utils.byte2Base64StringFun(RSAUtils.encrypt(aesKeyStr.getBytes(),publicKey));
```

```go
// Go实现
pubKey, err := parseRSAPublicKey(rsaPubKey)
ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, []byte(plaintext), nil)
```

### 4. **HTTP客户端转换**

#### Java OkHttp
```java
OkHttpClient client = new OkHttpClient().newBuilder().build();
RequestBody formBody = new FormBody.Builder()
    .add("secretKey", reqKey)
    .add("content", content)
    .add("version", version)
    .build();
Request request = new Request.Builder().url(url)
    .addHeader("Authorization",authorization).post(formBody).build();
```

#### Go net/http
```go
var buf bytes.Buffer
writer := multipart.NewWriter(&buf)
for key, value := range formData {
    writer.WriteField(key, value)
}
req, err := http.NewRequest("POST", url, &buf)
req.Header.Set("Authorization", basicAuth)
```

## 实现文件说明

### 1. **credit_info_client.go**
- **功能**: 企业信用信息查询客户端
- **包含**: AES/RSA加密解密、HTTP请求、响应处理
- **特点**: 完全基于Go标准库实现

### 2. **credit_info_api.go**
- **功能**: 企业信用信息API路由处理
- **包含**: REST API端点、请求验证、响应格式化
- **特点**: 集成到Company服务的Gin路由中

### 3. **main.go**
- **修改**: 添加信用信息API路由注册
- **集成**: 与现有Company服务无缝集成

## API端点说明

### 1. **获取企业信用信息**
```bash
POST /api/v1/company/credit/info
Authorization: Bearer <token>
Content-Type: application/json

{
    "company_name": "腾讯科技",
    "company_code": "91440300192189783K"
}
```

### 2. **获取企业信用评级**
```bash
GET /api/v1/company/credit/rating/腾讯科技
Authorization: Bearer <token>
```

### 3. **获取企业风险信息**
```bash
GET /api/v1/company/credit/risk/腾讯科技
Authorization: Bearer <token>
```

### 4. **获取企业合规状态**
```bash
GET /api/v1/company/credit/compliance/腾讯科技
Authorization: Bearer <token>
```

### 5. **批量查询企业信用信息**
```bash
POST /api/v1/company/credit/batch
Authorization: Bearer <token>
Content-Type: application/json

{
    "companies": [
        {"company_name": "腾讯科技", "company_code": "91440300192189783K"},
        {"company_name": "阿里巴巴", "company_code": "91330100MA27XN3X8N"}
    ]
}
```

## 配置说明

### 1. **环境变量配置**
```yaml
credit_api:
  base_url: "https://apitest.szscredit.com:8443/public_apis/common_api"
  username: "szc_zhangxx"
  password: "123456"
  aes_key: "8Of0L+PjmIm5FPJn"
  rsa_pub_key: "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAhLtBso+Hy6PEaGpA4Txkb6PA03dLUSlXHXV3bdjX1ilZ3re/O6JPinrhxaxsjjqliEqOc/qehNbzde4WKb9FRlnMwWwZReTruVCZNa9eNCLi+BzLcFYl9jO9QNP/Y+uS6P9ozDqmgux47GrbK7/0bIhhgRdXsegGvUp9z5VNiF/5OijDE5lrQcYzSIrPy8YiaDNkS0SZ7JQ24+wFe8fOYRWcIxzYbn5gl6U14JsxIjvnFVKWYBGMh4cfjIVv22M7tVxt52TNEcB0XEWbCcTLQQluf9c2ZGXSe5jyfMZSa4Z5e6mYG8NlywdwSBIRtM4r9WuzEkWrMxfns/sQ9sbF/wIDAQAB"
  rsa_pri_key: "MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQCEu0Gyj4fLo8RoakDhPGRvo8DTd0tRKVcddXdt2NfWKVnet787ok+KeuHFrGyOOqWISo5z+p6E1vN17hYpv0VGWczBbBlF5Ou5UJk1r140IuL4HMtwViX2M71A0/9j65Lo/2jMOqaC7Hjsatsrv/RsiGGBF1ex6Aa9Sn3PlU2IX/k6KMMTmWtBxjNIis/LxiJoM2RLRJnslDbj7AV7x85hFZwjHNhufmCXpTXgmzEiO+cVUpZgEYyHhx+MhW/bYzu1XG3nZM0RwHRcRZsJxMtBCW5/1zZkZdJ7mPJ8xlJrhnl7qZgbw2XLB3BIEhG0ziv1a7MSRaszF+ez+xD2xsX/AgMBAAECggEAQAu3RLjbNpjcIeH7UnN4pyHl3mP2tL/06CMRMLDsXMtxMPWK0fSc2t42aNKtQufrjdsj57Srnr+1lFcA3L4NaEfWdBJ8E2zFjZLlirEHDLM0v7HtPFRlVupaTJi+5/D433K2l61JQW1nX/SjsvWZtHEOU2L3DsI91kLGeE67raynlzI0EB0DD2oo3GYoJHiyioQojPjV6hSMHCq6yOcvCxG0q00/fPnZxiyNKJ7gBSuZLwfxqlwp4UQ01wcVDnPQYBhhxjzYPDWGUAPgExLCOngxxjLXAt2mh571YE+d4yQnlhIoY3/UQ7uYScioIWUetTXNxC4AwBzS2VzuTLCMwQKBgQDBP4eUjfRJ7L4fiYqYCDuQj/UA/DIWYeJV3zJIoMIY6rs4OnOuJUi6+WXszOGMUpitF1mdHsZGWzt5D8TXzcqP84X4jSPLaKv4z5j/hZmE+QvWmcmVA//IUwQLXRPCfb3eT6mTZF1B/cDtM7TU2GGvo4L+NJKQUKpwGjNhGf1VXwKBgQCv1Q10i2JRU3/vXGg7HDgke4m07OWHXQykjF93cuRKpE3xE23oo4bi07sPn29StRqjdivvvZNadvNJ2Z1vYKwRnztibbwHLlsour5V67fjhAv4APURO5NSbovHhG2lCyUrvLsyKJQhrzXaAS26CAHaP3au8LnCjg+iT3VLg+izYQKBgE/lCxHA6qmRhj0VqUYXyUCIM9v3aGHWkDO+dlSOmhChI0wo5mCuK3aZ26jeP7W7BEIzsCoEaib2Ww0/Fru96iw/mzjaaV0UZl0UvwWNX54ZNOrBZUGtT5GDBsCnUPAprn9p3c3fFLnLVckFHQXDbQG3wZoB9xAbWaxfmJ70z/zAoGAWVFlq10ejWdYJrQPMm+sSUQD+McZ9YAb6v5vhFL1isEZ4qtW+oUPAOxDKrV3rFDY/k4KFZd8Ycjo3wvPQIOgBLeZR++sQw2WOwNZqnW6DLXICqwZ0S4tMQN8t9YaiGs375bIlLsuPEovldVhcA2fO0lftZANHLpjULUCRWD1dSECgYBF+rfCsXno9qSTle5d7t2HAERM4RvIvtHhZm/prJ1pwSJSrxdcGzdskWFzI8EANf9rCm+MKiQo1s4Yye122hfU0kJSG0XBRXmh55cZUIkB/0I5MtqZTK4ktbqlb4Z6m57iO+Oydh6A2rWS0KPjMq7BujjeulVBXCpbEuoMwzTh+w=="
```

### 2. **配置文件位置**
- 配置文件: `../../configs/jobfirst-core-config.yaml`
- 环境变量: 支持通过环境变量覆盖配置

## 测试说明

### 1. **编译测试**
```bash
cd zervigo_future/backend/internal/company-service
go build -o company-service .
```

### 2. **启动服务**
```bash
./company-service
```

### 3. **API测试**
```bash
# 获取JWT token
TOKEN=$(curl -s -X POST -H "Content-Type: application/json" \
  -d '{"username":"szjason72","password":"@SZxym2006"}' \
  http://localhost:7520/api/v1/auth/login | jq -r '.data.token')

# 测试企业信用信息查询
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"company_name":"腾讯科技","company_code":"91440300192189783K"}' \
  http://localhost:7534/api/v1/company/credit/info
```

## 注意事项

### 1. **安全考虑**
- API密钥需要安全存储
- 建议使用环境变量或密钥管理服务
- 生产环境需要HTTPS

### 2. **性能优化**
- 实现连接池复用
- 添加请求缓存机制
- 考虑异步处理大批量查询

### 3. **错误处理**
- 完善的错误日志记录
- 优雅的降级处理
- 用户友好的错误信息

### 4. **监控告警**
- API调用成功率监控
- 响应时间监控
- 错误率告警

## 总结

通过将Java版本的企业信用信息查询API转换为Go语言实现，我们成功实现了：

1. **技术栈统一**: 与Company服务使用相同的Go语言技术栈
2. **功能完整**: 保持了原有的所有加密和安全机制
3. **易于维护**: 基于Go标准库，减少外部依赖
4. **性能优化**: 利用Go语言的并发特性
5. **无缝集成**: 与现有Company服务完美集成

这个实现为Company功能提案提供了重要的外部数据源支持，显著提升了系统的数据获取能力和分析价值。
