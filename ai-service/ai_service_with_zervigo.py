from dotenv import load_dotenv
load_dotenv()

#!/usr/bin/env python3
"""
JobFirst AI服务 - 集成Zervigo认证版本
处理文件上传后的内容解析和向量生成，集成zervigo认证系统
"""

import asyncio
import json
import logging
import os
import time
from datetime import datetime
from typing import List, Dict, Any

import httpx
import psycopg2
import psycopg2.extras
import requests
from sanic import Sanic, Request, json as sanic_json
from sanic.response import json as sanic_response

# 导入职位匹配服务
from job_matching_service import JobMatchingService

# 导入统一认证中间件
from unified_auth_client import unified_auth_middleware, unified_auth_client

# 导入权限获取模块
from get_user_permissions import get_user_permissions, has_permission, get_user_info_with_permissions

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# 创建Sanic应用
app = Sanic("ai-service-with-zervigo")

# 初始化职位匹配服务
job_matching_service = JobMatchingService(app)

# 异步初始化职位匹配服务
async def initialize_job_matching():
    await job_matching_service.initialize()

# 使用统一认证中间件
async def authenticate_user(request: Request):
    """使用统一认证的用户认证中间件"""
    try:
        auth_result = await unified_auth_middleware(request)
        if auth_result:
            return sanic_json(auth_result, status=401)
        return None  # 认证成功，继续处理
    except Exception as e:
        logger.error(f"统一认证异常: {e}")
        return sanic_json({
            "error": "认证异常",
            "code": "AUTH_ERROR",
            "message": "认证过程发生错误"
        }, status=500)

# 检查用户配额中间件
async def check_quota(request: Request, resource_type: str = "ai_requests"):
    """检查用户配额"""
    try:
        has_quota = await unified_auth_client.check_quota(request, resource_type)
        if not has_quota:
            return sanic_json({
                "error": "配额不足",
                "code": "QUOTA_EXCEEDED",
                "message": "您的AI服务使用配额已用完，请升级订阅或等待配额重置"
            }, status=429)
        return None
    except Exception as e:
        logger.error(f"配额检查异常: {e}")
        return sanic_json({
            "error": "配额检查失败",
            "code": "QUOTA_CHECK_ERROR",
            "message": "配额检查过程发生错误"
        }, status=500)

# 检查用户权限中间件
def check_permission(permission: str):
    """检查用户权限装饰器"""
    def decorator(func):
        async def wrapper(request: Request, *args, **kwargs):
            try:
                # 获取用户信息
                user_info = unified_auth_client.get_user_info(request)
                if not user_info:
                    return sanic_json({
                        "error": "用户信息获取失败",
                        "code": "USER_INFO_ERROR",
                        "message": "无法获取用户信息"
                    }, status=401)
                
                # 检查权限
                if not has_permission(user_info.user_id, permission):
                    return sanic_json({
                        "error": "权限不足",
                        "code": "PERMISSION_DENIED",
                        "message": f"您没有访问{permission}功能的权限"
                    }, status=403)
                
                return await func(request, *args, **kwargs)
            except Exception as e:
                logger.error(f"权限检查异常: {e}")
                return sanic_json({
                    "error": "权限检查失败",
                    "code": "PERMISSION_CHECK_ERROR",
                    "message": "权限检查过程发生错误"
                }, status=500)
        return wrapper
    return decorator

# 记录用户操作日志
async def log_user_action(request: Request, action: str, result: str):
    """记录用户操作日志"""
    try:
        user_info = unified_auth_client.get_user_info(request)
        if user_info:
            await unified_auth_client.log_access(
                user_id=user_info.user_id,
                action=action,
                resource="ai_service",
                result=result,
                ip_address=request.ip,
                user_agent=request.headers.get("User-Agent", "")
            )
    except Exception as e:
        logger.error(f"记录用户操作日志异常: {e}")

