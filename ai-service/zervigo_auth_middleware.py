#!/usr/bin/env python3
"""
Zervigo认证中间件
用于AI服务与zervigo认证系统的集成
"""

import asyncio
import json
import logging
import time
from typing import Dict, Any, Optional
from dataclasses import dataclass
from datetime import datetime, timedelta

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
    subscription_status: str
    subscription_type: str
    expires_at: Optional[datetime] = None
    is_active: bool = True
    last_login: Optional[datetime] = None
    created_at: Optional[datetime] = None

@dataclass
class PermissionInfo:
    """权限信息"""
    permission_id: int
    permission_name: str
    resource: str
    action: str
    is_allowed: bool

@dataclass
class QuotaInfo:
    """配额信息"""
    resource_type: str
    total_quota: int
    used_quota: int
    remaining_quota: int
    reset_time: datetime
    is_unlimited: bool

@dataclass
class AuthResult:
    """认证结果"""
    success: bool
    user: Optional[UserInfo] = None
    permissions: Optional[list[PermissionInfo]] = None
    quotas: Optional[list[QuotaInfo]] = None
    error: Optional[str] = None
    error_code: Optional[str] = None

class ZervigoAuthClient:
    """Zervigo认证客户端"""
    
    def __init__(self, base_url: str = "http://localhost:8207", timeout: int = 30):
        self.base_url = base_url.rstrip('/')
        self.timeout = timeout
        self.cache = {}
        self.cache_timeout = 300  # 5分钟缓存
        
    async def validate_jwt(self, token: str) -> AuthResult:
        """验证JWT token"""
        try:
            async with httpx.AsyncClient(timeout=self.timeout) as client:
                response = await client.post(
                    f"{self.base_url}/api/v1/auth/validate",
                    json={"token": token},
                    headers={"Content-Type": "application/json"}
                )
                
                if response.status_code == 200:
                    data = response.json()
                    if data.get("success"):
                        return self._parse_auth_result(data)
                    else:
                        return AuthResult(
                            success=False,
                            error=data.get("error", "认证失败"),
                            error_code=data.get("error_code", "AUTH_FAILED")
                        )
                else:
                    return AuthResult(
                        success=False,
                        error=f"认证服务错误: {response.status_code}",
                        error_code="AUTH_SERVICE_ERROR"
                    )
                    
        except httpx.TimeoutException:
            logger.error("认证服务超时")
            return AuthResult(
                success=False,
                error="认证服务超时",
                error_code="AUTH_TIMEOUT"
            )
        except Exception as e:
            logger.error(f"认证请求异常: {e}")
            return AuthResult(
                success=False,
                error=f"认证请求异常: {str(e)}",
                error_code="AUTH_ERROR"
            )
    
    async def check_permission(self, user_id: int, permission: str) -> bool:
        """检查用户权限"""
        try:
            async with httpx.AsyncClient(timeout=self.timeout) as client:
                response = await client.get(
                    f"{self.base_url}/api/v1/auth/permission",
                    params={"user_id": user_id, "permission": permission}
                )
                
                if response.status_code == 200:
                    data = response.json()
                    return data.get("has_permission", False)
                else:
                    logger.error(f"权限检查失败: {response.status_code}")
                    return False
                    
        except Exception as e:
            logger.error(f"权限检查异常: {e}")
            return False
    
    async def check_quota(self, user_id: int, resource_type: str = "ai_requests") -> Optional[QuotaInfo]:
        """检查用户配额"""
        try:
            async with httpx.AsyncClient(timeout=self.timeout) as client:
                response = await client.get(
                    f"{self.base_url}/api/v1/auth/quota",
                    params={"user_id": user_id, "resource_type": resource_type}
                )
                
                if response.status_code == 200:
                    data = response.json()
                    return QuotaInfo(
                        resource_type=data.get("resource_type", resource_type),
                        total_quota=data.get("total_quota", 0),
                        used_quota=data.get("used_quota", 0),
                        remaining_quota=data.get("remaining_quota", 0),
                        reset_time=datetime.fromisoformat(data.get("reset_time", "2025-01-01T00:00:00")),
                        is_unlimited=data.get("is_unlimited", False)
                    )
                else:
                    logger.error(f"配额检查失败: {response.status_code}")
                    return None
                    
        except Exception as e:
            logger.error(f"配额检查异常: {e}")
            return None
    
    async def validate_access(self, user_id: int, resource: str = "ai_service") -> AuthResult:
        """验证用户访问权限"""
        try:
            async with httpx.AsyncClient(timeout=self.timeout) as client:
                response = await client.post(
                    f"{self.base_url}/api/v1/auth/access",
                    json={"user_id": user_id, "resource": resource},
                    headers={"Content-Type": "application/json"}
                )
                
                if response.status_code == 200:
                    data = response.json()
                    return self._parse_auth_result(data)
                else:
                    return AuthResult(
                        success=False,
                        error=f"访问验证失败: {response.status_code}",
                        error_code="ACCESS_VALIDATION_ERROR"
                    )
                    
        except Exception as e:
            logger.error(f"访问验证异常: {e}")
            return AuthResult(
                success=False,
                error=f"访问验证异常: {str(e)}",
                error_code="ACCESS_VALIDATION_ERROR"
            )
    
    async def log_access(self, user_id: int, action: str, resource: str, result: str, 
                        ip_address: str = "", user_agent: str = ""):
        """记录访问日志"""
        try:
            async with httpx.AsyncClient(timeout=self.timeout) as client:
                await client.post(
                    f"{self.base_url}/api/v1/auth/log",
                    json={
                        "user_id": user_id,
                        "action": action,
                        "resource": resource,
                        "result": result,
                        "ip_address": ip_address,
                        "user_agent": user_agent
                    },
                    headers={"Content-Type": "application/json"}
                )
        except Exception as e:
            logger.error(f"记录访问日志失败: {e}")
    
    def _parse_auth_result(self, data: Dict[str, Any]) -> AuthResult:
        """解析认证结果"""
        user_data = data.get("user")
        if not user_data:
            return AuthResult(success=False, error="用户信息缺失", error_code="USER_INFO_MISSING")
        
        # 解析用户信息
        user = UserInfo(
            user_id=user_data.get("id", 0),  # 修复：认证服务返回的是id，不是user_id
            username=user_data.get("username", ""),
            email=user_data.get("email", ""),
            role=user_data.get("role", ""),
            subscription_status=user_data.get("subscription_status", ""),
            subscription_type=user_data.get("subscription_type", ""),
            is_active=user_data.get("is_active", True),
            expires_at=self._parse_datetime(user_data.get("expires_at")),
            last_login=self._parse_datetime(user_data.get("last_login")),
            created_at=self._parse_datetime(user_data.get("created_at"))
        )
        
        # 解析权限信息
        permissions = []
        for perm_data in data.get("permissions", []):
            permissions.append(PermissionInfo(
                permission_id=perm_data.get("permission_id", 0),
                permission_name=perm_data.get("permission_name", ""),
                resource=perm_data.get("resource", ""),
                action=perm_data.get("action", ""),
                is_allowed=perm_data.get("is_allowed", False)
            ))
        
        # 解析配额信息
        quotas = []
        for quota_data in data.get("quotas", []):
            quotas.append(QuotaInfo(
                resource_type=quota_data.get("resource_type", ""),
                total_quota=quota_data.get("total_quota", 0),
                used_quota=quota_data.get("used_quota", 0),
                remaining_quota=quota_data.get("remaining_quota", 0),
                reset_time=self._parse_datetime(quota_data.get("reset_time")),
                is_unlimited=quota_data.get("is_unlimited", False)
            ))
        
        return AuthResult(
            success=data.get("success", False),
            user=user,
            permissions=permissions,
            quotas=quotas,
            error=data.get("error"),
            error_code=data.get("error_code")
        )
    
    def _parse_datetime(self, datetime_str: Optional[str]) -> Optional[datetime]:
        """解析日期时间字符串"""
        if not datetime_str:
            return None
        
        try:
            # 尝试不同的日期时间格式
            formats = [
                "%Y-%m-%dT%H:%M:%S",
                "%Y-%m-%d %H:%M:%S",
                "%Y-%m-%dT%H:%M:%S.%f",
                "%Y-%m-%dT%H:%M:%S%z"
            ]
            
            for fmt in formats:
                try:
                    return datetime.strptime(datetime_str, fmt)
                except ValueError:
                    continue
            
            # 如果所有格式都失败，返回None
            return None
            
        except Exception:
            return None

