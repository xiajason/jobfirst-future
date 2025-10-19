#!/usr/bin/env python3
"""
授权管理器
符合《个人信息保护法》第13条 知情同意原则
"""
import sqlite3
import json
import logging
from datetime import datetime
from typing import Dict, Any, List, Optional

logger = logging.getLogger(__name__)

class ConsentManager:
    """授权管理器"""
    
    def __init__(self, db_path: str):
        """
        初始化授权管理器
        
        Args:
            db_path: 用户SQLite数据库路径
        """
        self.db_path = db_path
    
    def check_consent(self, user_id: int, service_type: str, 
                     data_types: List[str]) -> Dict[str, Any]:
        """
        检查用户授权
        
        符合《个人信息保护法》第13条：
        处理个人信息应当在事先充分告知的前提下取得个人同意
        
        Args:
            user_id: 简历所有者ID
            service_type: 服务类型
            data_types: 需要访问的数据类型列表
            
        Returns:
            授权状态和详情
        """
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        
        try:
            # 查询授权记录
            cursor.execute("""
                SELECT consent_level, expiry_timestamp, usage_count, status, data_types
                FROM user_consent_records 
                WHERE user_id = ? AND service_type = ?
                AND expiry_timestamp > datetime('now')
                AND status = 'active'
            """, (user_id, service_type))
            
            result = cursor.fetchone()
            
            if not result:
                logger.warning(f"用户{user_id}未授权服务{service_type}")
                return {
                    'has_consent': False,
                    'consent_level': 'no_consent',
                    'reason': '用户未授权此服务',
                    'required_action': '请用户授权后再使用',
                    'compliance_note': '符合《个人信息保护法》第13条 - 需要用户同意'
                }
            
            consent_level, expiry_timestamp, usage_count, status, allowed_data_types_json = result
            allowed_data_types = json.loads(allowed_data_types_json)
            
            # 检查数据类型是否在授权范围内
            for data_type in data_types:
                if data_type not in allowed_data_types and 'all_personal_data' not in allowed_data_types:
                    logger.warning(f"数据类型{data_type}未在授权范围内")
                    return {
                        'has_consent': False,
                        'consent_level': consent_level,
                        'reason': f'数据类型 {data_type} 未在授权范围内',
                        'allowed_data_types': allowed_data_types,
                        'requested_data_types': data_types,
                        'compliance_note': '符合《个人信息保护法》第14条 - 最小必要原则'
                    }
            
            logger.info(f"用户{user_id}已授权服务{service_type}，级别{consent_level}")
            
            return {
                'has_consent': True,
                'consent_level': consent_level,
                'expiry_timestamp': expiry_timestamp,
                'usage_count': usage_count,
                'allowed_data_types': allowed_data_types,
                'status': 'authorized'
            }
            
        finally:
            conn.close()
    
    def update_consent_usage(self, user_id: int, service_type: str):
        """更新授权使用记录"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        
        try:
            cursor.execute("""
                UPDATE user_consent_records 
                SET usage_count = usage_count + 1,
                    last_used = datetime('now')
                WHERE user_id = ? AND service_type = ?
            """, (user_id, service_type))
            
            conn.commit()
            logger.info(f"更新授权使用记录: user_id={user_id}, service={service_type}")
            
        except Exception as e:
            logger.error(f"更新授权使用记录失败: {e}")
        finally:
            conn.close()
    
    def log_privacy_audit(self, user_id: int, action_type: str, service_type: str,
                         data_type: str, privacy_level: str, anonymized: bool,
                         accessed_by_user_id: Optional[int] = None,
                         accessed_by_role: Optional[str] = None,
                         details: Optional[Dict] = None):
        """
        记录隐私审计日志
        
        符合《个人信息保护法》第24条：
        个人信息处理者应当建立个人信息处理记录
        """
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        
        try:
            cursor.execute("""
                INSERT INTO privacy_audit_log 
                (user_id, action_type, data_type, service_type, privacy_level,
                 anonymized, accessed_by_user_id, accessed_by_role, timestamp, details)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), ?)
            """, (
                user_id,
                action_type,
                data_type,
                service_type,
                privacy_level,
                anonymized,
                accessed_by_user_id,
                accessed_by_role,
                json.dumps(details) if details else None
            ))
            
            conn.commit()
            logger.info(f"记录隐私审计日志: user={user_id}, action={action_type}, service={service_type}")
            
        except Exception as e:
            logger.error(f"记录隐私审计日志失败: {e}")
        finally:
            conn.close()
    
    def update_usage_statistics(self, user_id: int, service_type: str, data_type: str):
        """更新使用统计"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        
        try:
            cursor.execute("""
                INSERT INTO usage_statistics 
                (user_id, service_type, data_type, usage_count, last_used, privacy_level)
                VALUES (?, ?, ?, 1, datetime('now'), 'partial')
                ON CONFLICT(user_id, service_type, data_type) DO UPDATE SET
                    usage_count = usage_count + 1,
                    last_used = datetime('now'),
                    updated_at = datetime('now')
            """, (user_id, service_type, data_type))
            
            conn.commit()
            logger.info(f"更新使用统计: user={user_id}, service={service_type}, data={data_type}")
            
        except Exception as e:
            logger.error(f"更新使用统计失败: {e}")
        finally:
            conn.close()
    
    def get_user_consents(self, user_id: int) -> List[Dict[str, Any]]:
        """获取用户所有授权记录"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        
        try:
            cursor.execute("""
                SELECT service_type, data_types, consent_level, 
                       consent_timestamp, expiry_timestamp, usage_count, status
                FROM user_consent_records
                WHERE user_id = ?
                ORDER BY consent_timestamp DESC
            """, (user_id,))
            
            results = cursor.fetchall()
            
            consents = []
            for row in results:
                consents.append({
                    'service_type': row[0],
                    'data_types': json.loads(row[1]),
                    'consent_level': row[2],
                    'consent_timestamp': row[3],
                    'expiry_timestamp': row[4],
                    'usage_count': row[5],
                    'status': row[6]
                })
            
            return consents
            
        finally:
            conn.close()
    
    def get_usage_history(self, user_id: int, limit: int = 100) -> List[Dict[str, Any]]:
        """获取数据使用历史"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        
        try:
            cursor.execute("""
                SELECT action_type, data_type, service_type, privacy_level,
                       anonymized, accessed_by_user_id, accessed_by_role, 
                       timestamp, details
                FROM privacy_audit_log
                WHERE user_id = ?
                ORDER BY timestamp DESC
                LIMIT ?
            """, (user_id, limit))
            
            results = cursor.fetchall()
            
            history = []
            for row in results:
                history.append({
                    'action_type': row[0],
                    'data_type': row[1],
                    'service_type': row[2],
                    'privacy_level': row[3],
                    'anonymized': bool(row[4]),
                    'accessed_by_user_id': row[5],
                    'accessed_by_role': row[6],
                    'timestamp': row[7],
                    'details': json.loads(row[8]) if row[8] else None
                })
            
            return history
            
        finally:
            conn.close()