# AI服务健康检查API
@app.route("/health", methods=["GET"])
async def health_check(request: Request):
    """健康检查"""
    try:
        # 检查zervigo认证服务状态
        zervigo_status = "unknown"
        try:
            async with httpx.AsyncClient(timeout=5) as client:
                response = await client.get("http://localhost:8207/health")
                if response.status_code == 200:
                    zervigo_status = "healthy"
                else:
                    zervigo_status = "unhealthy"
        except Exception:
            zervigo_status = "unreachable"
        
        return sanic_json({
            "status": "healthy",
            "service": "ai-service-with-zervigo",
            "timestamp": datetime.now().isoformat(),
            "version": "1.0.0",
            "unified_auth_client_status": zervigo_status,
            "job_matching_initialized": job_matching_service.initialized
        })
    except Exception as e:
        logger.error(f"健康检查异常: {e}")
        return sanic_json({
            "status": "unhealthy",
            "error": str(e),
            "timestamp": datetime.now().isoformat()
        }, status=500)

# 用户信息API
@app.route("/api/v1/ai/user-info", methods=["GET"])
async def get_user_info(request: Request):
    """获取用户信息"""
    try:
        # 认证检查
        auth_result = await authenticate_user(request)
        if auth_result:
            return auth_result
        
        user_info = unified_auth_client.get_user_info(request)
        if not user_info:
            return sanic_json({
                "error": "用户信息不存在",
                "code": "USER_INFO_NOT_FOUND"
            }, status=404)
        
        # 获取用户权限
        logger.info(f"开始获取用户 {user_info.user_id} 的权限...")
        try:
            permissions = get_user_permissions(user_info.user_id)
            logger.info(f"获取到权限: {permissions}")
        except Exception as e:
            logger.error(f"获取权限失败: {e}")
            permissions = []
        
        return sanic_json({
            "user_id": user_info.user_id,
            "username": user_info.username,
            "email": user_info.email,
            "role": user_info.role,
            "subscription_status": user_info.subscription_status,
            "subscription_type": user_info.subscription_type,
            "is_active": user_info.is_active,
            "permissions": permissions,
            "expires_at": user_info.expires_at.isoformat() if user_info.expires_at else None,
            "last_login": user_info.last_login.isoformat() if user_info.last_login else None
        })
        
    except Exception as e:
        logger.error(f"获取用户信息异常: {e}")
        return sanic_json({
            "error": "获取用户信息失败",
            "code": "USER_INFO_ERROR",
            "message": str(e)
        }, status=500)

# 用户权限API
@app.route("/api/v1/ai/permissions", methods=["GET"])
async def get_user_permissions_api(request: Request):
    """获取用户权限"""
    try:
        # 认证检查
        auth_result = await authenticate_user(request)
        if auth_result:
            return auth_result
        
        user_info = unified_auth_client.get_user_info(request)
        if not user_info:
            return sanic_json({
                "error": "用户信息不存在",
                "code": "USER_INFO_NOT_FOUND"
            }, status=404)
        
        # 获取用户权限
        try:
            permissions = get_user_permissions(user_info.user_id)
        except Exception as e:
            logger.error(f"获取权限失败: {e}")
            permissions = []
        
        return sanic_json({
            "user_id": user_info.user_id,
            "username": user_info.username,
            "permissions": permissions
        })
        
    except Exception as e:
        logger.error(f"获取用户权限异常: {e}")
        return sanic_json({
            "error": "获取用户权限失败",
            "code": "PERMISSIONS_ERROR",
            "message": str(e)
        }, status=500)

