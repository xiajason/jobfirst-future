#!/usr/bin/env python3
"""
数据匿名化处理器
符合《个人信息保护法》第51条 敏感个人信息处理规则
"""
import hashlib
import json
from typing import Dict, Any, List, Optional

class DataAnonymizer:
    """
    数据匿名化处理器
    
    符合《个人信息保护法》第51条：
    处理敏感个人信息应当取得个人的单独同意
    """
    
    @staticmethod
    def anonymize_personal_info(data: Dict[str, Any], level: str = 'partial') -> Dict[str, Any]:
        """
        匿名化个人信息
        
        Args:
            data: 原始个人信息
            level: 匿名化级别 ('none', 'partial', 'full')
            
        Returns:
            匿名化后的数据
        """
        if level == 'none':
            return data
        
        anonymized = data.copy() if isinstance(data, dict) else json.loads(data) if isinstance(data, str) else {}
        
        if level == 'full':
            # 完全匿名化 - 用于敏感个人信息
            anonymized['name'] = '***'
            anonymized['email'] = '***@***.***'
            anonymized['phone'] = '***-****-****'
            anonymized['location'] = '***'
            anonymized['address'] = '***'
            anonymized['id_card'] = '***'
            
        elif level == 'partial':
            # 部分匿名化 - 用于基础信息
            if 'name' in anonymized and anonymized['name']:
                name = str(anonymized['name'])
                anonymized['name'] = name[0] + '*' * max(0, len(name) - 1)
            
            if 'email' in anonymized and anonymized['email']:
                email = str(anonymized['email'])
                if '@' in email:
                    parts = email.split('@')
                    anonymized['email'] = parts[0][0] + '****@' + parts[1]
                else:
                    anonymized['email'] = email[0] + '****'
            
            if 'phone' in anonymized and anonymized['phone']:
                phone = str(anonymized['phone'])
                if len(phone) >= 7:
                    anonymized['phone'] = phone[:3] + '****' + phone[-4:]
                else:
                    anonymized['phone'] = '***-****'
            
            if 'location' in anonymized and anonymized['location']:
                # 只保留城市级别
                location = str(anonymized['location'])
                parts = location.replace('省', ',').replace('市', ',').replace('区', ',').split(',')
                if len(parts) >= 2:
                    anonymized['location'] = parts[0] + parts[1]
                else:
                    anonymized['location'] = location[:2] + '***'
        
        return anonymized
    
    @staticmethod
    def anonymize_work_experience(data: List[Dict[str, Any]], level: str = 'partial') -> List[Dict[str, Any]]:
        """
        匿名化工作经历
        
        保留职位和行业信息，隐藏公司名称和详细地址
        """
        if level == 'none':
            return data
        
        anonymized = []
        
        for exp in data if isinstance(data, list) else [data]:
            exp_copy = exp.copy() if isinstance(exp, dict) else json.loads(exp) if isinstance(exp, str) else {}
            
            if level == 'full':
                exp_copy['company'] = '***'
                exp_copy['location'] = '***'
                if 'responsibilities' in exp_copy:
                    exp_copy['responsibilities'] = '***'
                    
            elif level == 'partial':
                # 公司名称部分匿名化
                if 'company' in exp_copy and exp_copy['company']:
                    company = str(exp_copy['company'])
                    exp_copy['company'] = company[:2] + '***公司'
                
                # 位置泛化到城市
                if 'location' in exp_copy and exp_copy['location']:
                    location = str(exp_copy['location'])
                    parts = location.split(',')
                    exp_copy['location'] = parts[0] if parts else location[:2] + '***'
            
            anonymized.append(exp_copy)
        
        return anonymized
    
    @staticmethod
    def anonymize_education(data: List[Dict[str, Any]], level: str = 'partial') -> List[Dict[str, Any]]:
        """匿名化教育背景"""
        if level == 'none':
            return data
        
        anonymized = []
        
        for edu in data if isinstance(data, list) else [data]:
            edu_copy = edu.copy() if isinstance(edu, dict) else json.loads(edu) if isinstance(edu, str) else {}
            
            if level == 'full':
                edu_copy['school'] = '***'
                edu_copy['location'] = '***'
                    
            elif level == 'partial':
                # 学校名称部分匿名化
                if 'school' in edu_copy and edu_copy['school']:
                    school = str(edu_copy['school'])
                    edu_copy['school'] = school[:2] + '***大学'
                
                # 位置泛化
                if 'location' in edu_copy and edu_copy['location']:
                    location = str(edu_copy['location'])
                    edu_copy['location'] = location.split(',')[0] if ',' in location else location[:2] + '***'
            
            anonymized.append(edu_copy)
        
        return anonymized
    
    @staticmethod
    def get_anonymization_level_for_role(accessor_role: str, data_owner_id: int, 
                                        accessor_user_id: int) -> str:
        """
        根据访问者角色和关系确定匿名化级别
        
        Args:
            accessor_role: 访问者角色
            data_owner_id: 数据所有者ID
            accessor_user_id: 访问者ID
            
        Returns:
            匿名化级别 ('none', 'partial', 'full')
        """
        # 自己访问自己的数据 - 不匿名化
        if data_owner_id == accessor_user_id:
            return 'none'
        
        # 超级管理员访问他人数据 - 部分匿名化（保留管理需要的信息）
        if accessor_role == 'super_admin':
            return 'partial'
        
        # 系统管理员访问他人数据 - 部分匿名化
        if accessor_role == 'system_admin':
            return 'partial'
        
        # 普通用户访问他人数据 - 完全匿名化
        return 'full'
    
    @staticmethod
    def hash_sensitive_field(value: str) -> str:
        """对敏感字段进行哈希处理（不可逆）"""
        if not value:
            return ''
        return hashlib.sha256(value.encode()).hexdigest()[:16]

