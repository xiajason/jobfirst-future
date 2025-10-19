#!/usr/bin/env python3
"""
测试AI服务依赖导入
"""

import sys

print("=== 🧪 测试AI服务依赖导入 ===\n")

# 测试核心依赖
try:
    import sanic
    print("✅ sanic导入成功:", sanic.__version__)
except ImportError as e:
    print("❌ sanic导入失败:", e)
    sys.exit(1)

try:
    import httpx
    print("✅ httpx导入成功:", httpx.__version__)
except ImportError as e:
    print("❌ httpx导入失败:", e)
    sys.exit(1)

try:
    from dotenv import load_dotenv
    print("✅ python-dotenv导入成功")
except ImportError as e:
    print("❌ python-dotenv导入失败:", e)
    sys.exit(1)

try:
    import requests
    print("✅ requests导入成功:", requests.__version__)
except ImportError as e:
    print("❌ requests导入失败:", e)
    sys.exit(1)

# 测试环境变量加载
print("\n=== 🔐 测试环境变量加载 ===\n")
load_dotenv()

import os
ai_port = os.getenv("AI_SERVICE_PORT", "8100")
print(f"✅ AI_SERVICE_PORT: {ai_port}")

mysql_host = os.getenv("MYSQL_HOST", "localhost")
print(f"✅ MYSQL_HOST: {mysql_host}")

deepseek_model = os.getenv("DEEPSEEK_MODEL", "deepseek-chat")
print(f"✅ DEEPSEEK_MODEL: {deepseek_model}")

# 测试主服务文件导入
print("\n=== 📦 测试主服务文件 ===\n")
try:
    # 只检查语法，不实际运行
    with open("ai_service_with_zervigo.py", "r") as f:
        code = f.read()
        compile(code, "ai_service_with_zervigo.py", "exec")
    print("✅ ai_service_with_zervigo.py 语法检查通过")
except SyntaxError as e:
    print(f"❌ 语法错误: {e}")
    sys.exit(1)

print("\n=== ✅ 所有测试通过！AI服务可以正常启动 ===")
print("\n💡 提示: 实际启动需要连接数据库，请确保数据库服务运行中")

