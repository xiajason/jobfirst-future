#!/bin/bash

# AI服务启动脚本
# 用于阿里云服务器自动化部署

set -e

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== 🤖 启动AI服务 ===${NC}"
echo ""

# 获取脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# 检查虚拟环境
if [ ! -d "venv" ]; then
    echo -e "${YELLOW}⚠️  虚拟环境不存在，正在创建...${NC}"
    python3 -m venv venv
    echo -e "${GREEN}✅ 虚拟环境创建完成${NC}"
fi

# 激活虚拟环境
echo "激活Python虚拟环境..."
source venv/bin/activate

# 检查依赖
if [ -f "requirements.txt" ]; then
    echo "检查并安装依赖..."
    pip install --quiet --upgrade pip
    pip install --quiet -r requirements.txt
    echo -e "${GREEN}✅ 依赖安装完成${NC}"
fi

# 加载环境变量
if [ -f ".env" ]; then
    echo "加载环境变量..."
    export $(cat .env | grep -v '^#' | xargs)
    echo -e "${GREEN}✅ 环境变量加载完成${NC}"
else
    echo -e "${RED}❌ .env文件不存在${NC}"
    echo "请创建.env文件或从env.example复制"
    exit 1
fi

# 停止旧的服务进程
echo "停止旧的AI服务进程..."
pkill -f ai_service || true
sleep 2

# 创建日志目录
mkdir -p ../logs

# 启动AI服务
AI_PORT=${AI_SERVICE_PORT:-8100}
echo -e "${GREEN}启动AI服务 (端口: $AI_PORT)...${NC}"
nohup python ai_service_with_zervigo.py > ../logs/ai-service.log 2>&1 &
AI_PID=$!
echo $AI_PID > ../logs/ai-service.pid

# 等待服务启动
echo "等待服务启动..."
sleep 5

# 健康检查
echo "执行健康检查..."
if curl -f http://localhost:$AI_PORT/health 2>/dev/null; then
    echo -e "${GREEN}✅ AI服务启动成功！${NC}"
    echo "进程ID: $AI_PID"
    echo "端口: $AI_PORT"
    echo "日志: ../logs/ai-service.log"
else
    echo -e "${RED}❌ AI服务健康检查失败${NC}"
    echo "请查看日志: tail -f ../logs/ai-service.log"
    exit 1
fi

