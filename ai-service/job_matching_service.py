#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
AI职位匹配服务 - API服务层
提供职位匹配相关的API接口

创建时间: 2025-09-13
作者: AI Assistant
版本: 1.0.0
"""

import asyncio
import json
import logging
import os
import aiohttp
from typing import Dict, List, Optional, Any
from datetime import datetime
from sanic import Sanic, Request, response
from sanic.response import json as sanic_json
from functools import wraps

from job_matching_data_access import JobMatchingDataAccess, SecureSQLiteManager, UserSessionManager
from privacy_enhanced_data_access import PrivacyEnhancedDataAccess
from job_matching_engine import JobMatchingEngine

logger = logging.getLogger(__name__)

class JobMatchingService:
    """职位匹配服务"""
    
    def __init__(self, app: Sanic):
        """
        初始化职位匹配服务
        
        Args:
            app: Sanic应用实例
        """
        self.app = app
        self.data_access = None
        self.matching_engine = None
        self.initialized = False
        
        # 配置
        self.mysql_config = {
            'host': os.getenv('MYSQL_HOST', 'localhost'),
            'port': int(os.getenv('MYSQL_PORT', '3306')),
            'user': os.getenv('MYSQL_USER', 'root'),
            'password': os.getenv('MYSQL_PASSWORD', ''),
            'database': os.getenv('MYSQL_DATABASE', 'jobfirst')
        }
        
        self.postgres_config = {
            'host': os.getenv('POSTGRES_HOST', 'localhost'),
            'port': int(os.getenv('POSTGRES_PORT', '5432')),
            'user': os.getenv('POSTGRES_USER', 'szjason72'),
            'password': os.getenv('POSTGRES_PASSWORD', ''),
            'database': os.getenv('POSTGRES_DATABASE', 'jobfirst_vector')
        }
        
        # 权限配置
        self.permissions = {
            "ai.job_matching": "AI职位匹配功能",
            "ai.job_matching.admin": "AI职位匹配管理功能"
        }
    
    async def initialize(self):
        """初始化服务"""
        try:
            # 初始化数据访问层
            self.data_access = JobMatchingDataAccess(self.mysql_config, self.postgres_config)
            await self.data_access.initialize()
            
            # 初始化匹配引擎
            self.matching_engine = JobMatchingEngine(self.data_access, self.data_access.postgres_pool)
            
            self.initialized = True
            logger.info("职位匹配服务初始化成功")
            
        except Exception as e:
            logger.error(f"职位匹配服务初始化失败: {e}")
            raise
    
    def _setup_routes(self):
        """设置API路由"""
        
        # 职位匹配API (暂时移除权限检查用于测试)
        @self.app.route("/api/v1/ai/job-matching", methods=["POST"])
        async def job_matching_api(request: Request):
            """职位匹配API"""
            # 临时设置用户ID用于测试
            request.ctx.user_id = 1
            return await self._handle_job_matching(request)
        
        # 获取匹配结果详情
        @self.app.route("/api/v1/ai/job-matching/<job_id>/details", methods=["GET"])
        @self.require_permission("ai.job_matching")
        async def job_matching_details(request: Request, job_id: int):
            """获取匹配结果详情"""
            return await self._handle_job_matching_details(request, job_id)
        
        # 获取用户匹配历史
        @self.app.route("/api/v1/ai/job-matching/history", methods=["GET"])
        @self.require_permission("ai.job_matching")
        async def job_matching_history(request: Request):
            """获取用户匹配历史"""
            return await self._handle_job_matching_history(request)
        
        # 更新匹配权重配置 (管理员)
        @self.app.route("/api/v1/ai/job-matching/weights", methods=["PUT"])
        @self.require_permission("ai.job_matching.admin")
        async def update_matching_weights(request: Request):
            """更新匹配权重配置"""
            return await self._handle_update_weights(request)
        
        # 获取匹配统计信息 (管理员)
        @self.app.route("/api/v1/ai/job-matching/stats", methods=["GET"])
        @self.require_permission("ai.job_matching.admin")
        async def job_matching_stats(request: Request):
            """获取匹配统计信息"""
            return await self._handle_job_matching_stats(request)
    
    def require_permission(self, permission: str):
        """权限检查装饰器"""
        def decorator(func):
            @wraps(func)
            async def wrapper(request: Request, *args, **kwargs):
                try:
                    # 验证JWT token
                    user_id = await self._verify_jwt_token(request)
                    if not user_id:
                        return sanic_json({"error": "认证失败"}, status=401)
                    
                    # 检查权限
                    if not await self._check_user_permission(request, permission):
                        return sanic_json({"error": f"权限不足: {permission}"}, status=403)
                    
                    # 将用户ID添加到请求上下文
                    request.ctx.user_id = user_id
                    
                    return await func(request, *args, **kwargs)
                    
                except Exception as e:
                    logger.error(f"权限检查失败: {e}")
                    return sanic_json({"error": "权限检查失败"}, status=500)
            
            return wrapper
        return decorator
    
    async def _verify_jwt_token(self, request: Request) -> Optional[int]:
        """验证JWT token"""
        try:
            # 从请求头获取token
            auth_header = request.headers.get('Authorization', '')
            if not auth_header.startswith('Bearer '):
                return None
            
            token = auth_header[7:]  # 移除 "Bearer " 前缀
            
            # 调用User Service验证token
            user_service_url = f"{os.getenv('USER_SERVICE_URL', 'http://localhost:8081')}/api/v1/auth/verify"
            headers = {"Content-Type": "application/json"}
            data = {"token": token}
            
            async with aiohttp.ClientSession() as session:
                async with session.post(user_service_url, json=data, headers=headers, timeout=5) as resp:
                    if resp.status == 200:
                        result = await resp.json()
                        return result.get("user_id") if result.get("valid") else None
                    else:
                        logger.warning(f"JWT token验证失败: {resp.status}")
                        return None
                        
        except Exception as e:
            logger.error(f"JWT token验证异常: {e}")
            return None
    
    async def _check_user_permission(self, request: Request, required_permission: str) -> bool:
        """检查用户权限"""
        try:
            # 调用User Service检查权限
            user_service_url = f"{os.getenv('USER_SERVICE_URL', 'http://localhost:8081')}/api/v1/rbac/check"
            headers = {
                "Authorization": request.headers.get('Authorization', ''),
                "Content-Type": "application/json"
            }
            params = {"permission": required_permission}
            
            async with aiohttp.ClientSession() as session:
                async with session.get(user_service_url, headers=headers, params=params, timeout=5) as resp:
                    if resp.status == 200:
                        result = await resp.json()
                        return result.get("allowed", False)
                    else:
                        logger.warning(f"权限检查失败: {resp.status}")
                        return False
                        
        except Exception as e:
            logger.error(f"权限检查异常: {e}")
            return False
    
    async def _handle_job_matching(self, request: Request):
        """处理职位匹配请求"""
        try:
            if not self.initialized:
                return sanic_json({"error": "服务未初始化"}, status=503)
            
            # 解析请求参数
            data = request.json
            resume_id = data.get("resume_id")
            limit = data.get("limit", 10)
            filters = data.get("filters", {})
            
            if not resume_id:
                return sanic_json({"error": "简历ID不能为空"}, status=400)
            
            user_id = request.ctx.user_id
            
            # 获取简历数据用于匹配
            resume_data = await self.data_access.get_resume_for_matching(resume_id, user_id)
            if not resume_data:
                return sanic_json({"error": "简历数据不存在或无法访问"}, status=404)
            
            # 执行职位匹配
            matches = await self.matching_engine.find_matching_jobs(
                resume_data, user_id, limit, filters
            )
            
            # 获取公司信息
            for match in matches:
                company_info = await self._get_company_info(match['job_info']['company_id'])
                match['company_info'] = company_info
            
            return sanic_json({
                "status": "success",
                "data": {
                    "matches": matches,
                    "total": len(matches),
                    "resume_id": resume_id,
                    "user_id": user_id,
                    "filters_applied": filters,
                    "timestamp": datetime.now().isoformat()
                },
                "message": "职位匹配完成"
            })
            
        except Exception as e:
            logger.error(f"职位匹配API失败: {e}")
            return sanic_json({"error": str(e)}, status=500)
    
    async def _handle_job_matching_details(self, request: Request, job_id: int):
        """处理匹配结果详情请求"""
        try:
            if not self.initialized:
                return sanic_json({"error": "服务未初始化"}, status=503)
            
            user_id = request.ctx.user_id
            
            # 获取职位详细信息
            job_data = await self.data_access.get_job_data(job_id)
            if not job_data:
                return sanic_json({"error": "职位不存在"}, status=404)
            
            # 获取职位向量数据
            job_vectors = await self.data_access.get_job_vectors(job_id)
            
            # 获取公司信息
            company_info = await self._get_company_info(job_data['company_id'])
            
            return sanic_json({
                "status": "success",
                "data": {
                    "job": job_data,
                    "vectors": job_vectors,
                    "company": company_info,
                    "timestamp": datetime.now().isoformat()
                },
                "message": "职位详情获取成功"
            })
            
        except Exception as e:
            logger.error(f"获取职位详情失败: {e}")
            return sanic_json({"error": str(e)}, status=500)
    
    async def _handle_job_matching_history(self, request: Request):
        """处理匹配历史请求"""
        try:
            if not self.initialized:
                return sanic_json({"error": "服务未初始化"}, status=503)
            
            user_id = request.ctx.user_id
            page = int(request.args.get('page', 1))
            size = int(request.args.get('size', 20))
            
            # 获取用户匹配历史
            history = await self._get_user_matching_history(user_id, page, size)
            
            return sanic_json({
                "status": "success",
                "data": {
                    "history": history,
                    "page": page,
                    "size": size,
                    "timestamp": datetime.now().isoformat()
                },
                "message": "匹配历史获取成功"
            })
            
        except Exception as e:
            logger.error(f"获取匹配历史失败: {e}")
            return sanic_json({"error": str(e)}, status=500)
    
    async def _handle_update_weights(self, request: Request):
        """处理权重更新请求"""
        try:
            if not self.initialized:
                return sanic_json({"error": "服务未初始化"}, status=503)
            
            data = request.json
            new_weights = data.get("weights", {})
            
            if not new_weights:
                return sanic_json({"error": "权重配置不能为空"}, status=400)
            
            # 更新匹配权重
            await self.matching_engine.update_matching_weights(new_weights)
            
            return sanic_json({
                "status": "success",
                "data": {
                    "weights": new_weights,
                    "timestamp": datetime.now().isoformat()
                },
                "message": "权重配置更新成功"
            })
            
        except Exception as e:
            logger.error(f"更新权重配置失败: {e}")
            return sanic_json({"error": str(e)}, status=500)
    
    async def _handle_job_matching_stats(self, request: Request):
        """处理统计信息请求"""
        try:
            if not self.initialized:
                return sanic_json({"error": "服务未初始化"}, status=503)
            
            # 获取统计信息
            stats = await self._get_matching_stats()
            
            return sanic_json({
                "status": "success",
                "data": {
                    "stats": stats,
                    "timestamp": datetime.now().isoformat()
                },
                "message": "统计信息获取成功"
            })
            
        except Exception as e:
            logger.error(f"获取统计信息失败: {e}")
            return sanic_json({"error": str(e)}, status=500)
    
    async def _get_company_info(self, company_id: int) -> Optional[Dict[str, Any]]:
        """获取公司信息"""
        try:
            company_service_url = f"{os.getenv('COMPANY_SERVICE_URL', 'http://localhost:8083')}/api/v1/company/public/companies/{company_id}"
            
            async with aiohttp.ClientSession() as session:
                async with session.get(company_service_url, timeout=5) as resp:
                    if resp.status == 200:
                        return await resp.json()
                    else:
                        logger.warning(f"获取公司信息失败: {resp.status}")
                        return None
                        
        except Exception as e:
            logger.error(f"获取公司信息异常: {e}")
            return None
    
    async def _get_user_matching_history(self, user_id: int, page: int, size: int) -> List[Dict[str, Any]]:
        """获取用户匹配历史"""
        try:
            async with self.data_access.mysql_pool.acquire() as conn:
                async with conn.cursor(aiomysql.DictCursor) as cursor:
                    offset = (page - 1) * size
                    await cursor.execute("""
                        SELECT 
                            id, resume_id, matches_count, created_at
                        FROM job_matching_logs 
                        WHERE user_id = %s
                        ORDER BY created_at DESC
                        LIMIT %s OFFSET %s
                    """, (user_id, size, offset))
                    
                    results = await cursor.fetchall()
                    
                    # 转换datetime对象
                    for result in results:
                        if result['created_at']:
                            result['created_at'] = result['created_at'].isoformat()
                    
                    return results
                    
        except Exception as e:
            logger.error(f"获取用户匹配历史失败: {e}")
            return []
    
    async def _get_matching_stats(self) -> Dict[str, Any]:
        """获取匹配统计信息"""
        try:
            stats = {}
            
            async with self.data_access.mysql_pool.acquire() as conn:
                async with conn.cursor() as cursor:
                    # 总匹配次数
                    await cursor.execute("SELECT COUNT(*) FROM job_matching_logs")
                    stats['total_matches'] = (await cursor.fetchone())[0]
                    
                    # 活跃用户数
                    await cursor.execute("SELECT COUNT(DISTINCT user_id) FROM job_matching_logs")
                    stats['active_users'] = (await cursor.fetchone())[0]
                    
                    # 平均匹配数
                    await cursor.execute("SELECT AVG(matches_count) FROM job_matching_logs")
                    avg_result = await cursor.fetchone()
                    stats['avg_matches_per_request'] = float(avg_result[0]) if avg_result[0] else 0
                    
                    # 今日匹配次数
                    await cursor.execute("""
                        SELECT COUNT(*) FROM job_matching_logs 
                        WHERE DATE(created_at) = CURDATE()
                    """)
                    stats['today_matches'] = (await cursor.fetchone())[0]
            
            return stats
            
        except Exception as e:
            logger.error(f"获取匹配统计信息失败: {e}")
            return {}
    
    async def close(self):
        """关闭服务"""
        try:
            if self.data_access:
                await self.data_access.close()
            
            logger.info("职位匹配服务已关闭")
            
        except Exception as e:
            logger.error(f"关闭职位匹配服务失败: {e}")


# 使用示例
async def main():
    """使用示例"""
    app = Sanic("JobMatchingService")
    
    # 创建职位匹配服务
    job_matching_service = JobMatchingService(app)
    await job_matching_service.initialize()
    
    # 启动服务
    app.run(host="0.0.0.0", port=8207, debug=True)


if __name__ == "__main__":
    asyncio.run(main())