# 用户配额API
@app.route("/api/v1/ai/quotas", methods=["GET"])
async def get_user_quotas(request: Request):
    """获取用户配额"""
    try:
        # 认证检查
        auth_result = await authenticate_user(request)
        if auth_result:
            return auth_result
        
        user_info = unified_auth_client.get_user_info(request)
        if not user_info:
            return sanic_json({
                "error": "用户信息不存在",
                "code": "USER_INFO_NOT_FOUND"
            }, status=404)
        
        quotas = []
        if hasattr(request.ctx, 'quotas'):
            for quota in request.ctx.quotas:
                quotas.append({
                    "resource_type": quota.resource_type,
                    "total_quota": quota.total_quota,
                    "used_quota": quota.used_quota,
                    "remaining_quota": quota.remaining_quota,
                    "reset_time": quota.reset_time.isoformat(),
                    "is_unlimited": quota.is_unlimited
                })
        
        return sanic_json({
            "user_id": user_info.user_id,
            "username": user_info.username,
            "quotas": quotas
        })
        
    except Exception as e:
        logger.error(f"获取用户配额异常: {e}")
        return sanic_json({
            "error": "获取用户配额失败",
            "code": "QUOTAS_ERROR",
            "message": str(e)
        }, status=500)

# 职位匹配API - 集成zervigo认证
@app.route("/api/v1/ai/job-matching", methods=["POST"], name="job_matching_with_auth")
async def job_matching_api(request: Request):
    """职位匹配API - 需要ai_job_matching权限"""
    try:
        # 认证检查
        auth_result = await authenticate_user(request)
        if auth_result:
            return auth_result
        
        # 获取用户信息
        user_info = unified_auth_client.get_user_info(request)
        logger.info(f"收到职位匹配请求: {request.json}, 用户: {user_info.user_id if user_info else 'unknown'}")
        
        # 简化的权限检查 - 所有认证用户都可以使用AI匹配功能
        if not user_info or user_info.user_id == 0:
            return sanic_json({
                "error": "用户信息无效",
                "code": "INVALID_USER",
                "message": "无法获取用户信息"
            }, status=401)
        
        if not job_matching_service.initialized:
            logger.error("JobMatchingService未初始化")
            await log_user_action(request, "job_matching", "failed")
            return sanic_json({"error": "服务未初始化", "initialized": False}, status=503)
        
        # 设置用户ID到请求上下文
        request.ctx.user_id = user_info.user_id
        
        # 处理职位匹配请求
        result = await job_matching_service._handle_job_matching(request)
        
        # 记录配额消耗（简化版本）
        logger.info(f"用户 {user_info.user_id} 使用了AI匹配功能")
        
        # 记录成功日志
        await log_user_action(request, "job_matching", "success")
        
        return result
        
    except Exception as e:
        logger.error(f"职位匹配API异常: {e}")
        await log_user_action(request, "job_matching", "failed")
        return sanic_json({"error": f"服务器内部错误: {str(e)}"}, status=500)

# 简历分析API - 集成zervigo认证
@app.route("/api/v1/ai/resume-analysis", methods=["POST"], name="resume_analysis_with_auth")
async def resume_analysis_api(request: Request):
    """简历分析API - 需要ai_resume_analysis权限"""
    try:
        # 认证检查
        auth_result = await authenticate_user(request)
        if auth_result:
            return auth_result
        
        # 权限检查
        if not unified_auth_client.has_permission(request, "ai:resume_analysis"):
            return sanic_json({
                "error": "权限不足",
                "code": "PERMISSION_DENIED",
                "message": "您没有访问AI简历分析功能的权限"
            }, status=403)
        
        # 配额检查
        quota_result = await check_quota(request, "ai_requests")
        if quota_result:
            return quota_result
        
        user_info = unified_auth_client.get_user_info(request)
        logger.info(f"收到简历分析请求: {request.json}, 用户: {user_info.user_id}")
        
        # TODO: 实现简历分析功能
        # 这里需要集成实际的AI简历分析服务
        
        # 消耗配额
        await unified_auth_client.consume_quota(request, "ai_requests", 1)
        
        # 记录成功日志
        await log_user_action(request, "resume_analysis", "success")
        
        return sanic_json({
            "success": True,
            "message": "简历分析功能正在开发中",
            "user_id": user_info.user_id
        })
        
    except Exception as e:
        logger.error(f"简历分析API异常: {e}")
        await log_user_action(request, "resume_analysis", "failed")
        return sanic_json({"error": f"服务器内部错误: {str(e)}"}, status=500)

