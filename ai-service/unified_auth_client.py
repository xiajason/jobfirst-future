#!/usr/bin/env python3
"""
统一认证客户端
用于AI服务与Job Service认证系统的集成
"""

import asyncio
import json
import logging
import time
from datetime import datetime
from typing import Optional, Dict, Any
from dataclasses import dataclass

import httpx
from sanic import Request, json as sanic_json

logger = logging.getLogger(__name__)

@dataclass
class UserInfo:
    """用户信息"""
    user_id: int
    username: str
    email: str
    role: str
    status: str
    subscription_type: str = ""
    subscription_status: str = ""
    is_active: bool = True
    expires_at: Optional[datetime] = None
    last_login: Optional[datetime] = None
    created_at: Optional[datetime] = None
    permissions: list = None

    def __post_init__(self):
        if self.permissions is None:
            self.permissions = []

class UnifiedAuthClient:
    """统一认证客户端"""
    
    def __init__(self, auth_service_url: str = "http://localhost:8207"):
        self.auth_service_url = auth_service_url.rstrip('/')
        self.client = httpx.AsyncClient(timeout=10.0)
        self.cache = {}
        self.cache_timeout = 300  # 5分钟缓存
    
    async def validate_token(self, token: str) -> Optional[UserInfo]:
        """验证token并获取用户信息"""
        try:
            # 检查缓存
            cache_key = f"token_{hash(token)}"
            if cache_key in self.cache:
                cached_result = self.cache[cache_key]
                if time.time() - cached_result["timestamp"] < self.cache_timeout:
                    return cached_result["user_info"]
            
            # 调用认证服务验证token
            response = await self.client.post(
                f"{self.auth_service_url}/api/v1/auth/validate",
                json={"token": token}
            )
            
            if response.status_code == 200:
                data = response.json()
                if data.get("success"):
                    user_data = data.get("user", {})
                    user_info = UserInfo(
                        user_id=user_data.get("id", 0),
                        username=user_data.get("username", ""),
                        email=user_data.get("email", ""),
                        role=user_data.get("role", ""),
                        status=user_data.get("status", ""),
                        subscription_type=user_data.get("subscription_type", ""),
                        subscription_status=user_data.get("subscription_status", ""),
                        permissions=data.get("permissions", [])
                    )
                    
                    # 缓存结果
                    self.cache[cache_key] = {
                        "user_info": user_info,
                        "timestamp": time.time()
                    }
                    
                    return user_info
            
            logger.warning(f"Token validation failed: {response.status_code}")
            return None
            
        except httpx.TimeoutException:
            logger.error("Token validation timeout")
            return None
        except Exception as e:
            logger.error(f"Token validation error: {e}")
            return None
    
    async def sync_user_data(self, user_id: int) -> bool:
        """同步用户数据到AI服务"""
        try:
            # 这里可以添加用户数据同步逻辑
            # 例如：从统一认证服务获取用户信息并存储到AI服务的本地数据库
            logger.info(f"User data sync for user_id: {user_id}")
            return True
        except Exception as e:
            logger.error(f"User data sync error: {e}")
            return False
    
    def get_user_info(self, request: Request) -> Optional[UserInfo]:
        """从请求中获取用户信息"""
        try:
            # 从请求上下文中获取用户信息
            if hasattr(request, 'ctx'):
                # 优先检查 user_info（zervigo_auth_middleware 设置）
                if hasattr(request.ctx, 'user_info'):
                    user_data = request.ctx.user_info
                    # 如果user_data是UserInfo对象，直接返回
                    if isinstance(user_data, UserInfo):
                        return user_data
                    # 如果user_data是字典，转换为UserInfo对象
                    elif isinstance(user_data, dict):
                        return UserInfo(
                            user_id=user_data.get('user_id') or user_data.get('id', 0),
                            username=user_data.get('username', ''),
                            email=user_data.get('email', ''),
                            role=user_data.get('role', ''),
                            status=user_data.get('status', 'active'),
                            subscription_type=user_data.get('subscription_type', ''),
                            permissions=user_data.get('permissions', [])
                        )
                    # 如果user_data是字符串或其他类型，记录错误并返回None
                    else:
                        logger.error(f"Unexpected user_info type: {type(user_data)}, value: {user_data}")
                        return None
                
                # 检查 user（unified_auth_middleware 设置）
                elif hasattr(request.ctx, 'user'):
                    user_data = request.ctx.user
                    # 如果user_data是UserInfo对象，直接返回
                    if isinstance(user_data, UserInfo):
                        return user_data
                    # 如果user_data是字典，转换为UserInfo对象
                    elif isinstance(user_data, dict):
                        return UserInfo(
                            user_id=user_data.get('user_id') or user_data.get('id', 0),
                            username=user_data.get('username', ''),
                            email=user_data.get('email', ''),
                            role=user_data.get('role', ''),
                            status=user_data.get('status', 'active'),
                            subscription_type=user_data.get('subscription_type', ''),
                            permissions=user_data.get('permissions', [])
                        )
                    # 如果user_data是字符串或其他类型，记录错误并返回None
                    else:
                        logger.error(f"Unexpected user type: {type(user_data)}, value: {user_data}")
                        return None
            return None
        except Exception as e:
            logger.error(f"获取用户信息失败: {e}")
            return None
    
    def has_permission(self, request: Request, permission: str) -> bool:
        """检查用户是否有指定权限"""
        try:
            user_info = self.get_user_info(request)
            if not user_info:
                return False
            
            # 超级管理员拥有所有权限
            if user_info.role == 'super_admin':
                return True
            
            # 从数据库获取用户权限
            from get_user_permissions import has_permission as db_has_permission
            return db_has_permission(user_info.user_id, permission)
        except Exception as e:
            logger.error(f"权限检查失败: {e}")
            return False
    
    async def check_quota(self, request: Request, resource_type: str) -> bool:
        """检查用户配额"""
        try:
            user_info = self.get_user_info(request)
            if not user_info:
                return False
            
            # 这里可以实现配额检查逻辑
            # 暂时返回True，表示有配额
            return True
        except Exception as e:
            logger.error(f"配额检查失败: {e}")
            return False
    
    async def consume_quota(self, request: Request, resource_type: str, amount: int) -> bool:
        """消耗用户配额"""
        try:
            user_info = self.get_user_info(request)
            if not user_info:
                return False
            
            # 这里可以实现配额消耗逻辑
            logger.info(f"用户 {user_info.user_id} 消耗 {amount} 个 {resource_type} 配额")
            return True
        except Exception as e:
            logger.error(f"配额消耗失败: {e}")
            return False
    
    async def log_access(self, user_id: int, action: str, resource: str, result: str, ip_address: str = "", user_agent: str = ""):
        """记录用户访问日志"""
        try:
            # 这里可以记录用户访问日志到认证服务或本地日志
            logger.info(f"用户 {user_id} 执行 {action} 操作，资源: {resource}，结果: {result}")
            return True
        except Exception as e:
            logger.error(f"记录访问日志失败: {e}")
            return False

