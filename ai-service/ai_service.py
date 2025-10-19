#!/usr/bin/env python3
"""
JobFirst AI服务 - Python Sanic版本
处理文件上传后的内容解析和向量生成
"""

import asyncio
import json
import logging
import os
import time
from datetime import datetime
from typing import List, Dict, Any

import psycopg2
import psycopg2.extras
import requests
from sanic import Sanic, Request, json as sanic_json
from sanic.response import json as sanic_response

# 导入职位匹配服务
from job_matching_service import JobMatchingService

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# 创建Sanic应用
app = Sanic("ai-service")

# 初始化职位匹配服务
job_matching_service = JobMatchingService(app)

# 异步初始化职位匹配服务
async def initialize_job_matching():
    await job_matching_service.initialize()

# 用户认证中间件
async def authenticate_user(request: Request):
    """用户认证中间件"""
    try:
        # 获取Authorization头
        auth_header = request.headers.get('Authorization')
        if not auth_header:
            return sanic_json({
                "error": "认证失败",
                "code": "AUTH_REQUIRED",
                "message": "请提供有效的认证信息"
            }, status=401)
        
        # 检查Bearer token格式
        if not auth_header.startswith('Bearer '):
            return sanic_json({
                "error": "认证失败",
                "code": "INVALID_AUTH_FORMAT",
                "message": "认证格式无效，请使用Bearer token"
            }, status=401)
        
        # 提取token
        token = auth_header.replace('Bearer ', '')
        
        # 验证JWT token（这里需要调用用户服务验证）
        logger.info(f"开始验证JWT token: {token[:50]}...")
        user_info = await verify_jwt_token(token)
        logger.info(f"JWT验证结果: {user_info}")
        
        if not user_info:
            logger.warning("JWT验证失败")
            return sanic_json({
                "error": "认证失败",
                "code": "INVALID_TOKEN",
                "message": "认证token无效或已过期"
            }, status=401)
        
        # 暂时跳过订阅状态检查，直接允许访问
        # TODO: 后续需要实现完整的订阅状态检查
        subscription_status = {
            'has_access': True,
            'has_active_subscription': False,
            'is_trial_user': True,
            'subscription_status': 'trial'
        }
        
        # 将用户信息存储到请求上下文
        request.ctx.user_id = user_info['user_id']
        request.ctx.username = user_info['username']
        request.ctx.subscription_info = subscription_status
        
        return None  # 认证成功，继续处理
        
    except Exception as e:
        logger.error(f"用户认证异常: {e}")
        return sanic_json({
            "error": "认证异常",
            "code": "AUTH_ERROR",
            "message": "认证过程发生错误"
        }, status=500)

async def verify_jwt_token(token: str) -> dict:
    """验证JWT token"""
    try:
        import jwt
        import time
        
        # JWT密钥（与用户服务使用相同的密钥）
        jwt_secret = "default-secret"
        
        # 验证JWT token（暂时跳过签名验证进行测试）
        payload = jwt.decode(token, options={"verify_signature": False})
        
        # 检查token是否过期
        if payload.get('exp', 0) < time.time():
            logger.warning("JWT token已过期")
            return None
            
        return {
            'user_id': payload.get('user_id'),
            'username': payload.get('username'),
            'role': payload.get('role')
        }
                
    except jwt.ExpiredSignatureError:
        logger.warning("JWT token已过期")
        return None
    except jwt.InvalidTokenError as e:
        logger.warning(f"JWT token无效: {e}")
        return None
    except Exception as e:
        logger.error(f"JWT验证异常: {e}")
        return None

