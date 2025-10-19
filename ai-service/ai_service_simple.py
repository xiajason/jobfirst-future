#!/usr/bin/env python3
"""
AI服务简化版本 - 基础功能实现
提供简历分析、向量生成、职位匹配等AI服务
"""

import asyncio
import logging
import os
import json
import time
from typing import Dict, List, Optional, Any
from datetime import datetime, timedelta

from sanic import Sanic, Request, response
from sanic.response import json as json_response
import asyncpg
import mysql.connector
from mysql.connector import Error
import jwt
import aiohttp
import hashlib
import base64

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# 创建Sanic应用
app = Sanic("ai-service-containerized")

class DatabaseManager:
    """数据库管理器"""
    
    def __init__(self):
        self.pg_pool = None
        self.mysql_conn = None
        
    async def init_postgresql(self):
        """初始化PostgreSQL连接池"""
        try:
            self.pg_pool = await asyncpg.create_pool(
                host=os.getenv("DB_HOST", "host.docker.internal"),
                port=int(os.getenv("DB_PORT", "5432")),
                user=os.getenv("DB_USER", "postgres"),
                password=os.getenv("DB_PASSWORD", "postgres"),
                database=os.getenv("DB_NAME", "jobfirst"),
                min_size=1,
                max_size=10
            )
            logger.info("PostgreSQL连接池初始化成功")
        except Exception as e:
            logger.error(f"PostgreSQL连接池初始化失败: {e}")
            # 如果连接失败，继续运行但不使用PostgreSQL
    
    def init_mysql(self):
        """初始化MySQL连接"""
        try:
            self.mysql_conn = mysql.connector.connect(
                host=os.getenv("MYSQL_HOST", "host.docker.internal"),
                port=int(os.getenv("MYSQL_PORT", "3306")),
                user=os.getenv("MYSQL_USER", "root"),
                password=os.getenv("MYSQL_PASSWORD", ""),
                database=os.getenv("MYSQL_DB", "jobfirst"),
                autocommit=True
            )
            logger.info("MySQL连接初始化成功")
        except Error as e:
            logger.error(f"MySQL连接初始化失败: {e}")
            # 如果连接失败，继续运行但不使用MySQL
    
    async def close_connections(self):
        """关闭所有数据库连接"""
        if self.pg_pool:
            await self.pg_pool.close()
        if self.mysql_conn:
            self.mysql_conn.close()

class AuthManager:
    """认证管理器"""
    
    def __init__(self):
        self.jwt_secret = os.getenv("JWT_SECRET", "jobfirst-unified-auth-secret-key-2024")
        self.jwt_algorithm = "HS256"
        self.zervigo_auth_url = os.getenv("ZERVIGO_AUTH_URL", "http://host.docker.internal:8207")
    
    def verify_token(self, token: str) -> Optional[Dict]:
        """验证JWT token"""
        try:
            payload = jwt.decode(token, self.jwt_secret, algorithms=[self.jwt_algorithm])
            return payload
        except jwt.ExpiredSignatureError:
            logger.warning("Token已过期")
            return None
        except jwt.InvalidTokenError:
            logger.warning("无效的Token")
            return None
    
    async def verify_with_zervigo(self, token: str) -> Optional[Dict]:
        """通过Zervigo验证用户权限"""
        try:
            async with aiohttp.ClientSession() as session:
                headers = {"Authorization": f"Bearer {token}"}
                async with session.get(f"{self.zervigo_auth_url}/api/v1/auth/verify", headers=headers) as response:
                    if response.status == 200:
                        return await response.json()
                    else:
                        logger.warning(f"Zervigo认证失败: {response.status}")
                        return None
        except Exception as e:
            logger.error(f"Zervigo认证请求失败: {e}")
            return None

class SimpleAIManager:
    """简化AI管理器"""
    
    def __init__(self):
        self.model_loaded = True  # 简化版本，假设模型已加载
    
    async def generate_embedding(self, text: str) -> List[float]:
        """生成文本嵌入向量（简化版本）"""
        try:
            # 使用简单的哈希方法生成向量
            text_hash = hashlib.md5(text.encode()).hexdigest()
            
            # 将哈希转换为384维向量
            embedding = []
            for i in range(0, len(text_hash), 2):
                hex_pair = text_hash[i:i+2]
                value = int(hex_pair, 16) / 255.0  # 归一化到0-1
                embedding.append(value)
            
            # 填充到384维
            while len(embedding) < 384:
                embedding.append(0.0)
            
            return embedding[:384]
        except Exception as e:
            logger.error(f"生成嵌入向量失败: {e}")
            return [0.0] * 384

