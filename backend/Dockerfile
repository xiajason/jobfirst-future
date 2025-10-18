# 多阶段构建 - 构建阶段
FROM registry.cn-hangzhou.aliyuncs.com/library/golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 设置Go代理为国内镜像源
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o basic-server ./cmd/basic-server

# 运行阶段
FROM registry.cn-hangzhou.aliyuncs.com/library/alpine:latest

# 安装必要的包
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN adduser -D -s /bin/sh appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/basic-server .

# 复制配置文件
COPY --from=builder /app/configs ./configs

# 创建必要的目录
RUN mkdir -p logs uploads temp && \
    chown -R appuser:appuser /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动应用
CMD ["./basic-server"]