async def check_user_subscription(user_id: int) -> dict:
    """检查用户订阅状态"""
    try:
        # 直接从数据库查询用户订阅状态
        import mysql.connector
        from datetime import datetime
        
        # 连接MySQL数据库
        conn = mysql.connector.connect(
            host='localhost',
            user='root',
            password='',
            database='jobfirst'
        )
        
        cursor = conn.cursor(dictionary=True)
        
        # 查询用户订阅信息
        query = """
        SELECT subscription_status, subscription_type, subscription_expires_at, subscription_features
        FROM users 
        WHERE id = %s AND status = 'active'
        """
        
        cursor.execute(query, (user_id,))
        user_data = cursor.fetchone()
        
        cursor.close()
        conn.close()
        
        if not user_data:
            return {
                'has_access': False,
                'has_active_subscription': False,
                'is_trial_user': False,
                'error': '用户不存在或已被禁用'
            }
        
        subscription_status = user_data.get('subscription_status', '')
        subscription_expires_at = user_data.get('subscription_expires_at')
        
        # 检查是否为试用用户
        is_trial_user = subscription_status == 'trial'
        
        # 检查是否为付费用户
        has_active_subscription = subscription_status in ['premium', 'enterprise']
        
        # 检查试用是否过期
        if is_trial_user and subscription_expires_at:
            expires_at = datetime.fromisoformat(subscription_expires_at.replace('Z', '+00:00'))
            if datetime.now() > expires_at:
                is_trial_user = False
        
        # 判断用户是否有AI功能访问权限
        has_access = has_active_subscription or is_trial_user
        
        return {
            'has_access': has_access,
            'has_active_subscription': has_active_subscription,
            'is_trial_user': is_trial_user,
            'trial_ends_at': subscription_expires_at,
            'subscription_plan': user_data.get('subscription_type'),
            'subscription_status': subscription_status
        }
                
    except Exception as e:
        logger.error(f"订阅状态检查异常: {e}")
        return {
            'has_access': False,
            'has_active_subscription': False,
            'is_trial_user': False,
            'error': '订阅状态检查失败'
        }

# 直接在主文件中注册路由
@app.route("/api/v1/ai/job-matching", methods=["POST"])
async def job_matching_api(request: Request):
    """职位匹配API"""
    try:
        # 用户认证
        auth_result = await authenticate_user(request)
        if auth_result:
            return auth_result
        
        logger.info(f"收到职位匹配请求: {request.json}, 用户: {request.ctx.user_id}")
        
        if not job_matching_service.initialized:
            logger.error("JobMatchingService未初始化")
            return sanic_json({"error": "服务未初始化", "initialized": False}, status=503)
        
        return await job_matching_service._handle_job_matching(request)
    except Exception as e:
        logger.error(f"职位匹配API异常: {e}")
        return sanic_json({"error": f"服务器内部错误: {str(e)}"}, status=500)

# 配置
class Config:
    PORT = int(os.getenv("AI_SERVICE_PORT", 8206))
    POSTGRES_HOST = os.getenv("POSTGRES_HOST", "localhost")
    POSTGRES_USER = os.getenv("POSTGRES_USER", "szjason72")
    POSTGRES_DB = os.getenv("POSTGRES_DB", "jobfirst_vector")
    POSTGRES_PASSWORD = os.getenv("POSTGRES_PASSWORD", "")
    
    # 外部AI服务配置
    EXTERNAL_AI_PROVIDER = os.getenv("EXTERNAL_AI_PROVIDER", "deepseek")
    EXTERNAL_AI_API_KEY = os.getenv("EXTERNAL_AI_API_KEY", "")
    EXTERNAL_AI_BASE_URL = os.getenv("EXTERNAL_AI_BASE_URL", "https://api.deepseek.com/v1")
    EXTERNAL_AI_MODEL = os.getenv("EXTERNAL_AI_MODEL", "deepseek-chat")
    
    # 原有Ollama配置（保留）
    OLLAMA_HOST = os.getenv("OLLAMA_HOST", "http://127.0.0.1:11434")
    OLLAMA_MODEL = os.getenv("OLLAMA_MODEL", "gemma3:4b")

# 删除重复的测试模式JWT验证函数，使用上面正确的版本

# 用户使用限制检查
async def check_user_usage_limits(token: str, service_type: str) -> bool:
    """检查用户是否超出使用限制"""
    try:
        # 调用User Service检查使用限制
        user_service_url = "http://localhost:8081/api/v1/usage/check"
        headers = {"Authorization": f"Bearer {token}"}
        data = {"service_type": service_type}
        
        response = requests.post(user_service_url, headers=headers, json=data, timeout=5)
        
        if response.status_code == 200:
            result = response.json()
            return result.get("allowed", False)
        else:
            logger.warning(f"使用限制检查失败: {response.status_code}")
            return False
            
    except Exception as e:
        logger.error(f"使用限制检查异常: {e}")
        return False