class ResumeAnalysisService:
    """简历分析服务"""
    
    def __init__(self, db_manager: DatabaseManager, ai_manager: SimpleAIManager):
        self.db = db_manager
        self.ai = ai_manager
    
    async def analyze_resume(self, resume_data: Dict, user_id: int) -> Dict:
        """分析简历内容"""
        try:
            logger.info(f"开始分析简历，用户ID: {user_id}")
            
            # 1. 提取关键信息
            key_info = self._extract_key_info(resume_data)
            
            # 2. 生成向量嵌入
            full_text = self._combine_resume_text(resume_data)
            embedding = await self.ai.generate_embedding(full_text)
            
            # 3. 分析技能匹配度
            skills_analysis = self._analyze_skills(resume_data)
            
            # 4. 生成分析结果
            analysis_result = {
                "user_id": user_id,
                "key_info": key_info,
                "embedding": embedding,
                "skills_analysis": skills_analysis,
                "confidence_score": self._calculate_confidence(resume_data),
                "analysis_timestamp": datetime.now().isoformat(),
                "status": "completed"
            }
            
            # 5. 存储分析结果（如果数据库可用）
            await self._store_analysis_result(analysis_result)
            
            logger.info(f"简历分析完成，用户ID: {user_id}, 置信度: {analysis_result['confidence_score']}")
            return analysis_result
            
        except Exception as e:
            logger.error(f"简历分析失败，用户ID: {user_id}, 错误: {e}")
            raise
    
    def _extract_key_info(self, resume_data: Dict) -> Dict:
        """提取关键信息"""
        return {
            "name": resume_data.get("name", ""),
            "email": resume_data.get("email", ""),
            "phone": resume_data.get("phone", ""),
            "education": resume_data.get("education", []),
            "experience": resume_data.get("experience", []),
            "skills": resume_data.get("skills", []),
            "summary": resume_data.get("summary", "")
        }
    
    def _combine_resume_text(self, resume_data: Dict) -> str:
        """合并简历文本"""
        text_parts = []
        
        if resume_data.get("summary"):
            text_parts.append(resume_data["summary"])
        
        if resume_data.get("experience"):
            for exp in resume_data["experience"]:
                text_parts.append(f"{exp.get('title', '')} {exp.get('description', '')}")
        
        if resume_data.get("education"):
            for edu in resume_data["education"]:
                text_parts.append(f"{edu.get('degree', '')} {edu.get('school', '')}")
        
        if resume_data.get("skills"):
            text_parts.append(" ".join(resume_data["skills"]))
        
        return " ".join(text_parts)
    
    def _analyze_skills(self, resume_data: Dict) -> Dict:
        """分析技能"""
        skills = resume_data.get("skills", [])
        return {
            "total_skills": len(skills),
            "technical_skills": [s for s in skills if self._is_technical_skill(s)],
            "soft_skills": [s for s in skills if not self._is_technical_skill(s)],
            "skill_categories": self._categorize_skills(skills)
        }
    
    def _is_technical_skill(self, skill: str) -> bool:
        """判断是否为技术技能"""
        technical_keywords = ["python", "java", "javascript", "react", "vue", "mysql", "postgresql", "docker", "kubernetes"]
        return any(keyword in skill.lower() for keyword in technical_keywords)
    
    def _categorize_skills(self, skills: List[str]) -> Dict:
        """技能分类"""
        categories = {
            "programming": [],
            "database": [],
            "framework": [],
            "tool": [],
            "other": []
        }
        
        for skill in skills:
            skill_lower = skill.lower()
            if any(keyword in skill_lower for keyword in ["python", "java", "javascript", "go", "rust"]):
                categories["programming"].append(skill)
            elif any(keyword in skill_lower for keyword in ["mysql", "postgresql", "mongodb", "redis"]):
                categories["database"].append(skill)
            elif any(keyword in skill_lower for keyword in ["react", "vue", "angular", "django", "flask"]):
                categories["framework"].append(skill)
            elif any(keyword in skill_lower for keyword in ["docker", "kubernetes", "git", "jenkins"]):
                categories["tool"].append(skill)
            else:
                categories["other"].append(skill)
        
        return categories
    
    def _calculate_confidence(self, resume_data: Dict) -> float:
        """计算置信度分数"""
        score = 0.0
        
        # 基础信息完整性
        if resume_data.get("name"):
            score += 0.2
        if resume_data.get("email"):
            score += 0.2
        if resume_data.get("phone"):
            score += 0.1
        
        # 内容完整性
        if resume_data.get("summary"):
            score += 0.2
        if resume_data.get("experience"):
            score += 0.2
        if resume_data.get("education"):
            score += 0.1
        
        return min(score, 1.0)
    
    async def _store_analysis_result(self, result: Dict):
        """存储分析结果"""
        try:
            if self.db.pg_pool:
                async with self.db.pg_pool.acquire() as conn:
                    await conn.execute("""
                        INSERT INTO resume_analysis_results 
                        (user_id, key_info, embedding, skills_analysis, confidence_score, analysis_timestamp, status)
                        VALUES ($1, $2, $3, $4, $5, $6, $7)
                    """, 
                    result["user_id"],
                    json.dumps(result["key_info"]),
                    result["embedding"],
                    json.dumps(result["skills_analysis"]),
                    result["confidence_score"],
                    result["analysis_timestamp"],
                    result["status"]
                    )
        except Exception as e:
            logger.error(f"存储分析结果失败: {e}")

