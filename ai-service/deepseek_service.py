#!/usr/bin/env python3
"""
DeepSeek专业知识服务 - Layer 3
提供准确的职业咨询和行业建议
"""
import logging
import os
from typing import Dict, Any, List
try:
    from openai import AsyncOpenAI, OpenAI
except ImportError:
    print("需要安装openai库: pip install openai")
    import sys
    sys.exit(1)

logger = logging.getLogger(__name__)

class DeepSeekService:
    """DeepSeek专业知识服务（Layer 3）"""
    
    def __init__(self, api_key: str = None):
        self.api_key = api_key or os.getenv("DEEPSEEK_API_KEY", "")
        self.base_url = os.getenv("DEEPSEEK_BASE_URL", "https://api.deepseek.com")
        self.model = os.getenv("DEEPSEEK_MODEL", "deepseek-chat")
        
        if not self.api_key:
            logger.warning("⚠️  DeepSeek API Key未配置")
        
        # 使用同步客户端用于测试
        self.client = OpenAI(
            api_key=self.api_key or "test-key",
            base_url=self.base_url
        )
    
    def professional_consult(self, message: str, layer2_context: str) -> Dict[str, Any]:
        """
        专业职业咨询
        
        Args:
            message: 用户消息
            layer2_context: Layer 2生成的用户画像上下文
        """
        logger.info(f"🎓 Layer 3: DeepSeek专业咨询")
        
        # 构建系统提示词
        system_prompt = f"""你是JobFirst的资深AI职业顾问，拥有20年HR和职业规划经验。

{layer2_context}

你的职责:
1. 基于用户画像提供个性化建议
2. 考虑用户性格特质调整建议方式  
3. 结合用户能力现状给出可执行路径
4. 参考市场数据提供准确信息

回复要求:
- 专业准确（基于行业数据和最佳实践）
- 结构清晰（使用分点、标题）
- 可执行性强（具体步骤和时间线）
- 个性化（考虑用户性格和背景）
- 鼓励性（正向引导）
"""
        
        try:
            # 调用DeepSeek API
            response = self.client.chat.completions.create(
                model=self.model,
                messages=[
                    {"role": "system", "content": system_prompt},
                    {"role": "user", "content": message}
                ],
                temperature=0.7,
                max_tokens=1500
            )
            
            reply = response.choices[0].message.content
            usage = response.usage
            
            # 计算成本
            cost = self._calculate_cost(usage)
            
            logger.info(f"  ✅ DeepSeek回复完成，tokens={usage.total_tokens}, 成本=￥{cost:.4f}")
            
            return {
                'success': True,
                'reply': reply,
                'usage': {
                    'prompt_tokens': usage.prompt_tokens,
                    'completion_tokens': usage.completion_tokens,
                    'total_tokens': usage.total_tokens
                },
                'cost': cost
            }
            
        except Exception as e:
            logger.error(f"  ❌ DeepSeek调用失败: {e}")
            return {
                'success': False,
                'error': str(e),
                'fallback_reply': self._generate_fallback_reply(message)
            }
    
    def _calculate_cost(self, usage) -> float:
        """计算成本（人民币）"""
        # DeepSeek定价: 输入￥0.001/1K, 输出￥0.002/1K
        input_cost = (usage.prompt_tokens / 1000) * 0.001
        output_cost = (usage.completion_tokens / 1000) * 0.002
        return input_cost + output_cost
    
    def _generate_fallback_reply(self, message: str) -> str:
        """生成降级回复（当API不可用时）"""
        return f"""抱歉，AI服务暂时不可用。

您的问题是：{message}

建议：
1. 稍后重试
2. 或联系人工客服
3. 或查看帮助文档

我们会尽快恢复服务。"""

# 测试
if __name__ == "__main__":
    print("\n测试DeepSeek服务...")
    print("注意: 需要配置DEEPSEEK_API_KEY环境变量才能实际调用")
    
    service = DeepSeekService()
    
    # 模拟Layer 2上下文
    test_context = """【AI分身画像 - Layer 2深度分析】

✨ 性格特质 (INTJ - 建筑师):
- 核心特征: 战略思维, 独立性强, 分析能力强
- 沟通风格: 逻辑严谨、数据驱动、简洁直接
- 学习偏好: 理论先行、系统性学习、深度钻研

💼 职业状态:
- 当前职位: Python后端工程师
- 工作年限: 3年
- 职业阶段: 中级工程师
- 市场竞争力: 极具竞争力

🛠️ 技能评估:
- 总技能数: 6
- 热门技能: Python, Docker, Kubernetes, FastAPI
"""
    
    test_message = "我想转行做AI工程师，你觉得可行吗？"
    
    print(f"\n用户提问: {test_message}")
    print(f"\nLayer 2上下文:\n{test_context}")
    
    # 如果有API Key才调用
    if service.api_key and service.api_key != "test-key":
        print("\n调用DeepSeek API...")
        result = service.professional_consult(test_message, test_context)
        
        if result['success']:
            print(f"\n✅ DeepSeek回复:")
            print(result['reply'])
            print(f"\n成本: ￥{result['cost']:.4f}")
            print(f"Tokens: {result['usage']['total_tokens']}")
        else:
            print(f"\n❌ 调用失败: {result['error']}")
    else:
        print("\n⚠️  未配置API Key，跳过实际调用")
        print("  设置方法: export DEEPSEEK_API_KEY=sk-xxx")
    
    print("\n✅ DeepSeek服务测试完成")