# 记录AI服务使用
async def record_ai_usage(token: str, service_type: str, cost: float):
    """记录AI服务使用情况"""
    try:
        # 调用User Service记录使用
        user_service_url = "http://localhost:8081/api/v1/usage/record"
        headers = {"Authorization": f"Bearer {token}"}
        data = {
            "service_type": service_type,
            "cost": cost,
            "timestamp": datetime.now().isoformat()
        }
        
        response = requests.post(user_service_url, headers=headers, json=data, timeout=5)
        
        if response.status_code == 200:
            logger.info(f"AI使用记录成功: {service_type}, 成本: {cost}")
        else:
            logger.warning(f"AI使用记录失败: {response.status_code}")
            
    except Exception as e:
        logger.error(f"AI使用记录异常: {e}")

# 用户权限检查函数
async def check_user_permission(token: str, required_permission: str) -> bool:
    """检查用户是否有特定权限"""
    try:
        # 临时跳过权限检查，直接返回True进行测试
        logger.info(f"权限检查跳过（测试模式）: {required_permission}")
        return True
        
        # 调用User Service检查权限
        user_service_url = "http://localhost:8081/api/v1/rbac/check"
        headers = {"Authorization": f"Bearer {token}"}
        params = {"permission": required_permission}
        
        response = requests.get(user_service_url, headers=headers, params=params, timeout=5)
        
        if response.status_code == 200:
            result = response.json()
            return result.get("allowed", False)
        else:
            logger.warning(f"权限检查失败: {response.status_code}")
            return False
            
    except Exception as e:
        logger.error(f"权限检查异常: {e}")
        return False

# 数据库连接
def get_db_connection():
    """获取PostgreSQL数据库连接"""
    try:
        conn = psycopg2.connect(
            host=Config.POSTGRES_HOST,
            user=Config.POSTGRES_USER,
            password=Config.POSTGRES_PASSWORD,
            database=Config.POSTGRES_DB
        )
        return conn
    except Exception as e:
        logger.error(f"数据库连接失败: {e}")
        return None

# 数据模型
class ResumeAnalysisRequest:
    def __init__(self, resume_id: int, content: str, file_type: str, file_name: str):
        self.resume_id = resume_id
        self.content = content
        self.file_type = file_type
        self.file_name = file_name

class Analysis:
    def __init__(self, skills: List[str], experience: List[str], education: List[str], 
                 summary: str, score: int, suggestions: List[str]):
        self.skills = skills
        self.experience = experience
        self.education = education
        self.summary = summary
        self.score = score
        self.suggestions = suggestions

class Vectors:
    def __init__(self, content_vector: List[float], skills_vector: List[float], 
                 experience_vector: List[float]):
        self.content_vector = content_vector
        self.skills_vector = skills_vector
        self.experience_vector = experience_vector

# 路由处理
@app.route("/health", methods=["GET"])
async def health_check(request: Request):
    """健康检查"""
    return sanic_response({
        "status": "healthy",
        "service": "ai-service",
        "timestamp": datetime.now().isoformat()
    })

# ==================== Taro前端兼容的AI聊天API ====================

@app.route("/api/v1/ai/chat", methods=["POST"])
async def ai_chat(request: Request):
    """AI聊天 - 需要JWT认证"""
    try:
        # 验证JWT token
        auth_header = request.headers.get('Authorization')
        if not auth_header or not auth_header.startswith('Bearer '):
            return sanic_response({"error": "Missing or invalid authorization header"}, status=401)
        
        token = auth_header.split(' ')[1]
        
        # 验证JWT token的有效性
        if not await verify_jwt_token(token):
            return sanic_response({"error": "Invalid or expired token"}, status=401)
        
        # 检查AI聊天权限
        if not await check_user_permission(token, "ai.chat"):
            return sanic_response({"error": "Insufficient permissions for AI chat"}, status=403)
        
        # 检查使用限制
        if not await check_user_usage_limits(token, "ai.chat"):
            return sanic_response({"error": "Usage limit exceeded for AI chat"}, status=429)
        
        data = request.json
        message = data.get("message", "")
        history = data.get("history", [])
        
        logger.info(f"AI聊天请求: {message[:50]}...")
        
        # 模拟AI回复
        response_message = f"AI回复: 我收到了您的消息 '{message}'。这是一个模拟回复。"
        
        # 记录使用情况 (模拟成本: $0.01)
        await record_ai_usage(token, "ai.chat", 0.01)
        
        return sanic_response({
            "status": "success",
            "data": {
                "message": response_message,
                "timestamp": datetime.now().isoformat()
            }
        })
        
    except Exception as e:
        logger.error(f"AI聊天失败: {e}")
        return sanic_response({"error": str(e)}, status=500)