class JobMatchingService:
    """职位匹配服务"""
    
    def __init__(self, db_manager: DatabaseManager, ai_manager: SimpleAIManager):
        self.db = db_manager
        self.ai = ai_manager
    
    async def find_matching_jobs(self, user_id: int, limit: int = 10) -> List[Dict]:
        """查找匹配的职位"""
        try:
            logger.info(f"开始查找匹配职位，用户ID: {user_id}")
            
            # 1. 获取用户简历向量
            user_embedding = await self._get_user_embedding(user_id)
            if not user_embedding:
                return []
            
            # 2. 计算与所有职位的相似度
            matching_jobs = await self._calculate_similarity(user_embedding, limit)
            
            # 3. 排序并返回结果
            matching_jobs.sort(key=lambda x: x["similarity_score"], reverse=True)
            
            logger.info(f"职位匹配完成，用户ID: {user_id}, 匹配数量: {len(matching_jobs)}")
            return matching_jobs
            
        except Exception as e:
            logger.error(f"职位匹配失败，用户ID: {user_id}, 错误: {e}")
            raise
    
    async def _get_user_embedding(self, user_id: int) -> Optional[List[float]]:
        """获取用户简历向量"""
        try:
            if self.db.pg_pool:
                async with self.db.pg_pool.acquire() as conn:
                    row = await conn.fetchrow("""
                        SELECT embedding FROM resume_analysis_results 
                        WHERE user_id = $1 AND status = 'completed'
                        ORDER BY analysis_timestamp DESC LIMIT 1
                    """, user_id)
                    return row["embedding"] if row else None
        except Exception as e:
            logger.error(f"获取用户向量失败，用户ID: {user_id}, 错误: {e}")
            return None
    
    async def _calculate_similarity(self, user_embedding: List[float], limit: int) -> List[Dict]:
        """计算相似度"""
        try:
            # 这里应该从数据库获取职位向量，目前返回模拟数据
            matching_jobs = []
            for i in range(limit):
                matching_jobs.append({
                    "job_id": f"job_{i+1}",
                    "title": f"软件工程师 {i+1}",
                    "company": f"公司 {i+1}",
                    "similarity_score": 0.9 - (i * 0.05),
                    "match_reasons": ["技能匹配", "经验匹配", "教育背景匹配"]
                })
            return matching_jobs
        except Exception as e:
            logger.error(f"计算相似度失败: {e}")
            return []

# 全局服务实例
db_manager = DatabaseManager()
auth_manager = AuthManager()
ai_manager = SimpleAIManager()
resume_service = ResumeAnalysisService(db_manager, ai_manager)
job_matching_service = JobMatchingService(db_manager, ai_manager)

@app.before_server_start
async def setup_services(app, loop):
    """服务启动前初始化"""
    try:
        logger.info("开始初始化AI服务")
        
        # 初始化数据库连接
        await db_manager.init_postgresql()
        db_manager.init_mysql()
        
        logger.info("AI服务初始化完成")
    except Exception as e:
        logger.error(f"AI服务初始化失败: {e}")
        # 即使初始化失败也继续运行

