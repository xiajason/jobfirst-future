#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
AI职位匹配引擎 - 核心匹配算法实现
基于多维度向量相似度的智能匹配

创建时间: 2025-09-13
作者: AI Assistant
版本: 1.0.0
"""

import asyncio
import json
import logging
import numpy as np
from typing import Dict, List, Optional, Any, Tuple
from datetime import datetime
import asyncpg

logger = logging.getLogger(__name__)

class JobMatchingEngine:
    """职位匹配引擎"""
    
    def __init__(self, data_access, postgres_pool):
        """
        初始化匹配引擎
        
        Args:
            data_access: 数据访问层实例
            postgres_pool: PostgreSQL连接池
        """
        self.data_access = data_access
        self.postgres_pool = postgres_pool
        
        # 默认匹配权重配置
        self.default_weights = {
            'semantic': 0.35,      # 语义相似度
            'skills': 0.30,        # 技能匹配
            'experience': 0.20,    # 经验匹配
            'basic': 0.10,         # 基础条件
            'cultural': 0.05       # 文化匹配
        }
        
        # 行业特定权重配置
        self.industry_weights = {
            'technology': {
                'skills': 0.40,      # 技术行业更重视技能
                'semantic': 0.30,
                'experience': 0.20,
                'basic': 0.10
            },
            'finance': {
                'semantic': 0.40,    # 金融行业更重视经验描述
                'experience': 0.30,
                'skills': 0.20,
                'basic': 0.10
            },
            'marketing': {
                'semantic': 0.35,
                'cultural': 0.25,    # 营销行业更重视文化匹配
                'skills': 0.25,
                'experience': 0.15
            }
        }
    
    async def find_matching_jobs(self, resume_data: Dict[str, Any], user_id: int, 
                                limit: int = 10, filters: Optional[Dict] = None) -> List[Dict[str, Any]]:
        """
        基于多维度向量匹配职位 - 适配新架构
        
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
            
            # 3. 向量相似度计算
            vector_matches = await self._vector_similarity_search(
                resume_data['vectors'], basic_filtered_jobs
            )
            if not vector_matches:
                logger.info("向量相似度搜索无结果")
                return []
            
            # 4. 多维度评分
            scored_matches = []
            for job_vector in vector_matches:
                score_result = await self._calculate_multidimensional_score(
                    resume_data, job_vector
                )
                
                if score_result['overall'] > 0.3:  # 最低匹配度阈值
                    scored_matches.append({
                        'job_id': job_vector['job_id'],
                        'match_score': score_result['overall'],
                        'breakdown': score_result['breakdown'],
                        'confidence': score_result['confidence'],
                        'job_info': job_vector['job_info']
                    })
            
            # 5. 排序并返回结果
            scored_matches.sort(key=lambda x: x['match_score'], reverse=True)
            
            # 6. 记录匹配操作日志
            await self.data_access.log_job_matching_access(
                user_id, resume_data['resume_id'], len(scored_matches)
            )
            
            logger.info(f"职位匹配完成: user_id={user_id}, resume_id={resume_data['resume_id']}, "
                       f"candidates={len(basic_filtered_jobs)}, matches={len(scored_matches)}")
            
            return scored_matches[:limit]
            
        except Exception as e:
            logger.error(f"职位匹配失败: {e}")
            return []
    
    def _validate_resume_data(self, resume_data: Dict[str, Any]) -> bool:
        """验证简历数据完整性"""
        try:
            required_keys = ['metadata', 'parsed_data', 'vectors', 'user_id', 'resume_id']
            for key in required_keys:
                if key not in resume_data:
                    logger.error(f"简历数据缺少必要字段: {key}")
                    return False
            
            # 验证向量数据
            vectors = resume_data['vectors']
            vector_keys = ['content_vector', 'skills_vector', 'experience_vector']
            for key in vector_keys:
                if key not in vectors or not vectors[key]:
                    logger.error(f"向量数据缺少必要字段: {key}")
                    return False
            
            return True
            
        except Exception as e:
            logger.error(f"验证简历数据异常: {e}")
            return False
    
    async def _basic_filter(self, resume_data: Dict[str, Any], filters: Optional[Dict]) -> List[int]:
        """基础筛选 (硬性条件)"""
        try:
            # 获取活跃职位列表
            active_jobs = await self.data_access.get_active_jobs(limit=1000)
            if not active_jobs:
                return []
            
            # 应用筛选条件
            filtered_jobs = active_jobs
            
            if filters:
                # 行业筛选
                if 'industry' in filters:
                    filtered_jobs = await self._filter_by_industry(filtered_jobs, filters['industry'])
                
                # 地理位置筛选
                if 'location' in filters:
                    filtered_jobs = await self._filter_by_location(filtered_jobs, filters['location'])
                
                # 薪资范围筛选
                if 'salary_min' in filters or 'salary_max' in filters:
                    filtered_jobs = await self._filter_by_salary(filtered_jobs, filters)
                
                # 工作经验筛选
                if 'experience' in filters:
                    filtered_jobs = await self._filter_by_experience(filtered_jobs, filters['experience'])
            
            return filtered_jobs
            
        except Exception as e:
            logger.error(f"基础筛选失败: {e}")
            return []
    
    async def _filter_by_industry(self, job_ids: List[int], industry: str) -> List[int]:
        """按行业筛选"""
        try:
            async with self.data_access.mysql_pool.acquire() as conn:
                async with conn.cursor() as cursor:
                    await cursor.execute("""
                        SELECT id FROM jobs 
                        WHERE id IN ({}) AND industry = %s
                    """.format(','.join(['%s'] * len(job_ids))), 
                    job_ids + [industry])
                    
                    results = await cursor.fetchall()
                    return [row[0] for row in results]
                    
        except Exception as e:
            logger.error(f"按行业筛选失败: {e}")
            return job_ids
    
    async def _filter_by_location(self, job_ids: List[int], location: str) -> List[int]:
        """按地理位置筛选"""
        try:
            async with self.data_access.mysql_pool.acquire() as conn:
                async with conn.cursor() as cursor:
                    await cursor.execute("""
                        SELECT id FROM jobs 
                        WHERE id IN ({}) AND location LIKE %s
                    """.format(','.join(['%s'] * len(job_ids))), 
                    job_ids + [f'%{location}%'])
                    
                    results = await cursor.fetchall()
                    return [row[0] for row in results]
                    
        except Exception as e:
            logger.error(f"按地理位置筛选失败: {e}")
            return job_ids
    
    async def _filter_by_salary(self, job_ids: List[int], salary_filters: Dict) -> List[int]:
        """按薪资范围筛选"""
        try:
            query = "SELECT id FROM jobs WHERE id IN ({})".format(','.join(['%s'] * len(job_ids)))
            params = job_ids.copy()
            
            if 'salary_min' in salary_filters:
                query += " AND salary_max >= %s"
                params.append(salary_filters['salary_min'])
            
            if 'salary_max' in salary_filters:
                query += " AND salary_min <= %s"
                params.append(salary_filters['salary_max'])
            
            async with self.data_access.mysql_pool.acquire() as conn:
                async with conn.cursor() as cursor:
                    await cursor.execute(query, params)
                    results = await cursor.fetchall()
                    return [row[0] for row in results]
                    
        except Exception as e:
            logger.error(f"按薪资范围筛选失败: {e}")
            return job_ids
    
    async def _filter_by_experience(self, job_ids: List[int], experience: str) -> List[int]:
        """按工作经验筛选"""
        try:
            async with self.data_access.mysql_pool.acquire() as conn:
                async with conn.cursor() as cursor:
                    await cursor.execute("""
                        SELECT id FROM jobs 
                        WHERE id IN ({}) AND experience = %s
                    """.format(','.join(['%s'] * len(job_ids))), 
                    job_ids + [experience])
                    
                    results = await cursor.fetchall()
                    return [row[0] for row in results]
                    
        except Exception as e:
            logger.error(f"按工作经验筛选失败: {e}")
            return job_ids
    
    async def _vector_similarity_search(self, resume_vectors: Dict[str, Any], 
                                      job_ids: List[int]) -> List[Dict[str, Any]]:
        """向量相似度搜索"""
        try:
            if not job_ids:
                return []
            
            async with self.postgres_pool.acquire() as conn:
                # 构建向量相似度查询
                query = """
                SELECT 
                    jv.job_id,
                    jv.title_vector,
                    jv.description_vector,
                    jv.requirements_vector,
                    (jv.description_vector <=> $1::vector) as semantic_distance,
                    (jv.requirements_vector <=> $2::vector) as skills_distance,
                    (jv.requirements_vector <=> $3::vector) as experience_distance,
                    j.title, j.description, j.requirements, j.company_id,
                    j.industry, j.location, j.salary_min, j.salary_max
                FROM job_vectors jv
                JOIN jobs j ON jv.job_id = j.id
                WHERE jv.job_id = ANY($4::int[])
                ORDER BY 
                    (jv.description_vector <=> $1::vector) + 
                    (jv.requirements_vector <=> $2::vector) +
                    (jv.requirements_vector <=> $3::vector)
                LIMIT 100
                """
                
                results = await conn.fetch(query, 
                    resume_vectors['content_vector'],
                    resume_vectors['skills_vector'], 
                    resume_vectors['experience_vector'],
                    job_ids
                )
                
                # 转换为字典列表
                vector_matches = []
                for row in results:
                    vector_matches.append({
                        'job_id': row['job_id'],
                        'title_vector': row['title_vector'],
                        'description_vector': row['description_vector'],
                        'requirements_vector': row['requirements_vector'],
                        'semantic_distance': float(row['semantic_distance']),
                        'skills_distance': float(row['skills_distance']),
                        'experience_distance': float(row['experience_distance']),
                        'job_info': {
                            'title': row['title'],
                            'description': row['description'],
                            'requirements': row['requirements'],
                            'company_id': row['company_id'],
                            'industry': row['industry'],
                            'location': row['location'],
                            'salary_min': row['salary_min'],
                            'salary_max': row['salary_max']
                        }
                    })
                
                return vector_matches
                
        except Exception as e:
            logger.error(f"向量相似度搜索失败: {e}")
            return []
    
    async def _calculate_multidimensional_score(self, resume_data: Dict[str, Any], 
                                              job_vector: Dict[str, Any]) -> Dict[str, Any]:
        """计算多维度匹配分数"""
        try:
            # 获取权重配置
            industry = job_vector['job_info'].get('industry', '')
            weights = self._get_weights_for_industry(industry)
            
            # 1. 语义相似度 (内容匹配)
            semantic_score = 1 - job_vector['semantic_distance']
            
            # 2. 技能匹配度
            skills_score = 1 - job_vector['skills_distance']
            
            # 3. 经验匹配度
            experience_score = 1 - job_vector['experience_distance']
            
            # 4. 基础条件匹配 (硬性条件)
            basic_score = await self._calculate_basic_match_score(resume_data, job_vector)
            
            # 5. 文化匹配度 (软性条件)
            cultural_score = await self._calculate_cultural_match_score(resume_data, job_vector)
            
            # 6. 综合评分
            overall_score = (
                semantic_score * weights['semantic'] +
                skills_score * weights['skills'] +
                experience_score * weights['experience'] +
                basic_score * weights['basic'] +
                cultural_score * weights.get('cultural', 0.05)
            )
            
            # 计算置信度
            confidence = self._calculate_confidence(semantic_score, skills_score, experience_score)
            
            return {
                'overall': max(0, min(1, overall_score)),  # 限制在0-1范围内
                'breakdown': {
                    'semantic': semantic_score,
                    'skills': skills_score,
                    'experience': experience_score,
                    'basic': basic_score,
                    'cultural': cultural_score
                },
                'confidence': confidence,
                'weights_used': weights
            }
            
        except Exception as e:
            logger.error(f"计算多维度匹配分数失败: {e}")
            return {
                'overall': 0,
                'breakdown': {},
                'confidence': 0,
                'weights_used': {}
            }
    
    def _get_weights_for_industry(self, industry: str) -> Dict[str, float]:
        """获取行业特定权重配置"""
        industry_lower = industry.lower()
        
        # 检查是否有特定行业的权重配置
        for key, weights in self.industry_weights.items():
            if key in industry_lower:
                return {**self.default_weights, **weights}
        
        # 返回默认权重
        return self.default_weights.copy()
    
    async def _calculate_basic_match_score(self, resume_data: Dict[str, Any], 
                                         job_vector: Dict[str, Any]) -> float:
        """计算基础条件匹配分数"""
        try:
            score = 0.0
            job_info = job_vector['job_info']
            parsed_data = resume_data['parsed_data']['parsed']
            
            # 学历匹配 (如果有教育信息)
            if parsed_data.get('education') and job_info.get('education'):
                if self._education_match(parsed_data['education'], job_info['education']):
                    score += 0.3
            
            # 工作经验匹配
            if parsed_data.get('work_experience') and job_info.get('experience'):
                if self._experience_level_match(parsed_data['work_experience'], job_info['experience']):
                    score += 0.4
            
            # 地理位置匹配 (如果有个人信息)
            if parsed_data.get('personal_info') and job_info.get('location'):
                if self._location_match(parsed_data['personal_info'], job_info['location']):
                    score += 0.3
            
            return min(score, 1.0)
            
        except Exception as e:
            logger.error(f"计算基础条件匹配分数失败: {e}")
            return 0.0
    
    def _education_match(self, education_data: List[Dict], required_education: str) -> bool:
        """学历匹配检查"""
        try:
            if not education_data or not required_education:
                return True  # 如果没有要求，认为匹配
            
            # 简化的学历匹配逻辑
            education_levels = {
                '博士': ['博士', 'phd'],
                '硕士': ['硕士', '研究生', 'master'],
                '本科': ['本科', '学士', 'bachelor'],
                '大专': ['大专', '专科', 'college'],
                '高中': ['高中', '中专', 'high school']
            }
            
            required_level = required_education.lower()
            for edu in education_data:
                if isinstance(edu, dict) and 'degree' in edu:
                    degree = str(edu['degree']).lower()
                    for level, keywords in education_levels.items():
                        if any(keyword in degree for keyword in keywords):
                            # 检查是否满足要求
                            if self._is_education_level_sufficient(level, required_level):
                                return True
            
            return False
            
        except Exception as e:
            logger.error(f"学历匹配检查失败: {e}")
            return False
    
    def _is_education_level_sufficient(self, candidate_level: str, required_level: str) -> bool:
        """检查学历水平是否满足要求"""
        level_hierarchy = {
            '博士': 5,
            '硕士': 4,
            '本科': 3,
            '大专': 2,
            '高中': 1
        }
        
        candidate_score = level_hierarchy.get(candidate_level, 0)
        required_score = level_hierarchy.get(required_level, 0)
        
        return candidate_score >= required_score
    
    def _experience_level_match(self, work_experience: List[Dict], required_experience: str) -> bool:
        """工作经验匹配检查"""
        try:
            if not work_experience or not required_experience:
                return True
            
            # 计算总工作经验年限
            total_years = 0
            for exp in work_experience:
                if isinstance(exp, dict) and 'duration' in exp:
                    duration = str(exp['duration']).lower()
                    # 简化的年限提取
                    if '年' in duration:
                        try:
                            years = float(duration.split('年')[0])
                            total_years += years
                        except:
                            pass
            
            # 检查是否满足经验要求
            if '年以上' in required_experience:
                try:
                    required_years = float(required_experience.split('年以上')[0])
                    return total_years >= required_years
                except:
                    pass
            
            return total_years >= 1  # 默认至少1年经验
            
        except Exception as e:
            logger.error(f"工作经验匹配检查失败: {e}")
            return False
    
    def _location_match(self, personal_info: Dict, job_location: str) -> bool:
        """地理位置匹配检查"""
        try:
            if not personal_info or not job_location:
                return True
            
            # 获取个人地址信息
            address = personal_info.get('address', '') or personal_info.get('location', '')
            if not address:
                return True  # 如果没有地址信息，认为匹配
            
            # 简化的地理位置匹配
            job_city = job_location.split('市')[0] if '市' in job_location else job_location
            return job_city in address or address in job_location
            
        except Exception as e:
            logger.error(f"地理位置匹配检查失败: {e}")
            return False
    
    async def _calculate_cultural_match_score(self, resume_data: Dict[str, Any], 
                                            job_vector: Dict[str, Any]) -> float:
        """计算文化匹配度 (软性条件)"""
        try:
            # 这里可以实现更复杂的文化匹配逻辑
            # 目前返回一个基于行业和技能的简化评分
            
            job_info = job_vector['job_info']
            parsed_data = resume_data['parsed_data']['parsed']
            
            cultural_score = 0.0
            
            # 基于技能的软性匹配
            if parsed_data.get('skills') and job_info.get('requirements'):
                skill_match_count = 0
                total_skills = len(parsed_data['skills'])
                
                if total_skills > 0:
                    for skill in parsed_data['skills']:
                        if skill.lower() in job_info['requirements'].lower():
                            skill_match_count += 1
                    
                    cultural_score += (skill_match_count / total_skills) * 0.6
            
            # 基于行业经验的匹配
            industry = job_info.get('industry', '')
            if industry and parsed_data.get('work_experience'):
                for exp in parsed_data['work_experience']:
                    if isinstance(exp, dict) and 'industry' in exp:
                        if industry.lower() in str(exp['industry']).lower():
                            cultural_score += 0.4
                            break
            
            return min(cultural_score, 1.0)
            
        except Exception as e:
            logger.error(f"计算文化匹配度失败: {e}")
            return 0.0
    
    def _calculate_confidence(self, semantic_score: float, skills_score: float, 
                            experience_score: float) -> float:
        """计算匹配置信度"""
        try:
            # 基于各维度分数的加权平均
            weights = [0.4, 0.4, 0.2]  # 语义和技能更重要
            scores = [semantic_score, skills_score, experience_score]
            
            weighted_avg = sum(w * s for w, s in zip(weights, scores)) / sum(weights)
            
            # 添加分数稳定性因子
            score_variance = np.var(scores)
            stability_factor = max(0, 1 - score_variance)
            
            confidence = weighted_avg * stability_factor
            return max(0, min(1, confidence))
            
        except Exception as e:
            logger.error(f"计算匹配置信度失败: {e}")
            return 0.0
    
    async def update_matching_weights(self, new_weights: Dict[str, float]):
        """动态更新匹配权重配置"""
        try:
            self.default_weights.update(new_weights)
            
            # 这里可以保存到Redis缓存
            logger.info(f"匹配权重配置已更新: {new_weights}")
            
        except Exception as e:
            logger.error(f"更新匹配权重配置失败: {e}")


# 使用示例
async def main():
    """使用示例"""
    # 这里需要实际的数据库连接
    # data_access = JobMatchingDataAccess(MYSQL_CONFIG, POSTGRES_CONFIG)
    # await data_access.initialize()
    
    # matching_engine = JobMatchingEngine(data_access, postgres_pool)
    
    print("职位匹配引擎初始化完成")


if __name__ == "__main__":
    asyncio.run(main())