# AI聊天API - 集成zervigo认证
@app.route("/api/v1/ai/chat", methods=["POST"], name="ai_chat_with_auth")
async def ai_chat_api(request: Request):
    """AI聊天API - 需要ai_chat权限"""
    try:
        # 认证检查
        auth_result = await authenticate_user(request)
        if auth_result:
            return auth_result
        
        # 权限检查
        if not unified_auth_client.has_permission(request, "ai:chat"):
            return sanic_json({
                "error": "权限不足",
                "code": "PERMISSION_DENIED",
                "message": "您没有访问AI聊天功能的权限"
            }, status=403)
        
        # 配额检查
        quota_result = await check_quota(request, "ai_requests")
        if quota_result:
            return quota_result
        
        user_info = unified_auth_client.get_user_info(request)
        logger.info(f"收到AI聊天请求: {request.json}, 用户: {user_info.user_id}")
        
        # TODO: 实现AI聊天功能
        # 这里需要集成实际的AI聊天服务
        
        # 消耗配额
        await unified_auth_client.consume_quota(request, "ai_requests", 1)
        
        # 记录成功日志
        await log_user_action(request, "ai_chat", "success")
        
        return sanic_json({
            "success": True,
            "message": "AI聊天功能正在开发中",
            "user_id": user_info.user_id
        })
        
    except Exception as e:
        logger.error(f"AI聊天API异常: {e}")
        await log_user_action(request, "ai_chat", "failed")
        return sanic_json({"error": f"服务器内部错误: {str(e)}"}, status=500)

# 服务初始化事件
@app.before_server_start
async def setup_db(app, loop):
    """服务器启动前初始化"""
    try:
        logger.info("正在初始化AI服务...")
        await initialize_job_matching()
        logger.info("AI服务初始化完成")
    except Exception as e:
        logger.error(f"AI服务初始化失败: {e}")

# 配置
class Config:
    PORT = int(os.getenv("AI_SERVICE_PORT", 8206))
    ZERVIGO_AUTH_URL = os.getenv("ZERVIGO_AUTH_URL", "http://host.docker.internal:8207")
    POSTGRES_HOST = os.getenv("POSTGRES_HOST", "localhost")
    POSTGRES_USER = os.getenv("POSTGRES_USER", "szjason72")
    POSTGRES_DB = os.getenv("POSTGRES_DB", "jobfirst_vector")
    POSTGRES_PASSWORD = os.getenv("POSTGRES_PASSWORD", "")
    MYSQL_HOST = os.getenv("MYSQL_HOST", "localhost")
    MYSQL_PORT = int(os.getenv("MYSQL_PORT", 3306))
    MYSQL_USER = os.getenv("MYSQL_USER", "root")
    MYSQL_PASSWORD = os.getenv("MYSQL_PASSWORD", "")
    MYSQL_DATABASE = os.getenv("MYSQL_DATABASE", "jobfirst")
    
    # 外部AI服务配置
    EXTERNAL_AI_PROVIDER = os.getenv("EXTERNAL_AI_PROVIDER", "deepseek")
    EXTERNAL_AI_API_KEY = os.getenv("EXTERNAL_AI_API_KEY", "")

# 启动服务器
if __name__ == "__main__":
    logger.info(f"启动AI服务 (集成Zervigo认证) 在端口 {Config.PORT}")
    logger.info(f"Zervigo认证服务地址: {Config.ZERVIGO_AUTH_URL}")
    
    app.run(
        host="0.0.0.0",
        port=Config.PORT,
        debug=True,
        access_log=True
    )