# 全局认证客户端实例
import os
auth_service_url = os.getenv("ZERVIGO_AUTH_URL", "http://host.docker.internal:8207")
unified_auth_client = UnifiedAuthClient(auth_service_url)

async def unified_auth_middleware(request: Request):
    """统一的认证中间件"""
    try:
        # 提取token
        auth_header = request.headers.get("Authorization", "")
        if not auth_header.startswith("Bearer "):
            return {
                "error": "Invalid authorization header", 
                "code": "INVALID_AUTH_HEADER",
                "message": "请提供有效的Bearer token"
            }
        
        token = auth_header[7:]  # 移除 "Bearer " 前缀
        
        # 验证token
        user_info = await unified_auth_client.validate_token(token)
        if not user_info:
            return {
                "error": "Invalid token", 
                "code": "INVALID_TOKEN",
                "message": "认证失败，请重新登录"
            }
        
        # 同步用户数据
        await unified_auth_client.sync_user_data(user_info.user_id)
        
        # 将用户信息存储到请求上下文
        request.ctx.user = user_info
        
        return None  # 认证成功
        
    except Exception as e:
        logger.error(f"Authentication error: {e}")
        return {
            "error": "Authentication failed", 
            "code": "AUTH_ERROR",
            "message": "认证过程发生错误"
        }
