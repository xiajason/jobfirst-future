#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
增强版职位匹配引擎 - 基于Resume-Matcher最佳实践
结合现有架构和Resume-Matcher的成功经验

创建时间: 2025-09-18
作者: AI Assistant
版本: 2.0.0
"""

import asyncio
import json
import logging
import numpy as np
from typing import Dict, List, Optional, Any, Tuple
from datetime import datetime
import asyncpg
from sentence_transformers import SentenceTransformer
import structlog

logger = structlog.get_logger(__name__)

class EnhancedJobMatchingEngine:
    """增强版职位匹配引擎 - 借鉴Resume-Matcher"""
    
    def __init__(self, data_access, postgres_pool):
        """
        初始化增强版匹配引擎
        
        Args:
            data_access: 数据访问层实例
            postgres_pool: PostgreSQL连接池
        """
        self.data_access = data_access
        self.postgres_pool = postgres_pool
        
        # 初始化嵌入模型 (借鉴Resume-Matcher的模型选择)
        self.embedding_model = None
        self.model_loaded = False
        
        # 匹配维度权重配置 (基于Resume-Matcher最佳实践)
        self.matching_dimensions = {
            'semantic_similarity': 0.35,    # 语义相似度 (FastEmbed)
            'skills_match': 0.30,           # 技能匹配度
            'experience_match': 0.20,       # 经验匹配度
            'education_match': 0.10,        # 教育背景匹配
            'cultural_fit': 0.05            # 文化匹配度
        }
        
        # 行业特定权重调整 (基于Resume-Matcher的行业分析)
        self.industry_adjustments = {
            'technology': {
                'skills_match': 0.40,      # 技术行业更重视技能
                'semantic_similarity': 0.30,
                'experience_match': 0.20,
                'education_match': 0.10
            },
            'finance': {
                'semantic_similarity': 0.40,  # 金融行业更重视经验描述
                'experience_match': 0.30,
                'skills_match': 0.20,
                'education_match': 0.10
            },
            'marketing': {
                'semantic_similarity': 0.35,
                'cultural_fit': 0.25,      # 营销行业更重视文化匹配
                'skills_match': 0.25,
                'experience_match': 0.15
            },
            'healthcare': {
                'education_match': 0.35,   # 医疗行业更重视教育背景
                'experience_match': 0.30,
                'semantic_similarity': 0.25,
                'skills_match': 0.10
            }
        }
        
        # 匹配阈值配置
        self.match_thresholds = {
            'excellent': 0.85,    # 优秀匹配
            'good': 0.70,         # 良好匹配
            'fair': 0.55,         # 一般匹配
            'poor': 0.40          # 较差匹配
        }
    
    async def initialize(self):
        """初始化匹配引擎"""
        try:
            # 加载嵌入模型 (借鉴Resume-Matcher的模型选择策略)
            await self._load_embedding_model()
            
            # 初始化向量数据库索引
            await self._initialize_vector_indexes()
            
            logger.info("增强版职位匹配引擎初始化成功")
            
        except Exception as e:
            logger.error("增强版职位匹配引擎初始化失败", error=str(e))
            raise
    
    async def _load_embedding_model(self):
        """加载嵌入模型 - 借鉴Resume-Matcher的模型选择"""
        try:
            # 使用Resume-Matcher推荐的模型
            model_name = "sentence-transformers/all-MiniLM-L6-v2"
            logger.info("正在加载嵌入模型", model_name=model_name)
            
            self.embedding_model = SentenceTransformer(model_name)
            self.model_loaded = True
            
            logger.info("嵌入模型加载成功", model_name=model_name)
            
        except Exception as e:
            logger.error("嵌入模型加载失败", error=str(e))
            # 降级到备用模型
            try:
                backup_model = "sentence-transformers/all-MiniLM-L12-v2"
                logger.info("尝试加载备用模型", backup_model=backup_model)
                self.embedding_model = SentenceTransformer(backup_model)
                self.model_loaded = True
                logger.info("备用模型加载成功", backup_model=backup_model)
            except Exception as backup_error:
                logger.error("备用模型加载也失败", error=str(backup_error))
                raise
    
    async def _initialize_vector_indexes(self):
        """初始化向量数据库索引"""
        try:
            async with self.postgres_pool.acquire() as conn:
                # 创建向量扩展
                await conn.execute("CREATE EXTENSION IF NOT EXISTS vector")
                
                # 创建简历向量表
                await conn.execute("""
                    CREATE TABLE IF NOT EXISTS resume_vectors (
                        id SERIAL PRIMARY KEY,
                        resume_id INTEGER NOT NULL,
                        user_id INTEGER NOT NULL,
                        content_vector vector(384),
                        skills_vector vector(384),
                        experience_vector vector(384),
                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                    )
                """)
                
                # 创建职位向量表
                await conn.execute("""
                    CREATE TABLE IF NOT EXISTS job_vectors (
                        id SERIAL PRIMARY KEY,
                        job_id INTEGER NOT NULL,
                        company_id INTEGER NOT NULL,
                        description_vector vector(384),
                        requirements_vector vector(384),
                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                    )
                """)
                
                # 创建向量索引
                await conn.execute("""
                    CREATE INDEX IF NOT EXISTS idx_resume_content_vector 
                    ON resume_vectors USING ivfflat (content_vector vector_cosine_ops)
                """)
                
                await conn.execute("""
                    CREATE INDEX IF NOT EXISTS idx_job_description_vector 
                    ON job_vectors USING ivfflat (description_vector vector_cosine_ops)
                """)
                
                logger.info("向量数据库索引初始化成功")
                
        except Exception as e:
            logger.error("向量数据库索引初始化失败", error=str(e))
            raise
    
    async def find_matching_jobs(self, resume_data: Dict[str, Any], user_id: int, 
                                limit: int = 10, filters: Optional[Dict] = None) -> List[Dict[str, Any]]:
        """
        基于多维度向量匹配职位 - 增强版算法
        
        Args:
            resume_data: 简历数据（包含元数据、解析内容和向量数据）
            user_id: 用户ID
            limit: 返回结果数量限制
            filters: 筛选条件
            
        Returns:
            匹配结果列表
        """
        try:
            # 1. 验证数据完整性
            if not self._validate_resume_data(resume_data):
                raise ValueError("简历数据不完整或无效")
            
            # 2. 基础筛选 (硬性条件)
            basic_filtered_jobs = await self._basic_filter(resume_data, filters)
            if not basic_filtered_jobs:
                logger.info("基础筛选后无符合条件的职位")
                return []
            
            # 3. 生成简历向量 (如果不存在)
            resume_vectors = await self._ensure_resume_vectors(resume_data, user_id)
            
            # 4. 向量相似度搜索
            vector_matches = await self._vector_similarity_search(
                resume_vectors, basic_filtered_jobs, limit * 2  # 获取更多候选
            )
            if not vector_matches:
                logger.info("向量相似度搜索无结果")
                return []
            
            # 5. 多维度评分 (借鉴Resume-Matcher的评分体系)
            scored_matches = await self._multi_dimension_scoring(
                resume_data, vector_matches
            )
            
            # 6. 结果排序和过滤
            final_results = await self._rank_and_filter(scored_matches, limit)
            
            # 7. 生成推荐建议
            recommendations = await self._generate_recommendations(
                resume_data, final_results
            )
            
            # 8. 记录匹配日志
            await self._log_matching_result(user_id, resume_data.get('id'), final_results)
            
            return {
                'matches': final_results,
                'recommendations': recommendations,
                'metadata': {
                    'total_candidates': len(basic_filtered_jobs),
                    'vector_matches': len(vector_matches),
                    'final_results': len(final_results),
                    'processing_time': datetime.now().isoformat()
                }
            }
            
        except Exception as e:
            logger.error("职位匹配处理失败", error=str(e), user_id=user_id)
            raise
    
    async def _ensure_resume_vectors(self, resume_data: Dict[str, Any], user_id: int) -> Dict[str, List[float]]:
        """确保简历向量存在，如果不存在则生成"""
        try:
            resume_id = resume_data.get('id')
            if not resume_id:
                raise ValueError("简历ID不存在")
            
            async with self.postgres_pool.acquire() as conn:
                # 检查是否已有向量
                existing = await conn.fetchrow(
                    "SELECT * FROM resume_vectors WHERE resume_id = $1", resume_id
                )
                
                if existing:
                    return {
                        'content_vector': existing['content_vector'],
                        'skills_vector': existing['skills_vector'],
                        'experience_vector': existing['experience_vector']
                    }
                
                # 生成新向量
                vectors = await self._generate_resume_vectors(resume_data)
                
                # 存储向量
                await conn.execute("""
                    INSERT INTO resume_vectors (resume_id, user_id, content_vector, skills_vector, experience_vector)
                    VALUES ($1, $2, $3, $4, $5)
                """, resume_id, user_id, vectors['content_vector'], 
                    vectors['skills_vector'], vectors['experience_vector'])
                
                logger.info("简历向量生成并存储成功", resume_id=resume_id)
                return vectors
                
        except Exception as e:
            logger.error("简历向量处理失败", error=str(e), resume_id=resume_id)
            raise
    
    async def _generate_resume_vectors(self, resume_data: Dict[str, Any]) -> Dict[str, List[float]]:
        """生成简历向量 - 借鉴Resume-Matcher的向量化策略"""
        try:
            if not self.model_loaded:
                raise ValueError("嵌入模型未加载")
            
            # 1. 内容向量 (整体简历内容)
            content_text = self._extract_content_text(resume_data)
            content_vector = self.embedding_model.encode([content_text])[0].tolist()
            
            # 2. 技能向量 (技能相关文本)
            skills_text = self._extract_skills_text(resume_data)
            skills_vector = self.embedding_model.encode([skills_text])[0].tolist()
            
            # 3. 经验向量 (工作经验相关文本)
            experience_text = self._extract_experience_text(resume_data)
            experience_vector = self.embedding_model.encode([experience_text])[0].tolist()
            
            return {
                'content_vector': content_vector,
                'skills_vector': skills_vector,
                'experience_vector': experience_vector
            }
            
        except Exception as e:
            logger.error("简历向量生成失败", error=str(e))
            raise
    
    def _extract_content_text(self, resume_data: Dict[str, Any]) -> str:
        """提取简历内容文本"""
        parts = []
        
        # 基本信息
        if resume_data.get('name'):
            parts.append(f"姓名: {resume_data['name']}")
        if resume_data.get('summary'):
            parts.append(f"个人简介: {resume_data['summary']}")
        
        # 工作经验
        if resume_data.get('experience'):
            for exp in resume_data['experience']:
                exp_text = f"职位: {exp.get('title', '')} 公司: {exp.get('company', '')} 描述: {exp.get('description', '')}"
                parts.append(exp_text)
        
        # 教育背景
        if resume_data.get('education'):
            for edu in resume_data['education']:
                edu_text = f"学历: {edu.get('degree', '')} 学校: {edu.get('school', '')}"
                parts.append(edu_text)
        
        # 技能
        if resume_data.get('skills'):
            skills_text = "技能: " + ", ".join(resume_data['skills'])
            parts.append(skills_text)
        
        return " ".join(parts)
    
    def _extract_skills_text(self, resume_data: Dict[str, Any]) -> str:
        """提取技能相关文本"""
        parts = []
        
        # 技能列表
        if resume_data.get('skills'):
            parts.extend(resume_data['skills'])
        
        # 从工作经验中提取技能
        if resume_data.get('experience'):
            for exp in resume_data['experience']:
                if exp.get('description'):
                    parts.append(exp['description'])
        
        return " ".join(parts)
    
    def _extract_experience_text(self, resume_data: Dict[str, Any]) -> str:
        """提取工作经验相关文本"""
        parts = []
        
        if resume_data.get('experience'):
            for exp in resume_data['experience']:
                exp_parts = []
                if exp.get('title'):
                    exp_parts.append(exp['title'])
                if exp.get('company'):
                    exp_parts.append(exp['company'])
                if exp.get('description'):
                    exp_parts.append(exp['description'])
                if exp_parts:
                    parts.append(" ".join(exp_parts))
        
        return " ".join(parts)
    
    async def _vector_similarity_search(self, resume_vectors: Dict[str, List[float]], 
                                      candidate_jobs: List[Dict], limit: int) -> List[Dict]:
        """向量相似度搜索 - 借鉴Resume-Matcher的搜索策略"""
        try:
            if not candidate_jobs:
                return []
            
            job_ids = [job['id'] for job in candidate_jobs]
            
            async with self.postgres_pool.acquire() as conn:
                # 使用PostgreSQL的向量相似度搜索
                query = """
                    SELECT jv.job_id, jv.description_vector, jv.requirements_vector,
                           (jv.description_vector <=> $1::vector) as content_similarity,
                           (jv.requirements_vector <=> $2::vector) as skills_similarity
                    FROM job_vectors jv
                    WHERE jv.job_id = ANY($3::int[])
                    ORDER BY (jv.description_vector <=> $1::vector) + (jv.requirements_vector <=> $2::vector)
                    LIMIT $4
                """
                
                results = await conn.fetch(
                    query, 
                    resume_vectors['content_vector'],
                    resume_vectors['skills_vector'],
                    job_ids,
                    limit
                )
                
                # 构建结果
                vector_matches = []
                for row in results:
                    job_info = next((job for job in candidate_jobs if job['id'] == row['job_id']), None)
                    if job_info:
                        vector_matches.append({
                            'job_id': row['job_id'],
                            'job_info': job_info,
                            'content_similarity': 1 - row['content_similarity'],  # 转换为相似度
                            'skills_similarity': 1 - row['skills_similarity'],
                            'description_vector': row['description_vector'],
                            'requirements_vector': row['requirements_vector']
                        })
                
                logger.info("向量相似度搜索完成", 
                           candidates=len(candidate_jobs), 
                           matches=len(vector_matches))
                
                return vector_matches
                
        except Exception as e:
            logger.error("向量相似度搜索失败", error=str(e))
            return []
    
    async def _multi_dimension_scoring(self, resume_data: Dict[str, Any], 
                                     vector_matches: List[Dict]) -> List[Dict]:
        """多维度评分 - 借鉴Resume-Matcher的评分体系"""
        try:
            scored_matches = []
            
            for match in vector_matches:
                job_info = match['job_info']
                
                # 1. 语义相似度评分
                semantic_score = (match['content_similarity'] + match['skills_similarity']) / 2
                
                # 2. 技能匹配评分
                skills_score = self._calculate_skills_match(
                    resume_data.get('skills', []), 
                    job_info.get('required_skills', [])
                )
                
                # 3. 经验匹配评分
                experience_score = self._calculate_experience_match(
                    resume_data.get('experience', []),
                    job_info.get('experience_requirements', {})
                )
                
                # 4. 教育背景匹配评分
                education_score = self._calculate_education_match(
                    resume_data.get('education', []),
                    job_info.get('education_requirements', {})
                )
                
                # 5. 文化匹配评分
                cultural_score = self._calculate_cultural_fit(
                    resume_data.get('personality_traits', []),
                    job_info.get('company_culture', {})
                )
                
                # 6. 综合评分 (根据行业调整权重)
                industry = job_info.get('industry', 'general')
                weights = self.industry_adjustments.get(industry, self.matching_dimensions)
                
                final_score = (
                    semantic_score * weights['semantic_similarity'] +
                    skills_score * weights['skills_match'] +
                    experience_score * weights['experience_match'] +
                    education_score * weights['education_match'] +
                    cultural_score * weights['cultural_fit']
                )
                
                # 7. 置信度计算
                confidence = self._calculate_confidence({
                    'semantic_similarity': semantic_score,
                    'skills_match': skills_score,
                    'experience_match': experience_score,
                    'education_match': education_score,
                    'cultural_fit': cultural_score
                })
                
                scored_matches.append({
                    'job_id': match['job_id'],
                    'job_info': job_info,
                    'overall_score': final_score,
                    'confidence': confidence,
                    'breakdown': {
                        'semantic_similarity': semantic_score,
                        'skills_match': skills_score,
                        'experience_match': experience_score,
                        'education_match': education_score,
                        'cultural_fit': cultural_score
                    },
                    'match_level': self._get_match_level(final_score)
                })
            
            # 按综合评分排序
            scored_matches.sort(key=lambda x: x['overall_score'], reverse=True)
            
            logger.info("多维度评分完成", matches=len(scored_matches))
            return scored_matches
            
        except Exception as e:
            logger.error("多维度评分失败", error=str(e))
            return []
    
    def _calculate_skills_match(self, resume_skills: List[str], job_skills: List[str]) -> float:
        """计算技能匹配度"""
        if not job_skills:
            return 1.0
        
        if not resume_skills:
            return 0.0
        
        # 转换为小写进行比较
        resume_skills_lower = [skill.lower() for skill in resume_skills]
        job_skills_lower = [skill.lower() for skill in job_skills]
        
        # 计算匹配的技能数量
        matched_skills = set(resume_skills_lower) & set(job_skills_lower)
        match_ratio = len(matched_skills) / len(job_skills_lower)
        
        return min(match_ratio, 1.0)
    
    def _calculate_experience_match(self, resume_experience: List[Dict], 
                                  job_requirements: Dict) -> float:
        """计算经验匹配度"""
        if not job_requirements:
            return 1.0
        
        if not resume_experience:
            return 0.0
        
        # 计算总工作经验年限
        total_years = 0
        for exp in resume_experience:
            if exp.get('duration'):
                total_years += exp['duration']
        
        # 检查是否满足最低经验要求
        required_years = job_requirements.get('min_years', 0)
        if total_years >= required_years:
            return 1.0
        
        # 按比例计算匹配度
        return total_years / required_years if required_years > 0 else 1.0
    
    def _calculate_education_match(self, resume_education: List[Dict], 
                                 job_requirements: Dict) -> float:
        """计算教育背景匹配度"""
        if not job_requirements:
            return 1.0
        
        if not resume_education:
            return 0.0
        
        # 检查学历要求
        required_degree = job_requirements.get('degree_level', '')
        if not required_degree:
            return 1.0
        
        # 学历等级映射
        degree_levels = {
            '高中': 1, '中专': 1, '大专': 2, '本科': 3, 
            '硕士': 4, '博士': 5, '博士后': 6
        }
        
        required_level = degree_levels.get(required_degree, 0)
        if required_level == 0:
            return 1.0
        
        # 检查简历中的最高学历
        max_level = 0
        for edu in resume_education:
            degree = edu.get('degree', '')
            level = degree_levels.get(degree, 0)
            max_level = max(max_level, level)
        
        if max_level >= required_level:
            return 1.0
        
        return max_level / required_level
    
    def _calculate_cultural_fit(self, personality_traits: List[str], 
                              company_culture: Dict) -> float:
        """计算文化匹配度"""
        if not company_culture:
            return 0.5  # 默认中等匹配
        
        if not personality_traits:
            return 0.5
        
        # 简化的文化匹配计算
        culture_keywords = company_culture.get('keywords', [])
        if not culture_keywords:
            return 0.5
        
        # 计算匹配的文化关键词
        matched_keywords = 0
        for trait in personality_traits:
            for keyword in culture_keywords:
                if keyword.lower() in trait.lower():
                    matched_keywords += 1
                    break
        
        return min(matched_keywords / len(culture_keywords), 1.0)
    
    def _calculate_confidence(self, scores: Dict[str, float]) -> float:
        """计算匹配置信度"""
        # 基于各维度评分的方差计算置信度
        score_values = list(scores.values())
        if not score_values:
            return 0.0
        
        mean_score = np.mean(score_values)
        variance = np.var(score_values)
        
        # 置信度 = 平均分 - 方差 (分数越高且越稳定，置信度越高)
        confidence = max(0.0, min(1.0, mean_score - variance))
        
        return confidence
    
    def _get_match_level(self, score: float) -> str:
        """获取匹配等级"""
        if score >= self.match_thresholds['excellent']:
            return 'excellent'
        elif score >= self.match_thresholds['good']:
            return 'good'
        elif score >= self.match_thresholds['fair']:
            return 'fair'
        else:
            return 'poor'
    
    async def _generate_recommendations(self, resume_data: Dict[str, Any], 
                                      matches: List[Dict]) -> List[Dict]:
        """生成个性化推荐建议 - 借鉴Resume-Matcher的推荐策略"""
        try:
            recommendations = []
            
            if not matches:
                return recommendations
            
            # 分析匹配结果
            excellent_matches = [m for m in matches if m['match_level'] == 'excellent']
            good_matches = [m for m in matches if m['match_level'] == 'good']
            fair_matches = [m for m in matches if m['match_level'] == 'fair']
            
            # 1. 申请建议
            if excellent_matches:
                recommendations.append({
                    'type': 'application_advice',
                    'priority': 'high',
                    'title': '强烈推荐申请',
                    'content': f'发现{len(excellent_matches)}个高度匹配的职位，建议优先申请',
                    'matches': [m['job_id'] for m in excellent_matches[:3]]
                })
            
            # 2. 技能提升建议
            if fair_matches:
                recommendations.append({
                    'type': 'skill_improvement',
                    'priority': 'medium',
                    'title': '技能提升建议',
                    'content': '建议学习以下技能以提高匹配度',
                    'suggestions': self._get_skill_suggestions(resume_data, fair_matches)
                })
            
            # 3. 简历优化建议
            recommendations.append({
                'type': 'resume_optimization',
                'priority': 'low',
                'title': '简历优化建议',
                'content': '基于匹配分析，建议优化简历内容',
                'suggestions': self._get_resume_suggestions(resume_data, matches)
            })
            
            return recommendations
            
        except Exception as e:
            logger.error("推荐建议生成失败", error=str(e))
            return []
    
    def _get_skill_suggestions(self, resume_data: Dict[str, Any], 
                             fair_matches: List[Dict]) -> List[str]:
        """获取技能提升建议"""
        suggestions = []
        
        # 分析职位要求的技能
        all_required_skills = set()
        for match in fair_matches:
            job_skills = match['job_info'].get('required_skills', [])
            all_required_skills.update(job_skills)
        
        # 找出缺失的技能
        resume_skills = set(resume_data.get('skills', []))
        missing_skills = all_required_skills - resume_skills
        
        # 返回前5个建议
        suggestions = list(missing_skills)[:5]
        
        return suggestions
    
    def _get_resume_suggestions(self, resume_data: Dict[str, Any], 
                              matches: List[Dict]) -> List[str]:
        """获取简历优化建议"""
        suggestions = []
        
        # 基于匹配分析生成建议
        if not resume_data.get('summary'):
            suggestions.append("添加个人简介，突出核心优势")
        
        if not resume_data.get('experience'):
            suggestions.append("补充工作经历，详细描述项目经验")
        
        if not resume_data.get('skills'):
            suggestions.append("完善技能列表，包括技术技能和软技能")
        
        # 基于匹配分数生成建议
        avg_score = np.mean([m['overall_score'] for m in matches])
        if avg_score < 0.6:
            suggestions.append("整体匹配度较低，建议重新审视职业方向")
        
        return suggestions
    
    async def _log_matching_result(self, user_id: int, resume_id: int, 
                                 matches: List[Dict]):
        """记录匹配结果日志"""
        try:
            # 这里可以记录到数据库或日志系统
            logger.info("匹配结果记录", 
                       user_id=user_id, 
                       resume_id=resume_id, 
                       matches_count=len(matches))
            
        except Exception as e:
            logger.error("匹配结果记录失败", error=str(e))
    
    def _validate_resume_data(self, resume_data: Dict[str, Any]) -> bool:
        """验证简历数据完整性"""
        required_fields = ['id', 'name']
        return all(field in resume_data for field in required_fields)
    
    async def _basic_filter(self, resume_data: Dict[str, Any], 
                          filters: Optional[Dict]) -> List[Dict]:
        """基础条件筛选"""
        # 这里实现基础筛选逻辑
        # 例如：薪资范围、工作地点、工作类型等
        return []  # 简化实现
    
    async def _rank_and_filter(self, scored_matches: List[Dict], 
                             limit: int) -> List[Dict]:
        """结果排序和过滤"""
        # 按综合评分排序
        scored_matches.sort(key=lambda x: x['overall_score'], reverse=True)
        
        # 返回前N个结果
        return scored_matches[:limit]

    # ==============================================
    # 新增方法 - 适配AI服务API
    # ==============================================
    
    async def find_enhanced_matches(self, user_id: int, resume_id: int, 
                                  limit: int = 10, filters: Dict = None) -> List[Dict]:
        """
        增强版匹配方法 - 适配AI服务API
        
        Args:
            user_id: 用户ID
            resume_id: 简历ID
            limit: 返回结果数量限制
            filters: 筛选条件
            
        Returns:
            List[Dict]: 匹配结果列表
        """
        try:
            # 获取简历数据
            resume_data = await self._get_resume_data(resume_id)
            if not resume_data:
                logger.warning(f"简历数据不存在: resume_id={resume_id}")
                return []
            
            # 调用现有的匹配方法
            matches = await self.find_matching_jobs(resume_data, user_id, limit, filters or {})
            
            # 转换格式以适配AI服务响应
            enhanced_matches = []
            for match in matches:
                enhanced_match = {
                    "job_id": match.get("job_id"),
                    "match_score": match.get("overall_score", 0.0),
                    "breakdown": {
                        "semantic_similarity": match.get("semantic_similarity", 0.0),
                        "skills_match": match.get("skills_match", 0.0),
                        "experience_match": match.get("experience_match", 0.0),
                        "education_match": match.get("education_match", 0.0),
                        "cultural_fit": match.get("cultural_fit", 0.0)
                    },
                    "confidence": match.get("confidence", 0.0),
                    "job_info": match.get("job_info", {}),
                    "company_info": match.get("company_info", {}),
                    "reason": match.get("reason", "基于多维度匹配算法")
                }
                enhanced_matches.append(enhanced_match)
            
            return enhanced_matches
            
        except Exception as e:
            logger.error(f"增强版匹配失败: {e}")
            return []
    
    async def generate_recommendations(self, user_id: int, resume_id: int) -> Dict:
        """
        生成匹配推荐建议
        
        Args:
            user_id: 用户ID
            resume_id: 简历ID
            
        Returns:
            Dict: 推荐建议
        """
        try:
            # 获取简历数据
            resume_data = await self._get_resume_data(resume_id)
            if not resume_data:
                return {"error": "简历数据不存在"}
            
            # 生成基础推荐
            recommendations = await self._generate_recommendations(resume_data, user_id)
            
            # 添加个性化建议
            personalized_suggestions = {
                "skill_improvements": self._get_skill_suggestions(resume_data, []),
                "resume_optimizations": self._get_resume_suggestions(resume_data, []),
                "career_advice": [
                    "建议关注技能提升，特别是热门技术栈",
                    "考虑补充项目经验以提高竞争力",
                    "优化简历关键词以提高匹配度"
                ]
            }
            
            return {
                "user_id": user_id,
                "resume_id": resume_id,
                "recommendations": recommendations,
                "personalized_suggestions": personalized_suggestions,
                "generated_at": datetime.now().isoformat()
            }
            
        except Exception as e:
            logger.error(f"生成推荐建议失败: {e}")
            return {"error": str(e)}
    
    async def generate_analysis(self, user_id: int, resume_id: int) -> Dict:
        """
        生成匹配分析报告
        
        Args:
            user_id: 用户ID
            resume_id: 简历ID
            
        Returns:
            Dict: 分析报告
        """
        try:
            # 获取简历数据
            resume_data = await self._get_resume_data(resume_id)
            if not resume_data:
                return {"error": "简历数据不存在"}
            
            # 获取历史匹配数据
            history = await self._get_matching_history(user_id, resume_id)
            
            # 生成分析报告
            analysis = {
                "resume_analysis": {
                    "strengths": self._analyze_resume_strengths(resume_data),
                    "weaknesses": self._analyze_resume_weaknesses(resume_data),
                    "completeness_score": self._calculate_completeness_score(resume_data)
                },
                "matching_trends": {
                    "average_match_score": self._calculate_average_match_score(history),
                    "top_matching_industries": self._get_top_industries(history),
                    "skill_gaps": self._identify_skill_gaps(resume_data, history)
                },
                "improvement_areas": {
                    "priority_skills": self._get_priority_skills(resume_data),
                    "experience_gaps": self._identify_experience_gaps(resume_data),
                    "education_opportunities": self._get_education_opportunities(resume_data)
                },
                "market_insights": {
                    "demand_trends": self._get_demand_trends(),
                    "salary_expectations": self._get_salary_insights(resume_data),
                    "competition_analysis": self._get_competition_analysis(resume_data)
                }
            }
            
            return {
                "user_id": user_id,
                "resume_id": resume_id,
                "analysis": analysis,
                "generated_at": datetime.now().isoformat()
            }
            
        except Exception as e:
            logger.error(f"生成分析报告失败: {e}")
            return {"error": str(e)}
    
    # ==============================================
    # 辅助方法
    # ==============================================
    
    async def _get_resume_data(self, resume_id: int) -> Optional[Dict]:
        """获取简历数据"""
        try:
            # 这里应该从数据库获取简历数据
            # 暂时返回模拟数据
            return {
                "id": resume_id,
                "content": "模拟简历内容",
                "skills": ["Python", "Go", "JavaScript"],
                "experience": [{"title": "软件工程师", "duration": "2年"}],
                "education": [{"degree": "本科", "major": "计算机科学"}]
            }
        except Exception as e:
            logger.error(f"获取简历数据失败: {e}")
            return None
    
    async def _get_matching_history(self, user_id: int, resume_id: int) -> List[Dict]:
        """获取匹配历史"""
        try:
            # 这里应该从数据库获取匹配历史
            # 暂时返回空列表
            return []
        except Exception as e:
            logger.error(f"获取匹配历史失败: {e}")
            return []
    
    def _analyze_resume_strengths(self, resume_data: Dict) -> List[str]:
        """分析简历优势"""
        strengths = []
        if resume_data.get("skills"):
            strengths.append("技能丰富")
        if resume_data.get("experience"):
            strengths.append("有工作经验")
        return strengths
    
    def _analyze_resume_weaknesses(self, resume_data: Dict) -> List[str]:
        """分析简历劣势"""
        weaknesses = []
        if not resume_data.get("skills"):
            weaknesses.append("技能信息不足")
        if not resume_data.get("experience"):
            weaknesses.append("缺乏工作经验")
        return weaknesses
    
    def _calculate_completeness_score(self, resume_data: Dict) -> float:
        """计算简历完整度分数"""
        score = 0.0
        if resume_data.get("content"):
            score += 0.3
        if resume_data.get("skills"):
            score += 0.3
        if resume_data.get("experience"):
            score += 0.3
        if resume_data.get("education"):
            score += 0.1
        return score
    
    def _calculate_average_match_score(self, history: List[Dict]) -> float:
        """计算平均匹配分数"""
        if not history:
            return 0.0
        total_score = sum(match.get("score", 0) for match in history)
        return total_score / len(history)
    
    def _get_top_industries(self, history: List[Dict]) -> List[str]:
        """获取匹配最多的行业"""
        # 简化实现
        return ["互联网", "金融", "教育"]
    
    def _identify_skill_gaps(self, resume_data: Dict, history: List[Dict]) -> List[str]:
        """识别技能缺口"""
        return ["机器学习", "云计算", "DevOps"]
    
    def _get_priority_skills(self, resume_data: Dict) -> List[str]:
        """获取优先技能"""
        return ["Python", "Go", "Docker", "Kubernetes"]
    
    def _identify_experience_gaps(self, resume_data: Dict) -> List[str]:
        """识别经验缺口"""
        return ["团队管理", "项目管理", "架构设计"]
    
    def _get_education_opportunities(self, resume_data: Dict) -> List[str]:
        """获取教育机会"""
        return ["在线课程", "认证考试", "技术会议"]
    
    def _get_demand_trends(self) -> Dict:
        """获取需求趋势"""
        return {
            "hot_skills": ["AI/ML", "云计算", "区块链"],
            "growing_industries": ["新能源", "生物技术", "人工智能"]
        }
    
    def _get_salary_insights(self, resume_data: Dict) -> Dict:
        """获取薪资洞察"""
        return {
            "market_average": "15-25万",
            "recommended_range": "18-22万",
            "growth_potential": "高"
        }
    
    def _get_competition_analysis(self, resume_data: Dict) -> Dict:
        """获取竞争分析"""
        return {
            "competition_level": "中等",
            "differentiation_factors": ["技术深度", "项目经验"],
            "market_position": "中上"
        }