@app.after_server_stop
async def cleanup_services(app, loop):
    """服务停止后清理"""
    try:
        await db_manager.close_connections()
        logger.info("AI服务清理完成")
    except Exception as e:
        logger.error(f"AI服务清理失败: {e}")

@app.middleware('request')
async def auth_middleware(request: Request):
    """认证中间件"""
    # 跳过健康检查和公开端点
    if request.path in ['/health', '/api/v1/status']:
        return
    
    # 检查Authorization头
    auth_header = request.headers.get('Authorization')
    if not auth_header or not auth_header.startswith('Bearer '):
        return json_response({"error": "认证失败", "code": "AUTH_REQUIRED"}, status=401)
    
    token = auth_header.split(' ')[1]
    user_info = auth_manager.verify_token(token)
    
    if not user_info:
        return json_response({"error": "认证失败", "code": "INVALID_TOKEN"}, status=401)
    
    # 将用户信息添加到请求上下文
    request.ctx.user = user_info

@app.route("/health", methods=["GET"])
async def health_check(request: Request):
    """健康检查"""
    try:
        # 检查数据库连接
        db_status = "healthy"
        if not db_manager.pg_pool:
            db_status = "unhealthy"
        
        # 检查AI模型
        ai_status = "healthy" if ai_manager.model_loaded else "loading"
        
        return json_response({
            "status": "healthy",
            "service": "ai-service-containerized",
            "version": "1.0.0",
            "timestamp": datetime.now().isoformat(),
            "database_status": db_status,
            "ai_model_status": ai_status,
            "zervigo_auth_status": "integrated"
        })
    except Exception as e:
        logger.error(f"健康检查失败: {e}")
        return json_response({
            "status": "unhealthy",
            "error": str(e)
        }, status=500)

@app.route("/api/v1/status", methods=["GET"])
async def service_status(request: Request):
    """服务状态"""
    return json_response({
        "status": "success",
        "service": "ai-service-containerized",
        "version": "1.0.0",
        "features": [
            "resume_analysis",
            "job_matching",
            "vector_generation",
            "authentication"
        ],
        "database_connected": db_manager.pg_pool is not None,
        "ai_model_loaded": ai_manager.model_loaded
    })

@app.route("/api/v1/ai/resume-analysis", methods=["POST"])
async def analyze_resume(request: Request):
    """简历分析"""
    try:
        data = request.json
        user_id = request.ctx.user.get("user_id")
        
        if not data:
            return json_response({"error": "请求数据不能为空"}, status=400)
        
        # 执行简历分析
        result = await resume_service.analyze_resume(data, user_id)
        
        return json_response({
            "status": "success",
            "message": "简历分析完成",
            "result": result
        })
        
    except Exception as e:
        logger.error(f"简历分析API失败: {e}")
        return json_response({"error": str(e)}, status=500)

@app.route("/api/v1/ai/job-matching", methods=["POST"])
async def find_job_matches(request: Request):
    """职位匹配"""
    try:
        data = request.json or {}
        user_id = request.ctx.user.get("user_id")
        limit = data.get("limit", 10)
        
        # 查找匹配职位
        matches = await job_matching_service.find_matching_jobs(user_id, limit)
        
        return json_response({
            "success": True,
            "data": matches,
            "message": "职位匹配完成",
            "timestamp": datetime.now().isoformat(),
            "metadata": {
                "total": len(matches),
                "user_id": user_id,
                "limit": limit
            }
        })
        
    except Exception as e:
        logger.error(f"职位匹配API失败: {e}")
        return json_response({"error": str(e)}, status=500)