class ZervigoAuthMiddleware:
    """Zervigo认证中间件"""
    
    def __init__(self, zervigo_base_url: str = "http://localhost:8207"):
        self.client = ZervigoAuthClient(zervigo_base_url)
        self.cache = {}
        self.cache_timeout = 300  # 5分钟缓存
    
    async def authenticate(self, request: Request) -> Optional[Dict[str, Any]]:
        """用户认证中间件"""
        try:
            # 获取Authorization头
            auth_header = request.headers.get('Authorization')
            if not auth_header:
                return {
                    "error": "认证失败",
                    "code": "AUTH_REQUIRED",
                    "message": "请提供有效的认证信息"
                }
            
            # 检查Bearer token格式
            if not auth_header.startswith('Bearer '):
                return {
                    "error": "认证失败",
                    "code": "INVALID_TOKEN_FORMAT",
                    "message": "认证信息格式错误，请使用Bearer token"
                }
            
            token = auth_header[7:]  # 移除"Bearer "前缀
            
            # 检查缓存
            cache_key = f"auth_{token}"
            if cache_key in self.cache:
                cached_result = self.cache[cache_key]
                if time.time() - cached_result["timestamp"] < self.cache_timeout:
                    if cached_result["success"]:
                        request.ctx.user_info = cached_result["user"]
                        request.ctx.permissions = cached_result["permissions"]
                        request.ctx.quotas = cached_result["quotas"]
                        return None
                    else:
                        return cached_result["error"]
            
            # 调用zervigo认证API
            auth_result = await self.client.validate_jwt(token)
            
            # 缓存结果
            self.cache[cache_key] = {
                "success": auth_result.success,
                "user": auth_result.user,
                "permissions": auth_result.permissions,
                "quotas": auth_result.quotas,
                "error": {
                    "error": auth_result.error,
                    "code": auth_result.error_code,
                    "message": auth_result.error
                } if not auth_result.success else None,
                "timestamp": time.time()
            }
            
            if auth_result.success:
                # 将用户信息存储到请求上下文
                request.ctx.user_info = auth_result.user
                request.ctx.permissions = auth_result.permissions
                request.ctx.quotas = auth_result.quotas
                
                # 记录访问日志
                await self.client.log_access(
                    user_id=auth_result.user.user_id,
                    action="api_access",
                    resource="ai_service",
                    result="success",
                    ip_address=request.ip,
                    user_agent=request.headers.get("User-Agent", "")
                )
                
                return None
            else:
                # 记录失败的访问日志
                await self.client.log_access(
                    user_id=0,
                    action="api_access",
                    resource="ai_service",
                    result="failed",
                    ip_address=request.ip,
                    user_agent=request.headers.get("User-Agent", "")
                )
                
                return {
                    "error": auth_result.error or "认证失败",
                    "code": auth_result.error_code or "AUTH_FAILED",
                    "message": auth_result.error or "认证失败"
                }
                
        except Exception as e:
            logger.error(f"认证中间件异常: {e}")
            return {
                "error": "认证异常",
                "code": "AUTH_ERROR",
                "message": "认证过程发生错误"
            }
    
    async def check_quota(self, request: Request, resource_type: str = "ai_requests") -> bool:
        """检查用户配额"""
        try:
            if not hasattr(request.ctx, 'user_info'):
                return False
            
            user_id = request.ctx.user_info.user_id
            
            # 检查配额
            quota = await self.client.check_quota(user_id, resource_type)
            if not quota:
                return False
            
            # 检查是否无限制
            if quota.is_unlimited:
                return True
            
            # 检查是否还有剩余配额
            if quota.remaining_quota <= 0:
                # 记录配额超限日志
                await self.client.log_access(
                    user_id=user_id,
                    action="quota_exceeded",
                    resource=resource_type,
                    result="denied",
                    ip_address=request.ip,
                    user_agent=request.headers.get("User-Agent", "")
                )
                return False
            
            return True
            
        except Exception as e:
            logger.error(f"配额检查异常: {e}")
            return False
    
    async def consume_quota(self, request: Request, resource_type: str = "ai_requests", amount: int = 1):
        """消耗用户配额"""
        try:
            if not hasattr(request.ctx, 'user_info'):
                return
            
            user_id = request.ctx.user_info.user_id
            
            # 记录配额消耗日志
            await self.client.log_access(
                user_id=user_id,
                action="quota_consumed",
                resource=resource_type,
                result="success",
                ip_address=request.ip,
                user_agent=request.headers.get("User-Agent", "")
            )
            
        except Exception as e:
            logger.error(f"配额消耗记录异常: {e}")
    
    def has_permission(self, request: Request, permission: str) -> bool:
        """检查用户是否有特定权限"""
        try:
            if not hasattr(request.ctx, 'permissions'):
                return False
            
            for perm in request.ctx.permissions:
                if perm.permission_name == permission and perm.is_allowed:
                    return True
            
            return False
            
        except Exception as e:
            logger.error(f"权限检查异常: {e}")
            return False
    
    def get_user_info(self, request: Request) -> Optional[UserInfo]:
        """获取用户信息"""
        try:
            if hasattr(request.ctx, 'user_info'):
                return request.ctx.user_info
            return None
        except Exception as e:
            logger.error(f"获取用户信息异常: {e}")
            return None

# 全局认证中间件实例
zervigo_auth = ZervigoAuthMiddleware()
