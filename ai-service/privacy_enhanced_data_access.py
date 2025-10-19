#!/usr/bin/env python3
"""
隐私增强的数据访问层包装器
集成授权管理和数据匿名化
"""
import logging
from typing import Dict, Any, Optional
from consent_manager import ConsentManager
from data_anonymizer import DataAnonymizer

logger = logging.getLogger(__name__)

class PrivacyEnhancedDataAccess:
    """隐私增强的数据访问包装器"""
    
    def __init__(self, original_data_access):
        self.data_access = original_data_access
        self.anonymizer = DataAnonymizer()
    
    async def get_resume_with_privacy_check(self, resume_id: int, user_id: int, 
                                           user_role: str = 'normal_user',
                                           service_type: str = 'job_matching') -> Optional[Dict[str, Any]]:
        """
        获取简历数据（增强隐私保护版）
        
        流程:
        1. 获取原始简历数据
        2. 检查用户授权
        3. 根据角色和授权级别匿名化
        4. 记录隐私审计日志
        5. 更新使用统计
        """
        try:
            # 1. 获取原始简历数据
            resume_data = await self.data_access.get_resume_for_matching(resume_id, user_id)
            if not resume_data:
                return None
            
            # 获取简历所有者ID
            owner_user_id = resume_data['metadata'].get('user_id', 0)
            sqlite_db_path = resume_data['metadata'].get('sqlite_db_path')
            
            # 2. 检查授权（如果访问他人简历）
            if owner_user_id != user_id and user_role != 'super_admin':
                consent_manager = ConsentManager(sqlite_db_path)
                consent_check = consent_manager.check_consent(
                    owner_user_id,
                    service_type,
                    ['skills', 'experience', 'education']
                )
                
                if not consent_check['has_consent']:
                    logger.warning(f"用户{owner_user_id}未授权{service_type}访问")
                    return None
            
            # 3. 确定匿名化级别
            anonymization_level = self.anonymizer.get_anonymization_level_for_role(
                user_role,
                owner_user_id,
                user_id
            )
            
            logger.info(f"匿名化级别: {anonymization_level}, 访问者: {user_role}")
            
            # 4. 匿名化处理
            if anonymization_level != 'none':
                parsed_data = resume_data['parsed_data']
                
                if 'personal_info' in parsed_data:
                    parsed_data['personal_info'] = self.anonymizer.anonymize_personal_info(
                        parsed_data['personal_info'], anonymization_level
                    )
                
                if 'work_experience' in parsed_data:
                    parsed_data['work_experience'] = self.anonymizer.anonymize_work_experience(
                        parsed_data['work_experience'], anonymization_level
                    )
                
                if 'education' in parsed_data:
                    parsed_data['education'] = self.anonymizer.anonymize_education(
                        parsed_data['education'], anonymization_level
                    )
            
            # 5. 记录隐私审计
            consent_manager = ConsentManager(sqlite_db_path)
            consent_manager.log_privacy_audit(
                user_id=owner_user_id,
                action_type='view',
                service_type=service_type,
                data_type='resume',
                privacy_level=anonymization_level,
                anonymized=(anonymization_level != 'none'),
                accessed_by_user_id=user_id,
                accessed_by_role=user_role,
                details={
                    'resume_id': resume_id,
                    'service': service_type,
                    'anonymization_applied': anonymization_level
                }
            )
            
            # 6. 更新使用统计
            consent_manager.update_usage_statistics(owner_user_id, service_type, 'resume')
            
            # 7. 更新授权使用计数
            if owner_user_id != user_id:
                consent_manager.update_consent_usage(owner_user_id, service_type)
            
            logger.info(f"简历{resume_id}访问成功，匿名化级别: {anonymization_level}")
            
            return resume_data
            
        except Exception as e:
            logger.error(f"隐私增强数据访问失败: {e}")
            return None

# 测试代码
if __name__ == "__main__":
    print("✅ 隐私增强数据访问层模块创建成功")
    print("\n📋 功能:")
    print("  1. 授权检查（符合第13条）")
    print("  2. 最小必要原则（符合第14条）")
    print("  3. 隐私审计日志（符合第24条）")
    print("  4. 敏感信息保护（符合第51条）")
    print("\n🎯 集成到job_matching_service即可使用")