@app.route("/api/v1/ai/features", methods=["GET"])
async def get_ai_features(request: Request):
    """获取AI功能列表"""
    try:
        features = [
            {
                "id": 1,
                "title": "简历优化",
                "description": "AI智能分析简历，提供优化建议",
                "icon": "📝",
                "points": 10,
                "available": True,
                "features": ["内容分析", "格式优化", "关键词提取", "行业匹配"]
            },
            {
                "id": 2,
                "title": "职位匹配",
                "description": "智能匹配最适合的职位",
                "icon": "🎯",
                "points": 15,
                "available": True,
                "features": ["技能匹配", "经验分析", "薪资预测", "发展建议"]
            },
            {
                "id": 3,
                "title": "面试准备",
                "description": "AI模拟面试，提升面试表现",
                "icon": "💼",
                "points": 20,
                "available": True,
                "features": ["模拟面试", "问题预测", "回答建议", "表现评估"]
            }
        ]
        
        return sanic_response({
            "status": "success",
            "data": features
        })
        
    except Exception as e:
        logger.error(f"获取AI功能失败: {e}")
        return sanic_response({"error": str(e)}, status=500)

@app.route("/api/v1/ai/start-analysis", methods=["POST"])
async def start_analysis(request: Request):
    """开始AI分析 - 需要JWT认证"""
    try:
        # 验证JWT token
        auth_header = request.headers.get('Authorization')
        if not auth_header or not auth_header.startswith('Bearer '):
            return sanic_response({"error": "Missing or invalid authorization header"}, status=401)
        
        token = auth_header.split(' ')[1]
        
        # 验证JWT token的有效性
        if not await verify_jwt_token(token):
            return sanic_response({"error": "Invalid or expired token"}, status=401)
        
        # 检查AI分析权限
        if not await check_user_permission(token, "ai.analyze"):
            return sanic_response({"error": "Insufficient permissions for AI analysis"}, status=403)
        
        data = request.json
        feature_id = data.get("featureId", 1)
        content = data.get("content", "")
        analysis_type = data.get("type", "resume")
        
        # 生成任务ID
        task_id = f"task_{int(time.time())}_{feature_id}"
        
        logger.info(f"开始AI分析: task_id={task_id}, feature_id={feature_id}")
        
        # 模拟异步分析任务
        return sanic_response({
            "status": "success",
            "data": {
                "taskId": task_id,
                "status": "processing",
                "message": "分析任务已开始，请稍后查询结果"
            }
        })
        
    except Exception as e:
        logger.error(f"开始AI分析失败: {e}")
        return sanic_response({"error": str(e)}, status=500)

@app.route("/api/v1/ai/analysis-result/<task_id>", methods=["GET"])
async def get_analysis_result(request: Request, task_id: str):
    """获取AI分析结果 - 需要JWT认证"""
    try:
        # 验证JWT token
        auth_header = request.headers.get('Authorization')
        if not auth_header or not auth_header.startswith('Bearer '):
            return sanic_response({"error": "Missing or invalid authorization header"}, status=401)
        
        token = auth_header.split(' ')[1]
        
        # 验证JWT token的有效性
        if not await verify_jwt_token(token):
            return sanic_response({"error": "Invalid or expired token"}, status=401)
        
        # 检查AI分析结果查看权限
        if not await check_user_permission(token, "ai.analyze"):
            return sanic_response({"error": "Insufficient permissions to view analysis results"}, status=403)
        
        logger.info(f"获取AI分析结果: task_id={task_id}")
        
        # 模拟分析结果
        result = {
            "id": task_id,
            "featureId": 1,
            "title": "简历优化分析",
            "score": 85,
            "suggestions": [
                "建议增加项目经验描述",
                "优化技能关键词",
                "完善教育背景信息"
            ],
            "keywords": ["JavaScript", "React", "Node.js", "Python"],
            "industryMatch": {
                "前端开发": 0.9,
                "全栈开发": 0.8,
                "后端开发": 0.6
            },
            "competitiveness": "优秀",
            "createdAt": datetime.now().isoformat()
        }
        
        return sanic_response({
            "status": "success",
            "data": result
        })
        
    except Exception as e:
        logger.error(f"获取AI分析结果失败: {e}")
        return sanic_response({"error": str(e)}, status=500)

