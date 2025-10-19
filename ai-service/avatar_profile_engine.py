#!/usr/bin/env python3
"""
AI分身画像引擎 - Layer 2完整实现
整合MBIT + 简历 + 职位匹配数据
"""
import logging
import json
from datetime import datetime
from typing import Dict, Any
from mbit_analyzer import MBITAnalyzer
from resume_analyzer import ResumeDataAnalyzer

logger = logging.getLogger(__name__)

class AvatarProfileEngine:
    """AI分身画像引擎（Layer 2核心）"""
    
    def __init__(self):
        self.mbit_analyzer = MBITAnalyzer()
        self.resume_analyzer = ResumeDataAnalyzer()
    
    async def build_profile(self, user_data: dict) -> Dict[str, Any]:
        """
        构建完整用户画像
        
        Args:
            user_data: 包含user_id, resume_data, mbit_type等
        """
        user_id = user_data.get('user_id', 0)
        mbit_type = user_data.get('mbit_type', 'UNKNOWN')
        resume_data = user_data.get('resume_data', {})
        
        logger.info(f"🧠 Layer 2: 构建用户画像 user_id={user_id}")
        
        # 1. MBIT性格分析
        personality_profile = self.mbit_analyzer.get_personality_profile(mbit_type)
        communication_guide = self.mbit_analyzer.generate_communication_guide(mbit_type)
        
        # 2. 简历技能分析
        skills = resume_data.get('skills', [])
        if isinstance(skills, str):
            skills = json.loads(skills) if skills else []
        
        skills_analysis = self.resume_analyzer.analyze_skills(skills)
        
        # 3. 职业阶段判断
        years = resume_data.get('years_experience', 0)
        career_stage = self.resume_analyzer.determine_career_stage(years)
        
        # 4. 综合画像
        profile = {
            'user_id': user_id,
            'created_at': datetime.now().isoformat(),
            
            # 性格层
            'personality': {
                'mbit_type': mbit_type,
                'name': personality_profile['name'] if personality_profile else 'Unknown',
                'traits': personality_profile['traits'] if personality_profile else [],
                'communication_style': personality_profile['communication_style'] if personality_profile else '专业友好',
                'learning_preference': personality_profile['learning_preference'] if personality_profile else '实践学习',
                'communication_guide': communication_guide
            },
            
            # 能力层
            'capabilities': {
                'current_position': resume_data.get('current_position', ''),
                'years_experience': years,
                'career_stage': career_stage,
                'skills_analysis': skills_analysis,
                'competitiveness': skills_analysis.get('competitiveness', '一般')
            },
            
            # 偏好层（基于匹配历史）
            'preferences': user_data.get('preferences', {}),
            
            # 行为层
            'behavior': user_data.get('behavior', {})
        }
        
        logger.info(f"  ✅ 画像构建完成: {mbit_type} - {career_stage['name']}")
        
        return profile
    
    def generate_layer2_context(self, profile: Dict[str, Any]) -> str:
        """生成Layer 2上下文（给Layer 3使用）"""
        
        mbit = profile['personality']['mbit_type']
        stage = profile['capabilities']['career_stage']['name']
        skills_count = profile['capabilities']['skills_analysis']['total_count']
        hot_skills = profile['capabilities']['skills_analysis'].get('hot_skills', [])
        
        context = f"""【AI分身画像 - Layer 2深度分析】

✨ 性格特质 ({mbit} - {profile['personality']['name']}):
- 核心特征: {', '.join(profile['personality']['traits'][:3])}
- 沟通风格: {profile['personality']['communication_style']}
- 学习偏好: {profile['personality']['learning_preference']}

💼 职业状态:
- 当前职位: {profile['capabilities']['current_position']}
- 工作年限: {profile['capabilities']['years_experience']}年
- 职业阶段: {stage}
- 市场竞争力: {profile['capabilities']['competitiveness']}

🛠️ 技能评估:
- 总技能数: {skills_count}
- 热门技能: {', '.join(hot_skills[:5])}
- 技能市场价值: {profile['capabilities']['skills_analysis'].get('market_value_score', 0)}分

📋 建议风格要求:
{profile['capabilities']['career_stage']['advice_style']}
"""
        
        return context

# 测试
if __name__ == "__main__":
    print("\n测试AI分身画像引擎...")
    
    # 模拟用户数据
    test_user = {
        'user_id': 1,
        'mbit_type': 'INTJ',
        'resume_data': {
            'current_position': 'Python后端工程师',
            'years_experience': 3,
            'skills': ['Python', 'Docker', 'Kubernetes', 'MySQL', 'FastAPI', 'PyTorch']
        }
    }
    
    engine = AvatarProfileEngine()
    
    # 使用同步方式测试
    import asyncio
    async def test():
        profile = await engine.build_profile(test_user)
        
        print(f"\n用户画像:")
        print(f"  性格: {profile['personality']['mbit_type']} - {profile['personality']['name']}")
        print(f"  职位: {profile['capabilities']['current_position']}")
        print(f"  阶段: {profile['capabilities']['career_stage']['name']}")
        print(f"  竞争力: {profile['capabilities']['competitiveness']}")
        
        print(f"\n生成Layer 2上下文:")
        context = engine.generate_layer2_context(profile)
        print(context)
        
        print("\n✅ AI分身画像引擎测试完成")
    
    asyncio.run(test())