# 测试代码
if __name__ == "__main__":
    print("\n测试授权管理器...")
    
    # 测试简历#1的授权检查
    manager = ConsentManager('/tmp/sqlite_test_new/resume_1.db')
    
    # 测试1: 检查job_matching授权
    print("\n【测试1】检查job_matching授权...")
    result = manager.check_consent(1, 'job_matching', ['skills', 'experience'])
    print(f"  授权状态: {result['has_consent']}")
    print(f"  授权级别: {result.get('consent_level', 'N/A')}")
    if result['has_consent']:
        print(f"  允许的数据: {result['allowed_data_types']}")
    else:
        print(f"  原因: {result.get('reason', 'N/A')}")
    
    # 测试2: 检查未授权的服务
    print("\n【测试2】检查未授权服务...")
    result = manager.check_consent(1, 'full_ai_analysis', ['all_personal_data'])
    print(f"  授权状态: {result['has_consent']}")
    print(f"  原因: {result.get('reason', 'N/A')}")
    
    # 测试3: 记录审计日志
    print("\n【测试3】记录隐私审计日志...")
    manager.log_privacy_audit(
        user_id=1,
        action_type='view',
        service_type='job_matching',
        data_type='resume',
        privacy_level='partial',
        anonymized=True,
        accessed_by_user_id=1,
        accessed_by_role='super_admin',
        details={'test': 'privacy audit'}
    )
    print("  ✅ 审计日志记录成功")
    
    # 测试4: 更新使用统计
    print("\n【测试4】更新使用统计...")
    manager.update_usage_statistics(1, 'job_matching', 'resume')
    print("  ✅ 使用统计更新成功")
    
    # 测试5: 获取用户授权列表
    print("\n【测试5】获取用户授权列表...")
    consents = manager.get_user_consents(1)
    print(f"  ✅ 找到 {len(consents)} 条授权记录")
    for consent in consents:
        print(f"     - {consent['service_type']}: {consent['consent_level']}")
    
    # 测试6: 获取使用历史
    print("\n【测试6】获取数据使用历史...")
    history = manager.get_usage_history(1, limit=10)
    print(f"  ✅ 找到 {len(history)} 条使用记录")
    
    print("\n✅ 授权管理器测试完成！")