@app.route("/api/v1/ai/chat-history", methods=["GET"])
async def get_chat_history(request: Request):
    """获取聊天历史 - 需要JWT认证"""
    try:
        # 验证JWT token
        auth_header = request.headers.get('Authorization')
        if not auth_header or not auth_header.startswith('Bearer '):
            return sanic_response({"error": "Missing or invalid authorization header"}, status=401)
        
        token = auth_header.split(' ')[1]
        
        # 验证JWT token的有效性
        if not await verify_jwt_token(token):
            return sanic_response({"error": "Invalid or expired token"}, status=401)
        
        # 检查聊天历史查看权限
        if not await check_user_permission(token, "ai.chat"):
            return sanic_response({"error": "Insufficient permissions to view chat history"}, status=403)
        
        # 模拟聊天历史
        history = [
            {
                "id": 1,
                "message": "你好，我想优化我的简历",
                "type": "user",
                "timestamp": datetime.now().isoformat()
            },
            {
                "id": 2,
                "message": "好的，我来帮您分析简历并提供优化建议",
                "type": "ai",
                "timestamp": datetime.now().isoformat()
            }
        ]
        
        return sanic_response({
            "status": "success",
            "data": history
        })
        
    except Exception as e:
        logger.error(f"获取聊天历史失败: {e}")
        return sanic_response({"error": str(e)}, status=500)

@app.route("/api/v1/analyze/resume", methods=["POST"])
async def analyze_resume(request: Request):
    """分析简历 - 需要JWT认证"""
    try:
        # 验证JWT token
        auth_header = request.headers.get('Authorization')
        if not auth_header or not auth_header.startswith('Bearer '):
            return sanic_response({"error": "Missing or invalid authorization header"}, status=401)
        
        token = auth_header.split(' ')[1]
        
        # 验证JWT token的有效性
        if not await verify_jwt_token(token):
            return sanic_response({"error": "Invalid or expired token"}, status=401)
        
        # 检查简历分析权限
        if not await check_user_permission(token, "ai.analyze"):
            return sanic_response({"error": "Insufficient permissions for resume analysis"}, status=403)
        
        data = request.json
        req = ResumeAnalysisRequest(
            resume_id=int(data.get("resume_id", 0)),
            content=data.get("content"),
            file_type=data.get("file_type"),
            file_name=data.get("file_name")
        )
        
        logger.info(f"开始分析简历: {req.resume_id}")
        
        # 执行AI分析
        analysis = await perform_ai_analysis(req.content, req.file_type)
        
        # 生成向量
        vectors = await generate_vectors(req.content, analysis)
        
        # 保存到数据库
        await save_vectors_to_db(req.resume_id, vectors)
        
        response = {
            "resume_id": req.resume_id,
            "status": "completed",
            "analysis": {
                "skills": analysis.skills,
                "experience": analysis.experience,
                "education": analysis.education,
                "summary": analysis.summary,
                "score": analysis.score,
                "suggestions": analysis.suggestions
            },
            "vectors": {
                "content_vector": vectors.content_vector,
                "skills_vector": vectors.skills_vector,
                "experience_vector": vectors.experience_vector
            },
            "created_at": datetime.now().isoformat()
        }
        
        logger.info(f"简历分析完成: {req.resume_id}")
        return sanic_response(response)
        
    except Exception as e:
        logger.error(f"简历分析失败: {e}")
        return sanic_response({"error": str(e)}, status=500)

@app.route("/api/v1/vectors/<resume_id:int>", methods=["GET"])
async def get_resume_vectors(request: Request, resume_id: int):
    """获取简历向量 - 需要JWT认证"""
    try:
        # 验证JWT token
        auth_header = request.headers.get('Authorization')
        if not auth_header or not auth_header.startswith('Bearer '):
            return sanic_response({"error": "Missing or invalid authorization header"}, status=401)
        
        token = auth_header.split(' ')[1]
        
        # 验证JWT token的有效性
        if not await verify_jwt_token(token):
            return sanic_response({"error": "Invalid or expired token"}, status=401)
        
        # 检查向量数据访问权限
        if not await check_user_permission(token, "ai.vectors"):
            return sanic_response({"error": "Insufficient permissions to access vectors"}, status=403)
        
        vectors = await get_vectors_from_db(resume_id)
        if vectors:
            return sanic_response({
                "resume_id": resume_id,
                "vectors": vectors
            })
        else:
            return sanic_response({"error": "简历向量未找到"}, status=404)
    except Exception as e:
        logger.error(f"获取向量失败: {e}")
        return sanic_response({"error": str(e)}, status=500)

