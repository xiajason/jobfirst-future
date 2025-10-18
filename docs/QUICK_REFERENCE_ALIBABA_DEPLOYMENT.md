# Zervigo Future 阿里云部署快速参考

**服务器**: 47.115.168.107  
**部署目录**: /opt/services  
**最后更新**: 2025年10月18日

## 🎯 一分钟部署指南

### 自动部署 (推荐)
```bash
git push origin main
```

### 手动部署
```bash
# 1. 构建
cd zervigo_future/backend
go build -o bin/* ./...

# 2. 上传
scp bin/* root@47.115.168.107:/opt/services/backend/bin/

# 3. 部署
ssh root@47.115.168.107 'cd /opt/services && ./scripts/deploy-all.sh'
```

## 📊 服务端口映射 (完整版)

| 端口 | 服务 | 状态 | 类型 |
|------|------|------|------|
| **数据库层** ||||
| 3306 | MySQL | ✅ 已部署 | 容器 |
| 5432 | PostgreSQL | ✅ 已部署 | 容器 |
| 6379 | Redis | ✅ 已部署 | 容器 |
| 27017 | MongoDB | ✅ 已部署 | 容器 |
| **微服务层** ||||
| 8080 | API Gateway | 待部署 | Go |
| 8081 | User Service | 待部署 | Go |
| 8082 | Resume Service | 待部署 | Go |
| 8083 | Company Service | 待部署 | Go |
| 8084 | Notification Service | 待部署 | Go |
| 8085 | Template Service | 待部署 | Go |
| 8086 | Statistics Service | 待部署 | Go |
| 8087 | Banner Service | 待部署 | Go |
| 8088 | Dev Team Service | 待部署 | Go |
| 8089 | Job Service | 待部署 | Go |
| **AI服务层** ||||
| 8100 | AI Service | ✅ 已部署 | Python |

## 🔍 快速健康检查

```bash
# 一键检查所有服务
for port in 8080 8081 8082 8083 8084 8085 8086 8087 8088 8089 8100; do
    curl -f http://47.115.168.107:$port/health && echo "✅ Port $port OK" || echo "❌ Port $port Failed"
done

# 检查数据库
ssh root@47.115.168.107 'podman ps | grep migration'
```

## 🚀 启动时序

```
1. 网关层 (8080)          → 等待10秒
2. 认证层 (8081)          → 等待10秒
3. 核心业务 (8082-8083)   → 等待5秒
4. 支撑服务 (8084-8087)   → 等待3秒
5. 管理服务 (8088-8089)   → 等待3秒
```

## 🔒 数据库密码

| 数据库 | 密码 |
|--------|------|
| PostgreSQL | `JobFirst2025!PG` |
| MySQL | `JobFirst2025!MySQL` |
| MongoDB | `JobFirst2025!Mongo` |
| Redis | `JobFirst2025!Redis` |

## 📝 常用命令

### 服务管理
```bash
# 查看服务进程
ps aux | grep -E "(api-gateway|user-service|resume-service)"

# 查看端口监听
netstat -tlnp | grep -E "808[0-9]"

# 重启服务
pkill -f user-service
cd /opt/services/backend/bin
nohup ./user-service > ../../logs/user-service.log 2>&1 &
```

### 日志管理
```bash
# 查看日志
tail -f /opt/services/logs/api-gateway.log

# 查看错误
grep -i error /opt/services/logs/*.log

# 清理旧日志
find /opt/services/logs -name "*.log" -mtime +30 -delete
```

### 数据库操作
```bash
# MySQL
podman exec migration-mysql mysql -uroot -pJobFirst2025!MySQL

# PostgreSQL
podman exec migration-postgres psql -U postgres

# Redis
podman exec migration-redis redis-cli -a JobFirst2025!Redis

# MongoDB
podman exec migration-mongodb mongosh -u admin -p'JobFirst2025!Mongo' --authenticationDatabase admin
```

## 🆘 故障排除

### 问题1: 服务启动失败
```bash
# 查看日志
tail -100 /opt/services/logs/[service-name].log

# 检查端口占用
lsof -i :[port]

# 检查进程
ps aux | grep [service-name]
```

### 问题2: 数据库连接失败
```bash
# 检查容器状态
podman ps | grep migration

# 测试数据库连接
podman exec migration-mysql mysql -uroot -pJobFirst2025!MySQL -e "SELECT 1;"
```

### 问题3: 端口冲突
```bash
# 查找占用端口的进程
netstat -tlnp | grep [port]

# 杀死进程
kill -9 [PID]
```

## 📞 访问地址

- **API Gateway**: http://47.115.168.107:8080
- **User Service**: http://47.115.168.107:8081
- **Resume Service**: http://47.115.168.107:8082
- **Company Service**: http://47.115.168.107:8083
- **Notification Service**: http://47.115.168.107:8084
- **Template Service**: http://47.115.168.107:8085
- **Statistics Service**: http://47.115.168.107:8086
- **Banner Service**: http://47.115.168.107:8087
- **Dev Team Service**: http://47.115.168.107:8088
- **Job Service**: http://47.115.168.107:8089
- **AI Service**: http://47.115.168.107:8100

## 📚 详细文档

- [完整部署指南](zervigo_future/docs/guides/ALIBABA_MICROSERVICE_DEPLOYMENT_GUIDE.md)
- [CI/CD实现总结](ZERVIGO_FUTURE_CICD_IMPLEMENTATION_SUMMARY.md)
- [服务器现状报告](ALIYUN_SERVER_STATUS_REPORT_20251018.md)

---

**维护**: AI Assistant | **更新**: 2025-10-18
