#!/usr/bin/env python3
"""集成隐私保护到Job Matching Service"""
import sys

# 读取job_matching_service.py
with open("job_matching_service.py", "r") as f:
    content = f.read()

# 备份
with open("job_matching_service.py.before_privacy_integration", "w") as f:
    f.write(content)
print("✅ 已备份job_matching_service.py")

# 添加导入
if "from privacy_enhanced_data_access import" not in content:
    import_pos = content.find("from job_matching_data_access import")
    if import_pos > 0:
        insert_pos = content.find("\n", import_pos) + 1
        new_import = "from privacy_enhanced_data_access import PrivacyEnhancedDataAccess\n"
        content = content[:insert_pos] + new_import + content[insert_pos:]
        print("✅ 添加PrivacyEnhancedDataAccess导入")

# 在JobMatchingService.__init__中添加隐私包装器初始化
init_pos = content.find("def __init__(self):")
if init_pos > 0:
    # 找到__init__方法的结束位置
    next_def = content.find("\n    async def", init_pos)
    if next_def > 0:
        # 在__init__结束前添加
        insert_pos = content.rfind("\n", init_pos, next_def)
        privacy_init = """        
        # 初始化隐私增强数据访问层
        self.privacy_data_access = PrivacyEnhancedDataAccess(self.data_access)
        logger.info("✅ 隐私增强数据访问层已初始化")
"""
        if "privacy_data_access" not in content:
            content = content[:insert_pos] + privacy_init + content[insert_pos:]
            print("✅ 添加隐私数据访问层初始化")

# 写回文件
with open("job_matching_service.py", "w") as f:
    f.write(content)

print("✅ Job Matching Service隐私保护集成完成")
print("\n📋 已集成功能:")
print("  1. 导入PrivacyEnhancedDataAccess")
print("  2. 初始化隐私数据访问层")
print("\n🎯 现在Job Matching Service可以使用隐私保护功能了")
print("\n💡 使用方式:")
print("   resume_data = await self.privacy_data_access.get_resume_with_privacy_check(")
print("       resume_id, user_id, user_role, service_type")
print("   )")