# 测试代码
if __name__ == "__main__":
    print("\n测试数据匿名化处理器...")
    
    anonymizer = DataAnonymizer()
    
    # 测试数据
    test_personal_info = {
        'name': '张三',
        'email': 'zhangsan@example.com',
        'phone': '13812345678',
        'location': '北京市朝阳区',
        'address': '北京市朝阳区某某街道123号'
    }
    
    test_work_exp = [
        {
            'company': '阿里巴巴集团',
            'position': '高级软件工程师',
            'location': '北京市,朝阳区',
            'responsibilities': '负责核心业务系统开发'
        }
    ]
    
    test_education = [
        {
            'school': '北京大学',
            'major': '计算机科学',
            'location': '北京市,海淀区'
        }
    ]
    
    # 测试1: 不匿名化
    print("\n【测试1】不匿名化（用户访问自己的数据）...")
    result = anonymizer.anonymize_personal_info(test_personal_info, 'none')
    print(f"  姓名: {result.get('name', 'N/A')}")
    print(f"  邮箱: {result.get('email', 'N/A')}")
    
    # 测试2: 部分匿名化
    print("\n【测试2】部分匿名化（管理员访问）...")
    result = anonymizer.anonymize_personal_info(test_personal_info, 'partial')
    print(f"  姓名: {result.get('name', 'N/A')}")
    print(f"  邮箱: {result.get('email', 'N/A')}")
    print(f"  电话: {result.get('phone', 'N/A')}")
    print(f"  位置: {result.get('location', 'N/A')}")
    
    # 测试3: 完全匿名化
    print("\n【测试3】完全匿名化（普通用户访问他人数据）...")
    result = anonymizer.anonymize_personal_info(test_personal_info, 'full')
    print(f"  姓名: {result.get('name', 'N/A')}")
    print(f"  邮箱: {result.get('email', 'N/A')}")
    print(f"  电话: {result.get('phone', 'N/A')}")
    
    # 测试4: 工作经历匿名化
    print("\n【测试4】工作经历匿名化...")
    result = anonymizer.anonymize_work_experience(test_work_exp, 'partial')
    print(f"  公司: {result[0].get('company', 'N/A')}")
    print(f"  位置: {result[0].get('location', 'N/A')}")
    print(f"  职位: {result[0].get('position', 'N/A')}")  # 职位不匿名
    
    # 测试5: 教育背景匿名化
    print("\n【测试5】教育背景匿名化...")
    result = anonymizer.anonymize_education(test_education, 'partial')
    print(f"  学校: {result[0].get('school', 'N/A')}")
    print(f"  位置: {result[0].get('location', 'N/A')}")
    print(f"  专业: {result[0].get('major', 'N/A')}")  # 专业不匿名
    
    # 测试6: 角色级别判断
    print("\n【测试6】根据角色判断匿名化级别...")
    level = anonymizer.get_anonymization_level_for_role('super_admin', 1, 999)
    print(f"  超级管理员访问他人数据: {level}")
    
    level = anonymizer.get_anonymization_level_for_role('normal_user', 1, 1)
    print(f"  用户访问自己数据: {level}")
    
    level = anonymizer.get_anonymization_level_for_role('normal_user', 1, 2)
    print(f"  普通用户访问他人数据: {level}")
    
    print("\n✅ 数据匿名化处理器测试完成！")

