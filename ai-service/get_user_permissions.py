#!/usr/bin/env python3
"""
获取用户权限模块
从数据库直接获取用户权限信息
"""

import mysql.connector
import logging
from typing import List, Dict, Optional

logger = logging.getLogger(__name__)

# 数据库配置
import os
DB_CONFIG = {
    'host': os.getenv('MYSQL_HOST', 'host.docker.internal'),
    'user': os.getenv('MYSQL_USER', 'root'),
    'password': os.getenv('MYSQL_PASSWORD', ''),
    'database': os.getenv('MYSQL_DB', 'jobfirst'),
    'charset': 'utf8mb4'
}

def get_user_permissions(user_id: int) -> List[str]:
    """获取用户权限列表"""
    try:
        logger.info(f"尝试获取用户 {user_id} 的权限...")
        
        # 使用Docker容器内的数据库连接
        connection = mysql.connector.connect(**DB_CONFIG)
        cursor = connection.cursor()
        
        # 查询用户权限
        cursor.execute("""
            SELECT p.name 
            FROM permissions p
            INNER JOIN user_permissions up ON p.id = up.permission_id
            WHERE up.user_id = %s AND up.is_active = 1 AND p.is_active = 1
        """, (user_id,))
        
        permissions = [row[0] for row in cursor.fetchall()]
        
        cursor.close()
        connection.close()
        
        logger.info(f"用户 {user_id} 拥有 {len(permissions)} 个权限: {permissions}")
        return permissions
        
    except mysql.connector.Error as e:
        logger.error(f"数据库连接失败: {e}")
        # 如果数据库连接失败，使用硬编码权限作为备用
        if user_id == 1:  # 超级管理员
            permissions = ["ai:all", "ai:chat", "ai:job_matching", "ai:resume_analysis", "system:manage"]
        elif user_id == 4:  # 普通用户
            permissions = ["ai:chat"]  # 基本AI功能权限
        else:
            permissions = ["basic:access"]  # 默认基本权限
        
        logger.info(f"使用备用权限配置，用户 {user_id} 拥有 {len(permissions)} 个权限: {permissions}")
        return permissions
        
    except Exception as e:
        logger.error(f"获取用户权限异常: {e}")
        return []

def get_user_role(user_id: int) -> Optional[str]:
    """获取用户角色"""
    try:
        connection = mysql.connector.connect(**DB_CONFIG)
        cursor = connection.cursor()
        
        cursor.execute("SELECT role FROM users WHERE id = %s", (user_id,))
        result = cursor.fetchone()
        
        cursor.close()
        connection.close()
        
        if result:
            return result[0]
        return None
        
    except mysql.connector.Error as e:
        logger.error(f"获取用户角色失败: {e}")
        return None

def has_permission(user_id: int, permission: str) -> bool:
    """检查用户是否有特定权限"""
    try:
        # 获取用户角色
        role = get_user_role(user_id)
        if not role:
            return False
        
        # 超级管理员拥有所有权限
        if role == 'super_admin':
            return True
        
        # 检查具体权限
        permissions = get_user_permissions(user_id)
        return permission in permissions
        
    except Exception as e:
        logger.error(f"权限检查失败: {e}")
        return False

def get_user_info_with_permissions(user_id: int) -> Dict:
    """获取用户信息和权限"""
    try:
        connection = mysql.connector.connect(**DB_CONFIG)
        cursor = connection.cursor(dictionary=True)
        
        # 获取用户基本信息
        cursor.execute("""
            SELECT id, username, email, role, status, subscription_status, subscription_type
            FROM users 
            WHERE id = %s
        """, (user_id,))
        
        user_data = cursor.fetchone()
        if not user_data:
            return None
        
        # 获取用户权限
        permissions = get_user_permissions(user_id)
        
        cursor.close()
        connection.close()
        
        return {
            'user_id': user_data['id'],
            'username': user_data['username'],
            'email': user_data['email'],
            'role': user_data['role'],
            'status': user_data['status'],
            'subscription_status': user_data['subscription_status'],
            'subscription_type': user_data['subscription_type'],
            'permissions': permissions
        }
        
    except mysql.connector.Error as e:
        logger.error(f"获取用户信息失败: {e}")
        return None
