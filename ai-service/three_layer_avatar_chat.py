#!/usr/bin/env python3
"""
三层AI分身对话系统
Layer 1: WeClone (沟通风格) - 本地模型
Layer 2: MBIT + 简历 (数据分析) - Python分析
Layer 3: DeepSeek (专业知识) - Cloud API
"""
import logging
import json
import re
from datetime import datetime
from typing import Dict, Any, Optional
from avatar_profile_engine import AvatarProfileEngine
from deepseek_service import DeepSeekService

logger = logging.getLogger(__name__)

class ThreeLayerAvatarChat:
    """三层AI分身对话系统"""
    
    def __init__(self):
        # Layer 2: AI分身画像引擎
        self.profile_engine = AvatarProfileEngine()
        
        # Layer 3: DeepSeek专业服务
        self.deepseek = DeepSeekService()
        
        # Layer 1: WeClone (暂时用简单回复，下周部署实际模型)
        self.weclone_enabled = False
        
        logger.info("✅ 三层AI分身系统初始化完成")
    
    async def chat(self, user_data: dict, message: str) -> Dict[str, Any]:
        """
        三层AI分身对话主函数
        
        Args:
            user_data: 包含user_id, mbit_type, resume_data等
            message: 用户消息
        """
        user_id = user_data.get('user_id', 0)
        
        logger.info(f"\n{'='*60}")
        logger.info(f"🤖 三层AI分身对话 | user_id={user_id}")
        logger.info(f"📝 用户消息: {message}")
        logger.info(f"{'='*60}")
        
        # ═══════════════════════════════════════
        # Step 1: 意图识别
        # ═══════════════════════════════════════
        intent = self._analyze_intent(message)
        logger.info(f"📊 Step 1: 意图识别 → {intent}")
        
        # ═══════════════════════════════════════
        # Step 2: 路由决策  
        # ═══════════════════════════════════════
        route = self._decide_route(intent)
        logger.info(f"🎯 Step 2: 路由决策 → {route}")
        
        # ═══════════════════════════════════════
        # Step 3: Layer 2 - 构建用户画像
        # ═══════════════════════════════════════
        logger.info(f"🧠 Step 3: Layer 2 构建画像...")
        profile = await self.profile_engine.build_profile(user_data)
        layer2_context = self.profile_engine.generate_layer2_context(profile)
        logger.info(f"  ✅ 画像构建完成: {profile['personality']['mbit_type']} - {profile['capabilities']['career_stage']['name']}")
        
        # ═══════════════════════════════════════
        # Step 4: 根据路由调用相应Layer
        # ═══════════════════════════════════════
        
        layers_used = ['Layer 2: MBIT+简历分析']
        cost = 0.0
        
        if route == 'casual':
            # 闲聊 → Layer 1 (WeClone)
            logger.info(f"💬 Step 4: Layer 1处理（闲聊）")
            
            if self.weclone_enabled:
                reply = await self._chat_layer1_weclone(message, profile)
                layers_used.insert(0, 'Layer 1: WeClone风格')
            else:
                # WeClone未部署时的降级处理
                reply = self._generate_casual_reply(message, profile)
                layers_used.insert(0, 'Layer 1: 简单回复（待部署WeClone）')
            
        elif route == 'professional':
            # 专业咨询 → Layer 2 + Layer 3
            logger.info(f"🎓 Step 4: Layer 2+3处理（专业咨询）")
            
            result = self.deepseek.professional_consult(message, layer2_context)
            
            if result['success']:
                reply = result['reply']
                cost = result['cost']
                layers_used.append('Layer 3: DeepSeek专业知识')
                logger.info(f"  ✅ DeepSeek回复完成，成本=￥{cost:.4f}")
            else:
                reply = result.get('fallback_reply', '抱歉，服务暂时不可用')
                logger.warning(f"  ⚠️  DeepSeek失败，使用降级回复")
        
        else:  # hybrid
            # 混合 → 三层全部使用
            logger.info(f"🌟 Step 4: 三层全开（混合场景）")
            
            # Layer 3: DeepSeek生成专业内容
            result = self.deepseek.professional_consult(message, layer2_context)
            
            if result['success']:
                professional_reply = result['reply']
                cost = result['cost']
                layers_used.append('Layer 3: DeepSeek专业知识')
                
                # Layer 1: 风格转换（如果WeClone可用）
                if self.weclone_enabled:
                    reply = await self._style_transfer_layer1(
                        professional_reply,
                        profile['personality']['communication_style']
                    )
                    layers_used.insert(0, 'Layer 1: WeClone风格转换')
                else:
                    reply = professional_reply
                    layers_used.insert(0, 'Layer 1: 待部署WeClone')
            else:
                reply = result.get('fallback_reply', '抱歉，服务暂时不可用')
        
        # ═══════════════════════════════════════
        # Step 5: 返回结果
        # ═══════════════════════════════════════
        logger.info(f"✅ 对话完成: {len(reply)}字符，使用层级={len(layers_used)}")
        logger.info(f"{'='*60}\n")
        
        return {
            'reply': reply,
            'route': route,
            'layers_used': layers_used,
            'cost': cost,
            'profile_summary': {
                'mbit': profile['personality']['mbit_type'],
                'mbit_name': profile['personality']['name'],
                'career_stage': profile['capabilities']['career_stage']['name'],
                'competitiveness': profile['capabilities']['competitiveness']
            },
            'metadata': {
                'intent': intent,
                'timestamp': datetime.now().isoformat()
            }
        }
    
    def _analyze_intent(self, message: str) -> str:
        """意图识别"""
        
        # 闲聊模式
        casual_patterns = [
            r'^(你好|hi|hello|嗨|早|午安|晚安)',
            r'(哈哈|笑|😂|😄|😊)',
            r'(心情|感觉|情绪|开心|难过)',
            r'(聊天|闲聊|无聊)',
        ]
        
        # 专业咨询模式
        professional_patterns = [
            r'(转行|跳槽|换工作|离职)',
            r'(薪资|工资|薪水|待遇|收入)',
            r'(面试|准备|技巧|经验)',
            r'(学习|课程|培训|提升|进修)',
            r'(分析|评估|建议|规划)',
            r'(职位|工作|岗位|推荐|匹配)',
            r'(行业|市场|趋势|前景)',
        ]
        
        for pattern in casual_patterns:
            if re.search(pattern, message, re.IGNORECASE):
                return 'casual'
        
        for pattern in professional_patterns:
            if re.search(pattern, message, re.IGNORECASE):
                return 'professional'
        
        # 问号结尾且较长 → 专业咨询
        if message.endswith('?') or message.endswith('？'):
            if len(message) > 10:
                return 'professional'
        
        return 'mixed'
    
    def _decide_route(self, intent: str) -> str:
        """路由决策"""
        routing_map = {
            'casual': 'casual',              # Layer 1 (闲聊)
            'professional': 'professional',  # Layer 2+3 (专业)
            'mixed': 'hybrid'                # 三层全开
        }
        return routing_map.get(intent, 'hybrid')
    
    def _generate_casual_reply(self, message: str, profile: dict) -> str:
        """生成简单闲聊回复（WeClone未部署时的降级方案）"""
        
        mbit = profile['personality']['mbit_type']
        
        # 根据MBIT性格调整回复风格
        if mbit.startswith('E'):  # 外向型
            greetings = ['哈喽！', '嗨！', '你好呀！']
            tone = '热情'
        else:  # 内向型
            greetings = ['你好', 'Hi', '嗨']
            tone = '友好'
        
        # 简单模式匹配
        if re.search(r'(你好|hi|hello)', message, re.IGNORECASE):
            return f"{greetings[0]} 有什么可以帮你的吗？😊"
        elif re.search(r'(哈哈|笑)', message):
            return "哈哈，看来心情不错！😄"
        elif re.search(r'(心情不错|开心)', message):
            return "太好了！有什么开心的事吗？"
        elif re.search(r'(累|疲惫|辛苦)', message):
            return "辛苦啦！要不要聊聊天放松一下？"
        else:
            return "我在呢，有什么想聊的吗？"
    
    async def _chat_layer1_weclone(self, message: str, profile: dict) -> str:
        """Layer 1: WeClone风格对话（待实现）"""
        # TODO: 集成WeClone API
        return self._generate_casual_reply(message, profile)
    
    async def _style_transfer_layer1(self, professional_content: str, 
                                     communication_style: str) -> str:
        """Layer 1: 风格转换（待实现）"""
        # TODO: 使用WeClone进行风格转换
        return professional_content