@app.route("/api/v1/ai/enhanced-job-matching", methods=["POST"])
async def enhanced_job_matching(request: Request):
    """增强版职位匹配"""
    try:
        data = request.json or {}
        user_id = request.ctx.user.get("user_id")
        resume_id = data.get("resume_id")
        limit = data.get("limit", 10)
        filters = data.get("filters", {})
        
        if not resume_id:
            return json_response({
                "success": False,
                "message": "resume_id is required"
            }, status=400)
        
        # 使用增强版匹配引擎
        from enhanced_job_matching_engine import EnhancedJobMatchingEngine
        
        # 创建匹配引擎实例（使用简化初始化）
        matching_engine = EnhancedJobMatchingEngine(None, None)
        matches = await matching_engine.find_enhanced_matches(
            user_id=user_id,
            resume_id=resume_id,
            limit=limit,
            filters=filters
        )
        
        return json_response({
            "success": True,
            "data": matches,
            "message": "增强版职位匹配完成",
            "timestamp": datetime.now().isoformat(),
            "metadata": {
                "total": len(matches),
                "user_id": user_id,
                "resume_id": resume_id,
                "limit": limit,
                "filters": filters,
                "engine": "enhanced"
            }
        })
        
    except Exception as e:
        logger.error(f"增强版职位匹配API失败: {e}")
        return json_response({
            "success": False,
            "error": str(e)
        }, status=500)

@app.route("/api/v1/ai/matching/recommendations/<resume_id>", methods=["GET"])
async def get_matching_recommendations(request: Request, resume_id: int):
    """获取匹配推荐建议"""
    try:
        user_id = request.ctx.user.get("user_id")
        
        # 使用增强版匹配引擎生成推荐
        from enhanced_job_matching_engine import EnhancedJobMatchingEngine
        
        matching_engine = EnhancedJobMatchingEngine(None, None)
        recommendations = await matching_engine.generate_recommendations(
            user_id=user_id,
            resume_id=resume_id
        )
        
        return json_response({
            "success": True,
            "data": recommendations,
            "message": "匹配推荐建议获取成功",
            "timestamp": datetime.now().isoformat(),
            "metadata": {
                "user_id": user_id,
                "resume_id": resume_id
            }
        })
        
    except Exception as e:
        logger.error(f"获取匹配推荐建议失败: {e}")
        return json_response({
            "success": False,
            "error": str(e)
        }, status=500)

@app.route("/api/v1/ai/matching/analysis/<resume_id>", methods=["GET"])
async def get_matching_analysis(request: Request, resume_id: int):
    """获取匹配分析报告"""
    try:
        user_id = request.ctx.user.get("user_id")
        
        # 使用增强版匹配引擎生成分析
        from enhanced_job_matching_engine import EnhancedJobMatchingEngine
        
        matching_engine = EnhancedJobMatchingEngine(None, None)
        analysis = await matching_engine.generate_analysis(
            user_id=user_id,
            resume_id=resume_id
        )
        
        return json_response({
            "success": True,
            "data": analysis,
            "message": "匹配分析报告获取成功",
            "timestamp": datetime.now().isoformat(),
            "metadata": {
                "user_id": user_id,
                "resume_id": resume_id
            }
        })
        
    except Exception as e:
        logger.error(f"获取匹配分析报告失败: {e}")
        return json_response({
            "success": False,
            "error": str(e)
        }, status=500)

@app.route("/api/v1/ai/embedding", methods=["POST"])
async def generate_embedding(request: Request):
    """生成文本嵌入向量"""
    try:
        data = request.json
        text = data.get("text")
        
        if not text:
            return json_response({"error": "文本不能为空"}, status=400)
        
        # 生成嵌入向量
        embedding = await ai_manager.generate_embedding(text)
        
        return json_response({
            "status": "success",
            "embedding": embedding,
            "dimension": len(embedding)
        })
        
    except Exception as e:
        logger.error(f"生成嵌入向量API失败: {e}")
        return json_response({"error": str(e)}, status=500)

@app.route("/api/v1/ai/chat", methods=["POST"])
async def ai_chat(request: Request):
    """AI聊天"""
    try:
        data = request.json
        message = data.get("message")
        user_id = request.ctx.user.get("user_id")
        
        if not message:
            return json_response({"error": "消息不能为空"}, status=400)
        
        # 简单的AI回复逻辑
        response_text = f"收到您的消息：{message}。这是AI服务的回复。"
        
        return json_response({
            "status": "success",
            "response": response_text,
            "user_id": user_id,
            "timestamp": datetime.now().isoformat()
        })
        
    except Exception as e:
        logger.error(f"AI聊天API失败: {e}")
        return json_response({"error": str(e)}, status=500)

if __name__ == "__main__":
    # 启动服务
    logger.info("启动AI服务容器化版本，端口: 8206")
    app.run(
        host="0.0.0.0",
        port=8206,
        workers=1,
        debug=False,
        access_log=True
    )
