#!/usr/bin/env python3
"""
JobFirst AI服务 - DeepSeek API版本
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
import aiohttp
from sanic import Sanic, Request, json as sanic_json
from sanic.response import json as sanic_response

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# 创建Sanic应用
app = Sanic("ai-service")

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
class Analysis:
    def __init__(self, skills: List[str], experience: List[str], education: List[str], 
                 summary: str, score: int, suggestions: List[str]):
        self.skills = skills
        self.experience = experience
        self.education = education
        self.summary = summary
        self.score = score
        self.suggestions = suggestions

# DeepSeek API调用函数
async def call_deepseek_api(prompt: str) -> str:
    """调用DeepSeek API"""
    try:
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"{Config.EXTERNAL_AI_BASE_URL}/chat/completions",
                headers={
                    "Authorization": f"Bearer {Config.EXTERNAL_AI_API_KEY}",
                    "Content-Type": "application/json"
                },
                json={
                    "model": Config.EXTERNAL_AI_MODEL,
                    "messages": [{"role": "user", "content": prompt}],
                    "max_tokens": 1000,
                    "temperature": 0.7
                }
            ) as response:
                if response.status == 200:
                    result = await response.json()
                    return result['choices'][0]['message']['content']
                else:
                    error_text = await response.text()
                    logger.error(f"DeepSeek API调用失败: {response.status}, {error_text}")
                    raise Exception(f"AI服务调用失败: {response.status}")
    except Exception as e:
        logger.error(f"DeepSeek API调用异常: {e}")
        raise

# AI分析函数
async def perform_ai_analysis(content: str, file_type: str) -> Analysis:
    """执行AI分析（使用DeepSeek API）"""
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

        # 调用DeepSeek API
        ai_response = await call_deepseek_api(prompt)
        logger.info(f"DeepSeek响应: {ai_response}")
        
        # 解析JSON响应
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
                    score=parsed_data.get("score", 75),
                    suggestions=parsed_data.get("suggestions", [])
                )
            else:
                raise ValueError("未找到有效的JSON格式")
        except (json.JSONDecodeError, ValueError) as e:
            logger.warning(f"AI响应JSON解析失败: {e}，使用默认结果")
            return Analysis(
                skills=["技能分析中..."],
                experience=["经验分析中..."],
                education=["教育背景分析中..."],
                summary="AI分析完成，请查看详细建议",
                score=75,
                suggestions=["建议优化简历格式", "建议突出核心技能", "建议量化工作成果"]
            )
            
    except Exception as e:
        logger.error(f"AI分析失败: {e}")
        raise

# 向量生成函数
async def generate_vectors(content: str) -> List[float]:
    """生成内容向量（模拟实现）"""
    # 这里可以集成真实的向量生成服务
    # 目前返回模拟向量
    import random
    return [random.random() for _ in range(1536)]

# 向量存储函数
async def store_vectors(resume_id: str, content_vector: List[float], 
                       skills_vector: List[float], experience_vector: List[float]):
    """存储向量到数据库"""
    conn = get_db_connection()
    if not conn:
        raise Exception("数据库连接失败")
    
    try:
        cursor = conn.cursor()
        cursor.execute("""
            INSERT INTO resume_vectors (resume_id, content_vector, skills_vector, experience_vector, created_at, updated_at)
            VALUES (%s, %s, %s, %s, NOW(), NOW())
            ON CONFLICT (resume_id) 
            DO UPDATE SET 
                content_vector = EXCLUDED.content_vector,
                skills_vector = EXCLUDED.skills_vector,
                experience_vector = EXCLUDED.experience_vector,
                updated_at = NOW()
        """, (resume_id, content_vector, skills_vector, experience_vector))
        
        conn.commit()
        logger.info(f"向量存储成功: {resume_id}")
    except Exception as e:
        conn.rollback()
        logger.error(f"向量存储失败: {e}")
        raise
    finally:
        conn.close()

# 向量检索函数
async def get_vectors_from_db(resume_id: str) -> Dict[str, List[float]]:
    """从数据库获取向量"""
    conn = get_db_connection()
    if not conn:
        return None
    
    try:
        cursor = conn.cursor()
        cursor.execute("""
            SELECT content_vector, skills_vector, experience_vector
            FROM resume_vectors
            WHERE resume_id = %s
        """, (resume_id,))
        
        result = cursor.fetchone()
        if result:
            return {
                "content_vector": result[0],
                "skills_vector": result[1],
                "experience_vector": result[2]
            }
        return None
    except Exception as e:
        logger.error(f"向量检索失败: {e}")
        return None
    finally:
        conn.close()

# 相似简历搜索函数
async def search_similar_resumes_db(query_vector: List[float], limit: int = 10) -> List[Dict]:
    """搜索相似简历"""
    conn = get_db_connection()
    if not conn:
        return []
    
    try:
        cursor = conn.cursor()
        cursor.execute("""
            SELECT resume_id, content_vector, skills_vector, experience_vector
            FROM resume_vectors
            ORDER BY content_vector <-> %s
            LIMIT %s
        """, (query_vector, limit))
        
        results = []
        for row in cursor.fetchall():
            results.append({
                "resume_id": row[0],
                "content_vector": row[1],
                "skills_vector": row[2],
                "experience_vector": row[3]
            })
        
        return results
    except Exception as e:
        logger.error(f"相似简历搜索失败: {e}")
        return []
    finally:
        conn.close()

# API路由
@app.route("/health", methods=["GET"])
async def health_check(request: Request):
    """健康检查"""
    return sanic_response({
        "status": "healthy",
        "service": "ai-service",
        "timestamp": datetime.now().isoformat()
    })

@app.route("/api/v1/ai/features", methods=["GET"])
async def get_ai_features(request: Request):
    """获取AI功能列表"""
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

@app.route("/api/v1/ai/start-analysis", methods=["POST"])
async def start_analysis(request: Request):
    """开始AI分析 - 需要JWT认证"""
    try:
        # 验证JWT token
        auth_header = request.headers.get('Authorization')
        if not auth_header or not auth_header.startswith('Bearer '):
            return sanic_response({"error": "Missing or invalid authorization header"}, status=401)
        
        token = auth_header.split(' ')[1]
        # TODO: 验证JWT token的有效性
        
        data = request.json
        feature_id = data.get("featureId", 1)
        content = data.get("content", "")
        analysis_type = data.get("type", "resume")
        
        # 生成任务ID
        task_id = f"task_{int(time.time())}_{feature_id}"
        
        logger.info(f"开始AI分析: task_id={task_id}, feature_id={feature_id}")
        
        # 执行AI分析
        analysis_result = await perform_ai_analysis(content, analysis_type)
        
        # 生成向量
        content_vector = await generate_vectors(content)
        skills_vector = await generate_vectors(" ".join(analysis_result.skills))
        experience_vector = await generate_vectors(" ".join(analysis_result.experience))
        
        # 存储向量
        await store_vectors(task_id, content_vector, skills_vector, experience_vector)
        
        return sanic_response({
            "status": "success",
            "data": {
                "taskId": task_id,
                "status": "completed",
                "analysis": {
                    "skills": analysis_result.skills,
                    "experience": analysis_result.experience,
                    "education": analysis_result.education,
                    "summary": analysis_result.summary,
                    "score": analysis_result.score,
                    "suggestions": analysis_result.suggestions
                }
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
        # TODO: 验证JWT token的有效性
        
        # 从数据库获取分析结果
        vectors = await get_vectors_from_db(task_id)
        if vectors:
            return sanic_response({
                "status": "success",
                "data": {
                    "taskId": task_id,
                    "status": "completed",
                    "vectors": vectors
                }
            })
        else:
            return sanic_response({"error": "分析结果未找到"}, status=404)
            
    except Exception as e:
        logger.error(f"获取分析结果失败: {e}")
        return sanic_response({"error": str(e)}, status=500)

@app.route("/api/v1/ai/chat", methods=["POST"])
async def ai_chat(request: Request):
    """AI聊天功能 - 需要JWT认证"""
    try:
        # 验证JWT token
        auth_header = request.headers.get('Authorization')
        if not auth_header or not auth_header.startswith('Bearer '):
            return sanic_response({"error": "Missing or invalid authorization header"}, status=401)
        
        token = auth_header.split(' ')[1]
        # TODO: 验证JWT token的有效性
        
        data = request.json
        message = data.get("message", "")
        
        if not message:
            return sanic_response({"error": "消息内容不能为空"}, status=400)
        
        # 调用DeepSeek API进行聊天
        response = await call_deepseek_api(message)
        
        return sanic_response({
            "status": "success",
            "data": {
                "message": response,
                "timestamp": datetime.now().isoformat()
            }
        })
        
    except Exception as e:
        logger.error(f"AI聊天失败: {e}")
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
        # TODO: 验证JWT token的有效性
        
        # 返回模拟聊天历史
        chat_history = [
            {
                "id": 1,
                "message": "你好，我是JobFirst AI助手，有什么可以帮助您的吗？",
                "is_ai": True,
                "timestamp": datetime.now().isoformat()
            }
        ]
        
        return sanic_response({
            "status": "success",
            "data": chat_history
        })
        
    except Exception as e:
        logger.error(f"获取聊天历史失败: {e}")
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
        # TODO: 验证JWT token的有效性
        
        vectors = await get_vectors_from_db(str(resume_id))
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
        # TODO: 验证JWT token的有效性
        
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

# 启动函数
if __name__ == "__main__":
    logger.info(f"启动AI服务，端口: {Config.PORT}")
    app.run(
        host="0.0.0.0",
        port=Config.PORT,
        debug=True,  # 启用调试模式
        access_log=True
    )