# 测试
if __name__ == "__main__":
    print("\n测试三层AI分身系统...")
    
    # 模拟用户数据
    test_user = {
        'user_id': 1,
        'mbit_type': 'INTJ',
        'resume_data': {
            'current_position': 'Python后端工程师',
            'years_experience': 3,
            'skills': json.dumps(['Python', 'Docker', 'Kubernetes', 'MySQL', 'FastAPI', 'PyTorch'])
        }
    }
    
    chat_system = ThreeLayerAvatarChat()
    
    import asyncio
    
    async def test():
        # 测试1: 闲聊
        print("\n【测试1】闲聊场景:")
        result1 = await chat_system.chat(test_user, "今天心情不错😊")
        print(f"  路由: {result1['route']}")
        print(f"  层级: {result1['layers_used']}")
        print(f"  回复: {result1['reply']}")
        print(f"  成本: ￥{result1['cost']:.4f}")
        
        # 测试2: 专业咨询
        print("\n【测试2】专业咨询场景:")
        result2 = await chat_system.chat(test_user, "我想转行做AI工程师")
        print(f"  路由: {result2['route']}")
        print(f"  层级: {result2['layers_used']}")
        print(f"  回复长度: {len(result2['reply'])}字符")
        print(f"  成本: ￥{result2['cost']:.4f}")
        
        print("\n✅ 三层AI分身系统测试完成")
    
    asyncio.run(test())

