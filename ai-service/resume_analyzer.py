#!/usr/bin/env python3
"""
简历数据分析引擎 - Layer 2核心组件
从腾讯云从库读取LoomaCRM解析好的简历数据
"""
import logging
import json
from typing import Dict, Any, List

logger = logging.getLogger(__name__)

class ResumeDataAnalyzer:
    """简历数据分析引擎"""
    
    def __init__(self):
        self.skill_market_values = self._init_skill_values()
    
    def _init_skill_values(self) -> dict:
        """技能市场价值表（2025年10月）"""
        return {
            # 编程语言
            'Python': 95, 'Java': 90, 'Go': 92, 'JavaScript': 88,
            'TypeScript': 90, 'Rust': 85, 'C++': 82,
            
            # 前端
            'React': 92, 'Vue': 88, 'Next.js': 90, 'Angular': 75,
            
            # 后端框架
            'Django': 85, 'FastAPI': 90, 'Spring': 88, 'Gin': 87,
            
            # 数据库
            'MySQL': 88, 'PostgreSQL': 90, 'MongoDB': 85, 'Redis': 92,
            
            # 云原生
            'Docker': 95, 'Kubernetes': 98, '微服务': 94,
            
            # AI/ML
            'PyTorch': 96, 'TensorFlow': 94, '机器学习': 95, '深度学习': 97,
            
            # DevOps
            'CI/CD': 90, 'Git': 85, 'Linux': 88, 'AWS': 92, '阿里云': 88
        }
    
    def analyze_skills(self, skills: List[str]) -> Dict[str, Any]:
        """技能分析"""
        
        if not skills:
            return {'total_count': 0, 'categories': {}, 'market_value_score': 0}
        
        # 分类
        categories = {
            'programming': [],
            'frameworks': [],
            'databases': [],
            'cloud': [],
            'ai_ml': [],
            'devops': [],
            'other': []
        }
        
        total_value = 0
        hot_skills = []
        
        for skill in skills:
            # 获取市场价值
            value = self.skill_market_values.get(skill, 70)
            total_value += value
            
            if value >= 90:
                hot_skills.append(skill)
            
            # 分类
            category = self._categorize_skill(skill)
            categories[category].append({
                'name': skill,
                'market_value': value,
                'is_hot': value >= 90
            })
        
        # 计算平均市场价值
        avg_value = total_value / len(skills) if skills else 0
        
        return {
            'total_count': len(skills),
            'categories': categories,
            'hot_skills': hot_skills,
            'market_value_score': round(avg_value, 1),
            'competitiveness': self._rate_competitiveness(avg_value)
        }
    
    def _categorize_skill(self, skill: str) -> str:
        """技能分类"""
        skill_lower = skill.lower()
        
        if skill in ['Python', 'Java', 'Go', 'JavaScript', 'TypeScript', 'Rust', 'C++']:
            return 'programming'
        elif skill in ['React', 'Vue', 'Django', 'FastAPI', 'Spring', 'Gin']:
            return 'frameworks'
        elif skill in ['MySQL', 'PostgreSQL', 'MongoDB', 'Redis']:
            return 'databases'
        elif skill in ['Docker', 'Kubernetes', 'AWS', '阿里云', '微服务']:
            return 'cloud'
        elif skill in ['PyTorch', 'TensorFlow', '机器学习', '深度学习']:
            return 'ai_ml'
        elif skill in ['CI/CD', 'Git', 'Linux']:
            return 'devops'
        else:
            return 'other'
    
    def _rate_competitiveness(self, avg_value: float) -> str:
        """评估竞争力等级"""
        if avg_value >= 92:
            return "极具竞争力"
        elif avg_value >= 88:
            return "很有竞争力"
        elif avg_value >= 80:
            return "有竞争力"
        elif avg_value >= 70:
            return "一般竞争力"
        else:
            return "需要提升"
    
    def determine_career_stage(self, years: int) -> Dict[str, Any]:
        """判断职业发展阶段"""
        
        stages = {
            'entry': (0, 1, '职场新人', '技能积累、快速学习', '详细指导、鼓励为主'),
            'junior': (1, 3, '初级工程师', '深度提升、项目经验', '具体方法、案例参考'),
            'mid': (3, 5, '中级工程师', '专业深化、技术广度', '战略建议、方向指引'),
            'senior': (5, 8, '高级工程师', '架构能力、技术领导', '高层规划、行业洞察'),
            'expert': (8, 100, '技术专家', '技术战略、团队管理', '同行交流、战略研讨')
        }
        
        for key, (min_y, max_y, name, focus, advice) in stages.items():
            if min_y <= years < max_y:
                return {
                    'stage': key,
                    'name': name,
                    'focus': focus,
                    'advice_style': advice,
                    'years': years
                }
        
        return stages['expert'][2:]  # 默认专家级

# 测试
if __name__ == "__main__":
    analyzer = ResumeDataAnalyzer()
    
    print("\n测试简历分析引擎...")
    
    # 测试技能分析
    test_skills = ['Python', 'Docker', 'Kubernetes', 'MySQL', 'React', 'PyTorch']
    skills_analysis = analyzer.analyze_skills(test_skills)
    
    print(f"\n技能分析:")
    print(f"  总数: {skills_analysis['total_count']}")
    print(f"  热门技能: {', '.join(skills_analysis['hot_skills'])}")
    print(f"  市场价值: {skills_analysis['market_value_score']}")
    print(f"  竞争力: {skills_analysis['competitiveness']}")
    
    # 测试职业阶段
    stage = analyzer.determine_career_stage(3)
    print(f"\n职业阶段分析 (3年经验):")
    print(f"  阶段: {stage['name']}")
    print(f"  重点: {stage['focus']}")
    print(f"  建议风格: {stage['advice_style']}")
    
    print("\n✅ 简历分析引擎测试完成")
