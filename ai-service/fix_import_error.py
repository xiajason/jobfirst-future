#!/usr/bin/env python3
"""修复导入错误"""

# 读取文件
with open("job_matching_data_access.py", "r") as f:
    content = f.read()

# 删除错误的导入代码
error_block = """import sys
sys.path.append(os.path.dirname(os.path.abspath(__file__)))
from consent_manager import ConsentManager
from data_anonymizer import DataAnonymizer
"""

content = content.replace(error_block, "")

# 在正确的位置添加导入
import_os_pos = content.find("import os")
if import_pos > 0:
    # 找到import段落的结束
    import_end = content.find("logger = logging.getLogger(__name__)")
    if import_end > 0:
        # 在logger之前添加
        new_imports = """# 隐私保护模块
from consent_manager import ConsentManager
from data_anonymizer import DataAnonymizer

"""
        content = content[:import_end] + new_imports + content[import_end:]
        print("✅ 修复导入错误")

# 写回
with open("job_matching_data_access.py", "w") as f:
    f.write(content)

print("✅ 文件修复完成")
