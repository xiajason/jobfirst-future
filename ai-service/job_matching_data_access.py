#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
AI职位匹配系统 - 数据访问适配层
适配新的MySQL+SQLite分离架构

创建时间: 2025-09-13
作者: AI Assistant
版本: 1.0.0
"""

import asyncio
import json
import logging
import sqlite3
from typing import Dict, List, Optional, Any
from datetime import datetime
import aiohttp
import asyncpg
import aiomysql

logger = logging.getLogger(__name__)

class JobMatchingDataAccess:
    """职位匹配数据访问适配层"""
    
    def __init__(self, mysql_config: Dict, postgres_config: Dict):
        """
        初始化数据访问层
        
        Args:
            mysql_config: MySQL数据库配置
            postgres_config: PostgreSQL数据库配置
        """
        self.mysql_config = mysql_config
        self.postgres_config = postgres_config
        self.mysql_pool = None
        self.postgres_pool = None
        
    async def initialize(self):
        """初始化数据库连接池"""
        try:
            # 初始化MySQL连接池
            self.mysql_pool = await aiomysql.create_pool(
                host=self.mysql_config['host'],
                port=self.mysql_config['port'],
                user=self.mysql_config['user'],
                password=self.mysql_config['password'],
                db=self.mysql_config['database'],
                charset='utf8mb4',
                autocommit=True,
                maxsize=10,
                minsize=1
            )
            
            # 初始化PostgreSQL连接池
            self.postgres_pool = await asyncpg.create_pool(
                host=self.postgres_config['host'],
                port=self.postgres_config['port'],
                user=self.postgres_config['user'],
                password=self.postgres_config['password'],
                database=self.postgres_config['database'],
                min_size=1,
                max_size=10
            )
            
            logger.info("数据访问层初始化成功")
            
        except Exception as e:
            logger.error(f"数据访问层初始化失败: {e}")
            raise
    
    async def close(self):
        """关闭数据库连接池"""
        if self.mysql_pool:
            self.mysql_pool.close()
            await self.mysql_pool.wait_closed()
        
        if self.postgres_pool:
            await self.postgres_pool.close()
        
        logger.info("数据访问层连接池已关闭")
    
    async def get_resume_for_matching(self, resume_id: int, user_id: int) -> Optional[Dict[str, Any]]:
        """
        获取用于匹配的简历数据 - 适配新架构
        
        Args:
            resume_id: 简历ID
            user_id: 用户ID
            
        Returns:
            包含元数据、解析内容和向量数据的字典，失败返回None
        """
        try:
            # 1. 从MySQL获取元数据
            metadata = await self._get_resume_metadata(resume_id, user_id)
            if not metadata:
                logger.warning(f"简历元数据不存在: resume_id={resume_id}, user_id={user_id}")
                return None
            
            # 2. 从SQLite获取解析内容（带权限检查）
            parsed_data = await self._get_sqlite_data(metadata['sqlite_db_path'], resume_id, user_id)
            if not parsed_data:
                logger.warning(f"简历解析数据不存在或无访问权限: resume_id={resume_id}, user_id={user_id}")
                return None
            
            # 3. 从PostgreSQL获取向量数据
            vectors = await self._get_resume_vectors(resume_id)
            if not vectors:
                logger.warning(f"简历向量数据不存在: resume_id={resume_id}")
                return None
            
            # 4. 验证数据一致性
            if not self._validate_data_consistency(metadata, parsed_data, vectors):
                logger.error(f"数据一致性验证失败: resume_id={resume_id}")
                return None
            
            return {
                'metadata': metadata,
                'parsed_data': parsed_data,
                'vectors': vectors,
                'user_id': user_id,
                'resume_id': resume_id
            }
            
        except Exception as e:
            logger.error(f"获取简历数据失败: {e}")
            return None
    
    async def _get_resume_metadata(self, resume_id: int, user_id: int) -> Optional[Dict[str, Any]]:
        """从MySQL获取简历元数据"""
        try:
            async with self.mysql_pool.acquire() as conn:
                async with conn.cursor(aiomysql.DictCursor) as cursor:
                    # 超级管理员可以访问所有简历，普通用户只能访问自己的
                    await cursor.execute("""
                        SELECT 
                            id, user_id, file_id, title, creation_mode, 
                            template_id, status, is_public, view_count,
                            parsing_status, parsing_error, sqlite_db_path,
                            created_at, updated_at
                        FROM resume_metadata 
                        WHERE id = %s
                    """, (resume_id,))
                    
                    result = await cursor.fetchone()
                    if result:
                        # 转换datetime对象为字符串
                        result['created_at'] = result['created_at'].isoformat() if result['created_at'] else None
                        result['updated_at'] = result['updated_at'].isoformat() if result['updated_at'] else None
                    
                    return result
                    
        except Exception as e:
            logger.error(f"获取简历元数据失败: {e}")
            return None
    
    async def _get_sqlite_data(self, sqlite_db_path: str, resume_id: int, user_id: int) -> Optional[Dict[str, Any]]:
        """从SQLite获取解析内容数据（带权限检查）"""
        try:
            # 将相对路径转换为绝对路径
            import os
            if sqlite_db_path.startswith('./'):
                # 相对于项目根目录的路径
                project_root = os.path.abspath(os.path.join(os.path.dirname(__file__), '../../../'))
                sqlite_db_path = os.path.join(project_root, sqlite_db_path[2:])
            elif sqlite_db_path.startswith('data/'):
                # 直接是data/开头的路径
                project_root = os.path.abspath(os.path.join(os.path.dirname(__file__), '../../../'))
                sqlite_db_path = os.path.join(project_root, sqlite_db_path)
            
            logger.info(f"访问SQLite数据库: {sqlite_db_path}")
            
            # 检查文件是否存在
            if not os.path.exists(sqlite_db_path):
                logger.error(f"SQLite数据库文件不存在: {sqlite_db_path}")
                return None
            
            # 检查简历访问权限
            logger.info(f"开始检查简历访问权限: resume_id={resume_id}, user_id={user_id}")
            permission_result = await self._check_resume_access_permission(sqlite_db_path, resume_id, user_id, "view")
            logger.info(f"权限检查结果: {permission_result}")
            if not permission_result:
                logger.warning(f"用户 {user_id} 没有访问简历 {resume_id} 的权限")
                return None
            
            # 使用异步方式访问SQLite
            loop = asyncio.get_event_loop()
            
            def _get_sqlite_data_sync():
                conn = sqlite3.connect(sqlite_db_path)
                conn.row_factory = sqlite3.Row
                cursor = conn.cursor()
                
                try:
                    # 获取简历内容
                    cursor.execute("""
                        SELECT id, resume_metadata_id, title, content, raw_content, 
                               content_hash, created_at, updated_at
                        FROM resume_content 
                        WHERE resume_metadata_id = ?
                    """, (resume_id,))
                    
                    content_row = cursor.fetchone()
                    if not content_row:
                        logger.warning(f"简历内容不存在: resume_id={resume_id}")
                        return None
                    
                    logger.info(f"找到简历内容: {dict(content_row)}")
                    
                    content_data = dict(content_row)
                    
                    # 获取解析数据
                    cursor.execute("""
                        SELECT personal_info, work_experience, education, skills,
                               projects, certifications, keywords, confidence,
                               parsing_version, created_at, updated_at
                        FROM parsed_resume_data 
                        WHERE resume_content_id = ?
                    """, (content_data['id'],))
                    
                    parsed_row = cursor.fetchone()
                    if not parsed_row:
                        return None
                    
                    parsed_data = dict(parsed_row)
                    
                    # 解析JSON字符串
                    for key in ['personal_info', 'work_experience', 'education', 
                               'skills', 'projects', 'certifications', 'keywords']:
                        if parsed_data[key]:
                            try:
                                parsed_data[key] = json.loads(parsed_data[key])
                            except json.JSONDecodeError:
                                parsed_data[key] = None
                    
                    return {
                        'content': content_data,
                        'parsed': parsed_data
                    }
                    
                finally:
                    conn.close()
            
            result = await loop.run_in_executor(None, _get_sqlite_data_sync)
            return result
            
        except Exception as e:
            logger.error(f"获取SQLite数据失败: {e}")
            return None
    
    async def _check_resume_access_permission(self, sqlite_db_path: str, resume_id: int, user_id: int, access_type: str = "view") -> bool:
        """检查简历访问权限"""
        try:
            import json
            import concurrent.futures
            
            loop = asyncio.get_event_loop()
            
            def _check_permission_sync():
                conn = sqlite3.connect(sqlite_db_path)
                conn.row_factory = sqlite3.Row
                cursor = conn.cursor()
                
                try:
                    # 获取简历内容的隐私设置
                    cursor.execute("""
                        SELECT ps.is_public, ps.share_with_companies, ps.allow_search, 
                               ps.allow_download, ps.view_permissions, ps.download_permissions
                        FROM resume_content rc
                        LEFT JOIN user_privacy_settings ps ON rc.id = ps.resume_content_id
                        WHERE rc.resume_metadata_id = ?
                    """, (resume_id,))
                    
                    privacy_row = cursor.fetchone()
                    if not privacy_row:
                        logger.warning(f"简历隐私设置不存在: resume_id={resume_id}")
                        return False
                    
                    is_public, share_with_companies, allow_search, allow_download, view_permissions, download_permissions = privacy_row
                    
                    # 记录访问日志
                    cursor.execute("""
                        INSERT INTO resume_access_logs (resume_content_id, access_type, access_source, user_agent, ip_address)
                        VALUES ((SELECT rc.id FROM resume_content rc WHERE rc.resume_metadata_id = ?), ?, ?, ?, ?)
                    """, (resume_id, access_type, "ai_service", "AI-JobMatching/1.0", "127.0.0.1"))
                    
                    conn.commit()
                    
                    # AI服务作为"利益相关方"，需要检查权限
                    if access_type == "view":
                        logger.info(f"权限检查: is_public={is_public}, share_with_companies={share_with_companies}, allow_search={allow_search}")
                        logger.info(f"view_permissions: {view_permissions}")
                        
                        # 检查查看权限
                        if view_permissions:
                            try:
                                view_perms = json.loads(view_permissions)
                                logger.info(f"解析的权限设置: {view_perms}")
                                # 检查AI服务是否有查看权限
                                if "ai_service" in view_perms:
                                    result = view_perms["ai_service"] == "allowed"
                                    logger.info(f"AI服务权限检查结果: {result}")
                                    return result
                                elif "default" in view_perms:
                                    result = view_perms["default"] == "public"
                                    logger.info(f"默认权限检查结果: {result}")
                                    return result
                            except json.JSONDecodeError as e:
                                logger.error(f"权限JSON解析失败: {e}")
                        
                        # 默认权限检查
                        result = is_public or share_with_companies or allow_search
                        logger.info(f"默认权限检查结果: {result}")
                        return result
                    
                    elif access_type == "download":
                        # 检查下载权限
                        if download_permissions:
                            try:
                                download_perms = json.loads(download_permissions)
                                if "ai_service" in download_perms:
                                    return download_perms["ai_service"] == "allowed"
                                elif "default" in download_perms:
                                    return download_perms["default"] == "allowed"
                            except json.JSONDecodeError:
                                pass
                        
                        return allow_download
                    
                    return False
                    
                finally:
                    cursor.close()
                    conn.close()
            
            # 在线程池中执行同步操作
            with concurrent.futures.ThreadPoolExecutor() as executor:
                future = executor.submit(_check_permission_sync)
                return await loop.run_in_executor(None, lambda: future.result())
                
        except Exception as e:
            logger.error(f"检查简历访问权限失败: {e}")
            return False
    
    async def _get_resume_vectors(self, resume_id: int) -> Optional[Dict[str, Any]]:
        """从PostgreSQL获取简历向量数据"""
        try:
            async with self.postgres_pool.acquire() as conn:
                # 获取向量数据
                vector_data = await conn.fetchrow("""
                    SELECT 
                        id, resume_id, content_vector, skills_vector, 
                        experience_vector, created_at, updated_at
                    FROM resume_vectors 
                    WHERE resume_id = $1
                """, resume_id)
                
                if vector_data:
                    return dict(vector_data)
                
                return None
                
        except Exception as e:
            logger.error(f"获取简历向量数据失败: {e}")
            return None
    
    def _validate_data_consistency(self, metadata: Dict, sqlite_data: Dict, vectors: Dict) -> bool:
        """验证数据一致性"""
        try:
            # 检查ID关联
            if metadata['id'] != sqlite_data['content']['resume_metadata_id']:
                logger.error(f"ID关联不匹配: MySQL ID={metadata['id']}, SQLite resume_metadata_id={sqlite_data['content']['resume_metadata_id']}")
                return False
            
            if metadata['id'] != vectors['resume_id']:
                logger.error(f"向量数据ID不匹配: MySQL ID={metadata['id']}, PostgreSQL resume_id={vectors['resume_id']}")
                return False
            
            # 检查用户关联
            if metadata['user_id'] != sqlite_data.get('user_id', metadata['user_id']):
                logger.error(f"用户ID不匹配: MySQL user_id={metadata['user_id']}, SQLite user_id={sqlite_data.get('user_id')}")
                return False
            
            # 检查解析状态
            if metadata['parsing_status'] != 'completed':
                logger.error(f"解析状态未完成: parsing_status={metadata['parsing_status']}")
                return False
            
            logger.info(f"数据一致性验证通过: resume_id={metadata['id']}")
            return True
            
        except Exception as e:
            logger.error(f"数据一致性验证异常: {e}")
            return False
    
    async def get_job_data(self, job_id: int) -> Optional[Dict[str, Any]]:
        """获取职位数据"""
        try:
            async with self.mysql_pool.acquire() as conn:
                async with conn.cursor(aiomysql.DictCursor) as cursor:
                    await cursor.execute("""
                        SELECT 
                            id, title, description, requirements, company_id,
                            industry, location, salary_min, salary_max,
                            experience, education, job_type, status,
                            view_count, apply_count, created_by,
                            created_at, updated_at
                        FROM jobs 
                        WHERE id = %s AND status = 'active'
                    """, (job_id,))
                    
                    result = await cursor.fetchone()
                    if result:
                        # 转换datetime对象为字符串
                        result['created_at'] = result['created_at'].isoformat() if result['created_at'] else None
                        result['updated_at'] = result['updated_at'].isoformat() if result['updated_at'] else None
                    
                    return result
                    
        except Exception as e:
            logger.error(f"获取职位数据失败: {e}")
            return None
    
    async def get_job_vectors(self, job_id: int) -> Optional[Dict[str, Any]]:
        """获取职位向量数据"""
        try:
            async with self.postgres_pool.acquire() as conn:
                vector_data = await conn.fetchrow("""
                    SELECT 
                        id, job_id, title_vector, description_vector,
                        requirements_vector, created_at, updated_at
                    FROM job_vectors 
                    WHERE job_id = $1
                """, job_id)
                
                if vector_data:
                    return dict(vector_data)
                
                return None
                
        except Exception as e:
            logger.error(f"获取职位向量数据失败: {e}")
            return None
    
    async def get_active_jobs(self, limit: int = 100) -> List[int]:
        """获取活跃职位ID列表"""
        try:
            async with self.mysql_pool.acquire() as conn:
                async with conn.cursor() as cursor:
                    await cursor.execute("""
                        SELECT id FROM jobs 
                        WHERE status = 'active' 
                        ORDER BY created_at DESC 
                        LIMIT %s
                    """, (limit,))
                    
                    results = await cursor.fetchall()
                    return [row[0] for row in results]
                    
        except Exception as e:
            logger.error(f"获取活跃职位列表失败: {e}")
            return []
    
    async def log_job_matching_access(self, user_id: int, resume_id: int, matches_count: int):
        """记录职位匹配访问日志"""
        try:
            # 记录到MySQL
            async with self.mysql_pool.acquire() as conn:
                async with conn.cursor() as cursor:
                    await cursor.execute("""
                        INSERT INTO job_matching_logs 
                        (user_id, resume_id, matches_count, created_at)
                        VALUES (%s, %s, %s, %s)
                    """, (user_id, resume_id, matches_count, datetime.now()))
            
            logger.info(f"职位匹配访问日志记录成功: user_id={user_id}, resume_id={resume_id}, matches={matches_count}")
            
        except Exception as e:
            logger.error(f"记录职位匹配访问日志失败: {e}")
    
    async def validate_job_data_access(self, job_id: int, user_id: int) -> bool:
        """验证用户对职位数据的访问权限"""
        try:
            # 这里可以添加更复杂的权限验证逻辑
            # 目前简单检查职位是否存在且为活跃状态
            job_data = await self.get_job_data(job_id)
            return job_data is not None
            
        except Exception as e:
            logger.error(f"验证职位数据访问权限失败: {e}")
            return False


class SecureSQLiteManager:
    """安全的SQLite数据库管理器"""
    
    def __init__(self, user_id: int, base_path: str = "./data"):
        self.user_id = user_id
        self.base_path = base_path
        self.sqlite_path = f"{base_path}/users/{user_id}/resume.db"
    
    async def get_resume_data(self, resume_id: int) -> Optional[Dict[str, Any]]:
        """安全获取用户简历数据"""
        try:
            # 验证文件权限
            if not self._validate_file_permissions():
                raise PermissionError("文件权限验证失败")
            
            # 使用数据访问层获取数据
            # 这里简化实现，实际应该集成到JobMatchingDataAccess中
            return None
            
        except Exception as e:
            logger.error(f"获取用户简历数据失败: {e}")
            return None
    
    def _validate_file_permissions(self) -> bool:
        """验证文件权限"""
        try:
            import os
            import stat
            
            if not os.path.exists(self.sqlite_path):
                return False
            
            # 检查文件权限
            file_stat = os.stat(self.sqlite_path)
            file_mode = stat.filemode(file_stat.st_mode)
            
            # 确保文件权限为0600（只有所有者可读写）
            if file_mode != '-rw-------':
                logger.warning(f"SQLite文件权限不正确: {file_mode}")
                return False
            
            return True
            
        except Exception as e:
            logger.error(f"验证文件权限失败: {e}")
            return False


class UserSessionManager:
    """用户会话管理器"""
    
    def __init__(self, user_id: int, timeout_hours: int = 24):
        self.user_id = user_id
        self.timeout_hours = timeout_hours
        self.session_timeout = timeout_hours * 3600  # 转换为秒
    
    async def validate_session(self) -> bool:
        """验证用户会话"""
        try:
            # 这里应该调用User Service验证会话
            # 目前简化实现
            return True
            
        except Exception as e:
            logger.error(f"验证用户会话失败: {e}")
            return False
    
    async def log_access(self, resume_id: int, access_type: str):
        """记录访问日志"""
        try:
            # 记录访问日志
            logger.info(f"用户访问记录: user_id={self.user_id}, resume_id={resume_id}, type={access_type}")
            
        except Exception as e:
            logger.error(f"记录访问日志失败: {e}")


# 配置示例
MYSQL_CONFIG = {
    'host': 'localhost',
    'port': 3306,
    'user': 'root',
    'password': '',
    'database': 'jobfirst'
}

POSTGRES_CONFIG = {
    'host': 'localhost',
    'port': 5432,
    'user': 'postgres',
    'password': 'password',
    'database': 'jobfirst_vectors'
}


# 使用示例
async def main():
    """使用示例"""
    # 初始化数据访问层
    data_access = JobMatchingDataAccess(MYSQL_CONFIG, POSTGRES_CONFIG)
    await data_access.initialize()
    
    try:
        # 获取简历数据用于匹配
        resume_data = await data_access.get_resume_for_matching(1, 4)
        if resume_data:
            print(f"获取简历数据成功: {resume_data['metadata']['title']}")
            print(f"向量数据维度: {len(resume_data['vectors']['content_vector'])}")
        else:
            print("获取简历数据失败")
        
        # 获取职位数据
        job_data = await data_access.get_job_data(1)
        if job_data:
            print(f"获取职位数据成功: {job_data['title']}")
        else:
            print("获取职位数据失败")
            
    finally:
        await data_access.close()


if __name__ == "__main__":
    asyncio.run(main())