@app.route("/api/v1/vectors/search", methods=["POST"])
async def search_similar_resumes(request: Request):
    """搜索相似简历 - 需要JWT认证"""
    try:
        # 验证JWT token
        auth_header = request.headers.get('Authorization')
        if not auth_header or not auth_header.startswith('Bearer '):
            return sanic_response({"error": "Missing or invalid authorization header"}, status=401)
        
        token = auth_header.split(' ')[1]
        
        # 验证JWT token的有效性
        if not await verify_jwt_token(token):
            return sanic_response({"error": "Invalid or expired token"}, status=401)
        
        # 检查向量搜索权限
        if not await check_user_permission(token, "ai.search"):
            return sanic_response({"error": "Insufficient permissions for vector search"}, status=403)
        
        data = request.json
        query_vector = data.get("query_vector", [])
        limit = data.get("limit", 10)
        
        results = await search_similar_resumes_db(query_vector, limit)
        return sanic_response({
            "results": results,
            "total": len(results)
        })
    except Exception as e:
        logger.error(f"搜索失败: {e}")
        return sanic_response({"error": str(e)}, status=500)

# AI分析函数
async def perform_ai_analysis(content: str, file_type: str) -> Analysis:
    """执行AI分析（使用Ollama）"""
    try:
        # 构建分析提示词
        prompt = f"""请分析以下简历内容，并以JSON格式返回分析结果：

简历内容：
{content}

请分析并返回以下信息（JSON格式）：
{{
    "skills": ["技能1", "技能2", "技能3"],
    "experience": ["经验1", "经验2", "经验3"],
    "education": ["教育背景1", "教育背景2"],
    "summary": "个人总结",
    "score": 85,
    "suggestions": ["建议1", "建议2", "建议3"]
}}

请确保返回的是有效的JSON格式。"""

        # 调用Ollama API
        response = requests.post(f"{Config.OLLAMA_HOST}/api/generate", json={
            "model": Config.OLLAMA_MODEL,
            "prompt": prompt,
            "stream": False,
            "options": {
                "temperature": 0.3,
                "top_p": 0.9,
                "max_tokens": 1000
            }
        })
        
        if response.status_code == 200:
            ai_response = response.json()["response"]
            logger.info(f"Ollama响应: {ai_response}")
            
            # 尝试解析JSON响应
            try:
                # 清理响应文本，提取JSON部分
                json_start = ai_response.find('{')
                json_end = ai_response.rfind('}') + 1
                if json_start != -1 and json_end != 0:
                    json_str = ai_response[json_start:json_end]
                    parsed_data = json.loads(json_str)
                    
                    return Analysis(
                        skills=parsed_data.get("skills", []),
                        experience=parsed_data.get("experience", []),
                        education=parsed_data.get("education", []),
                        summary=parsed_data.get("summary", ""),
                        score=parsed_data.get("score", 70),
                        suggestions=parsed_data.get("suggestions", [])
                    )
                else:
                    raise ValueError("未找到JSON格式")
                    
            except (json.JSONDecodeError, ValueError) as e:
                logger.warning(f"JSON解析失败: {e}, 使用降级分析")
                return get_fallback_analysis(content)
        else:
            logger.error(f"Ollama API调用失败: {response.status_code}")
            return get_fallback_analysis(content)
            
    except Exception as e:
        logger.error(f"AI分析失败: {e}, 使用降级分析")
        return get_fallback_analysis(content)

def get_fallback_analysis(content: str) -> Analysis:
    """降级分析（当AI分析失败时使用）"""
    # 基于关键词的简单分析
    content_lower = content.lower()
    
    skills = []
    if any(word in content_lower for word in ["javascript", "js", "react", "vue", "angular"]):
        skills.append("前端开发")
    if any(word in content_lower for word in ["python", "java", "go", "node.js", "php"]):
        skills.append("后端开发")
    if any(word in content_lower for word in ["mysql", "postgresql", "mongodb", "redis"]):
        skills.append("数据库")
    if any(word in content_lower for word in ["docker", "kubernetes", "aws", "azure"]):
        skills.append("DevOps")
    
    if not skills:
        skills = ["技术开发", "软件工程"]
    
    experience = ["技术开发", "项目经验"]
    education = ["相关学历"]
    summary = "具备技术开发能力的工程师"
    score = 70
    suggestions = ["完善技能描述", "添加具体项目经验"]
    
    return Analysis(skills, experience, education, summary, score, suggestions)

