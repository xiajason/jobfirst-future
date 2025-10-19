#!/usr/bin/env python3
"""
MBIT性格分析引擎 - Layer 2核心组件
"""
import logging
from typing import Dict, Any, Optional

logger = logging.getLogger(__name__)

class MBITAnalyzer:
    """MBIT性格分析引擎"""
    
    def __init__(self):
        self.personality_profiles = self._init_mbit_profiles()
    
    def _init_mbit_profiles(self) -> dict:
        """初始化16型人格档案"""
        return {
            "INTJ": {
                "name": "建筑师",
                "traits": ["战略思维", "独立性强", "分析能力强", "追求完美"],
                "communication_style": "逻辑严谨、数据驱动、简洁直接",
                "learning_preference": "理论先行、系统性学习、深度钻研",
                "career_fit": ["架构师", "研究员", "数据科学家", "AI工程师", "技术专家"],
                "decision_style": "深思熟虑、基于数据和逻辑",
                "advice_approach": "提供系统性方案、强调长期规划、使用数据支撑"
            },
            "ENFP": {
                "name": "竞选者",
                "traits": ["热情洋溢", "创意丰富", "善于社交", "灵活变通"],
                "communication_style": "热情互动、鼓励性强、创意发散",
                "learning_preference": "实践导向、多样化学习、项目驱动",
                "career_fit": ["产品经理", "市场营销", "创意设计", "教育培训", "BD"],
                "decision_style": "直觉导向、快速决策",
                "advice_approach": "提供多种可能性、鼓励尝试、案例丰富"
            },
            "ISTJ": {
                "name": "物流师",
                "traits": ["务实可靠", "细节导向", "执行力强", "稳定性高"],
                "communication_style": "实用主义、具体方案、步骤清晰",
                "learning_preference": "步骤式学习、清晰路径、实用技能",
                "career_fit": ["测试工程师", "运维工程师", "项目管理", "质量管理"],
                "decision_style": "基于经验、稳健谨慎",
                "advice_approach": "提供具体步骤、强调可行性、风险提示"
            },
            "ENTP": {
                "name": "辩论家",
                "traits": ["创新思维", "辩论能力", "好奇心强", "挑战权威"],
                "communication_style": "思辨性强、多角度分析、启发思考",
                "learning_preference": "探索式学习、问题驱动、跨领域",
                "career_fit": ["产品经理", "创业者", "顾问", "创新总监"],
                "decision_style": "创新导向、敢于尝试",
                "advice_approach": "激发思考、提供多角度分析、鼓励创新"
            }
            # 其他12种可后续补充
        }
    
    def get_personality_profile(self, mbit_type: str) -> Optional[Dict]:
        """获取性格档案"""
        return self.personality_profiles.get(mbit_type)
    
    def generate_communication_guide(self, mbit_type: str) -> Dict[str, str]:
        """生成沟通指南（给AI使用）"""
        
        profile = self.personality_profiles.get(mbit_type)
        
        if not profile:
            return {
                'tone': '专业友好',
                'structure': '清晰分点',
                'emphasis': '实用建议'
            }
        
        return {
            'tone': profile['communication_style'],
            'approach': profile['advice_approach'],
            'learning_style': profile['learning_preference'],
            'decision_support': profile['decision_style']
        }

# 测试
if __name__ == "__main__":
    analyzer = MBITAnalyzer()
    
    print("\n测试MBIT分析引擎...")
    
    # 测试INTJ
    profile = analyzer.get_personality_profile("INTJ")
    print(f"\nINTJ档案:")
    print(f"  名称: {profile['name']}")
    print(f"  特质: {', '.join(profile['traits'])}")
    print(f"  沟通风格: {profile['communication_style']}")
    
    # 测试沟通指南
    guide = analyzer.generate_communication_guide("INTJ")
    print(f"\nINTJ沟通指南:")
    print(f"  语气: {guide['tone']}")
    print(f"  方法: {guide['approach']}")
    
    print("\n✅ MBIT分析引擎测试完成")
