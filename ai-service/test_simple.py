#!/usr/bin/env python3
"""
简化的AI服务测试版本
用于验证容器化功能
"""

import asyncio
import logging
from sanic import Sanic, Request, response
from sanic.response import json

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# 创建Sanic应用
app = Sanic("ai-service-test")

@app.route("/health", methods=["GET"])
async def health_check(request: Request):
    """健康检查"""
    return json({
        "status": "healthy",
        "service": "ai-service-test",
        "version": "1.0.0",
        "message": "AI服务容器化测试成功"
    })

@app.route("/api/v1/test", methods=["GET"])
async def test_endpoint(request: Request):
    """测试端点"""
    return json({
        "status": "success",
        "message": "AI服务容器化测试端点正常工作",
        "data": {
            "timestamp": "2025-09-14T13:30:00Z",
            "service": "ai-service-test",
            "containerized": True
        }
    })

@app.route("/api/v1/echo", methods=["POST"])
async def echo_endpoint(request: Request):
    """回显端点"""
    try:
        data = request.json
        return json({
            "status": "success",
            "echo": data,
            "message": "数据回显成功"
        })
    except Exception as e:
        return json({
            "status": "error",
            "message": f"处理请求失败: {str(e)}"
        }, status=500)

if __name__ == "__main__":
    # 启动服务
    logger.info("启动AI服务测试版本")
    app.run(host="0.0.0.0", port=8206, workers=1)