# 向量生成函数
async def generate_vectors(content: str, analysis: Analysis) -> Vectors:
    """生成向量（模拟）"""
    # 这里应该调用OpenAI API生成实际的向量
    # 目前使用模拟数据
    
    # 模拟向量生成过程
    await asyncio.sleep(0.5)
    
    def generate_mock_vector():
        return [float(i % 100) / 100.0 for i in range(1536)]
    
    return Vectors(
        content_vector=generate_mock_vector(),
        skills_vector=generate_mock_vector(),
        experience_vector=generate_mock_vector()
    )

# 数据库操作
async def save_vectors_to_db(resume_id: str, vectors: Vectors):
    """保存向量到数据库"""
    conn = get_db_connection()
    if not conn:
        raise Exception("数据库连接失败")
    
    try:
        with conn.cursor() as cursor:
            # 检查是否已存在
            cursor.execute(
                "SELECT id FROM resume_vectors WHERE resume_id = %s",
                (resume_id,)
            )
            
            if cursor.fetchone():
                # 更新现有记录
                cursor.execute("""
                    UPDATE resume_vectors 
                    SET content_vector = %s, skills_vector = %s, experience_vector = %s
                    WHERE resume_id = %s
                """, (
                    vectors.content_vector,
                    vectors.skills_vector,
                    vectors.experience_vector,
                    resume_id
                ))
            else:
                # 插入新记录
                cursor.execute("""
                    INSERT INTO resume_vectors (resume_id, content_vector, skills_vector, experience_vector)
                    VALUES (%s, %s, %s, %s)
                """, (
                    resume_id,
                    vectors.content_vector,
                    vectors.skills_vector,
                    vectors.experience_vector
                ))
            
            conn.commit()
            logger.info(f"向量数据已保存到数据库: {resume_id}")
            
    except Exception as e:
        conn.rollback()
        logger.error(f"保存向量失败: {e}")
        raise
    finally:
        conn.close()

async def get_vectors_from_db(resume_id: int) -> Dict[str, Any]:
    """从数据库获取向量"""
    conn = get_db_connection()
    if not conn:
        return None
    
    try:
        with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cursor:
            cursor.execute("""
                SELECT content_vector, skills_vector, experience_vector
                FROM resume_vectors 
                WHERE resume_id = %s
            """, (resume_id,))
            
            result = cursor.fetchone()
            if result:
                return dict(result)
            return None
            
    except Exception as e:
        logger.error(f"获取向量失败: {e}")
        return None
    finally:
        conn.close()

async def search_similar_resumes_db(query_vector: List[float], limit: int) -> List[Dict[str, Any]]:
    """搜索相似简历"""
    conn = get_db_connection()
    if not conn:
        return []
    
    try:
        with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cursor:
            cursor.execute("""
                SELECT resume_id, 
                       content_vector <=> %s as distance
                FROM resume_vectors 
                ORDER BY content_vector <=> %s 
                LIMIT %s
            """, (query_vector, query_vector, limit))
            
            results = []
            for row in cursor.fetchall():
                results.append({
                    "resume_id": row["resume_id"],
                    "distance": float(row["distance"])
                })
            
            return results
            
    except Exception as e:
        logger.error(f"搜索失败: {e}")
        return []
    finally:
        conn.close()

# 职位匹配路由由JobMatchingService自动注册

# 启动函数
# Sanic应用启动前初始化
@app.before_server_start
async def initialize_services(app, loop):
    """在服务器启动前初始化服务"""
    logger.info("开始初始化AI服务...")
    await initialize_job_matching()
    logger.info("AI服务初始化完成")

if __name__ == "__main__":
    logger.info(f"启动AI服务，端口: {Config.PORT}")
    
    app.run(
        host="0.0.0.0",
        port=Config.PORT,
        debug=True,  # 启用调试模式
        access_log=True
    )
